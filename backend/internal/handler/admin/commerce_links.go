package admin

import (
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

func (h *SettingHandler) GetCommerceLinks(c *gin.Context) {
	links, err := h.settingService.GetCommerceLinks(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	c.Header("Cache-Control", "no-store")
	response.Success(c, links)
}

func (h *SettingHandler) UpdateCommerceLinks(c *gin.Context) {
	var input service.CommerceLinks
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, "链接配置格式错误")
		return
	}
	links, err := h.settingService.SetCommerceLinks(c.Request.Context(), input)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	c.Header("Cache-Control", "no-store")
	response.Success(c, links)
}
