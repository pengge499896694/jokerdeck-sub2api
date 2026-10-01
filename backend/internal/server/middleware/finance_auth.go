package middleware

import (
	"errors"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

// NewFinanceAuthMiddleware allows admin and finance users to access only the
// dedicated finance API. It intentionally does not broaden admin routes.
func NewFinanceAuthMiddleware(
	authService *service.AuthService,
	userService *service.UserService,
	settingService *service.SettingService,
	auditService *service.AuditLogService,
) FinanceAuthMiddleware {
	return FinanceAuthMiddleware(func(c *gin.Context) {
		authHeader := strings.TrimSpace(c.GetHeader("Authorization"))
		if authHeader == "" {
			AbortWithError(c, 401, "UNAUTHORIZED", "Authorization required")
			return
		}
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || strings.TrimSpace(parts[1]) == "" {
			AbortWithError(c, 401, "UNAUTHORIZED", "Authorization required")
			return
		}

		claims, err := authService.ValidateToken(strings.TrimSpace(parts[1]))
		if err != nil {
			if errors.Is(err, service.ErrTokenExpired) {
				AbortWithError(c, 401, "TOKEN_EXPIRED", "Token has expired")
			} else {
				AbortWithError(c, 401, "INVALID_TOKEN", "Invalid token")
			}
			return
		}
		user, err := userService.GetByID(c.Request.Context(), claims.UserID)
		if err != nil {
			AbortWithError(c, 401, "USER_NOT_FOUND", "User not found")
			return
		}
		if !user.IsActive() {
			AbortWithError(c, 401, "USER_INACTIVE", "User account is not active")
			return
		}
		if claims.TokenVersion != user.TokenVersion {
			AbortWithError(c, 401, "TOKEN_REVOKED", "Token has been revoked")
			return
		}
		if !user.IsAdmin() && !user.IsFinance() {
			AbortWithError(c, 403, "FORBIDDEN", "Finance access required")
			return
		}
		if !enforceSessionBinding(c, authService, settingService, auditService, claims) {
			return
		}

		c.Set(string(ContextKeyUser), AuthSubject{UserID: user.ID, Concurrency: user.Concurrency})
		c.Set(string(ContextKeyUserRole), user.Role)
		c.Set(ContextKeyAuthEmail, user.Email)
		c.Set(ContextKeySessionID, claims.SessionID)
		c.Set("auth_method", "jwt")
		c.Next()
	})
}
