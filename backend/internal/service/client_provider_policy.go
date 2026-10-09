package service

import (
	"context"
	"errors"
	"strconv"
)

const clientProviderSwitchKey = "client_provider_switch_allowed"

type ClientProviderPolicy struct {
	AllowProviderSwitch     bool `json:"allow_provider_switch"`
	RequireRecharge         bool `json:"require_recharge"`
	ExternalProviderSupport bool `json:"external_provider_support"`
}

// Missing settings default to locked. Database errors must not grant access.
func (s *SettingService) GetClientProviderPolicy(ctx context.Context) (*ClientProviderPolicy, error) {
	raw, err := s.settingRepo.GetValue(ctx, clientProviderSwitchKey)
	if err != nil && !errors.Is(err, ErrSettingNotFound) {
		return nil, err
	}
	return &ClientProviderPolicy{AllowProviderSwitch: raw == "true", RequireRecharge: true, ExternalProviderSupport: false}, nil
}

func (s *SettingService) SetClientProviderPolicy(ctx context.Context, allowed bool) (*ClientProviderPolicy, error) {
	if err := s.settingRepo.Set(ctx, clientProviderSwitchKey, strconv.FormatBool(allowed)); err != nil {
		return nil, err
	}
	return &ClientProviderPolicy{AllowProviderSwitch: allowed, RequireRecharge: true, ExternalProviderSupport: false}, nil
}
