package handler

import (
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/gin-gonic/gin"
)

func (h *SettingHandler) GetClientNodePolicy(c *gin.Context) {
	if _, ok := middleware.GetAuthSubjectFromContext(c); !ok {
		response.Unauthorized(c, "authentication required")
		return
	}
	policy, err := h.settingService.GetClientNodePolicy(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, policy)
}

// GetClientNodeSubscription returns the merged subscription YAML with an
// automatic latency-based proxy group for clients that support url-test.
func (h *SettingHandler) GetClientNodeSubscription(c *gin.Context) {
	if _, ok := middleware.GetAuthSubjectFromContext(c); !ok {
		response.Unauthorized(c, "authentication required")
		return
	}
	policy, err := h.settingService.GetGlobalClientNodePolicy(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	if !policy.Enabled {
		c.JSON(403, gin.H{"error": "client node connections are disabled"})
		return
	}
	subscription, err := h.settingService.GetClientNodeSubscription(c.Request.Context())
	if err != nil {
		c.JSON(502, gin.H{"error": "client node subscription unavailable"})
		return
	}
	c.Header("Cache-Control", "no-store")
	c.Header("X-Content-Type-Options", "nosniff")
	c.Data(200, "application/yaml; charset=utf-8", subscription)
}
