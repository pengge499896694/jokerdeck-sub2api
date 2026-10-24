package handler

import (
	"bufio"
	"bytes"
	"database/sql"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/Wei-Shaw/sub2api/internal/pkg/ctxkey"
	"github.com/Wei-Shaw/sub2api/internal/pkg/requestmodel"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/tidwall/gjson"
)

const (
	conversationCaptureRequestMax  = 1 << 20
	conversationCaptureResponseMax = 128 << 10
	conversationQuestionMax        = 32 << 10
	conversationAnswerMax          = 64 << 10
)

// ConversationLogHandler exposes authorized conversation records and the capture middleware.
type ConversationLogHandler struct {
	service *service.ConversationLogService
}

func NewConversationLogHandler(logService *service.ConversationLogService) *ConversationLogHandler {
	return &ConversationLogHandler{service: logService}
}

func (h *ConversationLogHandler) CaptureMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if h == nil || h.service == nil {
			c.Next()
			return
		}
		if !isConversationCaptureRoute(c.Request.Method, c.Request.URL.Path) {
			c.Next()
			return
		}
		apiKey, ok := middleware.GetAPIKeyFromContext(c)
		if !ok || apiKey == nil || apiKey.User == nil {
			c.Next()
			return
		}

		startedAt := time.Now()
		requestBody, requestBodyTruncated := readConversationRequestBody(c.Request)
		question := extractConversationQuestion(requestBody)
		questionBytes := len(question)
		question, questionTruncated := truncateConversationText(question, conversationQuestionMax)
		questionTruncated = questionTruncated || requestBodyTruncated
		model := requestmodel.FromBodyForRoute(c.Request.URL.Path, c.GetHeader("Content-Type"), requestBody)
		stream := gjson.GetBytes(requestBody, "stream").Bool() || strings.Contains(strings.ToLower(c.Request.URL.Path), "streamgeneratecontent") || strings.Contains(strings.ToLower(c.GetHeader("Accept")), "text/event-stream")

		originalWriter := c.Writer
		captureWriter := &conversationCaptureWriter{ResponseWriter: originalWriter, maxBytes: conversationCaptureResponseMax}
		c.Writer = captureWriter
		c.Next()
		c.Writer = originalWriter

		answer, answerBytes, answerTruncated := extractConversationAnswer(captureWriter.Bytes(), captureWriter.Truncated())
		answer, answerLimitTruncated := truncateConversationText(answer, conversationAnswerMax)
		answerTruncated = answerTruncated || answerLimitTruncated
		statusCode := captureWriter.Status()
		groupID := apiKey.GroupID
		groupName := ""
		if apiKey.Group != nil {
			groupID = &apiKey.Group.ID
			groupName = apiKey.Group.Name
		}
		requestID, _ := c.Request.Context().Value(ctxkey.ClientRequestID).(string)
		if strings.TrimSpace(requestID) == "" {
			requestID = uuid.NewString()
		}
		h.service.Enqueue(service.ConversationLogInput{
			RequestID:         requestID,
			UserID:            apiKey.User.ID,
			UserEmail:         apiKey.User.Email,
			APIKeyID:          apiKey.ID,
			APIKeyName:        apiKey.Name,
			GroupID:           groupID,
			GroupName:         groupName,
			Model:             model,
			Endpoint:          c.Request.URL.Path,
			Stream:            stream,
			StatusCode:        statusCode,
			Success:           statusCode >= http.StatusOK && statusCode < http.StatusBadRequest,
			Question:          question,
			Answer:            answer,
			QuestionBytes:     questionBytes,
			AnswerBytes:       answerBytes,
			QuestionTruncated: questionTruncated,
			AnswerTruncated:   answerTruncated,
			DurationMs:        int(time.Since(startedAt).Milliseconds()),
			CreatedAt:         startedAt.UTC(),
		})
	}
}

