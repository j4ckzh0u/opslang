package security

import (
	"fmt"
	"time"
)

// RetryConfig configures retry behavior
type RetryConfig struct {
	MaxAttempts int           `json:"max_attempts"`
	Backoff     time.Duration `json:"backoff"`
}

// DefaultRetryConfig returns sensible default retry configuration
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts: 3,
		Backoff:     time.Second,
	}
}

// WithRetry executes fn with retry logic
func WithRetry(config RetryConfig, fn func() error) error {
	if config.MaxAttempts <= 0 {
		config.MaxAttempts = 1
	}

	var lastErr error
	for attempt := 1; attempt <= config.MaxAttempts; attempt++ {
		if err := fn(); err != nil {
			lastErr = err
			if attempt < config.MaxAttempts {
				// Wait before retrying
				time.Sleep(config.Backoff)
			}
			continue
		}
		// Success
		return nil
	}

	return fmt.Errorf("failed after %d attempts: %w", config.MaxAttempts, lastErr)
}

// WithRollback executes fn and rolls back on error
func WithRollback(fn func() error, rollback func()) error {
	if err := fn(); err != nil {
		// Attempt rollback
		rollback()
		return err
	}
	return nil
}

// RetryableError represents an error that indicates whether retry is appropriate
type RetryableError struct {
	Err       error
	Retryable bool
}

func (e *RetryableError) Error() string {
	return e.Err.Error()
}

// IsRetryable checks if an error is retryable
func IsRetryable(err error) bool {
	if err == nil {
		return false
	}
	if re, ok := err.(*RetryableError); ok {
		return re.Retryable
	}
	// By default, assume errors are retryable
	return true
}

// NewRetryableError creates a new retryable error
func NewRetryableError(err error, retryable bool) *RetryableError {
	return &RetryableError{
		Err:       err,
		Retryable: retryable,
	}
}
