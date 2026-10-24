package service

import (
	"context"
	"time"
)

// GroupHealth intentionally exposes no account identity, traffic volume or errors.
type GroupHealth struct {
	GroupID     int64     `json:"group_id"`
	Platform    string    `json:"platform"`
	Model       string    `json:"model,omitempty"`
	Status      string    `json:"status"`
	Reason      string    `json:"reason,omitempty"`
	DataThrough time.Time `json:"data_through"`
}

// GroupHealth uses the existing minute rollups, shared across desktop clients.
// Missing, disabled or stale monitoring never counts as evidence of an outage.
func (s *ChannelMonitorV2Service) GroupHealth(ctx context.Context) ([]GroupHealth, error) {
	s.groupHealthMu.Lock()
	defer s.groupHealthMu.Unlock()
	now := s.now().UTC()
	if now.Sub(s.groupHealthAt) < 30*time.Second {
		return append([]GroupHealth(nil), s.groupHealthRows...), nil
	}
	cfg, err := s.getEnabledConfig(ctx)
	if err != nil {
		return nil, err
	}
	end := now.Truncate(time.Minute)
	filter := ChannelMonitorV2Filter{Start: end.Add(-10 * time.Minute), End: end, Bucket: time.Minute}
	matrix, err := s.repo.GetMatrix(ctx, filter, *cfg, ChannelMonitorV2GroupByPlatformGroupModel, true)
	if err != nil {
		return nil, err
	}
	rows := []GroupHealth{}
	if matrix != nil {
		fresh := !matrix.Coverage.DataThrough.IsZero() && now.Sub(matrix.Coverage.DataThrough) <= 3*time.Minute
		for _, row := range matrix.Items {
			if row.GroupID == nil || *row.GroupID <= 0 {
				continue
			}
			status, reason := classifyGroupHealth(row.Metrics, cfg.HealthThresholds, fresh)
			rows = append(rows, GroupHealth{GroupID: *row.GroupID, Platform: row.Platform, Model: row.Model, Status: status, Reason: reason, DataThrough: matrix.Coverage.DataThrough})
		}
	}
	s.groupHealthRows, s.groupHealthAt = rows, now
	return append([]GroupHealth(nil), rows...), nil
}

func classifyGroupHealth(m ChannelMonitorV2Metric, thresholds ChannelMonitorV2HealthThresholds, fresh bool) (string, string) {
	minimum := thresholds.MinimumSample
	if minimum < 20 {
		minimum = 20
	}
	if !fresh || m.RequestCount < minimum {
		return "unknown", ""
	}
	// Cache hit rate affects cost, but is not evidence of upstream failure.
	if m.SuccessRequests == 0 && m.ErrorRate >= 0.9 {
		return "unavailable", "upstream_errors"
	}
	warningRate := thresholds.WarningErrorRate
	if warningRate <= 0 {
		warningRate = 0.05
	}
	if m.ErrorRate >= warningRate {
		return "degraded", "upstream_errors"
	}
	warningTTFT := thresholds.WarningTTFTMs
	if warningTTFT <= 0 {
		warningTTFT = 10000
	}
	if m.TTFT.SampleCount >= minimum && m.TTFT.P95Ms != nil && *m.TTFT.P95Ms >= warningTTFT {
		return "degraded", "slow_first_token"
	}
	return "healthy", ""
}
