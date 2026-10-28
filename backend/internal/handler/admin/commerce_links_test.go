package admin_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	userhandler "github.com/Wei-Shaw/sub2api/internal/handler"
	adminhandler "github.com/Wei-Shaw/sub2api/internal/handler/admin"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type commerceSettingsRepo struct {
	service.SettingRepository
	values map[string]string
}

func (r *commerceSettingsRepo) GetValue(_ context.Context, key string) (string, error) {
	value, ok := r.values[key]
	if !ok {
		return "", service.ErrSettingNotFound
	}
	return value, nil
}

func (r *commerceSettingsRepo) Set(_ context.Context, key, value string) error {
	r.values[key] = value
	return nil
}

func TestCommerceLinksAdminSaveAndUserRead(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &commerceSettingsRepo{values: map[string]string{"other_setting": "preserved"}}
	svc := service.NewSettingService(repo, &config.Config{})
	adminHandler := adminhandler.NewSettingHandler(svc, nil, nil, nil, nil, nil, nil)
	userHandler := userhandler.NewSettingHandler(svc, "test")
	router := gin.New()
	router.PUT("/admin/settings/commerce-links", adminHandler.UpdateCommerceLinks)
	router.GET("/commerce-links", userHandler.GetCommerceLinks)
	body := `{"xianyu_order_url":"https://m.tb.cn/new?a=1&b=2","store_url":"https://store.example/new","store_backup_url":"https://backup.example/shop"}`
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/admin/settings/commerce-links", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	require.Equal(t, "no-store", rec.Header().Get("Cache-Control"))
	saved := rec.Body.String()
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/commerce-links", nil))
	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, "no-store", rec.Header().Get("Cache-Control"))
	require.JSONEq(t, saved, rec.Body.String())
	require.Equal(t, "preserved", repo.values["other_setting"])
	for _, invalid := range []string{`{broken`, `{}`, `{"xianyu_order_url":"https://m.tb.cn/new","store_url":"javascript:alert(1)"}`} {
		rec = httptest.NewRecorder()
		req = httptest.NewRequest(http.MethodPut, "/admin/settings/commerce-links", strings.NewReader(invalid))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(rec, req)
		require.Equal(t, http.StatusBadRequest, rec.Code)
	}
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/commerce-links", nil))
	require.JSONEq(t, saved, rec.Body.String())
}