func isConversationCaptureRoute(method, path string) bool {
	if method != http.MethodPost {
		return false
	}
	path = strings.TrimRight(strings.ToLower(path), "/")
	if strings.HasSuffix(path, "/messages") || strings.HasSuffix(path, "/chat/completions") || strings.HasSuffix(path, "/responses") || strings.HasSuffix(path, "/systemone") {
		return !strings.HasSuffix(path, "/count_tokens")
	}
	if strings.Contains(path, "/models/") && (strings.Contains(path, ":generatecontent") || strings.Contains(path, ":streamgeneratecontent")) {
		return true
	}
	return false
}

func readConversationRequestBody(req *http.Request) ([]byte, bool) {
	if req == nil || req.Body == nil {
		return nil, false
	}
	body := req.Body
	prefetched, err := io.ReadAll(io.LimitReader(body, conversationCaptureRequestMax+1))
	truncated := len(prefetched) > conversationCaptureRequestMax
	captured := prefetched
	if truncated {
		captured = captured[:conversationCaptureRequestMax]
	}
	req.Body = io.NopCloser(io.MultiReader(bytes.NewReader(prefetched), body))
	if err != nil && !errors.Is(err, io.EOF) {
		return captured, true
	}
	return captured, truncated
}

type conversationCaptureWriter struct {
	gin.ResponseWriter
	body      bytes.Buffer
	maxBytes  int
	truncated bool
}

func (w *conversationCaptureWriter) Write(data []byte) (int, error) {
	n, err := w.ResponseWriter.Write(data)
	if n > 0 {
		remaining := w.maxBytes - w.body.Len()
		if remaining > 0 {
			written := n
			if written > remaining {
				written = remaining
				w.truncated = true
			}
			_, _ = w.body.Write(data[:written])
		} else {
			w.truncated = true
		}
	}
	return n, err
}

func (w *conversationCaptureWriter) WriteString(data string) (int, error) {
	return w.Write([]byte(data))
}
func (w *conversationCaptureWriter) Bytes() []byte   { return w.body.Bytes() }
func (w *conversationCaptureWriter) Truncated() bool { return w.truncated }

func extractConversationQuestion(body []byte) string {
	if !gjson.ValidBytes(body) {
		return ""
	}
	var parts []string
	appendText := func(value string) {
		value = strings.TrimSpace(strings.ToValidUTF8(value, ""))
		if value != "" {
			parts = append(parts, value)
		}
	}
	collectContent := func(value gjson.Result) { collectConversationContent(value, &parts) }
	for _, message := range gjson.GetBytes(body, "messages").Array() {
		role := strings.ToLower(strings.TrimSpace(message.Get("role").String()))
		if role == "user" {
			collectContent(message.Get("content"))
		}
	}
	input := gjson.GetBytes(body, "input")
	if input.Type == gjson.String {
		appendText(input.String())
	} else if input.IsArray() {
		for _, item := range input.Array() {
			role := strings.ToLower(strings.TrimSpace(item.Get("role").String()))
			typ := strings.ToLower(strings.TrimSpace(item.Get("type").String()))
			if typ == "message" && role != "user" {
				continue
			}
			if role == "user" || typ == "input_text" || typ == "message" {
				collectContent(item.Get("content"))
				if typ == "input_text" {
					appendText(item.Get("text").String())
				}
			}
		}
	}
	for _, content := range gjson.GetBytes(body, "contents").Array() {
		role := strings.ToLower(strings.TrimSpace(content.Get("role").String()))
		if role == "user" || role == "" {
			collectContent(content.Get("parts"))
		}
	}
	if len(parts) == 0 {
		for _, key := range []string{"prompt", "text", "query"} {
			value := gjson.GetBytes(body, key)
			if value.Type == gjson.String {
				appendText(value.String())
				if len(parts) > 0 {
					break
				}
			}
		}
	}
	return strings.Join(parts, "\n\n")
}

