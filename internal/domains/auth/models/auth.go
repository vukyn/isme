package models

import (
	"errors"

	pkgClaims "github.com/vukyn/isme/pkg/claims"

	"github.com/vukyn/kuery/validator"
)

type GetMeResponse struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Email   string `json:"email"`
	IsAdmin bool   `json:"is_admin,omitempty"`
}

type SignUpRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (r SignUpRequest) Validate() error {
	if r.Name == "" {
		return errors.New("name is required")
	}
	if r.Email == "" {
		return errors.New("email is required")
	}
	if !validator.IsEmail(r.Email) {
		return errors.New("invalid email")
	}
	if r.Password == "" {
		return errors.New("password is required")
	}
	return nil
}

type SignUpResponse struct {
	ID string `json:"id"`
}

type LoginRequest struct {
	Email     string `json:"email"`
	Password  string `json:"password"`
	SessionID string `json:"session_id"`
}

func (r LoginRequest) Validate() error {
	if r.Email == "" {
		return errors.New("invalid email or password")
	}
	if !validator.IsEmail(r.Email) {
		return errors.New("invalid email or password")
	}
	if r.Password == "" {
		return errors.New("invalid email or password")
	}
	return nil
}

type LoginResponse struct {
	AccessToken       string `json:"access_token"`
	RefreshToken      string `json:"refresh_token"`
	ExpiresAt         string `json:"expires_at"`
	RedirectURL       string `json:"redirect_url,omitempty"`
	AuthorizationCode string `json:"authorization_code,omitempty"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}

func (r RefreshTokenRequest) Validate() error {
	if r.RefreshToken == "" {
		return errors.New("refresh token is required")
	}
	return nil
}

type RefreshTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresAt    string `json:"expires_at"`
}

type VerifyTokenRequest struct {
	Token string `json:"token"`
}

func (r VerifyTokenRequest) Validate() error {
	if r.Token == "" {
		return errors.New("token is required")
	}
	return nil
}

type VerifyTokenResponse struct {
	Ok     bool             `json:"ok"`
	Claims pkgClaims.Claims `json:"claims"`
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

func (r ChangePasswordRequest) Validate() error {
	if r.OldPassword == "" {
		return errors.New("old password is required")
	}
	if r.NewPassword == "" {
		return errors.New("new password is required")
	}
	if len(r.NewPassword) < 6 {
		return errors.New("new password must be at least 6 characters")
	}
	return nil
}

type RequestLoginRequest struct {
	AppCode   string `json:"app_code"`
	AppSecret string `json:"app_secret"`
	CtxInfo   string `json:"ctx_info"`
}

func (r RequestLoginRequest) Validate() error {
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

type RequestLoginResponse struct {
	RedirectURL string `json:"redirect_url"`
}

type ExchangeCodeRequest struct {
	AuthorizationCode string `json:"authorization_code"`
}

func (r ExchangeCodeRequest) Validate() error {
	if r.AuthorizationCode == "" {
		return errors.New("authorization_code is required")
	}
	return nil
}

type ExchangeCodeResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresAt    string `json:"expires_at"`
}
