package finance

import (
	"encoding/csv"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

const (
	maxFinanceExportRows = 100000
)

type Handler struct {
	service *service.FinanceService
}

func NewHandler(financeService *service.FinanceService) *Handler {
	return &Handler{service: financeService}
}

type groupRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type groupMembersRequest struct {
	UserIDs []int64 `json:"user_ids"`
}

type financePage struct {
	Items       any                            `json:"items"`
	Total       int64                          `json:"total"`
	Page        int                            `json:"page"`
	PageSize    int                            `json:"page_size"`
	Pages       int                            `json:"pages"`
	TotalAmount float64                        `json:"total_amount"`
	Kind        string                         `json:"kind"`
	Filters     map[string]string              `json:"filters,omitempty"`
}

func (h *Handler) Users(c *gin.Context) {
	items, err := h.service.ListUsers(c.Request.Context(), c.Query("search"))
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"items": items})
}

func (h *Handler) Groups(c *gin.Context) {
	items, err := h.service.ListGroups(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"items": items})
}

func (h *Handler) CreateGroup(c *gin.Context) {
	var req groupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}
	name, description, err := normalizeGroupInput(req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok || subject.UserID <= 0 {
		response.Unauthorized(c, "User not found in context")
		return
	}
	item, err := h.service.CreateGroup(c.Request.Context(), subject.UserID, name, description)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Created(c, item)
}

func (h *Handler) UpdateGroup(c *gin.Context) {
	id, ok := parsePositiveID(c)
	if !ok {
		return
	}
	var req groupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}
	name, description, err := normalizeGroupInput(req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	item, err := h.service.UpdateGroup(c.Request.Context(), id, name, description)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, item)
}

func (h *Handler) DeleteGroup(c *gin.Context) {
	id, ok := parsePositiveID(c)
	if !ok {
		return
	}
	if err := h.service.DeleteGroup(c.Request.Context(), id); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"deleted": true})
}

func (h *Handler) ReplaceGroupMembers(c *gin.Context) {
	id, ok := parsePositiveID(c)
	if !ok {
		return
	}
	var req groupMembersRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}
	seen := make(map[int64]struct{}, len(req.UserIDs))
	for _, userID := range req.UserIDs {
		if userID <= 0 {
			response.BadRequest(c, "Invalid user id")
			return
		}
		if _, exists := seen[userID]; exists {
			response.BadRequest(c, "Duplicate user id")
			return
		}
		seen[userID] = struct{}{}
	}
	if err := h.service.ReplaceGroupMembers(c.Request.Context(), id, req.UserIDs); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"saved": true})
}

func (h *Handler) Recharges(c *gin.Context) {
	filters, err := parseFinanceFilters(c)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	items, page, totalAmount, err := h.service.ListRecharges(c.Request.Context(), filters)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, financePage{
		Items: items, Total: page.Total, Page: page.Page, PageSize: page.PageSize,
		Pages: page.Pages, TotalAmount: totalAmount, Kind: "recharge",
	})
}

func (h *Handler) Usage(c *gin.Context) {
	filters, err := parseFinanceFilters(c)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	items, page, totalAmount, err := h.service.ListUsage(c.Request.Context(), filters)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, financePage{
		Items: items, Total: page.Total, Page: page.Page, PageSize: page.PageSize,
		Pages: page.Pages, TotalAmount: totalAmount, Kind: "usage",
	})
}

