package models

import (
	"errors"
	"time"
)

type CreateRequest struct {
	UserID       string
	TokenID      string
	Email        string
	RefreshToken string
	ExpiresAt    time.Time
	ClientIP     string
	UserAgent    string
	// AppServiceID records the requesting app for SSO logins (empty for
	// first-party isme logins). Used at refresh time to decide whether the
	// new token stays aud-restricted to that app or spans all apps.
	AppServiceID string
}

func (r CreateRequest) Validate() error {
	if r.UserID == "" {
		return errors.New("user_id is required")
	}
	if r.TokenID == "" {
		return errors.New("token_id is required")
	}
	if r.Email == "" {
		return errors.New("email is required")
	}
	if r.RefreshToken == "" {
		return errors.New("refresh_token is required")
	}
	if r.ExpiresAt.IsZero() {
		return errors.New("expires_at is required")
	}
	return nil
}

type UpdateLastLoginRequest struct {
	ID      string
	// UserID owns the session being rotated; required so the matching
	// token_rotation_events row can be attributed to the user for the 24h count.
	UserID       string
	TokenID      string
	RefreshToken string
	ClientIP     string
	UserAgent    string
	ExpiresAt    time.Time
}

func (r UpdateLastLoginRequest) Validate() error {
	if r.ID == "" {
		return errors.New("id is required")
	}
	if r.UserID == "" {
		return errors.New("user_id is required")
	}
	if r.TokenID == "" {
		return errors.New("token_id is required")
	}
	if r.RefreshToken == "" {
		return errors.New("refresh_token is required")
	}
	if r.ExpiresAt.IsZero() {
		return errors.New("expires_at is required")
	}
	return nil
}
