package exec

import (
	"context"
	"errors"
	"testing"
)

var errTestSchedule = errors.New("schedule failure")

func TestScheduleCollectsResultsAndErrors(t *testing.T) {
	results, failures, err := Schedule(context.Background(), []string{"a", "b", ""}, 2, RetryPolicy{Attempts: 1}, func(_ context.Context, target string) (string, error) {
		if target == "b" {
			return "", errTestSchedule
		}
		return target + "-ok", nil
	})
	if err != nil || results["a"] != "a-ok" || failures["b"] == nil {
		t.Fatalf("results=%v failures=%v err=%v", results, failures, err)
	}
}

func TestScheduleRejectsInvalidInput(t *testing.T) {
	if _, _, err := Schedule[string](nil, nil, 1, RetryPolicy{}, func(context.Context, string) (string, error) { return "", nil }); err == nil {
		t.Fatal("expected nil context error")
	}
}
