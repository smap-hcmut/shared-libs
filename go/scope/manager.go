package scope

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/smap-hcmut/shared-libs/go/tracing"
)

// implManager implements Manager with trace integration
type implManager struct {
	secretKey string
	tracer    tracing.TraceContext
}

// New creates a new scope Manager with the provided secret key
func New(secretKey string) Manager {
	if secretKey == "" {
		panic("scope: secret key cannot be empty")
	}
	return &implManager{
		secretKey: secretKey,
		tracer:    tracing.NewTraceContext(),
	}
}

// NewWithTracer creates a new scope Manager with custom tracer
func NewWithTracer(secretKey string, tracer tracing.TraceContext) Manager {
	if secretKey == "" {
		panic("scope: secret key cannot be empty")
	}
	if tracer == nil {
		tracer = tracing.NewTraceContext()
	}
	return &implManager{
		secretKey: secretKey,
		tracer:    tracer,
	}
}

// Verify verifies the JWT token and returns the payload if valid
func (m *implManager) Verify(token string) (Payload, error) {
	return m.VerifyWithTrace(context.Background(), token)
}

// VerifyWithTrace verifies the JWT token with trace context
func (m *implManager) VerifyWithTrace(ctx context.Context, token string) (Payload, error) {
	if token == "" {
		return Payload{}, fmt.Errorf("%w: token is empty", ErrInvalidToken)
	}

	keyFunc := func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("%w: unexpected signing method: %v", ErrInvalidToken, t.Header["alg"])
		}
		return []byte(m.secretKey), nil
	}

	jwtToken, err := jwt.Parse(token, keyFunc)
	if err != nil {
		return Payload{}, fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}

	if !jwtToken.Valid {
		return Payload{}, fmt.Errorf("%w: token is not valid", ErrInvalidToken)
	}

	mapClaims, ok := jwtToken.Claims.(jwt.MapClaims)
	if !ok {
		return Payload{}, fmt.Errorf("%w: failed to parse claims", ErrInvalidToken)
	}

	return payloadFromMapClaims(mapClaims), nil
}

// CreateToken creates a new JWT token with the provided payload
func (m *implManager) CreateToken(payload Payload) (string, error) {
	return m.CreateTokenWithTrace(context.Background(), payload)
}

// CreateTokenWithTrace creates a new JWT token with trace context
func (m *implManager) CreateTokenWithTrace(ctx context.Context, payload Payload) (string, error) {
	now := time.Now()
	payload.StandardClaims = jwt.StandardClaims{
		ExpiresAt: now.Add(TokenExpirationDuration).Unix(),
		Id:        fmt.Sprintf("%d", now.UnixNano()),
		NotBefore: now.Unix(),
		IssuedAt:  now.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, payload)
	return token.SignedString([]byte(m.secretKey))
}

// VerifyScope parses and verifies scope header
func (m *implManager) VerifyScope(scopeHeader string) (Scope, error) {
	return m.VerifyScopeWithTrace(context.Background(), scopeHeader)
}

// VerifyScopeWithTrace parses and verifies scope header with trace context
func (m *implManager) VerifyScopeWithTrace(ctx context.Context, scopeHeader string) (Scope, error) {
	return ParseScopeHeader(scopeHeader)
}

// Helper functions for JWT claims parsing
func payloadFromMapClaims(claims jwt.MapClaims) Payload {
	payload := Payload{
		UserID:   getStringClaim(claims, "sub"),
		Username: firstNonEmptyClaim(claims, "username", "email"),
		Role:     getStringClaim(claims, "role"),
		Type:     getStringClaim(claims, "type"),
		Refresh:  getBoolClaim(claims, "refresh"),
	}
	payload.StandardClaims = jwt.StandardClaims{
		Audience:  getAudienceClaim(claims),
		ExpiresAt: getInt64Claim(claims, "exp"),
		Id:        firstNonEmptyClaim(claims, "jti", "id"),
		IssuedAt:  getInt64Claim(claims, "iat"),
		Issuer:    getStringClaim(claims, "iss"),
		NotBefore: getInt64Claim(claims, "nbf"),
		Subject:   getStringClaim(claims, "sub"),
	}
	return payload
}

func getStringClaim(claims jwt.MapClaims, key string) string {
	value, ok := claims[key]
	if !ok || value == nil {
		return ""
	}
	switch v := value.(type) {
	case string:
		return v
	case fmt.Stringer:
		return v.String()
	default:
		return fmt.Sprint(v)
	}
}

func firstNonEmptyClaim(claims jwt.MapClaims, keys ...string) string {
	for _, key := range keys {
		value := getStringClaim(claims, key)
		if value != "" {
			return value
		}
	}
	return ""
}

func getInt64Claim(claims jwt.MapClaims, key string) int64 {
	value, ok := claims[key]
	if !ok || value == nil {
		return 0
	}
	switch v := value.(type) {
	case float64:
		return int64(v)
	case float32:
		return int64(v)
	case int64:
		return v
	case int:
		return int64(v)
	case json.Number:
		n, _ := v.Int64()
		return n
	case string:
		n, err := strconv.ParseInt(v, 10, 64)
		if err == nil {
			return n
		}
	}
	return 0
}

func getBoolClaim(claims jwt.MapClaims, key string) bool {
	value, ok := claims[key]
	if !ok || value == nil {
		return false
	}
	switch v := value.(type) {
	case bool:
		return v
	case string:
		parsed, err := strconv.ParseBool(v)
		return err == nil && parsed
	default:
		return false
	}
}

func getAudienceClaim(claims jwt.MapClaims) string {
	value, ok := claims["aud"]
	if !ok || value == nil {
		return ""
	}
	switch v := value.(type) {
	case string:
		return v
	case []interface{}:
		for _, item := range v {
			if s, ok := item.(string); ok && s != "" {
				return s
			}
		}
	case []string:
		for _, item := range v {
			if item != "" {
				return item
			}
		}
	}
	return ""
}
