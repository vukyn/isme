package services

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/vukyn/isme/external/auth/constants"
	"github.com/vukyn/isme/external/auth/models"
	pkgBase "github.com/vukyn/kuery/http/base"
	pkgErr "github.com/vukyn/kuery/http/errors"

	"github.com/go-resty/resty/v2"
	"github.com/vukyn/kuery/log"
)

type service struct {
	endpoint string
}

func NewService(endpoint string) IService {
	return &service{
		endpoint: endpoint,
	}
}

func (s *service) rest(_ context.Context, retry int, retryInterval, timeout time.Duration) *resty.Client {
	return resty.New().
		SetRetryCount(retry).
		SetRetryWaitTime(retryInterval).
		SetTimeout(map[bool]time.Duration{true: timeout, false: constants.DEFAULT_TIMEOUT}[timeout > 0]).
		SetBaseURL(s.endpoint)
}

//lint:ignore U1000 For debugging purpose
func (s *service) restWithDebug(ctx context.Context, retry int, retryInterval, timeout time.Duration) *resty.Client {
	return s.rest(ctx, retry, retryInterval, timeout).
		SetDebug(true).
		SetLogger(log.New()).
		EnableGenerateCurlOnDebug() // Enable this to generate curl command on debug
}

func (s *service) GetMe(ctx context.Context, req *models.GetMeRequest) (*models.GetMeResponse, error) {
	var client *resty.Client
	if req.Debug {
		client = s.restWithDebug(ctx, req.Retry, req.RetryInterval, req.Timeout)
	} else {
		client = s.rest(ctx, req.Retry, req.RetryInterval, req.Timeout)
	}
	client.SetHeader("Content-Type", "application/json")
	client.SetHeader("Authorization", fmt.Sprintf("Bearer %s", req.AccessToken))

	apiResponse := &models.GetMeResponse{}
	resp, err := client.R().
		SetResult(apiResponse).
		Get(constants.API_AUTH_GET_ME)

	if err != nil {
		log.New().Errorf("Error get me from external auth: %v", err)
		return nil, pkgErr.InternalServerError(err.Error())
	}

	if resp.StatusCode() != http.StatusOK {
		log.New().Errorf("Error get me from external auth: %v", resp.String())
		return nil, handleResponseError(resp)
	}

	return apiResponse, nil
}

func (s *service) RequestLogin(ctx context.Context, req *models.RequestLoginRequest) (*models.RequestLoginResponse, error) {
	var client *resty.Client
	if req.Debug {
		client = s.restWithDebug(ctx, req.Retry, req.RetryInterval, req.Timeout)
	} else {
		client = s.rest(ctx, req.Retry, req.RetryInterval, req.Timeout)
	}
	client.SetHeader("Content-Type", "application/json")

	apiResponse := &models.RequestLoginResponse{}
	resp, err := client.R().
		SetBody(req).
		SetResult(apiResponse).
		Post(constants.API_AUTH_REQUEST_LOGIN)

	if err != nil {
		log.New().Errorf("Error request login from external auth: %v", err)
		return nil, pkgErr.InternalServerError(err.Error())
	}

	if resp.StatusCode() != http.StatusOK {
		log.New().Errorf("Error request login from external auth: %v", resp.String())
		return nil, handleResponseError(resp)
	}

	return apiResponse, nil
}

func (s *service) ExchangeCode(ctx context.Context, req *models.ExchangeCodeRequest) (*models.ExchangeCodeResponse, error) {
	var client *resty.Client
	if req.Debug {
		client = s.restWithDebug(ctx, req.Retry, req.RetryInterval, req.Timeout)
	} else {
		client = s.rest(ctx, req.Retry, req.RetryInterval, req.Timeout)
	}
	client.SetHeader("Content-Type", "application/json")

	apiResponse := &models.ExchangeCodeResponse{}
	resp, err := client.R().
		SetBody(req).
		SetResult(apiResponse).
		Post(constants.API_AUTH_EXCHANGE_CODE)

	if err != nil {
		log.New().Errorf("Error exchange code from external auth: %v", err)
		return nil, pkgErr.InternalServerError(err.Error())
	}

	if resp.StatusCode() != http.StatusOK {
		log.New().Errorf("Error exchange code from external auth: %v", resp.String())
		return nil, handleResponseError(resp)
	}

	return apiResponse, nil
}

func handleResponseError(resp *resty.Response) error {
	var baseErr pkgBase.Response
	err := json.Unmarshal(resp.Body(), &baseErr)
	if err != nil {
		return pkgErr.InternalServerError(resp.String())
	}
	return pkgErr.Forward(baseErr)
}
