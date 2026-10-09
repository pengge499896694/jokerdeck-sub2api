package admin

import (
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/gin-gonic/gin"
)

func (h *SettingHandler) GetClientProviderPolicy(c *gin.Context) {
	policy, err := h.settingService.GetClientProviderPolicy(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, policy)
}

func (h *SettingHandler) UpdateClientProviderPolicy(c *gin.Context) {
	var request struct {
		AllowProviderSwitch *bool `json:"allow_provider_switch" binding:"required"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		response.BadRequest(c, "allow_provider_switch 必须为 boolean")
		return
	}
	policy, err := h.settingService.SetClientProviderPolicy(c.Request.Context(), *request.AllowProviderSwitch)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, policy)
}
