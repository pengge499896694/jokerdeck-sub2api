package service

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

type clientNodeRepo struct {
	SettingRepository
	values  map[string]string
	writes  int
	bundles int
	err     error
}

func (r *clientNodeRepo) GetValue(_ context.Context, key string) (string, error) {
	if r.err != nil {
		return "", r.err
	}
	value, ok := r.values[key]
	if !ok {
		return "", ErrSettingNotFound
	}
	return value, nil
}
func (r *clientNodeRepo) Set(_ context.Context, key, value string) error {
	r.writes++
	r.values[key] = value
	return nil
}
func (r *clientNodeRepo) SetMultiple(_ context.Context, values map[string]string) error {
	r.bundles++
	for key, value := range values {
		r.values[key] = value
	}
	return nil
}

func TestClientNodeDefaultsAndPersistence(t *testing.T) {
	repo := &clientNodeRepo{values: map[string]string{}}
	svc := &SettingService{settingRepo: repo}
	policy, err := svc.GetClientNodePolicy(context.Background())
	require.NoError(t, err)
	require.Equal(t, &ClientNodePolicy{DisabledNodes: []string{}}, policy)
	content, err := svc.GetClientNodeSubscription(context.Background())
	require.NoError(t, err)
	require.Empty(t, content)
	err = svc.SetClientNodes(context.Background(), ClientNodePolicy{Enabled: true, DisabledNodes: []string{"a"}}, "proxies: []\n")
	require.NoError(t, err)
	require.Equal(t, 1, repo.bundles)
	require.Zero(t, repo.writes)
	policy, err = svc.GetClientNodePolicy(context.Background())
	require.NoError(t, err)
	require.Equal(t, []string{"a"}, policy.DisabledNodes)
	content, err = svc.GetClientNodeSubscription(context.Background())
	require.NoError(t, err)
	require.Equal(t, "proxies: []\n", content)
	repo.err = errors.New("database failed")
	_, err = svc.GetClientNodePolicy(context.Background())
	require.ErrorContains(t, err, "database failed")
}

func TestClientNodeSubscriptionValidation(t *testing.T) {
	repo := &clientNodeRepo{values: map[string]string{}}
	svc := &SettingService{settingRepo: repo}
	for _, invalid := range []string{"", "[broken", "a: &x value\nb: *x", "a: !custom value", "a: 1\n---\nb: 2", "a: 1\na: 2", strings.Repeat("a", MaxClientNodeContentSize+1)} {
		t.Run(invalid[:min(len(invalid), 12)], func(t *testing.T) {
			require.Error(t, svc.SetClientNodeSubscription(context.Background(), invalid))
			require.Zero(t, repo.writes)
			require.Error(t, svc.SetClientNodes(context.Background(), ClientNodePolicy{}, invalid))
			require.Zero(t, repo.bundles)
		})
	}
	require.NoError(t, svc.SetClientNodeSubscription(context.Background(), "proxies:\n  - name: edge\n    port: 443\n"))
	require.Equal(t, 1, repo.writes)
}

func TestClientNodeLegacyPolicyAndSubscriptionFallback(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("proxies:\n  - name: edge\n    port: 443\n"))
	}))
	defer server.Close()
	repo := &clientNodeRepo{values: map[string]string{
		legacyClientNodePolicyKey: `{"enabled":true,"disabled_nodes":["retired"],"subscription_url":"http://127.0.0.1:1/unavailable","backup_subscription_urls":["` + server.URL + `"]}`,
	}}
	svc := &SettingService{settingRepo: repo}
	policy, err := svc.GetClientNodePolicy(context.Background())
	require.NoError(t, err)
	require.Equal(t, &ClientNodePolicy{Enabled: true, DisabledNodes: []string{"retired"}}, policy)
	content, err := svc.GetClientNodeSubscription(context.Background())
	require.NoError(t, err)
	require.Contains(t, content, "name: edge")

	repo.values[clientNodePolicyKey] = `{"enabled":false,"disabled_nodes":[]}`
	repo.values[clientNodeSubscriptionKey] = "proxies: []\n"
	policy, err = svc.GetClientNodePolicy(context.Background())
	require.NoError(t, err)
	require.False(t, policy.Enabled)
	content, err = svc.GetClientNodeSubscription(context.Background())
	require.NoError(t, err)
	require.Equal(t, "proxies: []\n", content)
}
