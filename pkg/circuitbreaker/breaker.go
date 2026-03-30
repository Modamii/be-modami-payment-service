package circuitbreaker

import (
	"context"
	"time"

	"github.com/sony/gobreaker"
)

// Breaker wraps sony/gobreaker for payment gateway calls.
type Breaker struct {
	cb *gobreaker.CircuitBreaker
}

// Settings for the circuit breaker.
type Settings struct {
	Name        string
	MaxFailures uint32
	Interval    time.Duration
	Timeout     time.Duration
}

// DefaultSettings returns sensible defaults: 5 consecutive failures, 30s window, 60s open timeout.
func DefaultSettings(name string) Settings {
	return Settings{
		Name:        name,
		MaxFailures: 5,
		Interval:    30 * time.Second,
		Timeout:     60 * time.Second,
	}
}

// New creates a new Breaker with the given settings.
func New(s Settings) *Breaker {
	st := gobreaker.Settings{
		Name:        s.Name,
		MaxRequests: 1,
		Interval:    s.Interval,
		Timeout:     s.Timeout,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures >= s.MaxFailures
		},
	}
	return &Breaker{cb: gobreaker.NewCircuitBreaker(st)}
}

// Execute runs fn through the circuit breaker.
func (b *Breaker) Execute(ctx context.Context, fn func() (interface{}, error)) (interface{}, error) {
	return b.cb.Execute(func() (interface{}, error) {
		// Respect context cancellation.
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}
		return fn()
	})
}

// State returns the current state of the breaker.
func (b *Breaker) State() gobreaker.State {
	return b.cb.State()
}
