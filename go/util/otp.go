package util

import (
	"context"
	"crypto/rand"
	"math/big"
	"strings"
	"time"

	"github.com/smap-hcmut/shared-libs/go/tracing"
)

const (
	// OTPLength is the standard OTP length
	OTPLength = 6
	// OTPExpiry is the standard OTP expiry duration
	OTPExpiry = time.Hour * 24
)

// GenerateOTP generates a 6-digit OTP with expiry time and trace integration
func GenerateOTP() (string, time.Time) {
	var builder strings.Builder
	builder.Grow(OTPLength)

	for i := 0; i < OTPLength; i++ {
		n, _ := rand.Int(rand.Reader, big.NewInt(10))
		builder.WriteString(n.String())
	}

	return builder.String(), time.Now().Add(OTPExpiry)
}

// GenerateOTPWithTrace generates OTP with trace context
func GenerateOTPWithTrace(ctx context.Context) (string, time.Time) {
	tracer := tracing.NewTraceContext()
	if traceID := tracer.GetTraceID(ctx); traceID != "" {
		// Could add trace logging here if needed
	}
	return GenerateOTP()
}
