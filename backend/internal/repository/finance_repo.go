package repository

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

type financeRepository struct{ db *sql.DB }

func NewFinanceRepository(db *sql.DB) service.FinanceRepository {
	return &financeRepository{db: db}
}

func (r *financeRepository) ListUsers(ctx context.Context, search string) ([]service.FinanceUser, error) {
	args := []any{}
	where := "WHERE u.deleted_at IS NULL"
	if search = strings.TrimSpace(search); search != "" {
		args = append(args, "%"+search+"%")
		where += fmt.Sprintf(" AND (u.email ILIKE $%d OR COALESCE(u.username, '') ILIKE $%d)", len(args), len(args))
	}
	rows, err := r.db.QueryContext(ctx, `SELECT u.id, u.email, COALESCE(u.username, ''), COALESCE(u.balance, 0), COALESCE(u.concurrency, 0), fgm.group_id, COALESCE(fg.name, '')
		FROM users u
		LEFT JOIN finance_group_members fgm ON fgm.user_id = u.id
		LEFT JOIN finance_groups fg ON fg.id = fgm.group_id `+where+` ORDER BY u.email LIMIT 2000`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]service.FinanceUser, 0)
	for rows.Next() {
		var item service.FinanceUser
		if err := rows.Scan(&item.ID, &item.Email, &item.Username, &item.Balance, &item.Concurrency, &item.GroupID, &item.GroupName); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (r *financeRepository) ListBalances(ctx context.Context, f service.FinanceFilters) ([]service.FinanceUser, *pagination.PaginationResult, error) {
	conditions := []string{"u.deleted_at IS NULL"}
	args := []any{}
	if f.UserID > 0 {
		args = append(args, f.UserID)
		conditions = append(conditions, fmt.Sprintf("u.id = $%d", len(args)))
	}
	if f.GroupID > 0 {
		args = append(args, f.GroupID)
		conditions = append(conditions, fmt.Sprintf("fgm.group_id = $%d", len(args)))
	}
	if search := strings.TrimSpace(f.Search); search != "" {
		args = append(args, "%"+search+"%")
		conditions = append(conditions, fmt.Sprintf("(u.email ILIKE $%d OR COALESCE(u.username, '') ILIKE $%d)", len(args), len(args)))
	}
	base := ` FROM users u LEFT JOIN finance_group_members fgm ON fgm.user_id = u.id
		LEFT JOIN finance_groups fg ON fg.id = fgm.group_id WHERE ` + strings.Join(conditions, " AND ")
	var total int64
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*)`+base, args...).Scan(&total); err != nil {
		return nil, nil, err
	}
	pageArgs := append(append([]any{}, args...), f.PageSize, (f.Page-1)*f.PageSize)
	rows, err := r.db.QueryContext(ctx, `SELECT u.id, u.email, COALESCE(u.username, ''), COALESCE(u.balance, 0), COALESCE(u.concurrency, 0), fgm.group_id, COALESCE(fg.name, '')`+
		base+` ORDER BY u.id DESC LIMIT $`+fmt.Sprint(len(args)+1)+` OFFSET $`+fmt.Sprint(len(args)+2), pageArgs...)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()
	items := make([]service.FinanceUser, 0, f.PageSize)
	for rows.Next() {
		var item service.FinanceUser
		if err := rows.Scan(&item.ID, &item.Email, &item.Username, &item.Balance, &item.Concurrency, &item.GroupID, &item.GroupName); err != nil {
			return nil, nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, nil, err
	}
	return items, paginationResultFromTotal(total, pagination.PaginationParams{Page: f.Page, PageSize: f.PageSize}), nil
}

func (r *financeRepository) ListGroups(ctx context.Context) ([]service.FinanceGroup, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, name, description, sort_order FROM finance_groups ORDER BY sort_order, name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	groups := make([]service.FinanceGroup, 0)
	for rows.Next() {
		var g service.FinanceGroup
		if err := rows.Scan(&g.ID, &g.Name, &g.Description, &g.SortOrder); err != nil {
			return nil, err
		}
		g.Members = []service.FinanceUser{}
		groups = append(groups, g)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	for i := range groups {
		members, err := r.ListGroupMembers(ctx, groups[i].ID)
		if err != nil {
			return nil, err
		}
		groups[i].Members = members
	}
	return groups, nil
}

func (r *financeRepository) ListGroupMembers(ctx context.Context, groupID int64) ([]service.FinanceUser, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT u.id, u.email, COALESCE(u.username, ''), COALESCE(u.balance, 0), COALESCE(u.concurrency, 0), fgm.group_id, fg.name
		FROM finance_group_members fgm JOIN users u ON u.id = fgm.user_id
		JOIN finance_groups fg ON fg.id = fgm.group_id
		WHERE fgm.group_id = $1 AND u.deleted_at IS NULL ORDER BY u.email`, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]service.FinanceUser, 0)
	for rows.Next() {
		var item service.FinanceUser
		if err := rows.Scan(&item.ID, &item.Email, &item.Username, &item.Balance, &item.Concurrency, &item.GroupID, &item.GroupName); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (r *financeRepository) CreateGroup(ctx context.Context, actorID int64, name, description string) (*service.FinanceGroup, error) {
	var g service.FinanceGroup
	err := r.db.QueryRowContext(ctx, `INSERT INTO finance_groups(name, description, created_by) VALUES ($1, $2, $3)
		RETURNING id, name, description, sort_order`, name, description, actorID).
		Scan(&g.ID, &g.Name, &g.Description, &g.SortOrder)
	if err != nil {
		return nil, err
	}
	g.Members = []service.FinanceUser{}
	return &g, nil
}

func (r *financeRepository) UpdateGroup(ctx context.Context, id int64, name, description string) (*service.FinanceGroup, error) {
	var g service.FinanceGroup
	err := r.db.QueryRowContext(ctx, `UPDATE finance_groups SET name=$1, description=$2, updated_at=NOW()
		WHERE id=$3 RETURNING id, name, description, sort_order`, name, description, id).
		Scan(&g.ID, &g.Name, &g.Description, &g.SortOrder)
	if err != nil {
		return nil, err
	}
	g.Members, err = r.ListGroupMembers(ctx, id)
	return &g, err
}

func (r *financeRepository) DeleteGroup(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM finance_groups WHERE id=$1`, id)
	return err
}

func (r *financeRepository) ReplaceGroupMembers(ctx context.Context, groupID int64, userIDs []int64) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var exists bool
	if err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM finance_groups WHERE id=$1)`, groupID).Scan(&exists); err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("finance group not found")
	}
	if len(userIDs) == 0 {
		if _, err := tx.ExecContext(ctx, `DELETE FROM finance_group_members WHERE group_id=$1`, groupID); err != nil {
			return err
		}
	} else {
		if _, err := tx.ExecContext(ctx, `DELETE FROM finance_group_members WHERE group_id=$1 OR user_id = ANY($2)`, groupID, pqInt64Array(userIDs)); err != nil {
			return err
		}
	}
	for _, userID := range userIDs {
		if userID <= 0 {
			return fmt.Errorf("invalid user id")
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO finance_group_members(group_id, user_id) VALUES ($1, $2)`, groupID, userID); err != nil {
			return err
		}
	}
	return tx.Commit()
}

// pqInt64Array avoids adding a second SQL array dependency; PostgreSQL accepts
// the standard text array representation for ANY($1).
type pqInt64Array []int64

func (a pqInt64Array) Value() (driver.Value, error) {
	parts := make([]string, len(a))
	for i, v := range a {
		parts[i] = fmt.Sprintf("%d", v)
	}
	return "{" + strings.Join(parts, ",") + "}", nil
}

func (r *financeRepository) ListRecharges(ctx context.Context, f service.FinanceFilters) ([]service.FinanceRecharge, *pagination.PaginationResult, float64, error) {
	where, args := financeWhere("recharge.created_at", f)
	where = appendCondition(where, args, "recharge.value > 0")
	return r.queryRecharges(ctx, where, args, f)
}

func (r *financeRepository) queryRecharges(ctx context.Context, where string, args []any, f service.FinanceFilters) ([]service.FinanceRecharge, *pagination.PaginationResult, float64, error) {
	base := ` FROM (
		SELECT rc.id, rc.used_by AS user_id, rc.type, rc.value, rc.status, COALESCE(rc.notes, '') AS notes,
			rc.used_at, COALESCE(rc.used_at, rc.created_at) AS created_at
		FROM redeem_codes rc
		WHERE rc.used_by IS NOT NULL
			AND rc.type IN ('balance', 'admin_balance')
		UNION ALL
		SELECT po.id, po.user_id, 'payment' AS type, po.amount, po.status,
			COALESCE(NULLIF(po.payment_type, ''), 'online payment') AS notes,
			po.paid_at, COALESCE(po.paid_at, po.completed_at, po.created_at) AS created_at
		FROM payment_orders po
		WHERE po.order_type = 'balance'
			AND po.status IN ('PAID', 'RECHARGING', 'COMPLETED')
			AND po.paid_at IS NOT NULL
		UNION ALL
		SELECT xo.id, xo.user_id, 'xianyu_recharge' AS type, xo.credited_amount_usd, xo.status,
			COALESCE(NULLIF(xo.note, ''), COALESCE(xo.admin_note, '')) AS notes,
			xo.reviewed_at, COALESCE(xo.reviewed_at, xo.updated_at) AS created_at
		FROM xianyu_recharge_orders xo
		WHERE xo.status = 'credited'
			AND xo.credited_amount_usd IS NOT NULL
			AND xo.credited_amount_usd > 0
	) recharge
	JOIN users u ON u.id=recharge.user_id
		LEFT JOIN finance_group_members fgm ON fgm.user_id=u.id
		LEFT JOIN finance_groups fg ON fg.id=fgm.group_id ` + where
	var total int64
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*)`+base, args...).Scan(&total); err != nil {
		return nil, nil, 0, err
	}
	var sum float64
	if err := r.db.QueryRowContext(ctx, `SELECT COALESCE(SUM(recharge.value),0)`+base, args...).Scan(&sum); err != nil {
		return nil, nil, 0, err
	}
	argsPage := append(append([]any{}, args...), f.PageSize, (f.Page-1)*f.PageSize)
	rows, err := r.db.QueryContext(ctx, `SELECT recharge.id, u.id, u.email, COALESCE(u.username,''), COALESCE(fg.name,''), recharge.type, recharge.value, recharge.status, recharge.notes, recharge.used_at, recharge.created_at`+base+` ORDER BY recharge.created_at DESC, recharge.id DESC LIMIT $`+fmt.Sprint(len(args)+1)+` OFFSET $`+fmt.Sprint(len(args)+2), argsPage...)
	if err != nil {
		return nil, nil, 0, err
	}
	defer rows.Close()
	items := make([]service.FinanceRecharge, 0)
	for rows.Next() {
		var item service.FinanceRecharge
		if err := rows.Scan(&item.ID, &item.UserID, &item.Email, &item.Username, &item.GroupName, &item.Type, &item.Value, &item.Status, &item.Notes, &item.UsedAt, &item.CreatedAt); err != nil {
			return nil, nil, 0, err
		}
		items = append(items, item)
	}
	return items, paginationResultFromTotal(total, pagination.PaginationParams{Page: f.Page, PageSize: f.PageSize}), sum, rows.Err()
}

