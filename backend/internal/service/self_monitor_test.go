//go:build unit

package service

import (
	"context"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestSelfMonitorManualCheckIndependentOfScheduledSwitch(t *testing.T) {
	repo := &quotaModeRepoStub{monitor: &ChannelMonitor{ID: 1, PrimaryModel: "quota", CheckMode: MonitorCheckModeQuota}}
	svc := newQuotaModeService(repo)
	svc.SetRuntimeReader(channelMonitorRuntimeStub{rt: ChannelMonitorRuntime{Enabled: false, Mode: ChannelMonitorModeV2}})
	_, err := svc.RunCheck(context.Background(), 1)
	require.Error(t, err)
	results, err := svc.RunSelfCheck(context.Background(), 1)
	require.NoError(t, err)
	require.Len(t, results, 1)
	require.Len(t, repo.history, 1)
}
