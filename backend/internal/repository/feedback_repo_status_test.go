package repository

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestFeedbackSetStatus(t *testing.T) {
	for _, status := range []string{"pending_user", "resolved", "closed"} {
		t.Run(status, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()
			mock.ExpectExec(`UPDATE feedback_tickets\s+SET status=\$2::text,\s+closed_at=CASE WHEN \$2::text`).
				WithArgs(int64(42), status).
				WillReturnResult(sqlmock.NewResult(0, 1))

			require.NoError(t, NewFeedbackRepository(db).SetStatus(context.Background(), 42, status))
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestFeedbackSetStatusMissingTicket(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	mock.ExpectExec(`UPDATE feedback_tickets`).WithArgs(int64(42), "closed").
		WillReturnResult(sqlmock.NewResult(0, 0))

	require.ErrorIs(t, NewFeedbackRepository(db).SetStatus(context.Background(), 42, "closed"), service.ErrFeedbackNotFound)
	require.NoError(t, mock.ExpectationsWereMet())
}
