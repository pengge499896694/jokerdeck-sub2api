package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

const (
	clientNodePolicyKey       = "client_node_policy"
	clientNodeSubscriptionKey = "client_node_subscription"
	legacyClientNodePolicyKey = "client_node_global_policy"
	MaxClientNodeContentSize  = 1 << 20
)

type legacyClientNodePolicy struct {
	Enabled                bool     `json:"enabled"`
	DisabledNodes          []string `json:"disabled_nodes"`
	SubscriptionURL        string   `json:"subscription_url"`
	BackupSubscriptionURLs []string `json:"backup_subscription_urls"`
}

type ClientNodePolicy struct {
	Enabled       bool     `json:"enabled"`
	DisabledNodes []string `json:"disabled_nodes"`
}

func (s *SettingService) GetClientNodePolicy(ctx context.Context) (*ClientNodePolicy, error) {
	raw, err := s.settingRepo.GetValue(ctx, clientNodePolicyKey)
	if errors.Is(err, ErrSettingNotFound) {
		legacy, legacyErr := s.getLegacyClientNodePolicy(ctx)
		if legacyErr != nil {
			return nil, legacyErr
		}
		if legacy == nil {
			return &ClientNodePolicy{DisabledNodes: []string{}}, nil
		}
		if legacy.DisabledNodes == nil {
			legacy.DisabledNodes = []string{}
		}
		return &ClientNodePolicy{Enabled: legacy.Enabled, DisabledNodes: legacy.DisabledNodes}, nil
	}
	if err != nil {
		return nil, err
	}
	var policy ClientNodePolicy
	if err := json.Unmarshal([]byte(raw), &policy); err != nil {
		return nil, fmt.Errorf("invalid stored client node policy: %w", err)
	}
	if policy.DisabledNodes == nil {
		policy.DisabledNodes = []string{}
	}
	return &policy, nil
}

func (s *SettingService) SetClientNodePolicy(ctx context.Context, policy ClientNodePolicy) (*ClientNodePolicy, error) {
	if policy.DisabledNodes == nil {
		policy.DisabledNodes = []string{}
	}
	data, err := json.Marshal(policy)
	if err != nil {
		return nil, err
	}
	if len(data) > MaxClientNodeContentSize {
		return nil, fmt.Errorf("client node policy exceeds 1 MiB")
	}
	if err := s.settingRepo.Set(ctx, clientNodePolicyKey, string(data)); err != nil {
		return nil, err
	}
	return &policy, nil
}

func (s *SettingService) GetClientNodeSubscription(ctx context.Context) (string, error) {
	raw, err := s.settingRepo.GetValue(ctx, clientNodeSubscriptionKey)
	if errors.Is(err, ErrSettingNotFound) {
		legacy, legacyErr := s.getLegacyClientNodePolicy(ctx)
		if legacyErr != nil || legacy == nil {
			return "", legacyErr
		}
		if strings.TrimSpace(legacy.SubscriptionURL) == "" && len(legacy.BackupSubscriptionURLs) == 0 {
			return "", nil
		}
		for _, address := range append([]string{legacy.SubscriptionURL}, legacy.BackupSubscriptionURLs...) {
			content, fetchErr := fetchLegacyClientNodeSubscription(ctx, address)
			if fetchErr == nil {
				return content, nil
			}
		}
		return "", fmt.Errorf("client node subscription is unavailable")
	}
	return raw, err
}

func (s *SettingService) getLegacyClientNodePolicy(ctx context.Context) (*legacyClientNodePolicy, error) {
	raw, err := s.settingRepo.GetValue(ctx, legacyClientNodePolicyKey)
	if errors.Is(err, ErrSettingNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var policy legacyClientNodePolicy
	if err := json.Unmarshal([]byte(raw), &policy); err != nil {
		return nil, fmt.Errorf("invalid stored legacy client node policy: %w", err)
	}
	return &policy, nil
}

func fetchLegacyClientNodeSubscription(ctx context.Context, address string) (string, error) {
	parsed, err := url.Parse(strings.TrimSpace(address))
	if err != nil || parsed == nil || (parsed.Scheme != "https" && parsed.Scheme != "http") || parsed.Host == "" || parsed.User != nil {
		return "", fmt.Errorf("invalid legacy subscription address")
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, parsed.String(), nil)
	if err != nil {
		return "", err
	}
	client := &http.Client{Timeout: 12 * time.Second, CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
		return http.ErrUseLastResponse
	}}
	response, err := client.Do(request)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf("subscription returned status %d", response.StatusCode)
	}
	content, err := io.ReadAll(io.LimitReader(response.Body, MaxClientNodeContentSize+1))
	if err != nil || len(content) > MaxClientNodeContentSize || len(strings.TrimSpace(string(content))) == 0 {
		return "", fmt.Errorf("invalid legacy subscription response")
	}
	return string(content), nil
}

func (s *SettingService) SetClientNodeSubscription(ctx context.Context, content string) error {
	if err := validateClientNodeSubscription(content); err != nil {
		return err
	}
	return s.settingRepo.Set(ctx, clientNodeSubscriptionKey, content)
}

func (s *SettingService) SetClientNodes(ctx context.Context, policy ClientNodePolicy, content string) error {
	if err := validateClientNodeSubscription(content); err != nil {
		return err
	}
	if policy.DisabledNodes == nil {
		policy.DisabledNodes = []string{}
	}
	data, err := json.Marshal(policy)
	if err != nil {
		return err
	}
	if len(data) > MaxClientNodeContentSize {
		return fmt.Errorf("client node policy exceeds 1 MiB")
	}
	return s.settingRepo.SetMultiple(ctx, map[string]string{
		clientNodePolicyKey: string(data), clientNodeSubscriptionKey: content,
	})
}

func validateClientNodeSubscription(content string) error {
	if len(content) > MaxClientNodeContentSize {
		return fmt.Errorf("client node subscription exceeds 1 MiB")
	}
	if strings.TrimSpace(content) == "" {
		return fmt.Errorf("client node subscription must not be empty")
	}
	var node yaml.Node
	decoder := yaml.NewDecoder(strings.NewReader(content))
	if err := decoder.Decode(&node); err != nil {
		return fmt.Errorf("invalid YAML: %w", err)
	}
	var extra yaml.Node
	if err := decoder.Decode(&extra); err == nil {
		return fmt.Errorf("multiple YAML documents are not allowed")
	} else if !errors.Is(err, io.EOF) {
		return fmt.Errorf("invalid YAML: %w", err)
	}
	if err := rejectUnsafeYAML(&node); err != nil {
		return err
	}
	var decoded any
	if err := node.Decode(&decoded); err != nil {
		return fmt.Errorf("invalid YAML: %w", err)
	}
	return nil
}

func rejectUnsafeYAML(node *yaml.Node) error {
	if node == nil {
		return nil
	}
	if node.Kind == yaml.AliasNode || node.Tag != "" && node.Tag != "!!map" && node.Tag != "!!seq" && node.Tag != "!!str" && node.Tag != "!!int" && node.Tag != "!!float" && node.Tag != "!!bool" && node.Tag != "!!null" && node.Tag != "!!timestamp" {
		return fmt.Errorf("unsafe YAML tags or aliases are not allowed")
	}
	for _, child := range node.Content {
		if err := rejectUnsafeYAML(child); err != nil {
			return err
		}
	}
	return nil
}
