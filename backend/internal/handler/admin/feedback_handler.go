package admin

import (
	"errors"
	"strconv"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

type FeedbackHandler struct{ service *service.FeedbackService }
type feedbackReplyRequest struct {
	Body      string            `json:"body"`
	Model     string            `json:"model"`
	ErrorCode string            `json:"error_code"`
	RequestID string            `json:"request_id"`
	Metadata  map[string]string `json:"metadata"`
}

func NewFeedbackHandler(s *service.FeedbackService) *FeedbackHandler {
	return &FeedbackHandler{service: s}
}
func (h *FeedbackHandler) List(c *gin.Context) {
	page, size := response.ParsePagination(c)
	items, total, err := h.service.ListTickets(c, 0, true, strings.TrimSpace(c.Query("status")), page, size)
	if err != nil {
		response.Error(c, 500, err.Error())
		return
	}
	response.Paginated(c, items, total, page, size)
}
func (h *FeedbackHandler) Get(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		response.BadRequest(c, "Invalid ticket ID")
		return
	}
	t, err := h.service.GetTicket(c, id, 0, true)
	if err != nil {
		response.Error(c, 404, err.Error())
		return
	}
	response.Success(c, t)
}
func (h *FeedbackHandler) Reply(c *gin.Context) {
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "Admin not found in context")
		return
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		response.BadRequest(c, "Invalid ticket ID")
		return
	}
	var req feedbackReplyRequest
	if err = c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	if err = h.service.Reply(c, id, subject.UserID, true, req.Body, service.FeedbackContext{Model: req.Model, ErrorCode: req.ErrorCode, RequestID: req.RequestID, Metadata: req.Metadata}); err != nil {
		response.Error(c, 400, err.Error())
		return
	}
	t, err := h.service.GetTicket(c, id, 0, true)
	if err != nil {
		response.Error(c, 404, err.Error())
		return
	}
	response.Success(c, t)
}
func (h *FeedbackHandler) SetStatus(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		response.BadRequest(c, "Invalid ticket ID")
		return
	}
	var req struct {
		Status string `json:"status"`
	}
	if err = c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	if err = h.service.SetStatus(c, id, req.Status); err != nil {
		response.Error(c, 400, err.Error())
		return
	}
	response.Success(c, gin.H{"message": "ok"})
}

func (h *FeedbackHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		response.BadRequest(c, "Invalid ticket ID")
		return
	}
	if err := h.service.DeleteTicket(c, id); err != nil {
		if errors.Is(err, service.ErrFeedbackNotFound) {
			response.Error(c, 404, err.Error())
			return
		}
		response.Error(c, 400, err.Error())
		return
	}
	response.Success(c, gin.H{"message": "ok"})
}
