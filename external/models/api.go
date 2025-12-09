package models

import "time"

type ApiRequest struct {
	Debug         bool          `json:"-"`
	Retry         int           `json:"-"`
	RetryInterval time.Duration `json:"-"`
	Timeout       time.Duration `json:"-"`
}
