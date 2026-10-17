package service

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"gopkg.in/yaml.v3"
)

type clientNodeSettingRepo struct {
	SettingRepository
	values map[string]string
}

func (r *clientNodeSettingRepo) GetValue(_ context.Context, key string) (string, error) {
	value, ok := r.values[key]
	if !ok {
		return "", ErrSettingNotFound
	}
	return value, nil
}

func (r *clientNodeSettingRepo) Set(_ context.Context, key, value string) error {
	r.values[key] = value
	return nil
}

func (r *clientNodeSettingRepo) Delete(_ context.Context, key string) error {
	delete(r.values, key)
	return nil
}

func TestClientNodePolicyDefaultsAndDisabledNodes(t *testing.T) {
	ctx := context.Background()
	s := &SettingService{settingRepo: &clientNodeSettingRepo{values: map[string]string{}}}
	defaultPolicy, err := s.GetClientNodePolicy(ctx)
	if err != nil || !defaultPolicy.Enabled || defaultPolicy.SubscriptionURL != defaultClientNodeSubscription {
		t.Fatalf("unexpected default: %#v, %v", defaultPolicy, err)
	}
	_, err = s.SetGlobalClientNodePolicy(ctx, &ClientNodePolicy{
		Enabled: true, SubscriptionURL: "https://example.com/nodes.yaml", DisabledNodes: []string{"US-Slow"},
	})
	if err != nil {
		t.Fatal(err)
	}
	policy, err := s.GetClientNodePolicy(ctx)
	if err != nil || len(policy.DisabledNodes) != 1 || policy.DisabledNodes[0] != "US-Slow" || policy.SubscriptionEndpoint != "/api/v1/client/node-subscription" {
		t.Fatalf("unexpected policy: %#v, %v", policy, err)
	}
}

func TestClientNodePolicyBackupURLsAreValidatedAndRedacted(t *testing.T) {
	s := &SettingService{settingRepo: &clientNodeSettingRepo{values: map[string]string{}}}
	policy, err := s.SetGlobalClientNodePolicy(context.Background(), &ClientNodePolicy{
		Enabled:                true,
		SubscriptionURL:        "https://example.com/primary.yaml",
		BackupSubscriptionURLs: []string{"https://example.com/backup.yaml?token=secret", "https://example.net/backup.yaml"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(policy.BackupSubscriptionURLs) != 2 {
		t.Fatalf("backup URLs not persisted: %#v", policy.BackupSubscriptionURLs)
	}
	storedPolicy, err := s.GetGlobalClientNodePolicy(context.Background())
	if err != nil || len(storedPolicy.BackupSubscriptionURLs) != 2 || storedPolicy.BackupSubscriptionURLs[1] != "https://example.net/backup.yaml" {
		t.Fatalf("backup URLs not restored: %#v, %v", storedPolicy, err)
	}
	userPolicy, err := s.GetClientNodePolicy(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(userPolicy.BackupSubscriptionURLs) != 0 {
		t.Fatalf("backup URLs leaked to client policy: %#v", userPolicy.BackupSubscriptionURLs)
	}
	for _, backups := range [][]string{
		{"https://example.com/primary.yaml"},
		{"http://example.com/backup.yaml"},
		{"https://user:password@example.com/backup.yaml"},
	} {
		_, err := s.SetGlobalClientNodePolicy(context.Background(), &ClientNodePolicy{
			Enabled: true, SubscriptionURL: "https://example.com/primary.yaml", BackupSubscriptionURLs: backups,
		})
		if err == nil {
			t.Fatalf("unsafe or duplicate backup URL accepted: %#v", backups)
		}
	}
}

func TestFetchClientNodeDocumentWithFallback(t *testing.T) {
	calls := 0
	good := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.Header().Set("Content-Type", "text/yaml")
		_, _ = w.Write([]byte("proxies:\n  - name: JP-Fast\n    type: ss\n"))
	}))
	defer good.Close()
	client := good.Client()
	document, err := fetchClientNodeDocumentWithFallback(context.Background(), client, []string{"http://127.0.0.1:1/unreachable", good.URL})
	if err != nil {
		t.Fatal(err)
	}
	if calls != 1 || len(document.Proxies) != 1 || document.Proxies[0].Name != "JP-Fast" {
		t.Fatalf("unexpected fallback result: calls=%d document=%#v", calls, document)
	}
}

func TestFetchClientNodeSourceWithFallbackReturnsBackupYAML(t *testing.T) {
	primary := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "temporary failure", http.StatusBadGateway)
	}))
	defer primary.Close()
	backupYAML := []byte("proxies:\n  - name: US-Backup\n    type: vmess\n")
	backup := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(backupYAML)
	}))
	defer backup.Close()

	data, err := fetchClientNodeSourceWithFallback(context.Background(), backup.Client(), []string{primary.URL, backup.URL})
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != string(backupYAML) {
		t.Fatalf("unexpected backup YAML: %q", data)
	}
}

