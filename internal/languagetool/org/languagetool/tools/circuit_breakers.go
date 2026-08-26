package tools

import (
	"sync"
	"time"
)

// CircuitState is the breaker state.
type CircuitState int

const (
	CircuitClosed CircuitState = iota
	CircuitOpen
	CircuitHalfOpen
)

// CircuitBreaker is a minimal port of resilience4j-style breakers used by
// org.languagetool.tools.CircuitBreakers (no metrics backend).
type CircuitBreaker struct {
	mu               sync.Mutex
	name             string
	state            CircuitState
	failureThreshold int
	successThreshold int
	openDuration     time.Duration
	failures         int
	successes        int
	openedAt         time.Time
}

func NewCircuitBreaker(name string, failureThreshold int, openDuration time.Duration) *CircuitBreaker {
	if failureThreshold <= 0 {
		failureThreshold = 5
	}
	if openDuration <= 0 {
		openDuration = 5 * time.Second
	}
	return &CircuitBreaker{
		name:             name,
		state:            CircuitClosed,
		failureThreshold: failureThreshold,
		successThreshold: 1,
		openDuration:     openDuration,
	}
}

func (b *CircuitBreaker) Name() string { return b.name }

func (b *CircuitBreaker) State() CircuitState {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.maybeHalfOpenLocked(time.Now())
	return b.state
}

// Allow reports whether a call may proceed.
func (b *CircuitBreaker) Allow() bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.maybeHalfOpenLocked(time.Now())
	return b.state != CircuitOpen
}

func (b *CircuitBreaker) OnSuccess() {
	b.mu.Lock()
	defer b.mu.Unlock()
	switch b.state {
	case CircuitHalfOpen:
		b.successes++
		if b.successes >= b.successThreshold {
			b.state = CircuitClosed
			b.failures = 0
			b.successes = 0
		}
	case CircuitClosed:
		b.failures = 0
	}
}

func (b *CircuitBreaker) OnFailure() {
	b.mu.Lock()
	defer b.mu.Unlock()
	switch b.state {
	case CircuitHalfOpen:
		b.tripLocked(time.Now())
	case CircuitClosed:
		b.failures++
		if b.failures >= b.failureThreshold {
			b.tripLocked(time.Now())
		}
	}
}

func (b *CircuitBreaker) tripLocked(now time.Time) {
	b.state = CircuitOpen
	b.openedAt = now
	b.successes = 0
	b.failures = 0
}

func (b *CircuitBreaker) maybeHalfOpenLocked(now time.Time) {
	if b.state == CircuitOpen && now.Sub(b.openedAt) >= b.openDuration {
		b.state = CircuitHalfOpen
		b.successes = 0
	}
}

// CircuitBreakers is a named registry (ports tools.CircuitBreakers.registry).
type CircuitBreakers struct {
	mu   sync.Mutex
	byID map[string]*CircuitBreaker
}

var defaultCircuitBreakers = &CircuitBreakers{byID: map[string]*CircuitBreaker{}}

// CircuitBreakerRegistry returns the process-wide registry.
func CircuitBreakerRegistry() *CircuitBreakers { return defaultCircuitBreakers }

// GetOrCreate returns a breaker for name, creating with defaults if needed.
func (r *CircuitBreakers) GetOrCreate(name string) *CircuitBreaker {
	if r == nil {
		return NewCircuitBreaker(name, 5, 5*time.Second)
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.byID == nil {
		r.byID = map[string]*CircuitBreaker{}
	}
	if b, ok := r.byID[name]; ok {
		return b
	}
	b := NewCircuitBreaker(name, 5, 5*time.Second)
	r.byID[name] = b
	return b
}

// ResetCircuitBreakers clears the default registry (tests).
func ResetCircuitBreakers() {
	defaultCircuitBreakers = &CircuitBreakers{byID: map[string]*CircuitBreaker{}}
}
