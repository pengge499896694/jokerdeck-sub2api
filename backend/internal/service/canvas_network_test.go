//go:build unit

package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCanvasNetworkSettingsDefaultsAndRoundTrip(t *testing.T) {
	ctx := context.Background()
	repo := &clientProviderRepo{}
	svc := &SettingService{settingRepo: repo}
	settings, err := svc.GetCanvasNetworkSettings(ctx)
	require.NoError(t, err)
	require.False(t, settings.Enabled)
	require.False(t, settings.VisibleToUsers)
	require.Equal(t, defaultCanvasNetworkHosts, settings.AllowedHosts)

	proxyID := int64(42)
	input := CanvasNetworkSettings{Enabled: true, VisibleToUsers: false, ProxyID: &proxyID, AllowedHosts: []string{" raw.githubusercontent.com ", "github.com", "github.com"}}
	saved, err := svc.SetCanvasNetworkSettings(ctx, input)
	require.NoError(t, err)
	require.Equal(t, []string{"raw.githubusercontent.com", "github.com"}, saved.AllowedHosts)
	read, err := svc.GetCanvasNetworkSettings(ctx)
	require.NoError(t, err)
	require.Equal(t, saved, read)
}

func TestCanvasNetworkSettingsRejectUnsafeOrIncompleteValuesWithoutWrite(t *testing.T) {
	ctx := context.Background()
	repo := &clientProviderRepo{}
	svc := &SettingService{settingRepo: repo}
	_, err := svc.SetCanvasNetworkSettings(ctx, defaultCanvasNetworkSettings())
	require.NoError(t, err)
	original := repo.raw

	_, err = svc.SetCanvasNetworkSettings(ctx, CanvasNetworkSettings{Enabled: true, AllowedHosts: defaultCanvasNetworkHosts})
	require.Error(t, err)
	require.Equal(t, original, repo.raw)

	proxyID := int64(7)
	for _, hosts := range [][]string{{"https://github.com"}, {"github.com/path"}, {"github.com:443"}, {"github.com evil"}, {}} {
		_, err = svc.SetCanvasNetworkSettings(ctx, CanvasNetworkSettings{ProxyID: &proxyID, Enabled: true, AllowedHosts: hosts})
		require.Error(t, err)
		require.Equal(t, original, repo.raw)
	}
}
