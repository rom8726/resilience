# Resilience Package

This package provides circuit breaker and retry functionality for external service calls.

## Features

- Circuit Breaker pattern implementation
- Integration with [retry-go](https://github.com/avast/retry-go) for retry functionality
- Metrics for monitoring circuit breaker and retry operations
- **Thread-safe implementation** with proper read/write lock separation

## Thread Safety

The circuit breaker implementation is fully thread-safe and handles race conditions properly:

- **Read operations** (`State()`, `Metrics()`, `Name()`) use read locks (`RLock()`)
- **Write operations** (state transitions) use write locks (`Lock()`)
- **State transitions** in `State()` method use double-check pattern to avoid race conditions
- **Concurrent access** is tested with race detector (`-race` flag)

### Race Condition Prevention

The `State()` method properly handles the transition from `Open` to `HalfOpen` state:

```go
func (cb *circuitBreaker) State() State {
    cb.mutex.RLock()
    defer cb.mutex.RUnlock()

    // Check if we need to transition from Open to HalfOpen
    if cb.state == Open && time.Since(cb.stateChangeTime) > cb.openTimeout {
        // Release read lock and acquire write lock for state transition
        cb.mutex.RUnlock()
        cb.mutex.Lock()
        
        // Double-check pattern: verify state is still Open and timeout has elapsed
        if cb.state == Open && time.Since(cb.stateChangeTime) > cb.openTimeout {
            cb.toHalfOpen()
        }
        
        cb.mutex.Unlock()
        cb.mutex.RLock()
    }

    return cb.state
}
```

## Usage

### Circuit Breaker

```go
// Create a circuit breaker
cb := resilience.NewDefaultCircuitBreaker("test")

// Execute a function with circuit breaker protection
err := cb.Execute(ctx, func(ctx context.Context) error {
    // Your code here
    return nil
})

// Execute a function with circuit breaker protection and fallback
err := cb.ExecuteWithFallback(
    ctx,
    func(ctx context.Context) error {
        // Your code here
        return nil
    },
    func(ctx context.Context, err error) error {
        // Fallback code here
        return nil
    },
)
```

### Retry

```go
// Execute a function with retry
err := resilience.WithRetry(func() error {
    // Your code here
    return nil
}, resilience.DefaultRetryOptions()...)

// Execute a function with retry and context
err := resilience.WithRetryContext(ctx, func(ctx context.Context) error {
    // Your code here
    return nil
}, resilience.KafkaRetryOptions()...)
```

### Combined Circuit Breaker and Retry

```go
// Execute a function with circuit breaker and retry
err := resilience.WithCircuitBreakerAndRetry(
    ctx,
    cb,
    func(ctx context.Context) error {
        // Your code here
        return nil
    },
    resilience.KafkaRetryOptions()...,
)

// Execute a function with circuit breaker, retry, and fallback
err := resilience.WithCircuitBreakerAndRetryWithFallback(
    ctx,
    cb,
    func(ctx context.Context) error {
        // Your code here
        return nil
    },
    func(ctx context.Context, err error) error {
        // Fallback code here
        return nil
    },
    resilience.KafkaRetryOptions()...,
)
```

## Pre-configured Circuit Breakers

The package provides factory methods for creating pre-configured circuit breakers:

- `NewDefaultCircuitBreaker(name string)`: Circuit breaker with default settings

## Pre-configured Retry Options

The package provides factory methods for creating pre-configured retry options:

- `DefaultRetryOptions()`: Default retry options (3 attempts, 100ms delay)

## Metrics

The package provides Prometheus metrics for monitoring circuit breaker and retry operations:

- `warden_circuit_breaker_state`: Current state of the circuit breaker (0 - Closed, 1 - Open, 2 - HalfOpen)
- `warden_circuit_breaker_failures_total`: Total number of circuit breaker failures
- `warden_circuit_breaker_successes_total`: Total number of circuit breaker successes
- `warden_circuit_breaker_timeouts_total`: Total number of circuit breaker timeouts
- `warden_retry_attempts_total`: Total number of retry attempts
