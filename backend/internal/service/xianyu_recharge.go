package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"html"
	"log/slog"
	"strings"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/lib/pq"
)

const (
	XianyuRechargeStatusPending  = "pending"
	XianyuRechargeStatusCredited = "credited"
	XianyuRechargeStatusRejected = "rejected"

	xianyuAdminNotifyEmail = "3856645129@qq.com"
)

type XianyuRechargeOrder struct {
	ID                int64      `json:"id"`
	UserID            int64      `json:"user_id"`
	UserEmail         string     `json:"user_email,omitempty"`
	Username          string     `json:"username,omitempty"`
	XianyuOrderNo     string     `json:"xianyu_order_no"`
	PaidAmountCNY     float64    `json:"paid_amount_cny"`
	Note              string     `json:"note,omitempty"`
	Status            string     `json:"status"`
	CreditedAmountUSD *float64   `json:"credited_amount_usd,omitempty"`
	AdminNote         string     `json:"admin_note,omitempty"`
	ReviewedBy        *int64     `json:"reviewed_by,omitempty"`
	ReviewedAt        *time.Time `json:"reviewed_at,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

type CreateXianyuRechargeOrderInput struct {
	XianyuOrderNo string
	PaidAmountCNY float64
	Note          string
}

type XianyuRechargeService struct {
	entClient    *dbent.Client
	adminService AdminService
	userRepo     UserRepository
	emailService *EmailService
}

func NewXianyuRechargeService(entClient *dbent.Client, adminService AdminService, userRepo UserRepository, emailService *EmailService) *XianyuRechargeService {
	return &XianyuRechargeService{
		entClient:    entClient,
		adminService: adminService,
		userRepo:     userRepo,
		emailService: emailService,
	}
}

func (s *XianyuRechargeService) Create(ctx context.Context, userID int64, input CreateXianyuRechargeOrderInput) (*XianyuRechargeOrder, error) {
	orderNo := strings.TrimSpace(input.XianyuOrderNo)
	note := strings.TrimSpace(input.Note)
	if userID <= 0 {
		return nil, infraerrors.Unauthorized("USER_NOT_AUTHENTICATED", "user not authenticated")
	}
	if err := RequireRechargeTerms(ctx, s.entClient, userID); err != nil {
		return nil, err
	}
	if orderNo == "" || len(orderNo) > 128 {
		return nil, infraerrors.BadRequest("XIANYU_ORDER_NO_INVALID", "闲鱼订单号不能为空，且不能超过 128 个字符")
	}
	if input.PaidAmountCNY <= 0 || input.PaidAmountCNY > 1000000 {
		return nil, infraerrors.BadRequest("XIANYU_AMOUNT_INVALID", "实付金额必须大于 0")
	}
	if len(note) > 1000 {
		return nil, infraerrors.BadRequest("XIANYU_NOTE_TOO_LONG", "备注不能超过 1000 个字符")
	}

	rows, err := s.entClient.QueryContext(ctx, `
		INSERT INTO xianyu_recharge_orders
			(user_id, xianyu_order_no, paid_amount_cny, note)
		VALUES ($1, $2, $3, NULLIF($4, ''))
		RETURNING id, user_id, xianyu_order_no, paid_amount_cny, COALESCE(note, ''),
		          status, credited_amount_usd, COALESCE(admin_note, ''), reviewed_by,
		          reviewed_at, created_at, updated_at
	`, userID, orderNo, input.PaidAmountCNY, note)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			return nil, infraerrors.Conflict("XIANYU_ORDER_DUPLICATE", "该闲鱼订单号已经提交过")
		}
		return nil, fmt.Errorf("create xianyu recharge order: %w", err)
	}
	defer rows.Close()
	if !rows.Next() {
		return nil, fmt.Errorf("create xianyu recharge order: no row returned")
	}
	order, err := scanXianyuOrder(rows, false)
	if err != nil {
		return nil, err
	}

	user, userErr := s.userRepo.GetByID(ctx, userID)
	if userErr == nil {
		order.UserEmail = user.Email
		order.Username = user.Username
		s.sendAdminSubmissionEmail(*order)
	}
	return order, nil
}

func (s *XianyuRechargeService) ListForUser(ctx context.Context, userID int64) ([]XianyuRechargeOrder, error) {
	rows, err := s.entClient.QueryContext(ctx, `
		SELECT id, user_id, xianyu_order_no, paid_amount_cny, COALESCE(note, ''),
		       status, credited_amount_usd, COALESCE(admin_note, ''), reviewed_by,
		       reviewed_at, created_at, updated_at
		FROM xianyu_recharge_orders
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT 100
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("list user xianyu recharge orders: %w", err)
	}
	defer rows.Close()
	return scanXianyuOrders(rows, false)
}

