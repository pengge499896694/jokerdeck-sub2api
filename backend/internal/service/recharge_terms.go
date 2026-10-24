package service

import (
	"context"
	"fmt"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

const RechargeTermsVersion = "2026-10-04"

type RechargeTermsStatus struct {
	Version  string `json:"version"`
	Accepted bool   `json:"accepted"`
}

func RechargeTermsAccepted(ctx context.Context, db *dbent.Client, userID int64) (bool, error) {
	rows, err := db.QueryContext(ctx, `SELECT EXISTS (
		SELECT 1 FROM recharge_terms_acceptances WHERE user_id = $1 AND version = $2
	)`, userID, RechargeTermsVersion)
	if err != nil {
		return false, fmt.Errorf("read recharge terms acceptance: %w", err)
	}
	defer func() { _ = rows.Close() }()
	if !rows.Next() {
		return false, fmt.Errorf("read recharge terms acceptance: no row returned")
	}
	var accepted bool
	if err := rows.Scan(&accepted); err != nil {
		return false, fmt.Errorf("scan recharge terms acceptance: %w", err)
	}
	return accepted, rows.Err()
}

func RequireRechargeTerms(ctx context.Context, db *dbent.Client, userID int64) error {
	accepted, err := RechargeTermsAccepted(ctx, db, userID)
	if err != nil {
		return err
	}
	if !accepted {
		return infraerrors.Forbidden("RECHARGE_TERMS_REQUIRED", "请先阅读并确认当前充值服务条款")
	}
	return nil
}

func AcceptRechargeTerms(ctx context.Context, db *dbent.Client, userID int64, notMainlandSubject, outsideMainland, noInvoice bool) error {
	if userID <= 0 || !notMainlandSubject || !outsideMainland || !noInvoice {
		return infraerrors.BadRequest("RECHARGE_TERMS_INCOMPLETE", "请如实逐项确认全部充值服务条款")
	}
	rows, err := db.QueryContext(ctx, `
		INSERT INTO recharge_terms_acceptances
		    (user_id, version, not_mainland_subject, outside_mainland, no_invoice)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (user_id, version) DO NOTHING
		RETURNING user_id
	`, userID, RechargeTermsVersion, notMainlandSubject, outsideMainland, noInvoice)
	if err != nil {
		return fmt.Errorf("save recharge terms acceptance: %w", err)
	}
	defer func() { _ = rows.Close() }()
	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return fmt.Errorf("save recharge terms acceptance: %w", err)
		}
		return nil // The same account already accepted this version.
	}
	var savedUserID int64
	if err := rows.Scan(&savedUserID); err != nil {
		return fmt.Errorf("scan recharge terms acceptance: %w", err)
	}
	return rows.Err()
}
