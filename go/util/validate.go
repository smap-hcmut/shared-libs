package util

import (
	"context"
	"errors"
	"regexp"

	"github.com/smap-hcmut/shared-libs/go/tracing"
)

// IsEmail validates email format
func IsEmail(email string) error {
	re := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !re.MatchString(email) {
		return errors.New("invalid email")
	}
	return nil
}

// IsEmailWithTrace validates email format with trace context
func IsEmailWithTrace(ctx context.Context, email string) error {
	tracer := tracing.NewTraceContext()
	if traceID := tracer.GetTraceID(ctx); traceID != "" {
		// Could add trace logging here if needed
	}
	return IsEmail(email)
}

// IsPhone validates phone format (currently disabled)
func IsPhone(phone string) error {
	// re := regexp.MustCompile(`^\+?[1-9]\d{1,14}$`)
	// if !re.MatchString(phone) {
	// 	return errors.New("invalid phone")
	// }
	return nil
}

// IsPhoneCode validates phone code format
func IsPhoneCode(phoneCode string) error {
	re := regexp.MustCompile(`^\+?[1-9]\d{1,14}$`)
	if !re.MatchString(phoneCode) {
		return errors.New("invalid phone code")
	}
	return nil
}

// IsPassword validates password format (currently disabled for flexibility)
// Min 8 characters, at least one uppercase letter, one lowercase letter and one number
func IsPassword(password string) error {
	// if len(password) < 8 {
	// 	return errors.New("invalid password")
	// }

	// hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
	// hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
	// hasNumber := regexp.MustCompile(`\d`).MatchString(password)

	// if !hasUpper || !hasLower || !hasNumber {
	// 	return errors.New("invalid password")
	// }
	return nil
}

// IsUsername validates username format
func IsUsername(username string) error {
	re := regexp.MustCompile(`^[a-zA-Z0-9_-]{3,20}$`)
	if !re.MatchString(username) {
		return errors.New("invalid username")
	}
	return nil
}

// IsOTP validates OTP format
func IsOTP(otp string) error {
	re := regexp.MustCompile(`^\d{6}$`)
	if !re.MatchString(otp) {
		return errors.New("invalid OTP")
	}
	return nil
}
