//go:build unit

package service

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCommerceLinksDefaultsAndStorage(t *testing.T) {
	ctx := context.Background()
	for _, repo := range []*clientProviderRepo{{}, {err: ErrSettingNotFound}} {
		result, err := (&SettingService{settingRepo: repo}).GetCommerceLinks(ctx)
		require.NoError(t, err)
		require.Equal(t, defaultCommerceLinks(), *result)
	}
	repo := &clientProviderRepo{}
	svc := &SettingService{settingRepo: repo}
	input := CommerceLinks{" https://m.tb.cn/new?tk=a&x=2 ", "https://store.example/shop", "https://backup.example/shop"}
	saved, err := svc.SetCommerceLinks(ctx, input)
	require.NoError(t, err)
	require.Equal(t, strings.TrimSpace(input.XianyuOrderURL), saved.XianyuOrderURL)
	read, err := svc.GetCommerceLinks(ctx)
	require.NoError(t, err)
	require.Equal(t, saved, read)
	saved.StoreBackupURL = ""
	_, err = svc.SetCommerceLinks(ctx, *saved)
	require.NoError(t, err)
	read, err = svc.GetCommerceLinks(ctx)
	require.NoError(t, err)
	require.Empty(t, read.StoreBackupURL)
}

func TestCommerceLinksRejectInvalidWithoutPartialWrite(t *testing.T) {
	ctx := context.Background()
	repo := &clientProviderRepo{}
	svc := &SettingService{settingRepo: repo}
	_, err := svc.SetCommerceLinks(ctx, defaultCommerceLinks())
	require.NoError(t, err)
	original := repo.raw
	for _, invalid := range []string{"", "javascript:alert(1)", "data:text/html,hello", "//example.com", "https://user:pass@example.com", "https://example.com/a b", `https://example.com\@evil.example`, "https://", "https://example.com/" + strings.Repeat("x", 2048)} {
		t.Run(invalid[:min(len(invalid), 60)], func(t *testing.T) {
			input := defaultCommerceLinks()
			input.XianyuOrderURL = invalid
			_, err := svc.SetCommerceLinks(ctx, input)
			require.Error(t, err)
			require.Equal(t, original, repo.raw)
		})
	}
	input := defaultCommerceLinks()
	input.StoreBackupURL = "javascript:alert(1)"
	_, err = svc.SetCommerceLinks(ctx, input)
	require.Error(t, err)
	require.Equal(t, original, repo.raw)
}

func TestCommerceLinksStorageErrors(t *testing.T) {
	ctx := context.Background()
	for _, raw := range []string{"{broken", `{"store_url":"javascript:alert(1)"}`} {
		result, err := (&SettingService{settingRepo: &clientProviderRepo{raw: raw}}).GetCommerceLinks(ctx)
		require.Error(t, err)
		require.Nil(t, result)
	}
	dbErr := errors.New("database unavailable")
	svc := &SettingService{settingRepo: &clientProviderRepo{err: dbErr}}
	_, err := svc.GetCommerceLinks(ctx)
	require.ErrorIs(t, err, dbErr)
	_, err = svc.SetCommerceLinks(ctx, defaultCommerceLinks())
	require.ErrorIs(t, err, dbErr)
}