func (h *Handler) Export(c *gin.Context) {
	filters, err := parseFinanceFilters(c)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	filters.Page = 1
	filters.PageSize = 1000
	// Keep new records from shifting OFFSET pages during a long export.
	snapshotTime := time.Now()
	if filters.EndTime == nil || filters.EndTime.After(snapshotTime) {
		filters.EndTime = &snapshotTime
	}
	kind := strings.ToLower(strings.TrimSpace(c.DefaultQuery("type", "recharge")))
	if kind != "recharge" && kind != "usage" {
		response.BadRequest(c, "type must be recharge or usage")
		return
	}

	var rows []any
	for filters.Page <= 1000 {
		var pageTotal int64
		var pageAmount float64
		if kind == "usage" {
			items, page, amount, listErr := h.service.ListUsage(c.Request.Context(), filters)
			if listErr != nil {
				response.ErrorFrom(c, listErr)
				return
			}
			pageTotal, pageAmount = page.Total, amount
			if pageTotal > maxFinanceExportRows {
				response.BadRequest(c, "Export exceeds 100000 rows; narrow the filters")
				return
			}
			for _, item := range items {
				rows = append(rows, item)
			}
			if len(items) == 0 || len(rows) >= maxFinanceExportRows || int64(filters.Page*filters.PageSize) >= pageTotal {
				break
			}
		} else {
			items, page, amount, listErr := h.service.ListRecharges(c.Request.Context(), filters)
			if listErr != nil {
				response.ErrorFrom(c, listErr)
				return
			}
			pageTotal, pageAmount = page.Total, amount
			if pageTotal > maxFinanceExportRows {
				response.BadRequest(c, "Export exceeds 100000 rows; narrow the filters")
				return
			}
			for _, item := range items {
				rows = append(rows, item)
			}
			if len(items) == 0 || len(rows) >= maxFinanceExportRows || int64(filters.Page*filters.PageSize) >= pageTotal {
				break
			}
		}
		_ = pageAmount
		_ = pageTotal
		filters.Page++
	}
	if len(rows) > maxFinanceExportRows {
		rows = rows[:maxFinanceExportRows]
	}

	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="finance-%s-%s.csv"`, kind, time.Now().Format("20060102-150405")))
	c.Status(http.StatusOK)
	_, _ = c.Writer.Write([]byte{0xEF, 0xBB, 0xBF})
	writer := csv.NewWriter(c.Writer)
	if kind == "usage" {
		_ = writer.Write([]string{"记录ID", "用户ID", "邮箱", "用户名", "财务分组", "模型", "请求ID", "输入Token", "输出Token", "总Token", "实际费用USD", "时间"})
		for _, raw := range rows {
			item := raw.(service.FinanceUsage)
			_ = writer.Write([]string{
				strconv.FormatInt(item.ID, 10), strconv.FormatInt(item.UserID, 10), safeCSVText(item.Email), safeCSVText(item.Username),
				safeCSVText(item.GroupName), safeCSVText(item.Model), safeCSVText(item.RequestID), strconv.Itoa(item.InputTokens), strconv.Itoa(item.OutputTokens),
				strconv.Itoa(item.TotalTokens), formatMoney(item.ActualCost), item.CreatedAt.Format(time.RFC3339),
			})
		}
	} else {
		_ = writer.Write([]string{"记录ID", "用户ID", "邮箱", "用户名", "财务分组", "类型", "金额USD", "状态", "备注", "使用时间", "创建时间"})
		for _, raw := range rows {
			item := raw.(service.FinanceRecharge)
			usedAt := ""
			if item.UsedAt != nil {
				usedAt = item.UsedAt.Format(time.RFC3339)
			}
			_ = writer.Write([]string{
				strconv.FormatInt(item.ID, 10), strconv.FormatInt(item.UserID, 10), safeCSVText(item.Email), safeCSVText(item.Username),
				safeCSVText(item.GroupName), item.Type, formatMoney(item.Value), item.Status, safeCSVText(item.Notes), usedAt,
				item.CreatedAt.Format(time.RFC3339),
			})
		}
	}
	writer.Flush()
}

func parseFinanceFilters(c *gin.Context) (service.FinanceFilters, error) {
	page, pageSize := response.ParsePagination(c)
	f := service.FinanceFilters{Page: page, PageSize: pageSize, Search: strings.TrimSpace(c.Query("search"))}
	if raw := strings.TrimSpace(c.Query("user_id")); raw != "" {
		id, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || id <= 0 {
			return f, errors.New("invalid user_id")
		}
		f.UserID = id
	}
	if raw := strings.TrimSpace(c.Query("group_id")); raw != "" {
		id, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || id <= 0 {
			return f, errors.New("invalid group_id")
		}
		f.GroupID = id
	}
	month := strings.TrimSpace(c.Query("month"))
	startMonth := strings.TrimSpace(c.Query("start_month"))
	endMonth := strings.TrimSpace(c.Query("end_month"))
	if month != "" && (startMonth != "" || endMonth != "") {
		return f, errors.New("month cannot be combined with start_month or end_month")
	}
	if month != "" {
		start, err := time.ParseInLocation("2006-01", month, time.Local)
		if err != nil {
			return f, errors.New("month must use YYYY-MM")
		}
		next := start.AddDate(0, 1, 0)
		f.StartTime, f.EndTime = &start, &next
	}
	if startMonth != "" {
		start, err := time.ParseInLocation("2006-01", startMonth, time.Local)
		if err != nil {
			return f, errors.New("start_month must use YYYY-MM")
		}
		f.StartTime = &start
	}
	if endMonth != "" {
		end, err := time.ParseInLocation("2006-01", endMonth, time.Local)
		if err != nil {
			return f, errors.New("end_month must use YYYY-MM")
		}
		next := end.AddDate(0, 1, 0)
		f.EndTime = &next
	}
	if f.StartTime != nil && f.EndTime != nil && !f.StartTime.Before(*f.EndTime) {
		return f, errors.New("start_month must not be later than end_month")
	}
	return f, nil
}

func normalizeGroupInput(req groupRequest) (string, string, error) {
	name := strings.TrimSpace(req.Name)
	description := strings.TrimSpace(req.Description)
	if name == "" || len(name) > 100 {
		return "", "", errors.New("group name must be 1-100 characters")
	}
	if len(description) > 500 {
		return "", "", errors.New("group description is too long")
	}
	return name, description, nil
}

func parsePositiveID(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		response.BadRequest(c, "Invalid id")
		return 0, false
	}
	return id, true
}

func formatMoney(value float64) string {
	return strconv.FormatFloat(value, 'f', 10, 64)
}

func safeCSVText(value string) string {
	if strings.ContainsAny(value, "\r\n") {
		value = strings.NewReplacer("\r", " ", "\n", " ").Replace(value)
	}
	if trimmed := strings.TrimLeft(value, " \t"); trimmed != "" && strings.ContainsRune("=+-@", rune(trimmed[0])) {
		return "'" + value
	}
	return value
}