func (s *XianyuRechargeService) ListAdmin(ctx context.Context, page, pageSize int, status, search string) ([]XianyuRechargeOrder, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	status = strings.TrimSpace(status)
	search = strings.TrimSpace(search)
	if status != "" && status != XianyuRechargeStatusPending && status != XianyuRechargeStatusCredited && status != XianyuRechargeStatusRejected {
		return nil, 0, infraerrors.BadRequest("XIANYU_STATUS_INVALID", "无效的订单状态")
	}
	if len(search) > 128 {
		search = search[:128]
	}

	where := `WHERE ($1 = '' OR o.status = $1)
		AND ($2 = '' OR o.xianyu_order_no ILIKE '%' || $2 || '%' OR u.email ILIKE '%' || $2 || '%' OR u.username ILIKE '%' || $2 || '%')`
	countRows, err := s.entClient.QueryContext(ctx, `
		SELECT COUNT(*)
		FROM xianyu_recharge_orders o
		JOIN users u ON u.id = o.user_id
		`+where, status, search)
	if err != nil {
		return nil, 0, fmt.Errorf("count xianyu recharge orders: %w", err)
	}
	var total int64
	if countRows.Next() {
		err = countRows.Scan(&total)
	}
	countRows.Close()
	if err != nil {
		return nil, 0, fmt.Errorf("scan xianyu recharge order count: %w", err)
	}

	rows, err := s.entClient.QueryContext(ctx, `
		SELECT o.id, o.user_id, u.email, u.username, o.xianyu_order_no,
		       o.paid_amount_cny, COALESCE(o.note, ''), o.status,
		       o.credited_amount_usd, COALESCE(o.admin_note, ''), o.reviewed_by,
		       o.reviewed_at, o.created_at, o.updated_at
		FROM xianyu_recharge_orders o
		JOIN users u ON u.id = o.user_id
		`+where+`
		ORDER BY CASE WHEN o.status = 'pending' THEN 0 ELSE 1 END, o.created_at DESC
		LIMIT $3 OFFSET $4
	`, status, search, pageSize, (page-1)*pageSize)
	if err != nil {
		return nil, 0, fmt.Errorf("list admin xianyu recharge orders: %w", err)
	}
	defer rows.Close()
	orders, err := scanXianyuOrders(rows, true)
	return orders, total, err
}

