package services

import (
	"context"
	"net/http"
	"time"

	"github.com/vukyn/isme/external/auth/constants"
	"github.com/vukyn/isme/external/auth/models"
	pkgErr "github.com/vukyn/isme/pkg/http/errors"

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
		EnableTrace()
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
		return nil, pkgErr.InternalServerError(resp.String())
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
		return nil, pkgErr.InternalServerError(resp.String())
	}

	return apiResponse, nil
}