func collectConversationContent(value gjson.Result, out *[]string) {
	switch {
	case value.Type == gjson.String:
		text := strings.TrimSpace(value.String())
		if text != "" {
			*out = append(*out, text)
		}
	case value.IsArray():
		for _, item := range value.Array() {
			collectConversationContent(item, out)
		}
	case value.IsObject():
		for _, key := range []string{"text", "input_text"} {
			text := value.Get(key)
			if text.Type == gjson.String && strings.TrimSpace(text.String()) != "" {
				*out = append(*out, strings.TrimSpace(text.String()))
			}
		}
		if value.Get("text").Exists() || value.Get("input_text").Exists() {
			return
		}
		for _, key := range []string{"content", "parts"} {
			if nested := value.Get(key); nested.Exists() {
				collectConversationContent(nested, out)
			}
		}
	}
}

func truncateConversationText(value string, max int) (string, bool) {
	value = strings.ToValidUTF8(value, "")
	if len(value) <= max {
		return value, false
	}
	value = value[:max]
	for !utf8.ValidString(value) {
		value = value[:len(value)-1]
	}
	return value, true
}

func extractConversationAnswer(body []byte, rawTruncated bool) (string, int, bool) {
	if len(body) == 0 {
		return "", 0, rawTruncated
	}
	// SSE carries one JSON object per data line. Prefer deltas to avoid duplicating
	// final snapshots that some protocols emit after their streamed text chunks.
	var deltas, snapshots []string
	scanner := bufio.NewScanner(bytes.NewReader(body))
	scanner.Buffer(make([]byte, 4096), conversationCaptureResponseMax+1)
	seenDataLine := false
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(line, "data:") {
			continue
		}
		seenDataLine = true
		payload := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		if payload == "" || payload == "[DONE]" || !gjson.Valid(payload) {
			continue
		}
		text, delta := responseTextFromJSON([]byte(payload))
		if text == "" {
			continue
		}
		if delta {
			deltas = append(deltas, text)
		} else {
			snapshots = append(snapshots, text)
		}
	}
	if seenDataLine {
		combined := strings.Join(deltas, "")
		if combined == "" {
			combined = strings.Join(snapshots, "")
		}
		return combined, len(combined), rawTruncated
	}
	text, _ := responseTextFromJSON(body)
	return text, len(text), rawTruncated
}

func responseTextFromJSON(data []byte) (string, bool) {
	if !gjson.ValidBytes(data) {
		return "", false
	}
	var full, deltaParts []string
	isDeltaEvent := false
	root := gjson.ParseBytes(data)
	if root.IsArray() {
		for _, item := range root.Array() {
			text, delta := responseTextFromJSON([]byte(item.Raw))
			if text != "" {
				if delta {
					deltaParts = append(deltaParts, text)
				} else {
					full = append(full, text)
				}
			}
		}
	}
	for _, choice := range root.Get("choices").Array() {
		if content := choice.Get("delta.content"); content.Exists() {
			collectResponseContent(content, &deltaParts)
			isDeltaEvent = true
		}
		if content := choice.Get("message.content"); content.Exists() {
			collectResponseContent(content, &full)
		}
	}
	if value := root.Get("output_text.delta"); value.Type == gjson.String {
		deltaParts = append(deltaParts, value.String())
		isDeltaEvent = true
	}
	if strings.Contains(strings.ToLower(root.Get("type").String()), "delta") {
		delta := root.Get("delta")
		if delta.Type == gjson.String {
			deltaParts = append(deltaParts, delta.String())
			isDeltaEvent = true
		}
		if value := delta.Get("text"); value.Type == gjson.String {
			deltaParts = append(deltaParts, value.String())
			isDeltaEvent = true
		}
	}
	if value := root.Get("output_text"); value.Type == gjson.String {
		full = append(full, value.String())
	}
	for _, item := range root.Get("output").Array() {
		collectResponseContent(item.Get("content"), &full)
	}
	for _, candidate := range root.Get("candidates").Array() {
		collectResponseContent(candidate.Get("content.parts"), &full)
	}
	collectResponseContent(root.Get("content"), &full)
	if len(deltaParts) > 0 {
		return strings.Join(deltaParts, ""), true
	}
	return strings.Join(full, ""), isDeltaEvent
}

