package service

import (
	"context"
	"strconv"
)

// Self-monitor switches are independent of the scheduled monitor configuration.
func (s *SettingService) SelfMonitorEnabled(ctx context.Context, id int64) (bool, error) {
	key := "self_monitor_enabled"
	if id > 0 {
		key = "self_monitor_channel_" + strconv.FormatInt(id, 10)
	}
	values, err := s.settingRepo.GetMultiple(ctx, []string{key})
	if err != nil {
		return false, err
	}
	return values[key] != "false", nil
}

func (s *SettingService) SetSelfMonitorEnabled(ctx context.Context, id int64, enabled bool) error {
	key := "self_monitor_enabled"
	if id > 0 {
		key = "self_monitor_channel_" + strconv.FormatInt(id, 10)
	}
	return s.settingRepo.Set(ctx, key, strconv.FormatBool(enabled))
}

func (s *ChannelMonitorService) ListSelfMonitorViews(ctx context.Context) ([]*UserMonitorView, error) {
	monitors, _, err := s.repo.List(ctx, ChannelMonitorListParams{Page: 1, PageSize: 1000})
	if err != nil {
		return nil, err
	}
	ids, primary, extras := collectMonitorIndexes(monitors)
	summaries := s.BatchMonitorStatusSummary(ctx, ids, primary, extras)
	latest := s.batchLatest(ctx, ids)
	timeline := s.batchTimeline(ctx, ids, primary)
	views := make([]*UserMonitorView, 0, len(monitors))
	for _, m := range monitors {
		views = append(views, buildUserViewFromSummary(m, summaries[m.ID], pickLatest(latest[m.ID], m.PrimaryModel), timeline[m.ID]))
	}
	return views, nil
}
