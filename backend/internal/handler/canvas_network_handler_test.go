package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseCanvasRelayTargetAllowsConfiguredHostAndSubdomain(t *testing.T) {
	allowed := []string{"github.com", "raw.githubusercontent.com"}
	for _, raw := range []string{"https://github.com/a.png?x=1", "https://gist.github.com/a", "https://raw.githubusercontent.com/u/r/main/a.png"} {
		target, err := parseCanvasRelayTarget(raw, allowed)
		require.NoError(t, err)
		require.NotNil(t, target)
	}
}

func TestParseCanvasRelayTargetRejectsSSRFAndUnconfiguredHosts(t *testing.T) {
	allowed := []string{"github.com"}
	for _, raw := range []string{"http://127.0.0.1/a", "http://169.254.169.254/latest", "https://evil.example/a", "https://user:pass@github.com/a", "javascript:alert(1)"} {
		target, err := parseCanvasRelayTarget(raw, allowed)
		require.Error(t, err)
		require.Nil(t, target)
	}
}

// Config uses the same authenticated identity as the panel and only returns a
// short-lived cookie for the read-only relay, never a token in an asset URL.
func TestCanvasNetworkConfigAccessAndAssetCookie(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, tc := range []struct {
		role                    string
		public, enabled, access bool
	}{
		{"admin", false, false, true}, {"admin", false, true, true},
		{"user", false, true, false}, {"finance", false, true, false}, {"user", true, true, true},
	} {
		repo := &canvasNetworkHandlerRepo{raw: fmt.Sprintf(`{"enabled":%t,"visible_to_users":%t}`, tc.enabled, tc.public)}
		handler := NewCanvasNetworkHandler(service.NewSettingService(repo, nil))
		r := gin.New()
		r.GET("/config", func(c *gin.Context) { c.Set(string(middleware.ContextKeyUserRole), tc.role); handler.GetConfig(c) })
		req := httptest.NewRequest("GET", "/config", nil)
		req.Header.Set("Authorization", "Bearer test-session")
		req.Header.Set("X-Forwarded-Proto", "https")
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		require.Equal(t, 200, rec.Code)
		var result struct {
			Enabled bool `json:"enabled"`
			Access  bool `json:"can_access"`
		}
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &result))
		require.Equal(t, tc.access, result.Access)
		require.Equal(t, tc.enabled && tc.access, result.Enabled)
		cookies := rec.Result().Cookies()
		if result.Enabled {
			require.Len(t, cookies, 1)
			require.True(t, cookies[0].HttpOnly)
			require.True(t, cookies[0].Secure)
			require.Equal(t, "/api/v1/canvas/network/", cookies[0].Path)
			require.Equal(t, http.SameSiteStrictMode, cookies[0].SameSite)
			require.Equal(t, 600, cookies[0].MaxAge)
		} else {
			require.Empty(t, cookies)
		}
	}
}

type canvasNetworkHandlerRepo struct {
	service.SettingRepository
	raw string
}

func (r *canvasNetworkHandlerRepo) GetValue(context.Context, string) (string, error) {
	return r.raw, nil
}

func TestCanvasNetworkCookieCannotAuthenticateOtherAPIPaths(t *testing.T) {
	for _, tc := range []struct{ method, path, header, expected string }{
		{"GET", canvasNetworkRelayEndpoint, "", "Bearer test-session"},
		{"POST", canvasNetworkRelayEndpoint, "", ""},
		{"GET", "/api/v1/canvas/network-config", "", ""},
		{"GET", "/api/v1/user/profile", "", ""},
		{"GET", canvasNetworkRelayEndpoint, "Bearer explicit", "Bearer explicit"},
	} {
		r := gin.New()
		r.Use(CanvasNetworkAssetAuth())
		r.Handle(tc.method, tc.path, func(c *gin.Context) { c.String(200, c.GetHeader("Authorization")) })
		req := httptest.NewRequest(tc.method, tc.path, nil)
		req.AddCookie(&http.Cookie{Name: canvasNetworkCookie, Value: "test-session"})
		req.Header.Set("Authorization", tc.header)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		require.Equal(t, tc.expected, rec.Body.String())
	}
}
