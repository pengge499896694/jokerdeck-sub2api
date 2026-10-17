package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"gopkg.in/yaml.v3"
)

const (
	clientNodeGlobalKey           = "client_node_global_policy"
	defaultClientNodeSubscription = "https://234.qzz.io/fsllistyaml"
)

type ClientNodePolicy struct {
	Enabled                bool     `json:"enabled"`
	SubscriptionURL        string   `json:"subscription_url"`
	BackupSubscriptionURLs []string `json:"backup_subscription_urls,omitempty"`
	DisabledNodes          []string `json:"disabled_nodes"`
	SubscriptionEndpoint   string   `json:"subscription_endpoint,omitempty"`
}

type ClientNodeEntry struct {
	Name      string `json:"name"`
	Type      string `json:"type"`
	Enabled   bool   `json:"enabled"`
	LatencyMs *int64 `json:"latency_ms,omitempty"`
}

type ClientNodeSubscription struct {
	ID        string            `json:"id"`
	Name      string            `json:"name"`
	URL       string            `json:"url"`
	Primary   bool              `json:"primary"`
	Available bool              `json:"available"`
	Error     string            `json:"error,omitempty"`
	Nodes     []ClientNodeEntry `json:"nodes"`
}

type ClientNodeTestResult struct {
	SubscriptionID string `json:"subscription_id"`
	Name           string `json:"name"`
	Success        bool   `json:"success"`
	LatencyMs      int64  `json:"latency_ms,omitempty"`
	Message        string `json:"message,omitempty"`
}

func defaultClientNodePolicy() *ClientNodePolicy {
	return &ClientNodePolicy{Enabled: true, SubscriptionURL: defaultClientNodeSubscription, BackupSubscriptionURLs: []string{}, DisabledNodes: []string{}}
}

func validateClientNodeURL(raw string) error {
	u, err := url.Parse(raw)
	if err != nil || u.Scheme != "https" || u.Hostname() == "" || u.User != nil || u.Fragment != "" {
		return errors.New("subscription_url must be an HTTPS URL without credentials or fragment")
	}
	if ip := net.ParseIP(u.Hostname()); ip != nil && (!ip.IsGlobalUnicast() || isPrivateIP(ip)) {
		return errors.New("subscription_url must use a public host")
	}
	return nil
}

func validateClientNodeName(name string) error {
	if len(name) > 120 || strings.TrimSpace(name) != name || strings.ContainsAny(name, "\r\n\x00") {
		return errors.New("invalid node name")
	}
	return nil
}

func validateClientNodePolicy(policy *ClientNodePolicy) error {
	if policy == nil {
		return errors.New("policy is required")
	}
	if policy.SubscriptionURL != "" {
		if err := validateClientNodeURL(policy.SubscriptionURL); err != nil {
			return err
		}
	}
	if policy.SubscriptionURL == "" && len(policy.BackupSubscriptionURLs) == 0 {
		return errors.New("at least one subscription URL is required")
	}
	if len(policy.BackupSubscriptionURLs) > 5 {
		return errors.New("too many backup subscription URLs")
	}
	seenURLs := map[string]bool{policy.SubscriptionURL: true}
	for _, rawURL := range policy.BackupSubscriptionURLs {
		if err := validateClientNodeURL(rawURL); err != nil {
			return err
		}
		if seenURLs[rawURL] {
			return errors.New("subscription URLs must be unique")
		}
		seenURLs[rawURL] = true
	}
	if len(policy.DisabledNodes) > 500 {
		return errors.New("too many disabled nodes")
	}
	seen := make(map[string]bool, len(policy.DisabledNodes))
	for _, name := range policy.DisabledNodes {
		if name == "" || seen[name] {
			return errors.New("disabled_nodes contains an empty or duplicate name")
		}
		if err := validateClientNodeName(name); err != nil {
			return err
		}
		seen[name] = true
	}
	return nil
}