func collectResponseContent(value gjson.Result, out *[]string) {
	switch {
	case value.Type == gjson.String:
		if strings.TrimSpace(value.String()) != "" {
			*out = append(*out, value.String())
		}
	case value.IsArray():
		for _, item := range value.Array() {
			collectResponseContent(item, out)
		}
	case value.IsObject():
		if text := value.Get("text"); text.Type == gjson.String {
			*out = append(*out, text.String())
			return
		}
		for _, key := range []string{"content", "parts"} {
			if nested := value.Get(key); nested.Exists() {
				collectResponseContent(nested, out)
			}
		}
	}
}

func (h *ConversationLogHandler) requireViewAll(c *gin.Context) bool {
	if h.service == nil {
		response.InternalError(c, "Conversation logs are unavailable")
		return false
	}
	if role, _ := middleware.GetUserRoleFromContext(c); role == service.RoleAdmin {
		return true
	}
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return false
	}
	allowed, err := h.service.HasMenuPermission(c.Request.Context(), subject.UserID, service.ConversationLogsViewAll)
	if err != nil {
		response.InternalError(c, "Failed to check menu permission")
		return false
	}
	if !allowed {
		response.Forbidden(c, "Conversation log access is not granted")
		return false
	}
	return true
}

func (h *ConversationLogHandler) List(c *gin.Context) {
	if !h.requireViewAll(c) {
		return
	}
	filters, ok := parseConversationLogFilters(c)
	if !ok {
		return
	}
	page, pageSize := response.ParsePagination(c)
	items, total, err := h.service.List(c.Request.Context(), filters, page, pageSize)
	if err != nil {
		response.InternalError(c, "Failed to query conversation logs")
		return
	}
	response.Paginated(c, items, total, page, pageSize)
}

func (h *ConversationLogHandler) Get(c *gin.Context) {
	if !h.requireViewAll(c) {
		return
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		response.BadRequest(c, "Invalid conversation log id")
		return
	}
	item, err := h.service.Get(c.Request.Context(), id)
	if errors.Is(err, sql.ErrNoRows) {
		response.NotFound(c, "Conversation log not found")
		return
	}
	if err != nil {
		response.InternalError(c, "Failed to load conversation log")
		return
	}
	response.Success(c, item)
}

func (h *ConversationLogHandler) BatchDelete(c *gin.Context) {
	var req struct {
		IDs []int64 `json:"ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}
	deleted, err := h.service.BatchDelete(c.Request.Context(), req.IDs)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Success(c, gin.H{"deleted": deleted})
}

func (h *ConversationLogHandler) StorageStats(c *gin.Context) {
	stats, err := h.service.StorageStats(c.Request.Context())
	if err != nil {
		response.InternalError(c, "Failed to read conversation log storage stats")
		return
	}
	response.Success(c, stats)
}

func (h *ConversationLogHandler) MyMenuPermissions(c *gin.Context) {
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	permissions, err := h.service.ListUserMenuPermissions(c.Request.Context(), subject.UserID)
	if err != nil {
		response.InternalError(c, "Failed to load menu permissions")
		return
	}
	role, _ := middleware.GetUserRoleFromContext(c)
	canViewAll := role == service.RoleAdmin || containsPermission(permissions, service.ConversationLogsViewAll)
	response.Success(c, gin.H{"menu_permissions": permissions, "can_view_all_conversation_logs": canViewAll})
}

func (h *ConversationLogHandler) AdminGetUserMenuPermissions(c *gin.Context) {
	userID, ok := parseMenuPermissionUserID(c)
	if !ok {
		return
	}
	permissions, err := h.service.ListUserMenuPermissions(c.Request.Context(), userID)
	if err != nil {
		response.InternalError(c, "Failed to load menu permissions")
		return
	}
	response.Success(c, gin.H{"menu_permissions": permissions})
}

func (h *ConversationLogHandler) AdminSetUserMenuPermissions(c *gin.Context) {
	userID, ok := parseMenuPermissionUserID(c)
	if !ok {
		return
	}
	var req struct {
		MenuPermissions []string `json:"menu_permissions"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}
	if len(req.MenuPermissions) > 32 {
		response.BadRequest(c, "Too many menu permissions")
		return
	}
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	if err := h.service.SetUserMenuPermissions(c.Request.Context(), userID, subject.UserID, req.MenuPermissions); err != nil {
		if strings.Contains(err.Error(), "unsupported") {
			response.BadRequest(c, "Unsupported menu permission")
			return
		}
		response.ErrorFrom(c, err)
		return
	}
	permissions, err := h.service.ListUserMenuPermissions(c.Request.Context(), userID)
	if err != nil {
		response.InternalError(c, "Failed to load menu permissions")
		return
	}
	response.Success(c, gin.H{"menu_permissions": permissions})
}

