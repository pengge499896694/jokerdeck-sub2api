package admin_test

import (
	"encoding/json"
	adminhandler "github.com/Wei-Shaw/sub2api/internal/handler/admin"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCanvasNetworkSaveDefaultLineAndReportInvalidProxy(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &commerceSettingsRepo{values: map[string]string{}}
	svc := service.NewSettingService(repo, nil)
	h := adminhandler.NewSettingHandler(svc, nil, nil, nil, nil, nil, nil)
	router := gin.New()
	router.PUT("/settings", h.UpdateCanvasNetworkSettings)
	router.GET("/settings", h.GetCanvasNetworkSettings)
	request := func(method, body string) *httptest.ResponseRecorder {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(method, "/settings", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(rec, req)
		return rec
	}
	rec := request(http.MethodPut, `{"enabled":true,"visible_to_users":false,"proxy_id":null,"allowed_hosts":["raw.githubusercontent.com"]}`)
	require.Equal(t, 200, rec.Code, rec.Body.String())
	saved := rec.Body.String()
	require.Contains(t, saved, `"proxy_id":null`)
	require.JSONEq(t, saved, request(http.MethodGet, "").Body.String())
	rec = request(http.MethodPut, `{"enabled":true,"proxy_id":123,"allowed_hosts":["github.com"]}`)
	require.Equal(t, 400, rec.Code)
	var payload map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &payload))
	require.Contains(t, rec.Body.String(), "所选代理不可用")
	require.JSONEq(t, saved, request(http.MethodGet, "").Body.String())
}
