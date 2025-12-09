package constants

import "time"

const (
	API_AUTH_REQUEST_LOGIN = "/request-login" // POST
	API_AUTH_EXCHANGE_CODE = "/exchange-code" // POST
	DEFAULT_TIMEOUT        = 30 * time.Second
)