func (r *financeRepository) ListUsage(ctx context.Context, f service.FinanceFilters) ([]service.FinanceUsage, *pagination.PaginationResult, float64, error) {
	where, args := financeWhere("ul.created_at", f)
	base := ` FROM usage_logs ul JOIN users u ON u.id=ul.user_id
		LEFT JOIN finance_group_members fgm ON fgm.user_id=u.id
		LEFT JOIN finance_groups fg ON fg.id=fgm.group_id ` + where
	var total int64
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*)`+base, args...).Scan(&total); err != nil {
		return nil, nil, 0, err
	}
	var sum float64
	if err := r.db.QueryRowContext(ctx, `SELECT COALESCE(SUM(ul.actual_cost),0)`+base, args...).Scan(&sum); err != nil {
		return nil, nil, 0, err
	}
	argsPage := append(append([]any{}, args...), f.PageSize, (f.Page-1)*f.PageSize)
	rows, err := r.db.QueryContext(ctx, `SELECT ul.id, u.id, u.email, COALESCE(u.username,''), COALESCE(fg.name,''), COALESCE(NULLIF(ul.requested_model,''), ul.model), COALESCE(ul.request_id,''), ul.input_tokens, ul.output_tokens, ul.actual_cost, ul.created_at`+base+` ORDER BY ul.created_at DESC, ul.id DESC LIMIT $`+fmt.Sprint(len(args)+1)+` OFFSET $`+fmt.Sprint(len(args)+2), argsPage...)
	if err != nil {
		return nil, nil, 0, err
	}
	defer rows.Close()
	items := make([]service.FinanceUsage, 0)
	for rows.Next() {
		var item service.FinanceUsage
		if err := rows.Scan(&item.ID, &item.UserID, &item.Email, &item.Username, &item.GroupName, &item.Model, &item.RequestID, &item.InputTokens, &item.OutputTokens, &item.ActualCost, &item.CreatedAt); err != nil {
			return nil, nil, 0, err
		}
		item.TotalTokens = item.InputTokens + item.OutputTokens
		items = append(items, item)
	}
	return items, paginationResultFromTotal(total, pagination.PaginationParams{Page: f.Page, PageSize: f.PageSize}), sum, rows.Err()
}

func financeWhere(dateColumn string, f service.FinanceFilters) (string, []any) {
	conditions := []string{"u.deleted_at IS NULL"}
	args := []any{}
	if f.StartTime != nil {
		args = append(args, *f.StartTime)
		conditions = append(conditions, fmt.Sprintf("%s >= $%d", dateColumn, len(args)))
	}
	if f.EndTime != nil {
		args = append(args, *f.EndTime)
		conditions = append(conditions, fmt.Sprintf("%s < $%d", dateColumn, len(args)))
	}
	if f.UserID > 0 {
		args = append(args, f.UserID)
		conditions = append(conditions, fmt.Sprintf("u.id = $%d", len(args)))
	}
	if f.GroupID > 0 {
		args = append(args, f.GroupID)
		conditions = append(conditions, fmt.Sprintf("fgm.group_id = $%d", len(args)))
	}
	if search := strings.TrimSpace(f.Search); search != "" {
		args = append(args, "%"+search+"%")
		conditions = append(conditions, fmt.Sprintf("(u.email ILIKE $%d OR COALESCE(u.username,'') ILIKE $%d)", len(args), len(args)))
	}
	return " WHERE " + strings.Join(conditions, " AND "), args
}

func appendCondition(where string, args []any, condition string) string {
	_ = args
	if strings.TrimSpace(where) == "" {
		return " WHERE " + condition
	}
	return where + " AND " + condition
}
