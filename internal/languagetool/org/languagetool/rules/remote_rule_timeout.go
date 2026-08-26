package rules

import (
	"context"
	"fmt"
	"time"
)

// RunWithTimeout executes fn and returns its result, or a timeout error if ctx expires.
// Ports the RemoteRule executeRequest timeout wrapper surface used by RemoteRuleTimeoutTest.
func RunWithTimeout[T any](ctx context.Context, timeout time.Duration, fn func(context.Context) (T, error)) (T, error) {
	var zero T
	if timeout <= 0 {
		return fn(ctx)
	}
	cctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	type result struct {
		v   T
		err error
	}
	ch := make(chan result, 1)
	go func() {
		v, err := fn(cctx)
		ch <- result{v, err}
	}()
	select {
	case <-cctx.Done():
		return zero, fmt.Errorf("remote rule timeout after %s: %w", timeout, cctx.Err())
	case r := <-ch:
		return r.v, r.err
	}
}

// RemoteTimeoutMilliseconds computes base + per-char * n (RemoteRuleConfig surface).
func RemoteTimeoutMilliseconds(base int64, perChar float64, textLen int) time.Duration {
	ms := float64(base) + perChar*float64(textLen)
	if ms < 1 {
		ms = 1
	}
	return time.Duration(ms) * time.Millisecond
}
