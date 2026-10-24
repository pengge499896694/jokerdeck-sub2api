package handler

import (
	"errors"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

func (h *ChannelMonitorV2Handler) GroupHealth(c *gin.Context) {
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	groups, err := h.apiKeyService.GetAvailableGroups(c.Request.Context(), subject.UserID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	allowed := make(map[int64]bool, len(groups))
	for _, group := range groups {
		allowed[group.ID] = true
	}
	rows, err := h.service.GroupHealth(c.Request.Context())
	if err != nil && !errors.Is(err, service.ErrChannelMonitorDisabled) {
		response.ErrorFrom(c, err)
		return
	}
	visible := []service.GroupHealth{}
	for _, row := range rows {
		if allowed[row.GroupID] {
			visible = append(visible, row)
		}
	}
	c.Header("Cache-Control", "private, no-store")
	response.Success(c, gin.H{"checked_at": time.Now().UTC(), "groups": visible})
}
