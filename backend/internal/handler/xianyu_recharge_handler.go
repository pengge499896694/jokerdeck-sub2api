package handler

import (
	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

type XianyuRechargeHandler struct {
	service *service.XianyuRechargeService
	db      *dbent.Client
}

func NewXianyuRechargeHandler(rechargeService *service.XianyuRechargeService, db *dbent.Client) *XianyuRechargeHandler {
	return &XianyuRechargeHandler{service: rechargeService, db: db}
}

func (h *XianyuRechargeHandler) TermsStatus(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	accepted, err := service.RechargeTermsAccepted(c.Request.Context(), h.db, subject.UserID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, service.RechargeTermsStatus{Version: service.RechargeTermsVersion, Accepted: accepted})
}

func (h *XianyuRechargeHandler) AcceptTerms(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	var req struct {
		NotMainlandSubject bool   `json:"not_mainland_subject"`
		OutsideMainland    bool   `json:"outside_mainland"`
		NoInvoice          bool   `json:"no_invoice"`
		Typed              string `json:"typed" binding:"required"`
		Version            string `json:"version" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Version != service.RechargeTermsVersion ||
		!req.NotMainlandSubject || !req.OutsideMainland || !req.NoInvoice || req.Typed != "我已知晓并确认" {
		response.BadRequest(c, "请逐项阅读并确认当前版本的充值服务条款")
		return
	}
	if err := service.AcceptRechargeTerms(c.Request.Context(), h.db, subject.UserID, req.NotMainlandSubject, req.OutsideMainland, req.NoInvoice); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, service.RechargeTermsStatus{Version: service.RechargeTermsVersion, Accepted: true})
}

func (h *XianyuRechargeHandler) RequireTerms(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		c.Abort()
		return
	}
	if err := service.RequireRechargeTerms(c.Request.Context(), h.db, subject.UserID); err != nil {
		response.ErrorFrom(c, err)
		c.Abort()
	}
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
