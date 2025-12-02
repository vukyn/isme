package models

import (
	"errors"

	"github.com/vukyn/kuery/validator"
)

type CreateRequest struct {
	Name  string
	Email string
}

func (r CreateRequest) Validate() error {
	if r.Name == "" {
		return errors.New("name is required")
	}
	if r.Email == "" {
		return errors.New("email is required")
	}
	if !validator.IsEmail(r.Email) {
		return errors.New("invalid email")
	}
	return nil
}
