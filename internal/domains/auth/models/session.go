package models

// MySessionItem is a single active session belonging to the current user.
// It deliberately omits RefreshToken and TokenID so they are never exposed
// over the API.
type MySessionItem struct {
	ID           string `json:"id"`
	ClientIP     string `json:"client_ip"`
	UserAgent    string `json:"user_agent"`
	LastLoginAt  string `json:"last_login_at"`
	ExpiresAt    string `json:"expires_at"`
	AppServiceID string `json:"app_service_id"`
	// AppName/AppIcon/AppColor are enriched from the owning app_service for
	// SSO sessions. They are empty for first-party isme sessions (empty
	// AppServiceID) and for sessions whose app_service has been deleted.
	AppName  string `json:"app_name"`
	AppIcon  string `json:"app_icon"`
	AppColor string `json:"app_color"`
	Current  bool   `json:"current"`
}

// MySessionCount is the active-session summary used by the Welcome stat card.
type MySessionCount struct {
	Count    int `json:"count"`
	NewIn24h int `json:"new_in_24h"`
}
