package admin

import (
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

func (h *SettingHandler) GetCanvasNetworkSettings(c *gin.Context) {
	settings, err := h.settingService.GetCanvasNetworkSettings(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	c.Header("Cache-Control", "no-store")
	response.Success(c, settings)
}

func (h *SettingHandler) UpdateCanvasNetworkSettings(c *gin.Context) {
	var input service.CanvasNetworkSettings
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, "画布线路配置格式错误")
		return
	}
	settings, err := h.settingService.SetCanvasNetworkSettings(c.Request.Context(), input)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	c.Header("Cache-Control", "no-store")
	response.Success(c, settings)
}
