package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
)

type feedbackRepository struct{ db *sql.DB }

func NewFeedbackRepository(db *sql.DB) service.FeedbackRepository { return &feedbackRepository{db: db} }

func (r *feedbackRepository) CreateTicket(ctx context.Context, t *service.FeedbackTicket, body string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()
	err = tx.QueryRowContext(ctx, `INSERT INTO feedback_tickets (user_id,subject,category,status,priority,model,error_code,request_id,context)
VALUES ($1,$2,$3,'open','normal',NULLIF($4,''),NULLIF($5,''),NULLIF($6,''),$7::jsonb)
RETURNING id,created_at,updated_at,last_user_at`, t.UserID, t.Subject, t.Category, t.Model, t.ErrorCode, t.RequestID, string(t.Context)).Scan(&t.ID, &t.CreatedAt, &t.UpdatedAt, &t.LastUserAt)
	if err != nil {
		return err
	}
	m := &service.FeedbackMessage{TicketID: t.ID, AuthorType: "user", AuthorUserID: &t.UserID, Body: body, Context: []byte(`{}`)}
	err = tx.QueryRowContext(ctx, `INSERT INTO feedback_messages (ticket_id,author_type,author_user_id,body,context) VALUES ($1,'user',$2,$3,$4::jsonb) RETURNING id,created_at`, m.TicketID, t.UserID, body, string(m.Context)).Scan(&m.ID, &m.CreatedAt)
	if err != nil {
		return err
	}
	t.Messages = []service.FeedbackMessage{*m}
	err = tx.Commit()
	return err
}

func (r *feedbackRepository) ListTickets(ctx context.Context, userID int64, admin bool, status string, page, pageSize int) ([]service.FeedbackTicket, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	where := []string{"1=1"}
	args := []any{}
	if !admin {
		args = append(args, userID)
		where = append(where, fmt.Sprintf("t.user_id=$%d", len(args)))
	}
	if status != "" {
		args = append(args, status)
		where = append(where, fmt.Sprintf("t.status=$%d", len(args)))
	}
	cond := strings.Join(where, " AND ")
	var total int64
	if err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM feedback_tickets t WHERE "+cond, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	args = append(args, pageSize, (page-1)*pageSize)
	rows, err := r.db.QueryContext(ctx, `SELECT t.id,t.user_id,COALESCE(u.email,''),t.subject,t.category,t.status,t.priority,COALESCE(t.model,''),COALESCE(t.error_code,''),COALESCE(t.request_id,''),t.context,t.last_user_at,t.last_admin_at,t.last_reminded_at,t.closed_at,t.created_at,t.updated_at
FROM feedback_tickets t LEFT JOIN users u ON u.id=t.user_id WHERE `+cond+` ORDER BY t.updated_at DESC,t.id DESC LIMIT $`+fmt.Sprint(len(args)-1)+` OFFSET $`+fmt.Sprint(len(args)), args...)
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = rows.Close() }()
	items := make([]service.FeedbackTicket, 0)
	for rows.Next() {
		var t service.FeedbackTicket
		var contextBytes []byte
		var adminAt, remindedAt, closedAt sql.NullTime
		if err := rows.Scan(&t.ID, &t.UserID, &t.UserEmail, &t.Subject, &t.Category, &t.Status, &t.Priority, &t.Model, &t.ErrorCode, &t.RequestID, &contextBytes, &t.LastUserAt, &adminAt, &remindedAt, &closedAt, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, 0, err
		}
		t.Context = json.RawMessage(contextBytes)
		t.LastAdminAt = nullTimePtr(adminAt)
		t.LastRemindedAt = nullTimePtr(remindedAt)
		t.ClosedAt = nullTimePtr(closedAt)
		items = append(items, t)
	}
	return items, total, rows.Err()
}

func (r *feedbackRepository) GetTicket(ctx context.Context, id, userID int64, admin bool) (*service.FeedbackTicket, error) {
	args := []any{id}
	condition := "t.id=$1"
	if !admin {
		args = append(args, userID)
		condition += " AND t.user_id=$2"
	}
	var t service.FeedbackTicket
	var contextBytes []byte
	var adminAt, remindedAt, closedAt sql.NullTime
	err := r.db.QueryRowContext(ctx, `SELECT t.id,t.user_id,COALESCE(u.email,''),t.subject,t.category,t.status,t.priority,COALESCE(t.model,''),COALESCE(t.error_code,''),COALESCE(t.request_id,''),t.context,t.last_user_at,t.last_admin_at,t.last_reminded_at,t.closed_at,t.created_at,t.updated_at FROM feedback_tickets t LEFT JOIN users u ON u.id=t.user_id WHERE `+condition, args...).Scan(&t.ID, &t.UserID, &t.UserEmail, &t.Subject, &t.Category, &t.Status, &t.Priority, &t.Model, &t.ErrorCode, &t.RequestID, &contextBytes, &t.LastUserAt, &adminAt, &remindedAt, &closedAt, &t.CreatedAt, &t.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, service.ErrFeedbackNotFound
	}
	if err != nil {
		return nil, err
	}
	t.Context = json.RawMessage(contextBytes)
	t.LastAdminAt = nullTimePtr(adminAt)
	t.LastRemindedAt = nullTimePtr(remindedAt)
	t.ClosedAt = nullTimePtr(closedAt)
	rows, err := r.db.QueryContext(ctx, `SELECT m.id,m.ticket_id,m.author_type,m.author_user_id,COALESCE(u.email,''),m.body,m.context,m.created_at FROM feedback_messages m LEFT JOIN users u ON u.id=m.author_user_id WHERE m.ticket_id=$1 ORDER BY m.created_at ASC,m.id ASC`, id)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	for rows.Next() {
		var m service.FeedbackMessage
		var author sql.NullInt64
		var b []byte
		if err := rows.Scan(&m.ID, &m.TicketID, &m.AuthorType, &author, &m.AuthorEmail, &m.Body, &b, &m.CreatedAt); err != nil {
			return nil, err
		}
		if author.Valid {
			m.AuthorUserID = &author.Int64
		}
		m.Context = json.RawMessage(b)
		t.Messages = append(t.Messages, m)
	}
	return &t, rows.Err()
}

