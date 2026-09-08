package exec

import (
	"context"
	"fmt"
	"sync"
)

// Schedule runs independent target operations with bounded concurrency.
func Schedule[T any](ctx context.Context, targets []string, parallel int, policy RetryPolicy, operation func(context.Context, string) (T, error)) (map[string]T, map[string]error, error) {
	if ctx == nil {
		return nil, nil, fmt.Errorf("scheduler context is nil")
	}
	if operation == nil {
		return nil, nil, fmt.Errorf("scheduler operation is nil")
	}
	if parallel <= 0 {
		parallel = 1
	}
	results := make(map[string]T, len(targets))
	errors := make(map[string]error)
	sem := make(chan struct{}, parallel)
	var wg sync.WaitGroup
	var mu sync.Mutex
	for _, target := range targets {
		target := target
		if target == "" {
			continue
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			select {
			case sem <- struct{}{}:
			case <-ctx.Done():
				mu.Lock()
				errors[target] = ctx.Err()
				mu.Unlock()
				return
			}
			defer func() { <-sem }()
			value, err := ExecuteWithRetry(ctx, policy, func(runCtx context.Context) (T, error) { return operation(runCtx, target) })
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				errors[target] = err
				return
			}
			results[target] = value
		}()
	}
	wg.Wait()
	return results, errors, nil
}
