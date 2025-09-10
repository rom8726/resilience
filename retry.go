package resilience

import (
	"context"
	"log/slog"
	"time"

	"github.com/avast/retry-go"
)

// RetryableFunc is a function that can be retried.
type RetryableFunc func() error

// RetryableFuncWithContext is a function with context that can be retried.
type RetryableFuncWithContext func(ctx context.Context) error

// WithRetry executes the given function with retry logic.
func WithRetry(fn RetryableFunc, opts ...retry.Option) error {
	return retry.Do(retry.RetryableFunc(fn), opts...)
}

// WithRetryContext executes the given function with retry logic and context.
func WithRetryContext(ctx context.Context, fn RetryableFuncWithContext, opts ...retry.Option) error {
	wrappedFn := func() error {
		// Check if context is done before executing the function
		if ctx.Err() != nil {
			return ctx.Err()
		}

		return fn(ctx)
	}

	// Add context as a retry option
	opts = append(opts, retry.Context(ctx))

	// Add retry attempt tracking
	opts = append(opts, retry.OnRetry(func(n uint, err error) {
		// Increment retry attempts metric
		IncrementRetryAttempts("retry_context")
	}))

	return retry.Do(wrappedFn, opts...)
}

// DefaultRetryOptions returns the default retry options.
func DefaultRetryOptions() []retry.Option {
	return []retry.Option{
		retry.Attempts(3),
		retry.Delay(100 * time.Millisecond),
		retry.DelayType(retry.BackOffDelay),
		retry.OnRetry(func(n uint, err error) {
			slog.Warn("Retry attempt",
				"attempt", n+1,
				"error", err,
			)
		}),
	}
}

// WithCircuitBreakerAndRetry combines circuit breaker and retry patterns.
func WithCircuitBreakerAndRetry(
	ctx context.Context,
	cb CircuitBreaker,
	fn RetryableFuncWithContext,
	retryOpts ...retry.Option,
) error {
	return cb.Execute(ctx, func(ctx context.Context) error {
		return WithRetryContext(ctx, fn, retryOpts...)
	})
}

// WithCircuitBreakerAndRetryWithFallback combines circuit breaker and retry patterns with a fallback.
func WithCircuitBreakerAndRetryWithFallback(
	ctx context.Context,
	cb CircuitBreaker,
	fn RetryableFuncWithContext,
	fallback func(ctx context.Context, err error) error,
	retryOpts ...retry.Option,
) error {
	return cb.ExecuteWithFallback(
		ctx,
		func(ctx context.Context) error {
			return WithRetryContext(ctx, fn, retryOpts...)
		},
		fallback,
	)
}
