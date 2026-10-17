package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	clientNodePolicyKey       = "client_node_policy"
	clientNodeSubscriptionKey = "client_node_subscription"
	MaxClientNodeContentSize  = 1 << 20
)

type ClientNodePolicy struct {
	Enabled       bool     `json:"enabled"`
	DisabledNodes []string `json:"disabled_nodes"`
}

func (s *SettingService) GetClientNodePolicy(ctx context.Context) (*ClientNodePolicy, error) {
	raw, err := s.settingRepo.GetValue(ctx, clientNodePolicyKey)
	if errors.Is(err, ErrSettingNotFound) {
		return &ClientNodePolicy{DisabledNodes: []string{}}, nil
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
		return "", nil
	}
	return raw, err
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
