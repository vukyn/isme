package models

import (
	"isme/external/models"
)

type RequestSSOLoginRequest struct {
	models.ApiRequest
	AppCode   string
	AppSecret string
	CtxInfo   string
}

type RequestSSOLoginResponse struct {
	Code    int32  `json:"code"`
	Message string `json:"message"`
	Data    struct {
		RedirectURL string `json:"redirect_url"`
	} `json:"data"`
}
