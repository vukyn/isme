package models

import (
	"errors"

	pkgClaims "github.com/vukyn/kuery/claims"

	"github.com/vukyn/kuery/validator"
)

type GetMeResponse struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
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

// LoginResponse carries the result of a login.
//
//   - First-party isme login: AccessToken/RefreshToken/ExpiresAt are the
//     full-scope tokens; AuthorizationCode/RedirectURL are empty.
//   - SSO login (session_id set): AccessToken/RefreshToken/ExpiresAt carry the
//     full-scope IdP tokens (the browser writes these as isme cookies so silent
//     SSO can trigger later), while AuthorizationCode carries the app handoff —
//     a one-time code that ExchangeCode resolves to the APP-scoped, aud-restricted
//     tokens. The two token pairs are intentionally distinct and never crossed.
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

// SSOCheckRequest is the read-only silent-authorize probe. It resolves the SSO
// session and validates the caller's existing isme tokens WITHOUT rotating them.
// At least one of AccessToken / RefreshToken must be present (the access token
// may be expired, in which case the refresh token is probed instead).
type SSOCheckRequest struct {
	SessionID    string `json:"session_id"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func (r SSOCheckRequest) Validate() error {
	if r.SessionID == "" {
		return errors.New("session_id is required")
	}
	if r.AccessToken == "" && r.RefreshToken == "" {
		return errors.New("access_token or refresh_token is required")
	}
	return nil
}

// SSOCheckUser is the minimal identity shown on the consent screen. No avatar
// field exists on the user entity — the frontend derives initials from Name.
type SSOCheckUser struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// SSOCheckApp identifies the requesting app resolved from the session_id.
type SSOCheckApp struct {
	Name        string `json:"name"`
	RedirectURL string `json:"redirect_url"`
}

// SSOScope is one consent line item rendered on the consent screen.
type SSOScope struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

// SSOCheckResponse drives the consent screen. When Valid is false the frontend
// falls back to the password form; the probe never errors on an invalid/expired
// session so the page can degrade gracefully.
type SSOCheckResponse struct {
	Valid  bool         `json:"valid"`
	User   SSOCheckUser `json:"user"`
	App    SSOCheckApp  `json:"app"`
	Scopes []SSOScope   `json:"scopes"`
	// Nonce is a short-TTL single-use CSRF token that /sso/consent requires.
	Nonce string `json:"nonce"`
}

// SSOConsentRequest is the authorize step. It re-validates server-side and
// requires the single-use Nonce minted by /sso/check.
type SSOConsentRequest struct {
	SessionID    string `json:"session_id"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	Nonce        string `json:"nonce"`
}

func (r SSOConsentRequest) Validate() error {
	if r.SessionID == "" {
		return errors.New("session_id is required")
	}
	if r.Nonce == "" {
		return errors.New("nonce is required")
	}
	if r.AccessToken == "" && r.RefreshToken == "" {
		return errors.New("access_token or refresh_token is required")
	}
	return nil
}

// SSOConsentResponse mirrors the SSO fields of LoginResponse.
type SSOConsentResponse struct {
	RedirectURL       string `json:"redirect_url"`
	AuthorizationCode string `json:"authorization_code"`
}
