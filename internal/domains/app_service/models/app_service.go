package models

import (
	"errors"
	"net/url"
	"strings"

	"github.com/vukyn/isme/internal/domains/app_service/constants"
	roleConstants "github.com/vukyn/isme/internal/domains/role/constants"
)

// MaxAdditionalRedirectURLs caps the OAuth-style allowlist of EXTRA permitted
// callbacks. The primary redirect_url is separate and does NOT count toward
// this limit.
const MaxAdditionalRedirectURLs = 3

// ValidateRedirectURLList trims, validates, and dedupes the additional
// redirect-URL allowlist. Each entry must be a non-empty absolute URL (mirrors
// the single redirect_url validation), the list is deduped preserving first-seen
// order, and the result is capped at MaxAdditionalRedirectURLs. The returned
// slice is the cleaned, deduped list (never nil — empty input yields an empty,
// non-nil slice).
func ValidateRedirectURLList(urls []string) ([]string, error) {
	cleaned := make([]string, 0, len(urls))
	seen := make(map[string]struct{}, len(urls))
	for _, raw := range urls {
		trimmed := strings.TrimSpace(raw)
		if trimmed == "" {
			return nil, errors.New("redirect_urls entries must not be empty")
		}
		parsed, err := url.ParseRequestURI(trimmed)
		if err != nil || parsed.Scheme == "" || parsed.Host == "" {
			return nil, errors.New("each redirect_urls entry must be a valid URL")
		}
		if _, dup := seen[trimmed]; dup {
			continue
		}
		seen[trimmed] = struct{}{}
		cleaned = append(cleaned, trimmed)
	}
	if len(cleaned) > MaxAdditionalRedirectURLs {
		return nil, errors.New("redirect_urls allows at most 3 additional URLs")
	}
	return cleaned, nil
}

type RegisterRequest struct {
	AppCode      string   `json:"app_code"`
	AppName      string   `json:"app_name"`
	RedirectURL  string   `json:"redirect_url"`
	RedirectURLs []string `json:"redirect_urls"` // optional additional allowed callbacks (max 3)
	CtxInfo      string   `json:"ctx_info"`
	Icon         string   `json:"icon"`  // optional appearance icon key; empty = neutral
	Color        string   `json:"color"` // optional appearance palette key; empty = neutral
}

func (r RegisterRequest) Validate() error {
	if r.AppCode == "" {
		return errors.New("app_code is required")
	}
	if r.AppName == "" {
		return errors.New("app_name is required")
	}
	if r.RedirectURL == "" {
		return errors.New("redirect_url is required")
	}
	if len(r.RedirectURLs) > 0 {
		if _, err := ValidateRedirectURLList(r.RedirectURLs); err != nil {
			return err
		}
	}
	if r.CtxInfo == "" {
		return errors.New("ctx_info is required")
	}
	if !roleConstants.IsValidIcon(r.Icon) {
		return errors.New("icon is not a known icon key")
	}
	if !constants.IsValidColor(r.Color) {
		return errors.New("color is not a known palette key")
	}
	return nil
}

type RegisterResponse struct {
	AppSecret string `json:"app_secret"`
}

type VerifyRequest struct {
	AppCode   string `json:"app_code"`
	CtxInfo   string `json:"ctx_info"`
	AppSecret string `json:"app_secret"`
}

func (r VerifyRequest) Validate() error {
	if r.AppCode == "" {
		return errors.New("app_code is required")
	}
	if r.CtxInfo == "" {
		return errors.New("ctx_info is required")
	}
	if r.AppSecret == "" {
		return errors.New("app_secret is required")
	}
	return nil
}

type VerifyResponse struct {
	Ok bool `json:"ok"`
}

type RefreshRequest struct {
	AppCode   string `json:"app_code"`
	AppSecret string `json:"app_secret"`
	CtxInfo   string `json:"ctx_info"`
}

func (r RefreshRequest) Validate() error {
	if r.AppCode == "" {
		return errors.New("app_code is required")
	}
	if r.AppSecret == "" {
		return errors.New("app_secret is required")
	}
	if r.CtxInfo == "" {
		return errors.New("ctx_info is required")
	}
	return nil
}

type RefreshResponse struct {
	AppSecret string `json:"app_secret"`
}

// ListRequest for listing app services with pagination and filters
type ListRequest struct {
	Page     int    `json:"page" query:"page"`
	PageSize int    `json:"page_size" query:"page_size"`
	Search   string `json:"search" query:"search"`     // search by app_name or app_code
	Status   int32  `json:"status" query:"status"`     // 1=active, 2=inactive, 3=terminated
	CtxInfo  string `json:"ctx_info" query:"ctx_info"` // filter by ctx_info
}