func (s *XianyuRechargeService) PendingCount(ctx context.Context) (int64, error) {
	rows, err := s.entClient.QueryContext(ctx, `SELECT COUNT(*) FROM xianyu_recharge_orders WHERE status = 'pending'`)
	if err != nil {
		return 0, err
	}
	defer rows.Close()
	var count int64
	if !rows.Next() {
		return 0, sql.ErrNoRows
	}
	if err := rows.Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func (s *XianyuRechargeService) Credit(ctx context.Context, orderID, adminUserID int64, amountUSD float64, adminNote string) (*XianyuRechargeOrder, error) {
	if orderID <= 0 || adminUserID <= 0 {
		return nil, infraerrors.BadRequest("XIANYU_REVIEW_INVALID", "无效的订单或管理员")
	}
	if amountUSD <= 0 || amountUSD > 1000000 {
		return nil, infraerrors.BadRequest("XIANYU_CREDIT_AMOUNT_INVALID", "入账美元金额必须大于 0")
	}
	adminNote = strings.TrimSpace(adminNote)
	if len(adminNote) > 1000 {
		return nil, infraerrors.BadRequest("XIANYU_ADMIN_NOTE_TOO_LONG", "管理员备注不能超过 1000 个字符")
	}

	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()
	txCtx := dbent.NewTxContext(ctx, tx)
	txClient := tx.Client()

	rows, err := txClient.QueryContext(txCtx, `
		SELECT o.id, o.user_id, u.email, u.username, o.xianyu_order_no,
		       o.paid_amount_cny, COALESCE(o.note, ''), o.status,
		       o.credited_amount_usd, COALESCE(o.admin_note, ''), o.reviewed_by,
		       o.reviewed_at, o.created_at, o.updated_at
		FROM xianyu_recharge_orders o
		JOIN users u ON u.id = o.user_id
		WHERE o.id = $1
		FOR UPDATE OF o
	`, orderID)
	if err != nil {
		return nil, err
	}
	if !rows.Next() {
		rows.Close()
		return nil, infraerrors.NotFound("XIANYU_ORDER_NOT_FOUND", "充值订单不存在")
	}
	order, err := scanXianyuOrder(rows, true)
	rows.Close()
	if err != nil {
		return nil, err
	}
	if order.Status != XianyuRechargeStatusPending {
		return nil, infraerrors.Conflict("XIANYU_ORDER_ALREADY_REVIEWED", "该订单已经处理，不能重复入账")
	}

	notes := fmt.Sprintf("闲鱼订单 %s", order.XianyuOrderNo)
	if adminNote != "" {
		notes += "；" + adminNote
	}
	if _, err := s.adminService.UpdateUserBalance(txCtx, order.UserID, amountUSD, "add", notes); err != nil {
		return nil, fmt.Errorf("credit xianyu recharge balance: %w", err)
	}

	now := time.Now()
	updateRows, err := txClient.QueryContext(txCtx, `
		UPDATE xianyu_recharge_orders
		SET status = 'credited', credited_amount_usd = $2, admin_note = NULLIF($3, ''),
		    reviewed_by = $4, reviewed_at = $5, updated_at = $5
		WHERE id = $1 AND status = 'pending'
		RETURNING id, user_id, xianyu_order_no, paid_amount_cny, COALESCE(note, ''),
		          status, credited_amount_usd, COALESCE(admin_note, ''), reviewed_by,
		          reviewed_at, created_at, updated_at
	`, orderID, amountUSD, adminNote, adminUserID, now)
	if err != nil {
		return nil, err
	}
	if !updateRows.Next() {
		updateRows.Close()
		return nil, infraerrors.Conflict("XIANYU_ORDER_ALREADY_REVIEWED", "该订单已经处理，不能重复入账")
	}
	updated, err := scanXianyuOrder(updateRows, false)
	updateRows.Close()
	if err != nil {
		return nil, err
	}
	updated.UserEmail = order.UserEmail
	updated.Username = order.Username
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	s.sendUserReviewEmail(*updated, true)
	return updated, nil
}

func (s *XianyuRechargeService) Reject(ctx context.Context, orderID, adminUserID int64, reason string) (*XianyuRechargeOrder, error) {
	reason = strings.TrimSpace(reason)
	if reason == "" || len(reason) > 1000 {
		return nil, infraerrors.BadRequest("XIANYU_REJECT_REASON_INVALID", "请填写驳回原因，且不能超过 1000 个字符")
	}
	rows, err := s.entClient.QueryContext(ctx, `
		UPDATE xianyu_recharge_orders
		SET status = 'rejected', admin_note = $2, reviewed_by = $3,
		    reviewed_at = NOW(), updated_at = NOW()
		WHERE id = $1 AND status = 'pending'
		RETURNING id, user_id, xianyu_order_no, paid_amount_cny, COALESCE(note, ''),
		          status, credited_amount_usd, COALESCE(admin_note, ''), reviewed_by,
		          reviewed_at, created_at, updated_at
	`, orderID, reason, adminUserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	if !rows.Next() {
		return nil, infraerrors.Conflict("XIANYU_ORDER_ALREADY_REVIEWED", "订单不存在或已经处理")
	}
	order, err := scanXianyuOrder(rows, false)
	if err != nil {
		return nil, err
	}
	if user, userErr := s.userRepo.GetByID(ctx, order.UserID); userErr == nil {
		order.UserEmail = user.Email
		order.Username = user.Username
		s.sendUserReviewEmail(*order, false)
	}
	return order, nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanXianyuOrder(row rowScanner, withUser bool) (*XianyuRechargeOrder, error) {
	var order XianyuRechargeOrder
	var credited sql.NullFloat64
	var reviewedBy sql.NullInt64
	var reviewedAt sql.NullTime
	var err error
	if withUser {
		err = row.Scan(&order.ID, &order.UserID, &order.UserEmail, &order.Username,
			&order.XianyuOrderNo, &order.PaidAmountCNY, &order.Note, &order.Status,
			&credited, &order.AdminNote, &reviewedBy, &reviewedAt, &order.CreatedAt, &order.UpdatedAt)
	} else {
		err = row.Scan(&order.ID, &order.UserID, &order.XianyuOrderNo, &order.PaidAmountCNY,
			&order.Note, &order.Status, &credited, &order.AdminNote, &reviewedBy,
			&reviewedAt, &order.CreatedAt, &order.UpdatedAt)
	}
	if err != nil {
		return nil, fmt.Errorf("scan xianyu recharge order: %w", err)
	}
	if credited.Valid {
		order.CreditedAmountUSD = &credited.Float64
	}
	if reviewedBy.Valid {
		order.ReviewedBy = &reviewedBy.Int64
	}
	if reviewedAt.Valid {
		order.ReviewedAt = &reviewedAt.Time
	}
	return &order, nil
}

func scanXianyuOrders(rows *sql.Rows, withUser bool) ([]XianyuRechargeOrder, error) {
	orders := make([]XianyuRechargeOrder, 0)
	for rows.Next() {
		order, err := scanXianyuOrder(rows, withUser)
		if err != nil {
			return nil, err
		}
		orders = append(orders, *order)
	}
	return orders, rows.Err()
}

func (s *XianyuRechargeService) sendAdminSubmissionEmail(order XianyuRechargeOrder) {
	subject := fmt.Sprintf("[Sub2API] 新的闲鱼充值订单：%s", order.XianyuOrderNo)
	body := fmt.Sprintf(`<h2>新的闲鱼充值订单</h2>
<p>用户：%s</p><p>闲鱼订单号：%s</p><p>实付金额：¥%.2f</p><p>备注：%s</p>
<p>请登录 Sub2API 管理后台，在“充值订单”中审核。</p>`,
		html.EscapeString(order.UserEmail), html.EscapeString(order.XianyuOrderNo),
		order.PaidAmountCNY, html.EscapeString(order.Note))
	s.sendEmailAsync(xianyuAdminNotifyEmail, subject, body)
}

func (s *XianyuRechargeService) sendUserReviewEmail(order XianyuRechargeOrder, credited bool) {
	if strings.TrimSpace(order.UserEmail) == "" {
		return
	}
	subject := fmt.Sprintf("[Sub2API] 闲鱼充值订单已%s：%s", map[bool]string{true: "入账", false: "驳回"}[credited], order.XianyuOrderNo)
	detail := fmt.Sprintf("驳回原因：%s", html.EscapeString(order.AdminNote))
	if credited && order.CreditedAmountUSD != nil {
		detail = fmt.Sprintf("已入账金额：$%.2f", *order.CreditedAmountUSD)
	}
	body := fmt.Sprintf(`<h2>闲鱼充值订单处理结果</h2><p>订单号：%s</p><p>%s</p><p>可登录充值中心查看最新状态。</p>`,
		html.EscapeString(order.XianyuOrderNo), detail)
	s.sendEmailAsync(order.UserEmail, subject, body)
}

func (s *XianyuRechargeService) sendEmailAsync(to, subject, body string) {
	if s.emailService == nil || strings.TrimSpace(to) == "" {
		return
	}
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := s.emailService.SendEmail(ctx, to, subject, body); err != nil {
			slog.Warn("failed to send xianyu recharge email", "to", to, "error", err)
		}
	}()
}
