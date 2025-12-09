package models

import (
	"github.com/vukyn/isme/external/models"
)

type RequestLoginRequest struct {
	models.ApiRequest
	AppCode   string
	AppSecret string
	CtxInfo   string
}

type RequestLoginResponse struct {
	Code    int32  `json:"code"`
	Message string `json:"message"`
	Data    struct {
		RedirectURL string `json:"redirect_url"`
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
