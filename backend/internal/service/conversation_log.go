package service

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"unicode/utf8"
)

const (
	ConversationLogsViewAll      = "conversation_logs.view_all"
	ConversationLogRetentionDays = 90
	conversationQuestionMaxBytes = 32 * 1024
	conversationAnswerMaxBytes   = 64 * 1024
	conversationLogQueueSize     = 128
	conversationLogBatchSize     = 48
	conversationLogBatchWait     = 25 * time.Millisecond
)

var allowedConversationMenuPermissions = map[string]struct{}{
	ConversationLogsViewAll: {},
}

type ConversationLogInput struct {
	RequestID         string
	UserID            int64
	UserEmail         string
	APIKeyID          int64
	APIKeyName        string
	GroupID           *int64
	GroupName         string
	Model             string
	Endpoint          string
	Stream            bool
	StatusCode        int
	Success           bool
	Question          string
	Answer            string
	QuestionBytes     int
	AnswerBytes       int
	QuestionTruncated bool
	AnswerTruncated   bool
	DurationMs        int
	CreatedAt         time.Time
}

type ConversationLog struct {
	ID                int64     `json:"id"`
	RequestID         string    `json:"request_id"`
	UserID            int64     `json:"user_id"`
	UserEmail         string    `json:"user_email"`
	APIKeyID          int64     `json:"api_key_id"`
	APIKeyName        string    `json:"api_key_name"`
	GroupID           *int64    `json:"group_id"`
	GroupName         string    `json:"group_name"`
	Model             string    `json:"model"`
	Endpoint          string    `json:"endpoint"`
	Stream            bool      `json:"stream"`
	StatusCode        int       `json:"status_code"`
	Success           bool      `json:"success"`
	Question          string    `json:"question,omitempty"`
	Answer            string    `json:"answer,omitempty"`
	QuestionBytes     int       `json:"question_bytes"`
	AnswerBytes       int       `json:"answer_bytes"`
	QuestionTruncated bool      `json:"question_truncated"`
	AnswerTruncated   bool      `json:"answer_truncated"`
	InputTokens       int       `json:"input_tokens"`
	OutputTokens      int       `json:"output_tokens"`
	DurationMs        int       `json:"duration_ms"`
	CreatedAt         time.Time `json:"created_at"`
}

type ConversationLogFilters struct {
	UserID       *int64
	APIKeyID     *int64
	GroupID      *int64
	UserSearch   string
	APIKeySearch string
	GroupSearch  string
	Model        string
	Status       string
	StartTime    *time.Time
	EndTime      *time.Time
}

type ConversationLogStorageStats struct {
	Records       int64  `json:"records"`
	TextBytes     int64  `json:"text_bytes"`
	TableBytes    int64  `json:"table_bytes"`
	QueueDropped  uint64 `json:"queue_dropped"`
	RetentionDays int    `json:"retention_days"`
}

type ConversationLogService struct {
	db       *sql.DB
	queue    chan ConversationLogInput
	done     chan struct{}
	mu       sync.RWMutex
	stopping bool
	stopOnce sync.Once
	dropped  atomic.Uint64
}

func NewConversationLogService(db *sql.DB) *ConversationLogService {
	s := &ConversationLogService{db: db, queue: make(chan ConversationLogInput, conversationLogQueueSize), done: make(chan struct{})}
	if db != nil {
		go s.run()
	} else {
		close(s.done)
	}
	return s
}

