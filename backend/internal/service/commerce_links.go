package service

import (
	"context"
	"encoding/json"
	"errors"
	"net/url"
	"strings"
	"unicode"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

const commerceLinksKey = "commerce_links"

type CommerceLinks struct {
	XianyuOrderURL string `json:"xianyu_order_url"`
	StoreURL       string `json:"store_url"`
	StoreBackupURL string `json:"store_backup_url"`
}

func defaultCommerceLinks() CommerceLinks {
	return CommerceLinks{
		XianyuOrderURL: "https://m.tb.cn/h.8orR1uB?tk=qgivTTR6A22",
		StoreURL:       "https://catfk.com/shop/AYBIRLMN",
	}
}

func validateCommerceURL(raw string, required bool) (string, error) {
	value := strings.TrimSpace(raw)
	if value == "" && !required {
		return "", nil
	}
	u, err := url.Parse(value)
	if err != nil || len(value) > 2048 || value == "" ||
		strings.ContainsAny(value, "\\") || strings.IndexFunc(value, unicode.IsSpace) >= 0 ||
		(u.Scheme != "http" && u.Scheme != "https") || u.Hostname() == "" || u.User != nil {
		return "", infraerrors.BadRequest("COMMERCE_URL_INVALID", "请输入不含账号密码的完整 HTTP(S) 链接，最长 2048 字符")
	}
	return value, nil
}

func normalizeCommerceLinks(links CommerceLinks) (*CommerceLinks, error) {
	var err error
	if links.XianyuOrderURL, err = validateCommerceURL(links.XianyuOrderURL, true); err != nil {
		return nil, err
	}
	if links.StoreURL, err = validateCommerceURL(links.StoreURL, true); err != nil {
		return nil, err
	}
	if links.StoreBackupURL, err = validateCommerceURL(links.StoreBackupURL, false); err != nil {
		return nil, err
	}
	return &links, nil
}

func (s *SettingService) GetCommerceLinks(ctx context.Context) (*CommerceLinks, error) {
	raw, err := s.settingRepo.GetValue(ctx, commerceLinksKey)
	if err != nil && !errors.Is(err, ErrSettingNotFound) {
		return nil, err
	}
	links := defaultCommerceLinks()
	if err == nil && strings.TrimSpace(raw) != "" {
		if err := json.Unmarshal([]byte(raw), &links); err != nil {
			return nil, err
		}
	}
	return normalizeCommerceLinks(links)
}

func (s *SettingService) SetCommerceLinks(ctx context.Context, input CommerceLinks) (*CommerceLinks, error) {
	links, err := normalizeCommerceLinks(input)
	if err != nil {
		return nil, err
	}
	data, err := json.Marshal(links)
	if err != nil {
		return nil, err
	}
	if err := s.settingRepo.Set(ctx, commerceLinksKey, string(data)); err != nil {
		return nil, err
	}
	return links, nil
}
