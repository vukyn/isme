package constants

import "time"

const (
	API_AUTH_GET_ME        = "/auth/me"            // GET
	API_AUTH_REQUEST_LOGIN = "/auth/request-login" // POST
	API_AUTH_REFRESH_TOKEN = "/auth/refresh" // POST
	API_AUTH_EXCHANGE_CODE = "/auth/exchange-code" // POST
	API_AUTH_LOGOUT        = "/auth/logout"        // POST
	DEFAULT_TIMEOUT        = 30 * time.Second
)