func (s *SettingService) GetGlobalClientNodePolicy(ctx context.Context) (*ClientNodePolicy, error) {
	raw, err := s.settingRepo.GetValue(ctx, clientNodeGlobalKey)
	if errors.Is(err, ErrSettingNotFound) {
		return defaultClientNodePolicy(), nil
	}
	if err != nil {
		return nil, err
	}
	var policy ClientNodePolicy
	if err := json.Unmarshal([]byte(raw), &policy); err != nil {
		return nil, fmt.Errorf("parse client node policy: %w", err)
	}
	if err := validateClientNodePolicy(&policy); err != nil {
		return nil, err
	}
	if policy.DisabledNodes == nil {
		policy.DisabledNodes = []string{}
	}
	return &policy, nil
}

func (s *SettingService) SetGlobalClientNodePolicy(ctx context.Context, policy *ClientNodePolicy) (*ClientNodePolicy, error) {
	if err := validateClientNodePolicy(policy); err != nil {
		return nil, err
	}
	copy := *policy
	copy.SubscriptionEndpoint = ""
	if copy.DisabledNodes == nil {
		copy.DisabledNodes = []string{}
	}
	data, err := json.Marshal(copy)
	if err != nil {
		return nil, err
	}
	if err := s.settingRepo.Set(ctx, clientNodeGlobalKey, string(data)); err != nil {
		return nil, err
	}
	return &copy, nil
}

func (s *SettingService) GetClientNodePolicy(ctx context.Context) (*ClientNodePolicy, error) {
	policy, err := s.GetGlobalClientNodePolicy(ctx)
	if err != nil {
		return nil, err
	}
	// Backup URLs may contain provider access tokens in query parameters. They
	// are only returned by the admin endpoint, never to regular clients.
	policy.BackupSubscriptionURLs = nil
	policy.SubscriptionEndpoint = "/api/v1/client/node-subscription"
	return policy, nil
}

type clientNodeSubscriptionDocument struct {
	Proxies []clientNodeProxy `yaml:"proxies"`
}

type clientNodeProxy struct {
	Name   string `yaml:"name"`
	Type   string `yaml:"type"`
	Server string `yaml:"server"`
	Port   int    `yaml:"port"`
}

func fetchClientNodeSource(ctx context.Context, client *http.Client, subscriptionURL string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, subscriptionURL, nil)
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("subscription returned HTTP %d", resp.StatusCode)
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, 1024*1024+1))
	if err != nil {
		return nil, err
	}
	if len(data) > 1024*1024 {
		return nil, errors.New("subscription is too large")
	}
	return data, nil
}

func fetchClientNodeSourceWithFallback(ctx context.Context, client *http.Client, urls []string) ([]byte, error) {
	for _, subscriptionURL := range urls {
		data, err := fetchClientNodeSource(ctx, client, subscriptionURL)
		if err == nil {
			var document clientNodeSubscriptionDocument
			if yaml.Unmarshal(data, &document) == nil {
				for _, proxy := range document.Proxies {
					if strings.TrimSpace(proxy.Name) != "" && strings.TrimSpace(proxy.Type) != "" {
						return data, nil
					}
				}
			}
		}
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
	}
	return nil, errors.New("all client node subscriptions failed")
}

func fetchClientNodeDocumentWithFallback(ctx context.Context, client *http.Client, urls []string) (*clientNodeSubscriptionDocument, error) {
	data, err := fetchClientNodeSourceWithFallback(ctx, client, urls)
	if err != nil {
		return nil, err
	}
	var document clientNodeSubscriptionDocument
	if err := yaml.Unmarshal(data, &document); err != nil {
		return nil, err
	}
	return &document, nil
}

func newClientNodeHTTPClient() (*http.Client, *http.Transport) {
	dialer := &net.Dialer{Timeout: 5 * time.Second}
	transport := &http.Transport{DialContext: func(ctx context.Context, network, address string) (net.Conn, error) {
		host, port, err := net.SplitHostPort(address)
		if err != nil {
			return nil, err
		}
		ips, err := net.DefaultResolver.LookupIPAddr(ctx, host)
		if err != nil {
			return nil, err
		}
		for _, item := range ips {
			if item.IP.IsGlobalUnicast() && !isPrivateIP(item.IP) {
				return dialer.DialContext(ctx, network, net.JoinHostPort(item.IP.String(), port))
			}
		}
		return nil, errors.New("subscription host has no public address")
	}}
	client := &http.Client{Timeout: 12 * time.Second, Transport: transport, CheckRedirect: func(req *http.Request, via []*http.Request) error {
		if len(via) >= 3 || req.URL.Scheme != "https" {
			return errors.New("invalid subscription redirect")
		}
		return nil
	}}
	return client, transport
}

