package llm

import "fmt"

// MultiConfig holds configuration for multiple LLM providers.
// Each entry maps to a ProviderConfig. Empty API keys are skipped.
type MultiConfig struct {
	Providers []ProviderConfig
}

// NewFromConfig creates a FallbackProvider from MultiConfig.
// Providers with empty API keys are silently skipped.
// Returns an error if no providers could be created.
func NewFromConfig(cfg MultiConfig) (*FallbackProvider, error) {
	var providers []LLM

	for _, pc := range cfg.Providers {
		if pc.APIKey == "" {
			continue // skip unconfigured providers
		}
		p, err := NewProvider(pc)
		if err != nil {
			return nil, fmt.Errorf("llm: create provider %q: %w", pc.Name, err)
		}
		providers = append(providers, p)
	}

	if len(providers) == 0 {
		return nil, fmt.Errorf("llm: no providers configured (all API keys empty)")
	}

	return NewFallback(providers)
}
