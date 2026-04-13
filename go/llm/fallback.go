package llm

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// circuitState tracks a tripped (quota-exhausted) provider.
type circuitState struct {
	trippedAt time.Time
}

// FallbackProvider wraps multiple LLM providers and distributes requests
// round-robin with automatic fallback on errors. Providers with exhausted
// quotas are circuit-broken for 1 hour.
type FallbackProvider struct {
	providers []LLM
	counter   atomic.Uint64

	mu       sync.RWMutex
	circuits map[string]*circuitState // keyed by provider.Name()

	circuitTTL time.Duration
}

// NewFallback creates a FallbackProvider from the given providers.
// At least one provider is required.
func NewFallback(providers []LLM) (*FallbackProvider, error) {
	if len(providers) == 0 {
		return nil, fmt.Errorf("llm: at least one provider is required")
	}
	return &FallbackProvider{
		providers:  providers,
		circuits:   make(map[string]*circuitState),
		circuitTTL: 1 * time.Hour,
	}, nil
}

func (f *FallbackProvider) Name() string {
	names := make([]string, len(f.providers))
	for i, p := range f.providers {
		names[i] = p.Name()
	}
	return fmt.Sprintf("fallback[%s]", strings.Join(names, ","))
}

// Generate tries providers in round-robin order, falling back to the next
// provider on retryable errors. Quota-exhausted providers are skipped via
// circuit breaker (1h cooldown).
func (f *FallbackProvider) Generate(ctx context.Context, prompt string) (string, error) {
	n := len(f.providers)
	start := int(f.counter.Add(1) - 1) // round-robin offset

	var errs []string
	tried := 0

	for i := 0; i < n; i++ {
		idx := (start + i) % n
		p := f.providers[idx]

		// Skip circuit-broken providers
		if f.isCircuitOpen(p.Name()) {
			continue
		}

		tried++
		text, err := p.Generate(ctx, prompt)
		if err == nil {
			return text, nil
		}

		errs = append(errs, fmt.Sprintf("%s: %s", p.Name(), err.Error()))

		// Trip circuit breaker on quota exhaustion
		if isQuotaError(err) {
			f.tripCircuit(p.Name())
			continue
		}

		// Non-retryable → don't bother trying other providers with same class of error
		if !isRetryable(err) {
			// But DO try other providers — they may work fine
			continue
		}

		// Context cancelled → stop immediately
		if ctx.Err() != nil {
			return "", fmt.Errorf("context cancelled after trying %d providers: %s", tried, strings.Join(errs, "; "))
		}
	}

	if tried == 0 {
		// All providers are circuit-broken; reset circuits and try first provider
		f.resetCircuits()
		p := f.providers[start%n]
		return p.Generate(ctx, prompt)
	}

	return "", fmt.Errorf("llm: all %d providers failed: %s", tried, strings.Join(errs, "; "))
}

// ── Circuit breaker ─────────────────────────────────────────────────────────

func (f *FallbackProvider) isCircuitOpen(name string) bool {
	f.mu.RLock()
	defer f.mu.RUnlock()
	cs, ok := f.circuits[name]
	if !ok {
		return false
	}
	if time.Since(cs.trippedAt) > f.circuitTTL {
		return false // expired, will be cleaned up lazily
	}
	return true
}

func (f *FallbackProvider) tripCircuit(name string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.circuits[name] = &circuitState{trippedAt: time.Now()}
}

func (f *FallbackProvider) resetCircuits() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.circuits = make(map[string]*circuitState)
}

// ActiveProviders returns the names of providers not currently circuit-broken.
func (f *FallbackProvider) ActiveProviders() []string {
	f.mu.RLock()
	defer f.mu.RUnlock()
	var active []string
	for _, p := range f.providers {
		cs, ok := f.circuits[p.Name()]
		if !ok || time.Since(cs.trippedAt) > f.circuitTTL {
			active = append(active, p.Name())
		}
	}
	return active
}
