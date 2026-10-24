package admin

import (
	"strconv"

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

func (h *XianyuRechargeHandler) List(c *gin.Context) {
	page, pageSize := response.ParsePagination(c)
	orders, total, err := h.service.ListAdmin(c.Request.Context(), page, pageSize, c.Query("status"), c.Query("search"))
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Paginated(c, orders, total, page, pageSize)
}

func (h *XianyuRechargeHandler) PendingCount(c *gin.Context) {
	count, err := h.service.PendingCount(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"count": count})
}

type creditXianyuRechargeRequest struct {
	AmountUSD float64 `json:"amount_usd" binding:"required,gt=0"`
	Note      string  `json:"note" binding:"max=1000"`
}

func (h *XianyuRechargeHandler) Credit(c *gin.Context) {
	orderID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid order ID")
		return
	}
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "Admin not authenticated")
		return
	}
	var req creditXianyuRechargeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	order, err := h.service.Credit(c.Request.Context(), orderID, subject.UserID, req.AmountUSD, req.Note)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, order)
}

type rejectXianyuRechargeRequest struct {
	Reason string `json:"reason" binding:"required,max=1000"`
}

func (h *XianyuRechargeHandler) Reject(c *gin.Context) {
	orderID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid order ID")
		return
	}
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "Admin not authenticated")
		return
	}
	var req rejectXianyuRechargeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	order, err := h.service.Reject(c.Request.Context(), orderID, subject.UserID, req.Reason)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, order)
}
