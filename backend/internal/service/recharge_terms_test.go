package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAcceptRechargeTermsRequiresAllDeclarations(t *testing.T) {
	for _, tc := range []struct {
		name        string
		userID      int64
		notMainland bool
		outside     bool
		noInvoice   bool
	}{
		{"anonymous", 0, true, true, true},
		{"mainland_subject", 1, false, true, true},
		{"inside_mainland", 1, true, false, true},
		{"invoice_not_accepted", 1, true, true, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := AcceptRechargeTerms(context.Background(), nil, tc.userID, tc.notMainland, tc.outside, tc.noInvoice)
			require.ErrorContains(t, err, "请如实逐项确认")
		})
	}
}
