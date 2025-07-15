package validation

import (
	"errors"
	"net/mail"
)

func ValidateEmail(email string) error {
	if email == "" {
		return errors.New("Email is a required field.")
	}
	
	if _, err := mail.ParseAddress(email); err != nil {
		return errors.New("Invalid email format")
	}
	return nil
}