func TestFetchClientNodeSourceWithFallbackRedactsSourceURL(t *testing.T) {
	_, err := fetchClientNodeSourceWithFallback(context.Background(), &http.Client{Timeout: 10 * time.Millisecond}, []string{"http://127.0.0.1:1/nodes?token=secret"})
	if err == nil || strings.Contains(err.Error(), "secret") {
		t.Fatalf("source URL leaked in fetch error: %v", err)
	}
}

func TestFetchClientNodeSourceWithFallbackOnTimeoutAndInvalidYAML(t *testing.T) {
	backupYAML := []byte("proxies:\n  - name: JP-Backup\n    type: ss\n")
	backup := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(backupYAML)
	}))
	defer backup.Close()

	tests := []struct {
		name    string
		handler http.HandlerFunc
	}{
		{"timeout", func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(50 * time.Millisecond)
			_, _ = w.Write(backupYAML)
		}},
		{"invalid YAML", func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte("proxies: [unterminated"))
		}},
		{"empty proxies", func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte("proxies: []\n"))
		}},
		{"missing proxies", func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte("port: 7890\n"))
		}},
		{"proxies without usable names or types", func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte("proxies:\n  - name: '  '\n    type: ss\n  - name: JP-Empty\n"))
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			primary := httptest.NewServer(test.handler)
			defer primary.Close()
			client := primary.Client()
			client.Timeout = 10 * time.Millisecond
			data, err := fetchClientNodeSourceWithFallback(context.Background(), client, []string{primary.URL, backup.URL})
			if err != nil {
				t.Fatal(err)
			}
			if string(data) != string(backupYAML) {
				t.Fatalf("unexpected backup YAML: %q", data)
			}
		})
	}
}

func TestFetchClientNodeDocumentWithFallbackRejectsBadPrimaryAndUsesBackup(t *testing.T) {
	calls := 0
	bad := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		http.Error(w, "temporary failure", http.StatusBadGateway)
	}))
	defer bad.Close()
	good := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		_, _ = w.Write([]byte("proxies:\n  - name: US-Fast\n    type: vmess\n"))
	}))
	defer good.Close()
	document, err := fetchClientNodeDocumentWithFallback(context.Background(), bad.Client(), []string{bad.URL, good.URL})
	if err != nil {
		t.Fatal(err)
	}
	if calls != 2 || document.Proxies[0].Name != "US-Fast" {
		t.Fatalf("backup was not used: calls=%d document=%#v", calls, document)
	}
}

func TestClientNodePolicyRejectsUnsafeInputAndDBFailure(t *testing.T) {
	for _, address := range []string{"http://example.com/nodes", "https://127.0.0.1/nodes", "https://100.64.0.1/nodes", "https://user:password@example.com/nodes"} {
		if validateClientNodeURL(address) == nil {
			t.Fatalf("unsafe URL accepted: %q", address)
		}
	}
	s := &SettingService{settingRepo: &clientNodeSettingRepo{values: map[string]string{}}}
	_, err := s.SetGlobalClientNodePolicy(context.Background(), &ClientNodePolicy{
		Enabled: true, SubscriptionURL: defaultClientNodeSubscription, DisabledNodes: []string{"US-A", "US-A"},
	})
	if err == nil {
		t.Fatal("duplicate disabled node accepted")
	}
}

func TestClientNodeProbeRejectsSharedAddressSpace(t *testing.T) {
	if conn, err := dialPublicClientNode(context.Background(), "100.64.0.1", 443); err == nil {
		_ = conn.Close()
		t.Fatal("CGNAT address was accepted for node probing")
	}
}

func TestClientNodePolicyAllowsBackupOnly(t *testing.T) {
	s := &SettingService{settingRepo: &clientNodeSettingRepo{values: map[string]string{}}}
	policy, err := s.SetGlobalClientNodePolicy(context.Background(), &ClientNodePolicy{
		Enabled: true, BackupSubscriptionURLs: []string{"https://example.com/backup.yaml"},
	})
	if err != nil || policy.SubscriptionURL != "" {
		t.Fatalf("backup-only policy rejected: %#v, %v", policy, err)
	}
	if _, err := s.SetGlobalClientNodePolicy(context.Background(), &ClientNodePolicy{Enabled: true}); err == nil {
		t.Fatal("empty policy accepted")
	}
}

