package auth

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"sync"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/smap/shared-libs/go/tracing"
)

// Verifier verifies JWT tokens using public keys from JWKS endpoint with trace integration
type Verifier struct {
	jwksEndpoint string
	issuer       string
	audience     []string
	cacheTTL     time.Duration
	tracer       tracing.TraceContext

	// Public key cache
	publicKeys map[string]*rsa.PublicKey
	keysMutex  sync.RWMutex
	lastFetch  time.Time

	// HTTP client for fetching JWKS
	httpClient *http.Client
}

// VerifierConfig holds configuration for JWT verifier
type VerifierConfig struct {
	JWKSEndpoint string
	Issuer       string
	Audience     []string
	CacheTTL     time.Duration
	Tracer       tracing.TraceContext // Optional
}

// JWKS represents JSON Web Key Set
type JWKS struct {
	Keys []JWK `json:"keys"`
}

// JWK represents a JSON Web Key
type JWK struct {
	Kid string `json:"kid"`
	Kty string `json:"kty"`
	Use string `json:"use"`
	N   string `json:"n"`
	E   string `json:"e"`
}

// NewVerifier creates a new JWT verifier with trace integration
func NewVerifier(cfg VerifierConfig) (*Verifier, error) {
	if cfg.JWKSEndpoint == "" {
		return nil, fmt.Errorf("JWKS endpoint is required")
	}
	if cfg.Issuer == "" {
		return nil, fmt.Errorf("issuer is required")
	}
	if len(cfg.Audience) == 0 {
		return nil, fmt.Errorf("audience is required")
	}
	if cfg.CacheTTL == 0 {
		cfg.CacheTTL = 1 * time.Hour
	}
	if cfg.Tracer == nil {
		cfg.Tracer = tracing.NewTraceContext()
	}

	v := &Verifier{
		jwksEndpoint: cfg.JWKSEndpoint,
		issuer:       cfg.Issuer,
		audience:     cfg.Audience,
		cacheTTL:     cfg.CacheTTL,
		tracer:       cfg.Tracer,
		publicKeys:   make(map[string]*rsa.PublicKey),
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}

	// Fetch public keys on initialization
	if err := v.fetchPublicKeys(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to fetch initial public keys: %w", err)
	}

	// Start background refresh
	go v.backgroundRefresh()

	return v, nil
}

// VerifyToken verifies a JWT token and returns payload with trace integration
func (v *Verifier) VerifyToken(tokenString string) (Payload, error) {
	return v.VerifyTokenWithTrace(context.Background(), tokenString)
}

// VerifyTokenWithTrace verifies JWT token with trace context integration
func (v *Verifier) VerifyTokenWithTrace(ctx context.Context, tokenString string) (Payload, error) {
	// Parse token without verification first to get kid
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		// Get kid from header
		kid, ok := token.Header["kid"].(string)
		if !ok {
			return nil, fmt.Errorf("kid not found in token header")
		}

		// Get public key
		publicKey, err := v.getPublicKey(kid)
		if err != nil {
			return nil, err
		}

		return publicKey, nil
	})

	if err != nil {
		return Payload{}, fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return Payload{}, fmt.Errorf("token is invalid")
	}

	// Extract payload
	payload, err := v.extractPayload(token)
	if err != nil {
		return Payload{}, fmt.Errorf("failed to extract payload: %w", err)
	}

	// Validate payload
	if err := v.validatePayload(payload); err != nil {
		return Payload{}, fmt.Errorf("invalid payload: %w", err)
	}

	return payload, nil
}

// extractPayload extracts payload from JWT token
func (v *Verifier) extractPayload(token *jwt.Token) (Payload, error) {
	mapClaims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return Payload{}, fmt.Errorf("invalid claims type")
	}

	return payloadFromMapClaims(mapClaims), nil
}

// validatePayload validates JWT payload
func (v *Verifier) validatePayload(payload Payload) error {
	// Check expiration
	if payload.ExpiresAt > 0 && time.Now().Unix() > payload.ExpiresAt {
		return fmt.Errorf("token is expired")
	}

	// Check issuer
	if payload.Issuer != v.issuer {
		return fmt.Errorf("invalid issuer: expected %s, got %s", v.issuer, payload.Issuer)
	}

	// Check audience
	validAudience := false
	for _, expectedAud := range v.audience {
		if payload.Audience == expectedAud {
			validAudience = true
			break
		}
	}
	if !validAudience {
		return fmt.Errorf("invalid audience")
	}

	return nil
}

// getPublicKey retrieves public key by kid
func (v *Verifier) getPublicKey(kid string) (*rsa.PublicKey, error) {
	v.keysMutex.RLock()

	// Check if cache is expired
	if time.Since(v.lastFetch) > v.cacheTTL {
		v.keysMutex.RUnlock()
		// Refresh keys
		if err := v.fetchPublicKeys(context.Background()); err != nil {
			return nil, fmt.Errorf("failed to refresh public keys: %w", err)
		}
		v.keysMutex.RLock()
	}

	publicKey, ok := v.publicKeys[kid]
	v.keysMutex.RUnlock()

	if !ok {
		return nil, fmt.Errorf("public key not found for kid: %s", kid)
	}

	return publicKey, nil
}

// fetchPublicKeys fetches public keys from JWKS endpoint
func (v *Verifier) fetchPublicKeys(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "GET", v.jwksEndpoint, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := v.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to fetch JWKS: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("JWKS endpoint returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	var jwks JWKS
	if err := json.Unmarshal(body, &jwks); err != nil {
		return fmt.Errorf("failed to parse JWKS: %w", err)
	}

	// Parse and store public keys
	newKeys := make(map[string]*rsa.PublicKey)
	for _, key := range jwks.Keys {
		if key.Kty != "RSA" {
			continue
		}

		publicKey, err := parseRSAPublicKey(key.N, key.E)
		if err != nil {
			return fmt.Errorf("failed to parse public key for kid %s: %w", key.Kid, err)
		}

		newKeys[key.Kid] = publicKey
	}

	// Update cache
	v.keysMutex.Lock()
	v.publicKeys = newKeys
	v.lastFetch = time.Now()
	v.keysMutex.Unlock()

	return nil
}

// backgroundRefresh periodically refreshes public keys
func (v *Verifier) backgroundRefresh() {
	ticker := time.NewTicker(v.cacheTTL / 2) // Refresh at half TTL
	defer ticker.Stop()

	for range ticker.C {
		if err := v.fetchPublicKeys(context.Background()); err != nil {
			// Log error but continue (would use shared logger in real implementation)
			_ = err
		}
	}
}

// parseRSAPublicKey parses RSA public key from JWK n and e values
func parseRSAPublicKey(nStr, eStr string) (*rsa.PublicKey, error) {
	// Decode base64url encoded n and e
	n, err := base64.RawURLEncoding.DecodeString(nStr)
	if err != nil {
		return nil, fmt.Errorf("failed to decode n: %w", err)
	}

	e, err := base64.RawURLEncoding.DecodeString(eStr)
	if err != nil {
		return nil, fmt.Errorf("failed to decode e: %w", err)
	}

	// Convert to big.Int
	nInt := new(big.Int)
	nInt.SetBytes(n)

	eInt := 0
	for _, b := range e {
		eInt = eInt*256 + int(b)
	}

	return &rsa.PublicKey{
		N: nInt,
		E: eInt,
	}, nil
}
