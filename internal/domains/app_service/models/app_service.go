package models

import (
	"errors"
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
