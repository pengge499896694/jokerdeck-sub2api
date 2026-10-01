package routes

import (
	"encoding/json"
	"fmt"
	"html"
	"io"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

const clientReleaseBaseURL = "https://github.com/pengge499896694/jokerdeck-client/releases/latest/download/"
const clientReleaseManifestURL = clientReleaseBaseURL + "latest.json"
const codexLocalizationReleaseAPI = "https://api.github.com/repos/xqnode/codex-zh-CN/releases/tags/v0.1.2"
const codexLocalizationReleaseBaseURL = "https://github.com/xqnode/codex-zh-CN/releases/download/v0.1.2/"
const codexLocalizationAsset = "codex-zh-CN-v0.1.2.zip"
const maxClientManifestSize = 2 << 20
const maxClientAssetSize = 300 << 20
const maxCodexLocalizationSize = 12 << 20

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
	// access to GitHub. The server is also the public download page.
	r.GET("/client", clientDownloadPage)
	r.GET("/client/", clientDownloadPage)
	r.GET("/client/latest.json", proxyClientManifest)
	r.GET("/client/download/:asset", proxyClientAsset)
	r.GET("/client/codex-localization/latest.json", proxyCodexLocalizationManifest)
	r.GET("/client/codex-localization/download/:asset", proxyCodexLocalizationAsset)
}

func clientReleaseHTTPClient() *http.Client {
	return &http.Client{Timeout: 90 * time.Second}
}

func proxyClientManifest(c *gin.Context) {
	c.Header("Cache-Control", "no-store, no-cache, must-revalidate")
	manifest, err := fetchClientManifest()
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "client update source unavailable"})
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

func fetchClientManifest() (map[string]any, error) {
	body, err := fetchClientRelease(clientReleaseManifestURL, maxClientManifestSize)
	if err != nil {
		return nil, err
	}
	var manifest map[string]any
	if err := json.Unmarshal(body, &manifest); err != nil {
		return nil, err
	}
	if _, ok := manifest["platforms"].(map[string]any); !ok {
		return nil, fmt.Errorf("client update manifest has no platforms")
	}
	return manifest, nil
}

func clientDownloadPage(c *gin.Context) {
	manifest, err := fetchClientManifest()
	if err != nil {
		c.String(http.StatusBadGateway, "客户端发布信息暂时不可用，请稍后重试")
		return
	}
	platforms := manifest["platforms"].(map[string]any)
	base := clientPublicBaseURL(c)
	links := make([]string, 0, len(platforms))
	for _, platform := range []string{"windows-x86_64", "darwin-x86_64", "darwin-aarch64"} {
		raw, ok := platforms[platform]
		if !ok {
			continue
		}
		asset := assetURLValue(raw)
		if asset == "" {
			continue
		}
		links = append(links, fmt.Sprintf(
			`<li><a href="%s">%s</a></li>`,
			html.EscapeString(proxyAssetURL(base, asset)),
			html.EscapeString(platformLabel(platform)),
		))
	}
	page := `<!doctype html>
<html lang="zh-CN">
<head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1">
<title>jokerdeck 客户端下载</title>
<style>
body{margin:0;padding:48px 20px;background:#f5f7fb;color:#172033;font:16px/1.6 system-ui,-apple-system,"Microsoft YaHei",sans-serif}
main{max-width:640px;margin:auto;background:#fff;border:1px solid #dce2ec;border-radius:10px;padding:32px;box-shadow:0 8px 30px #17203312}
h1{margin:0 0 8px;font-size:28px}p{color:#5d687b}ul{padding-left:22px}a{color:#1769e0}
</style></head>
<body><main><h1>jokerdeck 客户端</h1>
<p>从服务器下载最新版本，安装包由 jokerdeck 官方服务器提供。</p>
<ul>` + strings.Join(links, "") + `</ul>
<p><a href="/client/latest.json">查看更新清单</a></p>
</main></body></html>`
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(page))
}

func assetURLValue(value any) string {
	switch value := value.(type) {
	case string:
		return value
	case map[string]any:
		if url, ok := value["url"].(string); ok {
			return url
		}
		if url, ok := value["download_url"].(string); ok {
			return url
		}
	}
	return ""
}

func platformLabel(platform string) string {
	switch platform {
	case "windows-x86_64":
		return "Windows 64 位安装包"
	case "darwin-x86_64":
		return "macOS Intel 安装包"
	case "darwin-aarch64":
		return "macOS Apple Silicon 安装包"
	default:
		return platform
	}
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

func proxyCodexLocalizationManifest(c *gin.Context) {
	c.Header("Cache-Control", "public, max-age=300")
	body, err := fetchClientRelease(codexLocalizationReleaseAPI, 1<<20)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "codex localization source unavailable"})
		return
	}

	var manifest map[string]any
	if err := json.Unmarshal(body, &manifest); err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "codex localization manifest is invalid"})
		return
	}
	assets, ok := manifest["assets"].([]any)
	if !ok {
		c.JSON(http.StatusBadGateway, gin.H{"error": "codex localization manifest has no assets"})
		return
	}
	base := clientPublicBaseURL(c)
	for _, raw := range assets {
		asset, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		if name, ok := asset["name"].(string); ok {
			asset["browser_download_url"] = base + "/client/codex-localization/download/" + path.Base(name)
		}
	}
	data, err := json.Marshal(manifest)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "codex localization manifest cannot be encoded"})
		return
	}
	c.Data(http.StatusOK, "application/json; charset=utf-8", data)
}

func proxyCodexLocalizationAsset(c *gin.Context) {
	asset := c.Param("asset")
	if asset != codexLocalizationAsset || path.Base(asset) != asset || strings.ContainsAny(asset, `/\`) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid codex localization asset"})
		return
	}
	body, err := fetchClientRelease(codexLocalizationReleaseBaseURL+asset, maxCodexLocalizationSize)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "codex localization asset unavailable"})
		return
	}
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, asset))
	c.Data(http.StatusOK, "application/octet-stream", body)
}

func fetchClientRelease(url string, maxSize int64) ([]byte, error) {
	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("User-Agent", "jokerdeck-server")
	request.Header.Set("Accept", "application/vnd.github+json")
	resp, err := clientReleaseHTTPClient().Do(request)
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