// Validate performs pure checks only; normalization is handled by Normalize.
func (r ListRequest) Validate() error {
	if r.Status != 0 &&
		r.Status != constants.AppServiceStatusActive &&
		r.Status != constants.AppServiceStatusInactive &&
		r.Status != constants.AppServiceStatusTerminated {
		return errors.New("invalid status, must be 1 (active), 2 (inactive) or 3 (terminated)")
	}
	if r.CtxInfo != "" {
		if _, ok := constants.AllowedCtxInfos[r.CtxInfo]; !ok {
			return errors.New("invalid ctx_info")
		}
	}
	return nil
}

// Normalize applies pagination defaults; pointer receiver so callers keep the changes.
func (r *ListRequest) Normalize() {
	if r.Page < 1 {
		r.Page = 1
	}
	if r.PageSize < 1 || r.PageSize > 100 {
		r.PageSize = 10
	}
}

// AppServiceListItem represents an app service in the list response.
// AppSecret is intentionally absent — it must never leave the usecase layer.
type AppServiceListItem struct {
	ID             string   `json:"id"`
	AppCode        string   `json:"app_code"`
	AppName        string   `json:"app_name"`
	RedirectURL    string   `json:"redirect_url"`
	RedirectURLs   []string `json:"redirect_urls"` // additional allowed callbacks (never nil)
	CtxInfo        string   `json:"ctx_info"`
	Status         int32    `json:"status"`
	Icon           string   `json:"icon"`
	Color          string   `json:"color"`
	CreatedAt      string   `json:"created_at"`
	CreatedBy      string   `json:"created_by"`       // creator user id
	CreatedByEmail string   `json:"created_by_email"` // resolved creator email (empty when unresolvable)
	UpdatedAt      string   `json:"updated_at"`
	UpdatedBy      string   `json:"updated_by"`
}

// ListResponse for app service list endpoint
type ListResponse struct {
	Items []AppServiceListItem `json:"items"`
	Total int64                `json:"total"`
	Page  int                  `json:"page"`
}

// UpdateStatusRequest for updating app service status
type UpdateStatusRequest struct {
	Status int32 `json:"status"`
}

func (r UpdateStatusRequest) Validate() error {
	if r.Status != constants.AppServiceStatusActive &&
		r.Status != constants.AppServiceStatusInactive &&
		r.Status != constants.AppServiceStatusTerminated {
		return errors.New("invalid status, must be 1 (active), 2 (inactive) or 3 (terminated)")
	}
	return nil
}

// UpdateAppearanceRequest is a partial update of an app service's display and
// SSO-contract fields. All fields are optional pointers; nil means "leave
// unchanged". At least one field must be present.
type UpdateAppearanceRequest struct {
	AppName      *string   `json:"app_name"`
	RedirectURL  *string   `json:"redirect_url"`
	RedirectURLs *[]string `json:"redirect_urls"` // nil = unchanged, [] = clear, else replace (max 3)
	Icon         *string   `json:"icon"`
	Color        *string   `json:"color"`
}

func (r UpdateAppearanceRequest) Validate() error {
	if r.AppName == nil && r.RedirectURL == nil && r.RedirectURLs == nil && r.Icon == nil && r.Color == nil {
		return errors.New("at least one field is required")
	}
	if r.AppName != nil && *r.AppName == "" {
		return errors.New("app_name must not be empty")
	}
	// redirect_url is part of the SSO contract. An empty string is allowed to
	// clear it (prod app_services were seeded empty); a non-empty value must be
	// a valid absolute URL.
	if r.RedirectURL != nil && *r.RedirectURL != "" {
		parsed, err := url.ParseRequestURI(*r.RedirectURL)
		if err != nil || parsed.Scheme == "" || parsed.Host == "" {
			return errors.New("redirect_url must be a valid URL")
		}
	}
	// redirect_urls is the additional-callback allowlist. nil = leave unchanged,
	// an empty slice clears it, and a non-nil slice is validated/deduped/capped.
	if r.RedirectURLs != nil {
		if _, err := ValidateRedirectURLList(*r.RedirectURLs); err != nil {
			return err
		}
	}
	if r.Icon != nil && !roleConstants.IsValidIcon(*r.Icon) {
		return errors.New("icon is not a known icon key")
	}
	if r.Color != nil && !constants.IsValidColor(*r.Color) {
		return errors.New("color is not a known palette key")
	}
	return nil
}