func parseMenuPermissionUserID(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		response.BadRequest(c, "Invalid user id")
		return 0, false
	}
	return id, true
}

func containsPermission(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func parseConversationLogFilters(c *gin.Context) (service.ConversationLogFilters, bool) {
	var filters service.ConversationLogFilters
	for name, target := range map[string]**int64{"user_id": &filters.UserID, "api_key_id": &filters.APIKeyID, "group_id": &filters.GroupID} {
		raw := strings.TrimSpace(c.Query(name))
		if raw == "" {
			continue
		}
		value, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || value <= 0 {
			response.BadRequest(c, "Invalid "+name)
			return filters, false
		}
		*target = &value
	}
	filters.Model = strings.TrimSpace(c.Query("model"))
	filters.UserSearch = strings.TrimSpace(c.Query("user"))
	filters.APIKeySearch = strings.TrimSpace(c.Query("api_key"))
	filters.GroupSearch = strings.TrimSpace(c.Query("group"))
	if len(filters.UserSearch) > 255 || len(filters.APIKeySearch) > 255 || len(filters.GroupSearch) > 255 || len(filters.Model) > 255 {
		response.BadRequest(c, "Filter value is too long")
		return filters, false
	}
	filters.Status = strings.TrimSpace(c.Query("status"))
	if filters.Status != "" && filters.Status != "success" && filters.Status != "failed" {
		response.BadRequest(c, "Invalid status")
		return filters, false
	}
	tz := strings.TrimSpace(c.Query("timezone"))
	if raw := strings.TrimSpace(c.Query("start_date")); raw != "" {
		value, err := parseConversationDate(raw, tz, false)
		if err != nil {
			response.BadRequest(c, "Invalid start_date")
			return filters, false
		}
		filters.StartTime = &value
	}
	if raw := strings.TrimSpace(c.Query("end_date")); raw != "" {
		value, err := parseConversationDate(raw, tz, true)
		if err != nil {
			response.BadRequest(c, "Invalid end_date")
			return filters, false
		}
		filters.EndTime = &value
	}
	return filters, true
}

func parseConversationDate(raw, timezoneName string, end bool) (time.Time, error) {
	if value, err := time.Parse(time.RFC3339, raw); err == nil {
		return value, nil
	}
	location := time.UTC
	if timezoneName != "" {
		if parsed, err := time.LoadLocation(timezoneName); err == nil {
			location = parsed
		}
	}
	value, err := time.ParseInLocation("2006-01-02", raw, location)
	if err != nil {
		return time.Time{}, err
	}
	if end {
		value = value.AddDate(0, 0, 1)
	}
	return value.UTC(), nil
}
