package exec

import (
	"context"
	"errors"
	"testing"
)

func TestExecuteWithRetryRetriesAndReturns(t *testing.T) {
	attempts := 0
	value, err := ExecuteWithRetry(context.Background(), RetryPolicy{Attempts: 3}, func(context.Context) (string, error) {
		attempts++
		if attempts < 3 {
			return "", errors.New("temporary")
		}
		return "ok", nil
	})
	if err != nil || value != "ok" || attempts != 3 {
		t.Fatalf("value=%q err=%v attempts=%d", value, err, attempts)
	}
}

func TestExecuteWithRetryRejectsInvalidInputs(t *testing.T) {
	if _, err := ExecuteWithRetry[string](nil, RetryPolicy{}, func(context.Context) (string, error) { return "", nil }); err == nil {
		t.Fatal("expected nil context error")
	}
	if _, err := ExecuteWithRetry[string](context.Background(), RetryPolicy{}, nil); err == nil {
		t.Fatal("expected nil operation error")
	}
}
