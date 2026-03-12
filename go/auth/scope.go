package auth

import (
	"encoding/base64"
	"encoding/json"
)

// CreateScopeHeader encodes scope as base64 JSON header value.
// This function is compatible with existing service implementations.
func CreateScopeHeader(scope Scope) (string, error) {
	jsonData, err := json.Marshal(scope)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(jsonData), nil
}

// ParseScopeHeader decodes scope from base64 JSON header value.
// This function is compatible with existing service implementations.
func ParseScopeHeader(scopeHeader string) (Scope, error) {
	jsonData, err := base64.StdEncoding.DecodeString(scopeHeader)
	if err != nil {
		return Scope{}, err
	}
	var scope Scope
	if err := json.Unmarshal(jsonData, &scope); err != nil {
		return Scope{}, err
	}
	return scope, nil
}
