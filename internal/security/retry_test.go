package security

import (
	"errors"
	"testing"
	"time"
)

func TestWithRetrySuccess(t *testing.T) {
	config := RetryConfig{
		MaxAttempts: 3,
		Backoff:     time.Millisecond,
	}

	callCount := 0
	err := WithRetry(config, func() error {
		callCount++
		return nil
	})

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if callCount != 1 {
		t.Errorf("Expected 1 call, got %d", callCount)
	}
}

func TestWithRetrySuccessOnSecondAttempt(t *testing.T) {
	config := RetryConfig{
		MaxAttempts: 3,
		Backoff:     time.Millisecond,
	}

	callCount := 0
	err := WithRetry(config, func() error {
		callCount++
		if callCount == 1 {
			return errors.New("first attempt fails")
		}
		return nil
	})

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if callCount != 2 {
		t.Errorf("Expected 2 calls, got %d", callCount)
	}
}

func TestWithRetryExhausted(t *testing.T) {
	config := RetryConfig{
		MaxAttempts: 3,
		Backoff:     time.Millisecond,
	}

	callCount := 0
	expectedErr := errors.New("persistent error")
	err := WithRetry(config, func() error {
		callCount++
		return expectedErr
	})

	if err == nil {
		t.Error("Expected error after exhausting retries")
	}
	if callCount != 3 {
		t.Errorf("Expected 3 calls, got %d", callCount)
	}
	if !errors.Is(err, expectedErr) {
		t.Errorf("Expected error to wrap %v, got %v", expectedErr, err)
	}
}

func TestWithRetryZeroAttempts(t *testing.T) {
	config := RetryConfig{
		MaxAttempts: 0,
		Backoff:     time.Millisecond,
	}

	callCount := 0
	err := WithRetry(config, func() error {
		callCount++
		return nil
	})

	// Should default to 1 attempt
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if callCount != 1 {
		t.Errorf("Expected 1 call, got %d", callCount)
	}
}

func TestWithRetryNegativeAttempts(t *testing.T) {
	config := RetryConfig{
		MaxAttempts: -1,
		Backoff:     time.Millisecond,
	}

	callCount := 0
	err := WithRetry(config, func() error {
		callCount++
		return nil
	})

	// Should default to 1 attempt
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if callCount != 1 {
		t.Errorf("Expected 1 call, got %d", callCount)
	}
}

func TestWithRetrySingleAttempt(t *testing.T) {
	config := RetryConfig{
		MaxAttempts: 1,
		Backoff:     time.Millisecond,
	}

	callCount := 0
	expectedErr := errors.New("single attempt error")
	err := WithRetry(config, func() error {
		callCount++
		return expectedErr
	})

	if err == nil {
		t.Error("Expected error")
	}
	if callCount != 1 {
		t.Errorf("Expected 1 call, got %d", callCount)
	}
}

func TestDefaultRetryConfig(t *testing.T) {
	config := DefaultRetryConfig()

	if config.MaxAttempts <= 0 {
		t.Errorf("MaxAttempts = %d, want > 0", config.MaxAttempts)
	}
	if config.Backoff <= 0 {
		t.Errorf("Backoff = %v, want > 0", config.Backoff)
	}
}

func TestWithRollbackSuccess(t *testing.T) {
	rollbackCalled := false
	err := WithRollback(
		func() error { return nil },
		func() { rollbackCalled = true },
	)

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if rollbackCalled {
		t.Error("Rollback should not be called on success")
	}
}

func TestWithRollbackError(t *testing.T) {
	rollbackCalled := false
	expectedErr := errors.New("operation failed")
	err := WithRollback(
		func() error { return expectedErr },
		func() { rollbackCalled = true },
	)

	if err == nil {
		t.Error("Expected error")
	}
	if !errors.Is(err, expectedErr) {
		t.Errorf("Expected error %v, got %v", expectedErr, err)
	}
	if !rollbackCalled {
		t.Error("Rollback should be called on error")
	}
}

func TestIsRetryable(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"nil error", nil, false},
		{"regular error retryable", errors.New("error"), true},
		{"explicitly retryable", NewRetryableError(errors.New("error"), true), true},
		{"explicitly non-retryable", NewRetryableError(errors.New("error"), false), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsRetryable(tt.err); got != tt.want {
				t.Errorf("IsRetryable() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRetryableError(t *testing.T) {
	inner := errors.New("inner error")
	re := NewRetryableError(inner, true)

	if re.Error() != "inner error" {
		t.Errorf("Error() = %v, want %v", re.Error(), "inner error")
	}
	if !re.Retryable {
		t.Error("Retryable should be true")
	}
}

func TestWithRetryBackoff(t *testing.T) {
	config := RetryConfig{
		MaxAttempts: 3,
		Backoff:     10 * time.Millisecond,
	}

	start := time.Now()
	callCount := 0
	_ = WithRetry(config, func() error {
		callCount++
		if callCount < 3 {
			return errors.New("not yet")
		}
		return nil
	})
	elapsed := time.Since(start)

	// Should have waited at least 2 * backoff (2 retries)
	minExpected := 2 * config.Backoff
	if elapsed < minExpected {
		t.Errorf("Elapsed time %v < expected minimum %v", elapsed, minExpected)
	}
}
