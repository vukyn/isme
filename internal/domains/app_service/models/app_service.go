package models

import (
	"errors"

	"github.com/vukyn/isme/internal/domains/app_service/constants"
)

type RegisterRequest struct {
	AppCode     string `json:"app_code"`
	AppName     string `json:"app_name"`
	RedirectURL string `json:"redirect_url"`
	CtxInfo     string `json:"ctx_info"`
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
	if r.CtxInfo == "" {
		return errors.New("ctx_info is required")
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
	ID             string `json:"id"`
	AppCode        string `json:"app_code"`
	AppName        string `json:"app_name"`
	RedirectURL    string `json:"redirect_url"`
	CtxInfo        string `json:"ctx_info"`
	Status         int32  `json:"status"`
	CreatedAt      string `json:"created_at"`
	CreatedBy      string `json:"created_by"`       // creator user id
	CreatedByEmail string `json:"created_by_email"` // resolved creator email (empty when unresolvable)
	UpdatedAt      string `json:"updated_at"`
	UpdatedBy      string `json:"updated_by"`
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
