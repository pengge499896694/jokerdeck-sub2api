package routes

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type routeNodeSettings struct {
	service.SettingRepository
	values map[string]string
}

func (r *routeNodeSettings) GetValue(_ context.Context, key string) (string, error) {
	value, ok := r.values[key]
	if !ok {
		return "", service.ErrSettingNotFound
	}
	return value, nil
}
func (r *routeNodeSettings) Set(_ context.Context, key, value string) error {
	r.values[key] = value
	return nil
}
func (r *routeNodeSettings) SetMultiple(_ context.Context, values map[string]string) error {
	for key, value := range values {
		r.values[key] = value
	}
	return nil
}

func TestClientNodeRoutesAuthenticationAndContent(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	// Seed the admin compliance acknowledgement for user 1; otherwise
	// AdminComplianceGuard aborts every /admin/client write with 423 Locked
	// before the handlers under test are reached.
	ack, err := json.Marshal(service.AdminComplianceAcknowledgement{
		Version:     service.AdminComplianceVersion,
		AdminUserID: 1,
	})
	require.NoError(t, err)
	store := &routeNodeSettings{values: map[string]string{
		"admin_compliance_acknowledgement:1": string(ack),
	}}
	jwt := middleware.JWTAuthMiddleware(func(c *gin.Context) {
		if c.GetHeader("Authorization") == "" {
			middleware.AbortWithError(c, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
			return
		}
		c.Next()
	})
	admin := middleware.AdminAuthMiddleware(func(c *gin.Context) {
		if c.GetHeader("Authorization") == "" {
			middleware.AbortWithError(c, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
			return
		}
		role := service.RoleUser
		if c.GetHeader("Authorization") == "admin" {
			role = service.RoleAdmin
		}
		c.Set(string(middleware.ContextKeyUserRole), role)
		c.Set(string(middleware.ContextKeyUser), middleware.AuthSubject{UserID: 1})
		c.Next()
	})
	RegisterClientNodeRoutes(r.Group("/api/v1"), jwt, admin, middleware.AuditLogMiddleware(func(c *gin.Context) { c.Next() }), service.NewSettingService(store, nil), nil)
	req := func(method, path, role, body string) *httptest.ResponseRecorder {
		rec := httptest.NewRecorder()
		request := httptest.NewRequest(method, "/api/v1"+path, bytes.NewBufferString(body))
		request.Header.Set("Authorization", role)
		r.ServeHTTP(rec, request)
		return rec
	}
	require.Equal(t, http.StatusUnauthorized, req(http.MethodGet, "/client/node-policy", "", "").Code)
	require.Equal(t, http.StatusUnauthorized, req(http.MethodGet, "/client/node-subscription", "", "").Code)
	require.Equal(t, http.StatusForbidden, req(http.MethodPut, "/admin/client/nodes", "user", `{}`).Code)
	require.Equal(t, http.StatusForbidden, req(http.MethodPut, "/admin/client/node-policy", "user", `{}`).Code)
	require.Equal(t, http.StatusForbidden, req(http.MethodPut, "/admin/client/node-subscription", "user", `{}`).Code)
	require.Equal(t, http.StatusBadRequest, req(http.MethodPut, "/admin/client/nodes", "admin", `{"enabled":true,"disabled_nodes":[],"content":"a: *alias"}`).Code)
	require.Equal(t, http.StatusOK, req(http.MethodPut, "/admin/client/nodes", "admin", `{"enabled":true,"disabled_nodes":["edge"],"content":"proxies: []\n"}`).Code)
	policy := req(http.MethodGet, "/client/node-policy", "user", "")
	require.Equal(t, http.StatusOK, policy.Code)
	require.JSONEq(t, `{"enabled":true,"disabled_nodes":["edge"]}`, policy.Body.String())
	content := req(http.MethodGet, "/client/node-subscription", "user", "")
	require.Equal(t, http.StatusOK, content.Code)
	require.Equal(t, "proxies: []\n", content.Body.String())
	require.Contains(t, content.Header().Get("Content-Type"), "text/yaml")
	require.Contains(t, req(http.MethodGet, "/admin/client/node-subscription", "admin", "").Body.String(), `"content":"proxies: []\n"`)
}
