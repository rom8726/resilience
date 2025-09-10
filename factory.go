package resilience

import (
	"context"
	"time"
)

// NewDefaultCircuitBreaker creates a new circuit breaker with default settings.
func NewDefaultCircuitBreaker(name string) CircuitBreaker {
	return NewCircuitBreaker(Config{
		Name:                     name,
		ErrorThreshold:           0.5, // 50% error threshold
		MinRequests:              10,
		OpenTimeout:              30 * time.Second,
		HalfOpenSuccessThreshold: 2,
		MaxHalfOpenRequests:      5,
		Timeout:                  5 * time.Second,
		IgnoredErrors:            []error{context.Canceled, context.DeadlineExceeded},
	})
}
