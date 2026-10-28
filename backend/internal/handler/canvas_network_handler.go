package handler

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/proxyurl"
	"github.com/Wei-Shaw/sub2api/internal/pkg/proxyutil"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

const canvasNetworkRelayEndpoint = "/api/v1/canvas/network/fetch"

type CanvasNetworkHandler struct {
	settings *service.SettingService
}

func NewCanvasNetworkHandler(settings *service.SettingService) *CanvasNetworkHandler {
	return &CanvasNetworkHandler{settings: settings}
}

func (h *CanvasNetworkHandler) GetConfig(c *gin.Context) {
	settings, err := h.settings.GetCanvasNetworkSettings(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Canvas network settings unavailable"})
		return
	}
	role, _ := middleware.GetUserRoleFromContext(c)
	enabled := settings.Enabled && (settings.VisibleToUsers || role == service.RoleAdmin)
	c.Header("Cache-Control", "no-store")
	c.JSON(http.StatusOK, gin.H{
		"enabled":  enabled,
		"endpoint": canvasNetworkRelayEndpoint,
	})
}

func canvasRelayHostAllowed(host string, allowed []string) bool {
	host = strings.ToLower(strings.TrimSuffix(strings.TrimSpace(host), "."))
	for _, suffix := range allowed {
		suffix = strings.ToLower(strings.TrimSuffix(strings.TrimSpace(suffix), "."))
		if host == suffix || strings.HasSuffix(host, "."+suffix) {
			return true
		}
	}
	return false
}

func (h *CanvasNetworkHandler) Fetch(c *gin.Context) {
	settings, err := h.settings.GetCanvasNetworkSettings(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Canvas network settings unavailable"})
		return
	}
	role, _ := middleware.GetUserRoleFromContext(c)
	if !settings.Enabled || (!settings.VisibleToUsers && role != service.RoleAdmin) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Canvas relay is disabled"})
		return
	}
	target, err := parseCanvasRelayTarget(c.Query("url"), settings.AllowedHosts)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Only configured public asset hosts can use the canvas relay"})
		return
	}
	proxyRaw, err := h.settings.CanvasNetworkProxyURL(c.Request.Context(), settings.ProxyID)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "Canvas relay proxy is unavailable"})
		return
	}
	_, proxyURL, err := proxyurl.Parse(proxyRaw)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "Canvas relay proxy is invalid"})
		return
	}
	transport := &http.Transport{Proxy: nil}
	if err := proxyutil.ConfigureTransportProxy(transport, proxyURL); err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "Canvas relay proxy is invalid"})
		return
	}
	client := &http.Client{Transport: transport, Timeout: 90 * time.Second}
	client.CheckRedirect = func(req *http.Request, _ []*http.Request) error {
		if req.URL.Scheme != "http" && req.URL.Scheme != "https" || req.URL.User != nil || !canvasRelayHostAllowed(req.URL.Hostname(), settings.AllowedHosts) {
			return http.ErrUseLastResponse
		}
		return nil
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 90*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target.String(), nil)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid canvas asset URL"})
		return
	}
	req.Header.Set("Accept", c.GetHeader("Accept"))
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "Canvas asset fetch failed"})
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		c.JSON(http.StatusBadGateway, gin.H{"error": "Canvas asset source returned an error"})
		return
	}
	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	c.Header("Content-Type", contentType)
	c.Header("Cache-Control", "private, max-age=300")
	c.Header("X-Content-Type-Options", "nosniff")
	if resp.ContentLength >= 0 {
		c.Header("Content-Length", formatContentLength(resp.ContentLength))
	}
	c.Status(http.StatusOK)
	_, _ = io.Copy(c.Writer, resp.Body)
}

func parseCanvasRelayTarget(raw string, allowed []string) (*url.URL, error) {
	target, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || target.User != nil || (target.Scheme != "http" && target.Scheme != "https") || target.Hostname() == "" || !canvasRelayHostAllowed(target.Hostname(), allowed) {
		return nil, errInvalidCanvasRelayTarget
	}
	return target, nil
}

var errInvalidCanvasRelayTarget = &canvasRelayTargetError{}

type canvasRelayTargetError struct{}

func (*canvasRelayTargetError) Error() string { return "invalid canvas relay target" }

func formatContentLength(value int64) string {
	return strconv.FormatInt(value, 10)
}
