package repository

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

func TestFinanceListUsageFiltersByAPIKeyAndReturnsKeyName(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()

	countQuery := regexp.QuoteMeta("SELECT COUNT(*)") + `(?s).*ul\.api_key_id = \$1`
	mock.ExpectQuery(countQuery).WithArgs(int64(17)).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT COALESCE(SUM(ul.actual_cost),0)") + `(?s).*ul\.api_key_id = \$1`).
		WithArgs(int64(17)).WillReturnRows(sqlmock.NewRows([]string{"sum"}).AddRow(2.75))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT ul.id, u.id, ul.api_key_id")+`(?s).*LEFT JOIN api_keys.*ul\.api_key_id = \$1.*LIMIT \$2 OFFSET \$3`).
		WithArgs(int64(17), 20, 0).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "user_id", "api_key_id", "api_key_name", "email", "username", "group_name", "model", "request_id",
			"input_tokens", "output_tokens", "actual_cost", "created_at",
		}).AddRow(31, 9, 17, "production", "user@example.com", "user", "finance", "gpt-test", "req-1", 4, 6, 2.75, time.Date(2026, 10, 12, 0, 0, 0, 0, time.UTC)))

	repo := NewFinanceRepository(db)
	items, page, totalAmount, err := repo.ListUsage(context.Background(), service.FinanceFilters{
		APIKeyID: 17,
		Page:     1,
		PageSize: 20,
	})
	if err != nil {
		t.Fatal(err)
	}
	if page.Total != 1 || totalAmount != 2.75 {
		t.Fatalf("summary = total %d, amount %v; want total 1, amount 2.75", page.Total, totalAmount)
	}
	if len(items) != 1 || items[0].APIKeyID != 17 || items[0].APIKeyName != "production" || items[0].TotalTokens != 10 {
		t.Fatalf("usage row = %+v", items)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
