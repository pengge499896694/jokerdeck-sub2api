package handler

import (
	"context"
	"io"
	"net"
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
const canvasNetworkCookie = "canvas_network_session"

// Asset elements cannot attach an Authorization header. Accept the short-lived
// HttpOnly cookie ONLY on this read-only relay path, then use the normal JWT
// validation (including expiry, revocation, status and session binding).
func CanvasNetworkAssetAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == http.MethodGet && c.Request.URL.Path == canvasNetworkRelayEndpoint && c.GetHeader("Authorization") == "" {
			if token, err := c.Cookie(canvasNetworkCookie); err == nil && token != "" {
				c.Request.Header.Set("Authorization", "Bearer "+token)
			}
		}
		c.Next()
	}
}

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
	// The config call is authenticated with the panel's Bearer token. The cookie
	// is path-scoped and never sent to the source host or exposed in an asset URL.
	if enabled {
		if auth := c.GetHeader("Authorization"); strings.HasPrefix(auth, "Bearer ") {
			http.SetCookie(c.Writer, &http.Cookie{Name: canvasNetworkCookie, Value: strings.TrimPrefix(auth, "Bearer "), Path: "/api/v1/canvas/network/", MaxAge: 600, HttpOnly: true, Secure: c.Request.TLS != nil || c.GetHeader("X-Forwarded-Proto") == "https", SameSite: http.SameSiteStrictMode})
		}
	}
	c.Header("Cache-Control", "no-store")
	c.JSON(http.StatusOK, gin.H{
		"enabled":       enabled,
		"endpoint":      canvasNetworkRelayEndpoint,
		"can_access":    role == service.RoleAdmin || settings.VisibleToUsers,
		"allowed_hosts": settings.AllowedHosts,
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
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.Proxy = nil
	if proxyURL == nil {
		transport.DialContext = canvasRelayPublicDial
	}
	if err := proxyutil.ConfigureTransportProxy(transport, proxyURL); err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "Canvas relay proxy is invalid"})
		return
	}
	defer transport.CloseIdleConnections()
	client := &http.Client{Transport: transport, Timeout: 90 * time.Second}
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		if len(via) >= 5 {
			return http.ErrUseLastResponse
		}
		if _, err := parseCanvasRelayTarget(req.URL.String(), settings.AllowedHosts); err != nil {
			return http.ErrUseLastResponse
		}
		if proxyURL != nil {
			if _, err := canvasRelayPublicAddresses(req.Context(), req.URL.Hostname()); err != nil {
				return err
			}
		}
		return nil
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 90*time.Second)
	defer cancel()
	if proxyURL != nil {
		if _, err := canvasRelayPublicAddresses(ctx, target.Hostname()); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Canvas asset host must resolve to public addresses"})
			return
		}
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target.String(), nil)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid canvas asset URL"})
		return
	}
	req.Header.Set("Accept", c.GetHeader("Accept"))
	if value := c.GetHeader("Range"); value != "" {
		req.Header.Set("Range", value)
	}
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
	c.Header("Cache-Control", "private, no-store")
	c.Header("Content-Security-Policy", "sandbox; default-src 'none'")
	c.Header("X-Content-Type-Options", "nosniff")
	if resp.ContentLength >= 0 {
		c.Header("Content-Length", formatContentLength(resp.ContentLength))
	}
	if value := resp.Header.Get("Content-Range"); value != "" {
		c.Header("Content-Range", value)
	}
	c.Status(resp.StatusCode)
	_, _ = io.Copy(c.Writer, resp.Body)
}

func parseCanvasRelayTarget(raw string, allowed []string) (*url.URL, error) {
	target, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || len(raw) > 8192 || strings.ContainsAny(raw, "\\\r\n") || target.User != nil || (target.Scheme != "http" && target.Scheme != "https") || target.Hostname() == "" || (target.Port() != "" && target.Port() != "443" && target.Port() != "80") || net.ParseIP(target.Hostname()) != nil || !canvasRelayHostAllowed(target.Hostname(), allowed) {
		return nil, errInvalidCanvasRelayTarget
	}
	return target, nil
}

func canvasRelayPublicDial(ctx context.Context, network, address string) (net.Conn, error) {
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return nil, err
	}
	ips, err := canvasRelayPublicAddresses(ctx, host)
	if err != nil {
		return nil, err
	}
	dialer := &net.Dialer{Timeout: 10 * time.Second, KeepAlive: 30 * time.Second}
	for _, item := range ips {
		conn, err := dialer.DialContext(ctx, network, net.JoinHostPort(item.IP.String(), port))
		if err == nil {
			return conn, nil
		}
	}
	return nil, errInvalidCanvasRelayTarget
}

func canvasRelayPublicAddresses(ctx context.Context, host string) ([]net.IPAddr, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	ips, err := net.DefaultResolver.LookupIPAddr(ctx, host)
	if err != nil {
		return nil, err
	}
	for _, item := range ips {
		if !item.IP.IsGlobalUnicast() || item.IP.IsPrivate() || item.IP.IsLoopback() || item.IP.IsLinkLocalUnicast() {
			return nil, errInvalidCanvasRelayTarget
		}
	}
	if len(ips) == 0 {
		return nil, errInvalidCanvasRelayTarget
	}
	return ips, nil
}

var errInvalidCanvasRelayTarget = &canvasRelayTargetError{}

type canvasRelayTargetError struct{}

func (*canvasRelayTargetError) Error() string { return "invalid canvas relay target" }

func formatContentLength(value int64) string {
	return strconv.FormatInt(value, 10)
}