func TestMergeClientNodeSourcesPreservesPrimaryPriority(t *testing.T) {
	primary := []byte("proxies:\n  - name: JP-Fast\n    type: ss\n    server: primary.example.com\n    port: 443\n  - name: JP-Off\n    type: ss\n    server: primary.example.com\n    port: 443\n")
	backup := []byte("proxies:\n  - name: JP-Fast\n    type: ss\n    server: backup.example.com\n    port: 443\n  - name: SG-Backup\n    type: ss\n    server: backup.example.com\n    port: 443\n")
	sources := []clientNodeSourceDocument{
		{Primary: true, ID: "primary", Data: primary, Doc: &clientNodeSubscriptionDocument{Proxies: []clientNodeProxy{{Name: "JP-Fast", Type: "ss"}}}},
		{ID: "backup-1", Data: backup, Doc: &clientNodeSubscriptionDocument{Proxies: []clientNodeProxy{{Name: "SG-Backup", Type: "ss"}}}},
	}
	data, err := mergeClientNodeSources(sources, []string{"JP-Off"})
	if err != nil {
		t.Fatal(err)
	}
	var document struct {
		Proxies []struct {
			Name string `yaml:"name"`
		} `yaml:"proxies"`
		Groups []struct {
			Name    string   `yaml:"name"`
			Type    string   `yaml:"type"`
			Proxies []string `yaml:"proxies"`
		} `yaml:"proxy-groups"`
		Rules   []string `yaml:"rules"`
		Sources []struct {
			ID    string   `yaml:"id"`
			Role  string   `yaml:"role"`
			Nodes []string `yaml:"nodes"`
		} `yaml:"x-jokerdeck-sources"`
	}
	if err := yaml.Unmarshal(data, &document); err != nil {
		t.Fatal(err)
	}
	if len(document.Proxies) != 3 || document.Proxies[1].Name != "backup-1/JP-Fast/1" ||
		len(document.Groups) != 3 || document.Groups[0].Type != "url-test" ||
		document.Groups[1].Type != "url-test" || document.Groups[2].Type != "fallback" ||
		document.Groups[2].Proxies[0] != "CLIENT-PRIMARY" || document.Groups[2].Proxies[1] != "CLIENT-BACKUP" ||
		len(document.Rules) != 1 || document.Rules[0] != "MATCH,CLIENT-AUTO" {
		t.Fatalf("unexpected merged document: %#v", document)
	}
	if len(document.Sources) != 2 || document.Sources[0].Role != "primary" ||
		document.Sources[1].ID != "backup-1" || document.Sources[1].Nodes[0] != "backup-1/JP-Fast/1" {
		t.Fatalf("unexpected subscription groups: %#v", document.Sources)
	}
}

func TestMergeClientNodeSourcesUsesBackupWhenPrimaryUnavailable(t *testing.T) {
	backup := []byte("proxies:\n  - name: SG-Backup\n    type: ss\n")
	sources := []clientNodeSourceDocument{
		{Primary: true, ID: "primary"},
		{ID: "backup-1", Data: backup, Doc: &clientNodeSubscriptionDocument{Proxies: []clientNodeProxy{{Name: "SG-Backup", Type: "ss"}}}},
	}
	data, err := mergeClientNodeSources(sources, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "CLIENT-BACKUP") || strings.Contains(string(data), "CLIENT-PRIMARY") {
		t.Fatalf("backup-only document is incorrect: %s", data)
	}
}

func TestClientNodeSubscriptionTreeRetainsFailedSources(t *testing.T) {
	primary := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "unavailable", http.StatusBadGateway)
	}))
	defer primary.Close()
	backup := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("proxies:\n  - name: Backup\n    type: ss\n"))
	}))
	defer backup.Close()
	policy := &ClientNodePolicy{SubscriptionURL: primary.URL, BackupSubscriptionURLs: []string{backup.URL}}
	client := primary.Client()
	sources, err := fetchClientNodeSources(context.Background(), client, policy)
	if err != nil || len(sources) != 2 || sources[0].Doc != nil || sources[1].Doc == nil {
		t.Fatalf("failed source was not retained: %#v, %v", sources, err)
	}
}

func TestDialPublicClientNodeRejectsLocalAddresses(t *testing.T) {
	for _, host := range []string{"127.0.0.1", "10.0.0.1", "localhost", "::1"} {
		if conn, err := dialPublicClientNode(context.Background(), host, 443); err == nil {
			_ = conn.Close()
			t.Fatalf("local address accepted: %s", host)
		}
	}
}