// GetClientNodeSubscription returns the validated YAML selected by ordered
// primary/backup failover. It is called only from an authenticated endpoint.
func (s *SettingService) GetClientNodeSubscription(ctx context.Context) ([]byte, error) {
	policy, err := s.GetGlobalClientNodePolicy(ctx)
	if err != nil {
		return nil, err
	}
	if !policy.Enabled {
		return nil, errors.New("client node connections are disabled")
	}
	client, transport := newClientNodeHTTPClient()
	defer transport.CloseIdleConnections()
	sources, err := fetchClientNodeSources(ctx, client, policy)
	if err != nil {
		return nil, err
	}
	return mergeClientNodeSources(sources, policy.DisabledNodes)
}

type clientNodeSourceDocument struct {
	ID      string
	Name    string
	URL     string
	Primary bool
	Data    []byte
	Doc     *clientNodeSubscriptionDocument
}

func clientNodeSources(policy *ClientNodePolicy) []clientNodeSourceDocument {
	sources := []clientNodeSourceDocument{}
	if policy.SubscriptionURL != "" {
		sources = append(sources, clientNodeSourceDocument{ID: "primary", Name: "主订阅", URL: policy.SubscriptionURL, Primary: true})
	}
	for index, subscriptionURL := range policy.BackupSubscriptionURLs {
		sources = append(sources, clientNodeSourceDocument{
			ID: fmt.Sprintf("backup-%d", index+1), Name: fmt.Sprintf("备用订阅 %d", index+1), URL: subscriptionURL,
		})
	}
	return sources
}

func fetchClientNodeSources(ctx context.Context, client *http.Client, policy *ClientNodePolicy) ([]clientNodeSourceDocument, error) {
	sources := loadClientNodeSources(ctx, client, policy)
	for _, source := range sources {
		if source.Doc != nil {
			return sources, nil
		}
	}
	return nil, errors.New("all client node subscriptions failed")
}

func loadClientNodeSources(ctx context.Context, client *http.Client, policy *ClientNodePolicy) []clientNodeSourceDocument {
	sources := clientNodeSources(policy)
	var wg sync.WaitGroup
	for index := range sources {
		index := index
		wg.Add(1)
		go func() {
			defer wg.Done()
			data, err := fetchClientNodeSource(ctx, client, sources[index].URL)
			if err == nil {
				var document clientNodeSubscriptionDocument
				if yaml.Unmarshal(data, &document) == nil && hasUsableClientNodes(&document) {
					sources[index].Data = data
					sources[index].Doc = &document
				}
			}
		}()
	}
	wg.Wait()
	return sources
}

func hasUsableClientNodes(document *clientNodeSubscriptionDocument) bool {
	for _, proxy := range document.Proxies {
		if strings.TrimSpace(proxy.Name) != "" && strings.TrimSpace(proxy.Type) != "" {
			return true
		}
	}
	return false
}

func sourceNodeEntries(source clientNodeSourceDocument, disabled map[string]bool) []ClientNodeEntry {
	if source.Doc == nil {
		return []ClientNodeEntry{}
	}
	entries := make([]ClientNodeEntry, 0, len(source.Doc.Proxies))
	seen := make(map[string]bool)
	for _, node := range source.Doc.Proxies {
		if strings.TrimSpace(node.Name) == "" || strings.TrimSpace(node.Type) == "" || seen[node.Name] {
			continue
		}
		seen[node.Name] = true
		entries = append(entries, ClientNodeEntry{Name: node.Name, Type: node.Type, Enabled: !disabled[node.Name]})
	}
	return entries
}

