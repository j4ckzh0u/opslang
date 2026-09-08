package exec

import (
	"context"
	"fmt"
	"time"
)

// RetryPolicy bounds transient task retries and their backoff.
type RetryPolicy struct {
	Attempts int
	Backoff  time.Duration
}

func (p RetryPolicy) normalize() RetryPolicy {
	if p.Attempts <= 0 {
		p.Attempts = 1
	}
	if p.Backoff < 0 {
		p.Backoff = 0
	}
	return p
}

// ExecuteWithRetry retries an operation while preserving context cancellation.
func ExecuteWithRetry[T any](ctx context.Context, policy RetryPolicy, operation func(context.Context) (T, error)) (T, error) {
	var zero T
	if ctx == nil {
		return zero, fmt.Errorf("retry context is nil")
	}
	if operation == nil {
		return zero, fmt.Errorf("retry operation is nil")
	}
	policy = policy.normalize()
	var lastErr error
	for attempt := 0; attempt < policy.Attempts; attempt++ {
		value, err := operation(ctx)
		if err == nil {
			return value, nil
		}
		lastErr = err
		if attempt+1 == policy.Attempts {
			break
		}
		timer := time.NewTimer(policy.Backoff)
		select {
		case <-ctx.Done():
			timer.Stop()
			return zero, ctx.Err()
		case <-timer.C:
		}
	}
	return zero, fmt.Errorf("operation failed after %d attempts: %w", policy.Attempts, lastErr)
}
