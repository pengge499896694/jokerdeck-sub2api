package admin

import (
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

func (h *SettingHandler) GetClientNodePolicy(c *gin.Context) {
	policy, err := h.settingService.GetGlobalClientNodePolicy(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, policy)
}

func (h *SettingHandler) UpdateClientNodePolicy(c *gin.Context) {
	var request service.ClientNodePolicy
	if err := c.ShouldBindJSON(&request); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	policy, err := h.settingService.SetGlobalClientNodePolicy(c.Request.Context(), &request)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Success(c, policy)
}

func (h *SettingHandler) ListClientNodes(c *gin.Context) {
	nodes, err := h.settingService.ListClientNodeSubscriptions(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, nodes)
}

func (h *SettingHandler) TestClientNodes(c *gin.Context) {
	var request struct {
		SubscriptionID string   `json:"subscription_id"`
		Names          []string `json:"names"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	results, err := h.settingService.TestClientNodeNodes(c.Request.Context(), request.SubscriptionID, request.Names)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, results)
}