// ListClientNodeSubscriptions keeps each subscription as a tree root so the
// admin panel can distinguish primary and backup sources and their nodes.
func (s *SettingService) ListClientNodeSubscriptions(ctx context.Context) ([]ClientNodeSubscription, error) {
	policy, err := s.GetGlobalClientNodePolicy(ctx)
	if err != nil {
		return nil, err
	}
	client, transport := newClientNodeHTTPClient()
	defer transport.CloseIdleConnections()
	sources := loadClientNodeSources(ctx, client, policy)
	disabled := make(map[string]bool, len(policy.DisabledNodes))
	for _, name := range policy.DisabledNodes {
		disabled[name] = true
	}
	result := make([]ClientNodeSubscription, 0, len(sources))
	for _, source := range sources {
		entry := ClientNodeSubscription{ID: source.ID, Name: source.Name, URL: source.URL, Primary: source.Primary, Available: source.Doc != nil, Nodes: sourceNodeEntries(source, disabled)}
		if source.Doc == nil {
			entry.Error = "订阅读取失败或没有可用节点"
		}
		result = append(result, entry)
	}
	return result, nil
}

// ListClientNodes remains as a flattened compatibility helper for internal callers.
func (s *SettingService) ListClientNodes(ctx context.Context) ([]ClientNodeEntry, error) {
	subscriptions, err := s.ListClientNodeSubscriptions(ctx)
	if err != nil {
		return nil, err
	}
	var nodes []ClientNodeEntry
	for _, subscription := range subscriptions {
		nodes = append(nodes, subscription.Nodes...)
	}
	return nodes, nil
}

func mergeClientNodeSources(sources []clientNodeSourceDocument, disabledNames []string) ([]byte, error) {
	disabled := make(map[string]bool, len(disabledNames))
	for _, name := range disabledNames {
		disabled[name] = true
	}
	primaryNames := make([]string, 0)
	backupNames := make([]string, 0)
	sourceGroups := make([]map[string]interface{}, 0, len(sources))
	seen := map[string]bool{"CLIENT-PRIMARY": true, "CLIENT-BACKUP": true, "CLIENT-AUTO": true}
	var merged []interface{}
	for _, source := range sources {
		if source.Doc == nil {
			continue
		}
		sourceNames := make([]string, 0, len(source.Doc.Proxies))
		var raw map[string]interface{}
		if yaml.Unmarshal(source.Data, &raw) != nil {
			continue
		}
		if values, ok := raw["proxies"].([]interface{}); ok {
			for _, value := range values {
				item, ok := value.(map[string]interface{})
				if !ok {
					continue
				}
				name, _ := item["name"].(string)
				if name == "" || disabled[name] {
					continue
				}
				if seen[name] {
					original := name
					for suffix := 1; seen[name]; suffix++ {
						name = fmt.Sprintf("%s/%s/%d", source.ID, original, suffix)
					}
					item["name"] = name
				}
				seen[name] = true
				merged = append(merged, item)
				sourceNames = append(sourceNames, name)
				if source.Primary {
					primaryNames = append(primaryNames, name)
				} else {
					backupNames = append(backupNames, name)
				}
			}
		}
		if len(sourceNames) > 0 {
			sourceGroups = append(sourceGroups, map[string]interface{}{
				"id": source.ID, "filename": source.Name, "role": map[bool]string{true: "primary", false: "backup"}[source.Primary], "nodes": sourceNames,
			})
		}
	}
	if len(merged) == 0 {
		return nil, errors.New("all client node subscriptions have no enabled nodes")
	}
	result := map[string]interface{}{"proxies": merged, "x-jokerdeck-sources": sourceGroups}
	const testURL = "https://www.gstatic.com/generate_204"
	groups := make([]interface{}, 0, 3)
	order := make([]string, 0, 2)
	if len(primaryNames) > 0 {
		groups = append(groups, map[string]interface{}{"name": "CLIENT-PRIMARY", "type": "url-test", "proxies": primaryNames, "url": testURL, "interval": 300})
		order = append(order, "CLIENT-PRIMARY")
	}
	if len(backupNames) > 0 {
		groups = append(groups, map[string]interface{}{"name": "CLIENT-BACKUP", "type": "url-test", "proxies": backupNames, "url": testURL, "interval": 300})
		order = append(order, "CLIENT-BACKUP")
	}
	if len(order) == 2 {
		groups = append(groups, map[string]interface{}{"name": "CLIENT-AUTO", "type": "fallback", "proxies": order, "url": testURL, "interval": 300})
	} else {
		groups = append(groups, map[string]interface{}{"name": "CLIENT-AUTO", "type": "select", "proxies": order})
	}
	result["proxy-groups"] = groups
	result["rules"] = []string{"MATCH,CLIENT-AUTO"}
	return yaml.Marshal(result)
}