func (r *feedbackRepository) AddMessage(ctx context.Context, id int64, m *service.FeedbackMessage, admin bool) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()
	var exists int
	query, args := "SELECT 1 FROM feedback_tickets WHERE id=$1", []any{id}
	if !admin && m.AuthorUserID != nil {
		args = append(args, *m.AuthorUserID)
		query += " AND user_id=$2"
	}
	if err = tx.QueryRowContext(ctx, query, args...).Scan(&exists); errors.Is(err, sql.ErrNoRows) {
		return service.ErrFeedbackNotFound
	}
	if err != nil {
		return err
	}
	if err = tx.QueryRowContext(ctx, `INSERT INTO feedback_messages (ticket_id,author_type,author_user_id,body,context) VALUES ($1,$2,$3,$4,$5::jsonb) RETURNING id,created_at`, id, m.AuthorType, m.AuthorUserID, m.Body, string(m.Context)).Scan(&m.ID, &m.CreatedAt); err != nil {
		return err
	}
	if admin {
		_, err = tx.ExecContext(ctx, "UPDATE feedback_tickets SET last_admin_at=NOW(),status='pending_user',updated_at=NOW(),last_reminded_at=NULL WHERE id=$1", id)
	} else {
		_, err = tx.ExecContext(ctx, "UPDATE feedback_tickets SET last_user_at=NOW(),status='pending_admin',updated_at=NOW(),last_reminded_at=NULL WHERE id=$1", id)
	}
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (r *feedbackRepository) SetStatus(ctx context.Context, id int64, status string) error {
	result, err := r.db.ExecContext(ctx, `UPDATE feedback_tickets
		SET status=$2::text,
			closed_at=CASE WHEN $2::text IN ('resolved','closed') THEN NOW() ELSE NULL END,
			updated_at=NOW()
		WHERE id=$1`, id, status)
	if err != nil {
		return err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return service.ErrFeedbackNotFound
	}
	return nil
}

func (r *feedbackRepository) DeleteTicket(ctx context.Context, id int64) error {
	result, err := r.db.ExecContext(ctx, "DELETE FROM feedback_tickets WHERE id=$1", id)
	if err != nil {
		return err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return service.ErrFeedbackNotFound
	}
	return nil
}

func (r *feedbackRepository) ListUnanswered(ctx context.Context, before time.Time) ([]service.FeedbackTicket, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id,user_id,subject,category,status,priority,COALESCE(model,''),COALESCE(error_code,''),COALESCE(request_id,''),context,last_user_at,last_admin_at,last_reminded_at,closed_at,created_at,updated_at FROM feedback_tickets WHERE status IN ('open','pending_admin') AND last_user_at < $1 AND (last_admin_at IS NULL OR last_admin_at < last_user_at) AND (last_reminded_at IS NULL OR last_reminded_at < $1) ORDER BY last_user_at ASC LIMIT 100`, before)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	out := make([]service.FeedbackTicket, 0)
	for rows.Next() {
		var t service.FeedbackTicket
		var b []byte
		var aa, ra, ca sql.NullTime
		if err := rows.Scan(&t.ID, &t.UserID, &t.Subject, &t.Category, &t.Status, &t.Priority, &t.Model, &t.ErrorCode, &t.RequestID, &b, &t.LastUserAt, &aa, &ra, &ca, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		t.Context = json.RawMessage(b)
		t.LastAdminAt = nullTimePtr(aa)
		t.LastRemindedAt = nullTimePtr(ra)
		t.ClosedAt = nullTimePtr(ca)
		out = append(out, t)
	}
	return out, rows.Err()
}
func (r *feedbackRepository) MarkReminded(ctx context.Context, id int64, at time.Time) error {
	_, err := r.db.ExecContext(ctx, "UPDATE feedback_tickets SET last_reminded_at=$2 WHERE id=$1", id, at)
	return err
}
func nullTimePtr(v sql.NullTime) *time.Time {
	if !v.Valid {
		return nil
	}
	t := v.Time
	return &t
}
