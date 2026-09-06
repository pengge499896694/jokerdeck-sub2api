package handler

import (
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

type XianyuRechargeHandler struct {
	service *service.XianyuRechargeService
}

func NewXianyuRechargeHandler(service *service.XianyuRechargeService) *XianyuRechargeHandler {
	return &XianyuRechargeHandler{service: service}
}

type createXianyuRechargeRequest struct {
	XianyuOrderNo string  `json:"xianyu_order_no" binding:"required,max=128"`
	PaidAmountCNY float64 `json:"paid_amount_cny" binding:"required,gt=0"`
	Note          string  `json:"note" binding:"max=1000"`
}

func (h *XianyuRechargeHandler) Create(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	var req createXianyuRechargeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	order, err := h.service.Create(c.Request.Context(), subject.UserID, service.CreateXianyuRechargeOrderInput{
		XianyuOrderNo: req.XianyuOrderNo,
		PaidAmountCNY: req.PaidAmountCNY,
		Note:          req.Note,
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, order)
}

func (h *XianyuRechargeHandler) List(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	orders, err := h.service.ListForUser(c.Request.Context(), subject.UserID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, orders)
}

