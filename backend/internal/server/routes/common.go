package routes

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

const clientReleaseBaseURL = "https://github.com/pengge499896694/jokerdeck-client/releases/latest/download/"
const clientReleaseManifestURL = clientReleaseBaseURL + "latest.json"
const maxClientManifestSize = 2 << 20
const maxClientAssetSize = 300 << 20

// RegisterCommonRoutes 注册通用路由（健康检查、状态等）
func RegisterCommonRoutes(r *gin.Engine) {
	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Claude Code 遥测日志（忽略，直接返回200）
	r.POST("/api/event_logging/batch", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	// Setup status endpoint (always returns needs_setup: false in normal mode)
	// This is used by the frontend to detect when the service has restarted after setup
	r.GET("/setup/status", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"code": 0,
			"data": gin.H{
				"needs_setup": false,
				"step":        "completed",
			},
		})
	})

	// Proxy desktop update metadata and assets so clients do not need direct
	// access to GitHub. GitHub remains the authoritative release source.
	r.GET("/client/latest.json", proxyClientManifest)
	r.GET("/client/download/:asset", proxyClientAsset)
}

func clientReleaseHTTPClient() *http.Client {
	return &http.Client{Timeout: 90 * time.Second}
}

func proxyClientManifest(c *gin.Context) {
	c.Header("Cache-Control", "no-store, no-cache, must-revalidate")
	body, err := fetchClientRelease(clientReleaseManifestURL, maxClientManifestSize)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "client update source unavailable"})
		return
	}

	var manifest map[string]any
	if err := json.Unmarshal(body, &manifest); err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "client update manifest is invalid"})
		return
	}
	platforms, ok := manifest["platforms"].(map[string]any)
	if !ok {
		c.JSON(http.StatusBadGateway, gin.H{"error": "client update manifest has no platforms"})
		return
	}
	base := clientPublicBaseURL(c)
	for platform, raw := range platforms {
		switch value := raw.(type) {
		case string:
			platforms[platform] = proxyAssetURL(base, value)
		case map[string]any:
			if url, ok := value["url"].(string); ok {
				value["url"] = proxyAssetURL(base, url)
			}
			if url, ok := value["download_url"].(string); ok {
				value["download_url"] = proxyAssetURL(base, url)
			}
		}
	}
	data, err := json.Marshal(manifest)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "client update manifest cannot be encoded"})
		return
	}
	c.Data(http.StatusOK, "application/json; charset=utf-8", data)
}

func proxyClientAsset(c *gin.Context) {
	asset := c.Param("asset")
	if asset == "" || path.Base(asset) != asset || strings.ContainsAny(asset, `/\`) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid client asset"})
		return
	}
	body, err := fetchClientRelease(clientReleaseBaseURL+asset, maxClientAssetSize)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "client asset unavailable"})
		return
	}
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, asset))
	c.Data(http.StatusOK, "application/octet-stream", body)
}

func fetchClientRelease(url string, maxSize int64) ([]byte, error) {
	resp, err := clientReleaseHTTPClient().Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("release source returned HTTP %d", resp.StatusCode)
	}
	if resp.ContentLength > maxSize {
		return nil, fmt.Errorf("release asset exceeds size limit")
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxSize+1))
	if err != nil {
		return nil, err
	}
	if int64(len(body)) > maxSize {
		return nil, fmt.Errorf("release asset exceeds size limit")
	}
	return body, nil
}

func clientPublicBaseURL(c *gin.Context) string {
	scheme := c.GetHeader("X-Forwarded-Proto")
	if scheme != "http" && scheme != "https" {
		scheme = "https"
		if c.Request.TLS != nil {
			scheme = "https"
		}
	}
	return scheme + "://" + c.Request.Host
}

func proxyAssetURL(base, asset string) string {
	if strings.HasPrefix(asset, clientReleaseBaseURL) {
		return base + "/client/download/" + path.Base(asset)
	}
	return asset
}