func (s *SettingService) TestClientNodeNodes(ctx context.Context, subscriptionID string, names []string) ([]ClientNodeTestResult, error) {
	if len(names) > 200 {
		return nil, errors.New("too many nodes to test")
	}
	policy, err := s.GetGlobalClientNodePolicy(ctx)
	if err != nil {
		return nil, err
	}
	client, transport := newClientNodeHTTPClient()
	defer transport.CloseIdleConnections()
	sources, err := fetchClientNodeSources(ctx, client, policy)
	if err != nil {
		return nil, err
	}
	wanted := make(map[string]bool, len(names))
	for _, name := range names {
		wanted[name] = true
	}
	type target struct {
		clientNodeProxy
		subscriptionID string
	}
	var targets []target
	seen := map[string]bool{}
	for _, source := range sources {
		if source.Doc == nil || (subscriptionID != "" && subscriptionID != source.ID) {
			continue
		}
		for _, proxy := range source.Doc.Proxies {
			if proxy.Name == "" || disabledNode(policy.DisabledNodes, proxy.Name) || seen[source.ID+"/"+proxy.Name] || (len(wanted) > 0 && !wanted[proxy.Name]) {
				continue
			}
			seen[source.ID+"/"+proxy.Name] = true
			targets = append(targets, target{clientNodeProxy: proxy, subscriptionID: source.ID})
		}
	}
	if len(targets) > 200 {
		return nil, errors.New("too many nodes to test")
	}
	results := make([]ClientNodeTestResult, len(targets))
	var wg sync.WaitGroup
	limit := make(chan struct{}, 8)
	for index := range targets {
		index := index
		wg.Add(1)
		go func() {
			defer wg.Done()
			limit <- struct{}{}
			defer func() { <-limit }()
			started := time.Now()
			result := ClientNodeTestResult{SubscriptionID: targets[index].subscriptionID, Name: targets[index].Name}
			if targets[index].Server == "" || targets[index].Port <= 0 || targets[index].Port > 65535 {
				result.Message = "节点没有可测试的 server/port"
				results[index] = result
				return
			}
			conn, dialErr := dialPublicClientNode(ctx, targets[index].Server, targets[index].Port)
			if dialErr != nil {
				result.Message = dialErr.Error()
			} else {
				result.Success = true
				result.LatencyMs = time.Since(started).Milliseconds()
				_ = conn.Close()
			}
			results[index] = result
		}()
	}
	wg.Wait()
	return results, nil
}

func dialPublicClientNode(ctx context.Context, host string, port int) (net.Conn, error) {
	if strings.ContainsAny(host, "/@\\\x00") {
		return nil, errors.New("invalid node host")
	}
	addresses, err := net.DefaultResolver.LookupIPAddr(ctx, host)
	if err != nil {
		return nil, err
	}
	dialer := net.Dialer{Timeout: 3 * time.Second}
	for _, address := range addresses {
		if address.IP.IsGlobalUnicast() && !isPrivateIP(address.IP) {
			return dialer.DialContext(ctx, "tcp", net.JoinHostPort(address.IP.String(), strconv.Itoa(port)))
		}
	}
	return nil, errors.New("node host has no public address")
}

func disabledNode(disabled []string, name string) bool {
	for _, item := range disabled {
		if item == name {
			return true
		}
	}
	return false
}
