package models

import (
	"github.com/vukyn/isme/external/models"
)

type GetMeRequest struct {
	models.ApiRequest
	AccessToken string
}

type GetMeResponse struct {
	Code    int32  `json:"code"`
	Message string `json:"message"`
	Data    struct {
		ID    string `json:"id"`
		Name  string `json:"name"`
		Email string `json:"email"`
	} `json:"data"`
}

type RequestLoginRequest struct {
	models.ApiRequest
	AppCode   string `json:"app_code"`
	AppSecret string `json:"app_secret"`
	CtxInfo   string `json:"ctx_info"`
}

type RequestLoginResponse struct {
	Code    int32  `json:"code"`
	Message string `json:"message"`
	Data    struct {
		RedirectURL string `json:"redirect_url"`
	} `json:"data"`
}

type RefreshTokenRequest struct {
	models.ApiRequest
	RefreshToken string `json:"refresh_token"`
}

type RefreshTokenResponse struct {
	Code    int32  `json:"code"`
	Message string `json:"message"`
	Data    struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresAt    string `json:"expires_at"`
	} `json:"data"`
}

type ExchangeCodeRequest struct {
	models.ApiRequest
	AuthorizationCode string `json:"authorization_code"`
}

type ExchangeCodeResponse struct {
	Code    int32  `json:"code"`
	Message string `json:"message"`
	Data    struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresAt    string `json:"expires_at"`
	} `json:"data"`
}

type LogoutRequest struct {
	models.ApiRequest
	AccessToken string
}

type LogoutResponse struct {
	Code    int32  `json:"code"`
	Message string `json:"message"`
}
