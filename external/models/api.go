package models

import "time"

type ApiRequest struct {
	Retry         int           `json:"-"`
	RetryInterval time.Duration `json:"-"`
	Timeout       time.Duration `json:"-"`
}