// Enqueue is intentionally non-blocking: request logging must not hold a gateway response open.
func (s *ConversationLogService) Enqueue(entry ConversationLogInput) {
	if s == nil || s.db == nil || entry.UserID <= 0 || entry.APIKeyID <= 0 {
		return
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.stopping {
		return
	}
	entry.Question, entry.QuestionBytes, entry.QuestionTruncated = clampConversationText(entry.Question, entry.QuestionBytes, entry.QuestionTruncated, conversationQuestionMaxBytes)
	entry.Answer, entry.AnswerBytes, entry.AnswerTruncated = clampConversationText(entry.Answer, entry.AnswerBytes, entry.AnswerTruncated, conversationAnswerMaxBytes)
	if entry.CreatedAt.IsZero() {
		entry.CreatedAt = time.Now().UTC()
	}
	select {
	case s.queue <- entry:
	default:
		s.dropped.Add(1)
	}
}

func clampConversationText(value string, originalBytes int, truncated bool, maxBytes int) (string, int, bool) {
	value = strings.ToValidUTF8(value, "")
	if originalBytes < len(value) {
		originalBytes = len(value)
	}
	if len(value) > maxBytes {
		value = value[:maxBytes]
		for !utf8.ValidString(value) {
			value = value[:len(value)-1]
		}
		truncated = true
	}
	return value, originalBytes, truncated
}

func (s *ConversationLogService) run() {
	defer close(s.done)
	retentionTicker := time.NewTicker(time.Hour)
	defer retentionTicker.Stop()
	for {
		select {
		case first, ok := <-s.queue:
			if !ok {
				return
			}
			batch := []ConversationLogInput{first}
			timer := time.NewTimer(conversationLogBatchWait)
		collect:
			for len(batch) < conversationLogBatchSize {
				select {
				case next, ok := <-s.queue:
					if !ok {
						break collect
					}
					batch = append(batch, next)
				case <-timer.C:
					break collect
				}
			}
			if !timer.Stop() {
				select {
				case <-timer.C:
				default:
				}
			}
			s.persist(batch)
		case <-retentionTicker.C:
			s.cleanupExpired()
		}
	}
}

func (s *ConversationLogService) persist(batch []ConversationLogInput) {
	if len(batch) == 0 || s.db == nil {
		return
	}
	const columns = 20
	values := make([]string, 0, len(batch))
	args := make([]any, 0, len(batch)*columns)
	for _, entry := range batch {
		start := len(args) + 1
		placeholders := make([]string, columns)
		for i := range placeholders {
			placeholders[i] = fmt.Sprintf("$%d", start+i)
		}
		values = append(values, "("+strings.Join(placeholders, ",")+")")
		args = append(args,
			entry.RequestID, entry.UserID, entry.UserEmail, entry.APIKeyID, entry.APIKeyName,
			entry.GroupID, entry.GroupName, entry.Model, entry.Endpoint, entry.Stream,
			entry.StatusCode, entry.Success, entry.Question, entry.Answer, entry.QuestionBytes,
			entry.AnswerBytes, entry.QuestionTruncated, entry.AnswerTruncated, entry.DurationMs, entry.CreatedAt,
		)
	}
	query := `INSERT INTO conversation_logs (
		request_id, user_id, user_email, api_key_id, api_key_name, group_id, group_name,
		model, endpoint, stream, status_code, success, question, answer, question_bytes,
		answer_bytes, question_truncated, answer_truncated, duration_ms, created_at
	) VALUES ` + strings.Join(values, ",")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if _, err := s.db.ExecContext(ctx, query, args...); err != nil {
		log.Printf("[ConversationLogs] persist batch failed: count=%d err=%v", len(batch), err)
	}
}

func (s *ConversationLogService) cleanupExpired() {
	if s == nil || s.db == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	for {
		result, err := s.db.ExecContext(ctx, `WITH expired AS (
			SELECT id FROM conversation_logs
			WHERE created_at < NOW() - ($1 * INTERVAL '1 day')
			ORDER BY created_at ASC LIMIT 5000
		) DELETE FROM conversation_logs cl USING expired WHERE cl.id = expired.id`, ConversationLogRetentionDays)
		if err != nil {
			log.Printf("[ConversationLogs] retention cleanup failed: %v", err)
			return
		}
		deleted, err := result.RowsAffected()
		if err != nil || deleted < 5000 {
			return
		}
	}
}

func (s *ConversationLogService) Stop() {
	if s == nil || s.queue == nil {
		return
	}
	s.stopOnce.Do(func() {
		s.mu.Lock()
		s.stopping = true
		close(s.queue)
		s.mu.Unlock()
	})
	select {
	case <-s.done:
	case <-time.After(8 * time.Second):
		log.Printf("[ConversationLogs] shutdown drain timed out")
	}
}

func (s *ConversationLogService) List(ctx context.Context, filters ConversationLogFilters, page, pageSize int) ([]ConversationLog, int64, error) {
	if s == nil || s.db == nil {
		return nil, 0, fmt.Errorf("conversation log service not ready")
	}
	where, args := buildConversationLogWhere(filters)
	var total int64
	if err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM conversation_logs cl "+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * pageSize
	query := `SELECT cl.id, cl.request_id, cl.user_id, cl.user_email, cl.api_key_id, cl.api_key_name,
		cl.group_id, cl.group_name, cl.model, cl.endpoint, cl.stream, cl.status_code, cl.success,
		cl.question_bytes, cl.answer_bytes, cl.question_truncated, cl.answer_truncated,
		COALESCE(ul.input_tokens, 0), COALESCE(ul.output_tokens, 0), cl.duration_ms, cl.created_at
	FROM conversation_logs cl
	LEFT JOIN LATERAL (
		SELECT input_tokens, output_tokens FROM usage_logs WHERE request_id = cl.request_id AND api_key_id = cl.api_key_id ORDER BY id DESC LIMIT 1
	) ul ON TRUE ` + where + fmt.Sprintf(" ORDER BY cl.created_at DESC, cl.id DESC LIMIT $%d OFFSET $%d", len(args)+1, len(args)+2)
	args = append(args, pageSize, offset)
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = rows.Close() }()
	items := make([]ConversationLog, 0)
	for rows.Next() {
		var item ConversationLog
		var groupID sql.NullInt64
		if err := rows.Scan(&item.ID, &item.RequestID, &item.UserID, &item.UserEmail, &item.APIKeyID, &item.APIKeyName,
			&groupID, &item.GroupName, &item.Model, &item.Endpoint, &item.Stream, &item.StatusCode, &item.Success,
			&item.QuestionBytes, &item.AnswerBytes, &item.QuestionTruncated, &item.AnswerTruncated,
			&item.InputTokens, &item.OutputTokens, &item.DurationMs, &item.CreatedAt); err != nil {
			return nil, 0, err
		}
		if groupID.Valid {
			v := groupID.Int64
			item.GroupID = &v
		}
		items = append(items, item)
	}
	return items, total, rows.Err()
}

