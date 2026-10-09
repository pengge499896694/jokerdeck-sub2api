//go:build unit

package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

type clientProviderRepo struct {
	SettingRepository
	raw string
	err error
}

func (r *clientProviderRepo) GetValue(context.Context, string) (string, error) { return r.raw, r.err }
func (r *clientProviderRepo) Set(_ context.Context, _ string, value string) error {
	if r.err != nil {
		return r.err
	}
	r.raw = value
	return nil
}

func TestClientProviderPolicyDefaultsAndFailure(t *testing.T) {
	for _, raw := range []string{"", "false", "garbage", "TRUE"} {
		svc := &SettingService{settingRepo: &clientProviderRepo{raw: raw}}
		policy, err := svc.GetClientProviderPolicy(context.Background())
		require.NoError(t, err)
		require.False(t, policy.AllowProviderSwitch)
		require.True(t, policy.RequireRecharge)
		require.False(t, policy.ExternalProviderSupport)
	}
	svc := &SettingService{settingRepo: &clientProviderRepo{err: ErrSettingNotFound}}
	policy, err := svc.GetClientProviderPolicy(context.Background())
	require.NoError(t, err)
	require.False(t, policy.AllowProviderSwitch)
	svc.settingRepo = &clientProviderRepo{err: errors.New("db unavailable")}
	policy, err = svc.GetClientProviderPolicy(context.Background())
	require.Error(t, err)
	require.Nil(t, policy)
}

func TestClientProviderPolicyAdminUpdates(t *testing.T) {
	svc := &SettingService{settingRepo: &clientProviderRepo{}}
	for _, allowed := range []bool{true, false} {
		policy, err := svc.SetClientProviderPolicy(context.Background(), allowed)
		require.NoError(t, err)
		require.Equal(t, allowed, policy.AllowProviderSwitch)
		stored, err := svc.GetClientProviderPolicy(context.Background())
		require.NoError(t, err)
		require.Equal(t, policy, stored)
	}
}
