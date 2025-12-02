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
	ID           string
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