func (s *ConversationLogService) Get(ctx context.Context, id int64) (*ConversationLog, error) {
	if s == nil || s.db == nil {
		return nil, fmt.Errorf("conversation log service not ready")
	}
	var item ConversationLog
	var groupID sql.NullInt64
	err := s.db.QueryRowContext(ctx, `SELECT cl.id, cl.request_id, cl.user_id, cl.user_email, cl.api_key_id, cl.api_key_name,
		cl.group_id, cl.group_name, cl.model, cl.endpoint, cl.stream, cl.status_code, cl.success,
		cl.question, cl.answer, cl.question_bytes, cl.answer_bytes, cl.question_truncated, cl.answer_truncated,
		COALESCE(ul.input_tokens, 0), COALESCE(ul.output_tokens, 0), cl.duration_ms, cl.created_at
	FROM conversation_logs cl LEFT JOIN LATERAL (
		SELECT input_tokens, output_tokens FROM usage_logs WHERE request_id = cl.request_id AND api_key_id = cl.api_key_id ORDER BY id DESC LIMIT 1
	) ul ON TRUE WHERE cl.id = $1`, id).Scan(
		&item.ID, &item.RequestID, &item.UserID, &item.UserEmail, &item.APIKeyID, &item.APIKeyName,
		&groupID, &item.GroupName, &item.Model, &item.Endpoint, &item.Stream, &item.StatusCode, &item.Success,
		&item.Question, &item.Answer, &item.QuestionBytes, &item.AnswerBytes, &item.QuestionTruncated, &item.AnswerTruncated,
		&item.InputTokens, &item.OutputTokens, &item.DurationMs, &item.CreatedAt)
	if err != nil {
		return nil, err
	}
	if groupID.Valid {
		v := groupID.Int64
		item.GroupID = &v
	}
	return &item, nil
}

