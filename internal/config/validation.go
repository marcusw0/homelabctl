package config

import (
	"errors"
	"time"
)

// ValidateExpectedStatus accepts zero as the default HTTP status.
func ValidateExpectedStatus(status int) error {
	if status != 0 && (status < 100 || status > 599) {
		return errors.New("expect_status must be between 100 and 599 (or 0 for the default)")
	}
	return nil
}

// ValidateTimeout accepts zero as the default check timeout.
func ValidateTimeout(timeout time.Duration) error {
	if timeout < 0 {
		return errors.New("timeout must not be negative")
	}
	return nil
}
