package service

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestGroupHealthBandsUseFractionalErrorRates(t *testing.T) {
	for _, tc := range []struct {
		name           string
		count, success int64
		rate           float64
		fresh          bool
		want           string
	}{
		{"healthy", 100, 100, 0, true, "healthy"},
		{"error warning", 100, 90, 0.1, true, "degraded"},
		{"all failed", 100, 0, 1, true, "unavailable"},
		{"low sample", 2, 0, 1, true, "unknown"},
		{"stale", 100, 0, 1, false, "unknown"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			status, _ := classifyGroupHealth(ChannelMonitorV2Metric{RequestCount: tc.count, SuccessRequests: tc.success, ErrorRate: tc.rate}, ChannelMonitorV2HealthThresholds{}, tc.fresh)
			require.Equal(t, tc.want, status)
		})
	}
}