func buildConversationLogWhere(filters ConversationLogFilters) (string, []any) {
	conditions := make([]string, 0, 6)
	args := make([]any, 0, 6)
	add := func(condition string, value any) {
		args = append(args, value)
		conditions = append(conditions, fmt.Sprintf(condition, len(args)))
	}
	if filters.UserID != nil {
		add("cl.user_id = $%d", *filters.UserID)
	}
	if filters.APIKeyID != nil {
		add("cl.api_key_id = $%d", *filters.APIKeyID)
	}
	if filters.GroupID != nil {
		add("cl.group_id = $%d", *filters.GroupID)
	}
	if value := strings.TrimSpace(filters.UserSearch); value != "" {
		add("cl.user_email ILIKE '%%' || $%d || '%%'", value)
	}
	if value := strings.TrimSpace(filters.APIKeySearch); value != "" {
		add("cl.api_key_name ILIKE '%%' || $%d || '%%'", value)
	}
	if value := strings.TrimSpace(filters.GroupSearch); value != "" {
		add("cl.group_name ILIKE '%%' || $%d || '%%'", value)
	}
	if model := strings.TrimSpace(filters.Model); model != "" {
		add("cl.model = $%d", model)
	}
	switch filters.Status {
	case "success":
		conditions = append(conditions, "cl.success = TRUE")
	case "failed":
		conditions = append(conditions, "cl.success = FALSE")
	}
	if filters.StartTime != nil {
		add("cl.created_at >= $%d", *filters.StartTime)
	}
	if filters.EndTime != nil {
		add("cl.created_at < $%d", *filters.EndTime)
	}
	if len(conditions) == 0 {
		return "", args
	}
	return "WHERE " + strings.Join(conditions, " AND "), args
}

func (s *ConversationLogService) BatchDelete(ctx context.Context, ids []int64) (int64, error) {
	if s == nil || s.db == nil {
		return 0, fmt.Errorf("conversation log service not ready")
	}
	if len(ids) == 0 || len(ids) > 500 {
		return 0, fmt.Errorf("select between 1 and 500 conversation logs")
	}
	placeholders := make([]string, len(ids))
	args := make([]any, len(ids))
	for i, id := range ids {
		if id <= 0 {
			return 0, fmt.Errorf("invalid conversation log id")
		}
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = id
	}
	result, err := s.db.ExecContext(ctx, "DELETE FROM conversation_logs WHERE id IN ("+strings.Join(placeholders, ",")+")", args...)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func (s *ConversationLogService) StorageStats(ctx context.Context) (*ConversationLogStorageStats, error) {
	if s == nil || s.db == nil {
		return nil, fmt.Errorf("conversation log service not ready")
	}
	stats := &ConversationLogStorageStats{QueueDropped: s.dropped.Load(), RetentionDays: ConversationLogRetentionDays}
	err := s.db.QueryRowContext(ctx, `SELECT COUNT(*), COALESCE(SUM(octet_length(question) + octet_length(answer)), 0),
		pg_total_relation_size('conversation_logs') FROM conversation_logs`).Scan(&stats.Records, &stats.TextBytes, &stats.TableBytes)
	return stats, err
}

func (s *ConversationLogService) ListUserMenuPermissions(ctx context.Context, userID int64) ([]string, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT menu_key FROM user_menu_permissions WHERE user_id = $1 ORDER BY menu_key`, userID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	permissions := make([]string, 0, 1)
	for rows.Next() {
		var key string
		if err := rows.Scan(&key); err != nil {
			return nil, err
		}
		permissions = append(permissions, key)
	}
	return permissions, rows.Err()
}

func (s *ConversationLogService) HasMenuPermission(ctx context.Context, userID int64, key string) (bool, error) {
	var granted bool
	err := s.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM user_menu_permissions WHERE user_id = $1 AND menu_key = $2)`, userID, key).Scan(&granted)
	return granted, err
}

func (s *ConversationLogService) SetUserMenuPermissions(ctx context.Context, userID, grantedBy int64, permissions []string) error {
	if userID <= 0 || grantedBy <= 0 {
		return fmt.Errorf("invalid user or grantor")
	}
	unique := make(map[string]struct{}, len(permissions))
	for _, permission := range permissions {
		permission = strings.TrimSpace(permission)
		if _, ok := allowedConversationMenuPermissions[permission]; !ok {
			return fmt.Errorf("unsupported menu permission")
		}
		unique[permission] = struct{}{}
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := tx.ExecContext(ctx, `DELETE FROM user_menu_permissions WHERE user_id = $1`, userID); err != nil {
		return err
	}
	for permission := range unique {
		if _, err := tx.ExecContext(ctx, `INSERT INTO user_menu_permissions (user_id, menu_key, granted_by) VALUES ($1, $2, $3)`, userID, permission, grantedBy); err != nil {
			return err
		}
	}
	return tx.Commit()
}
