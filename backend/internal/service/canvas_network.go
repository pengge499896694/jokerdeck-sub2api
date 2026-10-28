package service

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"net/url"
	"strings"
	"time"
	"unicode"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

const canvasNetworkSettingsKey = "canvas_network"

// CanvasNetworkSettings controls the optional same-origin relay used by the
// embedded canvas for reference images and media downloads. The relay never
// accepts arbitrary destinations; AllowedHosts is matched against the request
// host before any outbound connection is made.
type CanvasNetworkSettings struct {
	Enabled        bool     `json:"enabled"`
	VisibleToUsers bool     `json:"visible_to_users"`
	ProxyID        *int64   `json:"proxy_id"`
	AllowedHosts   []string `json:"allowed_hosts"`
}

var defaultCanvasNetworkHosts = []string{
	"github.com",
	"raw.githubusercontent.com",
	"githubusercontent.com",
	"githubassets.com",
	"github.io",
	"jsdelivr.net",
	"unpkg.com",
}

func defaultCanvasNetworkSettings() CanvasNetworkSettings {
	return CanvasNetworkSettings{AllowedHosts: append([]string(nil), defaultCanvasNetworkHosts...)}
}

func normalizeCanvasNetworkHosts(values []string) ([]string, error) {
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, raw := range values {
		host := strings.ToLower(strings.TrimSpace(raw))
		host = strings.TrimSuffix(host, ".")
		if host == "" || !strings.Contains(host, ".") || net.ParseIP(host) != nil || len(host) > 253 || strings.ContainsAny(host, "/*\\:@?#") || strings.IndexFunc(host, unicode.IsSpace) >= 0 {
			return nil, infraerrors.BadRequest("CANVAS_NETWORK_HOST_INVALID", "画布线路允许域名格式错误")
		}
		if _, err := url.Parse("https://" + host); err != nil {
			return nil, infraerrors.BadRequest("CANVAS_NETWORK_HOST_INVALID", "画布线路允许域名格式错误")
		}
		if _, ok := seen[host]; ok {
			continue
		}
		seen[host] = struct{}{}
		result = append(result, host)
	}
	if len(result) == 0 || len(result) > 32 {
		return nil, infraerrors.BadRequest("CANVAS_NETWORK_HOSTS_INVALID", "画布线路至少需要一个允许域名，最多 32 个")
	}
	return result, nil
}

func normalizeCanvasNetworkSettings(input CanvasNetworkSettings) (*CanvasNetworkSettings, error) {
	hosts, err := normalizeCanvasNetworkHosts(input.AllowedHosts)
	if err != nil {
		return nil, err
	}
	if input.ProxyID != nil && *input.ProxyID <= 0 {
		return nil, infraerrors.BadRequest("CANVAS_NETWORK_PROXY_INVALID", "请选择服务器默认线路或有效代理")
	}
	copy := input
	copy.AllowedHosts = hosts
	return &copy, nil
}

func (s *SettingService) GetCanvasNetworkSettings(ctx context.Context) (*CanvasNetworkSettings, error) {
	if s == nil || s.settingRepo == nil {
		return nil, errors.New("settings repository unavailable")
	}
	raw, err := s.settingRepo.GetValue(ctx, canvasNetworkSettingsKey)
	if err != nil && !errors.Is(err, ErrSettingNotFound) {
		return nil, err
	}
	settings := defaultCanvasNetworkSettings()
	if strings.TrimSpace(raw) != "" {
		if err := json.Unmarshal([]byte(raw), &settings); err != nil {
			return nil, err
		}
	}
	return normalizeCanvasNetworkSettings(settings)
}

func (s *SettingService) SetCanvasNetworkSettings(ctx context.Context, input CanvasNetworkSettings) (*CanvasNetworkSettings, error) {
	settings, err := normalizeCanvasNetworkSettings(input)
	if err != nil {
		return nil, err
	}
	if s == nil || s.settingRepo == nil {
		return nil, errors.New("settings repository unavailable")
	}
	if settings.ProxyID != nil {
		if _, err := s.CanvasNetworkProxyURL(ctx, settings.ProxyID); err != nil {
			return nil, infraerrors.BadRequest("CANVAS_NETWORK_PROXY_INVALID", "所选代理不可用或已过期，请刷新线路列表后重试")
		}
	}
	data, err := json.Marshal(settings)
	if err != nil {
		return nil, err
	}
	if err := s.settingRepo.Set(ctx, canvasNetworkSettingsKey, string(data)); err != nil {
		return nil, err
	}
	return settings, nil
}

func (s *SettingService) CanvasNetworkProxyURL(ctx context.Context, id *int64) (string, error) {
	if id == nil {
		return "", nil
	}
	if s == nil || s.proxyRepo == nil {
		return "", errors.New("proxy repository unavailable")
	}
	proxy, err := s.proxyRepo.GetByID(ctx, *id)
	if err != nil {
		return "", err
	}
	if proxy == nil || !proxy.IsActive() || proxy.IsExpired(time.Now()) {
		return "", errors.New("canvas relay proxy is inactive")
	}
	return proxy.URL(), nil
}
