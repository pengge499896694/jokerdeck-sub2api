package routes

import (
	"github.com/Wei-Shaw/sub2api/internal/handler"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/gin-gonic/gin"
)

func RegisterFinanceRoutes(
	v1 *gin.RouterGroup,
	h *handler.Handlers,
	financeAuth middleware.FinanceAuthMiddleware,
	auditLog middleware.AuditLogMiddleware,
	panelRateLimiter *middleware.PanelRateLimiter,
) {
	finance := v1.Group("/finance")
	finance.Use(gin.HandlerFunc(financeAuth))
	finance.Use(panelRateLimiter.Global())
	finance.Use(gin.HandlerFunc(auditLog))
	{
		finance.GET("/users", h.Finance.Users)
		finance.GET("/balances", h.Finance.Balances)
		finance.GET("/users/:id/balance-history", h.Finance.BalanceHistory)
		finance.GET("/groups", h.Finance.Groups)
		finance.POST("/groups", h.Finance.CreateGroup)
		finance.PUT("/groups/:id", h.Finance.UpdateGroup)
		finance.DELETE("/groups/:id", h.Finance.DeleteGroup)
		finance.PUT("/groups/:id/members", h.Finance.ReplaceGroupMembers)
		finance.GET("/recharges", h.Finance.Recharges)
		finance.GET("/usage", h.Finance.Usage)
		finance.GET("/export", h.Finance.Export)
	}
}
