package clearance

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

func TestClearanceDeadlineReachesVerifier(t *testing.T) {
	missing := errors.New("deadline not propagated")
	v := VerifierFunc(func(ctx context.Context, id string) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(80 * time.Millisecond):
			return missing
		}
	})
	p := NewPipeline(NewRepository(), v)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	err := p.Clear(ctx, "C1")
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("err=%v", err)
	}
}
func TestClearanceRequestsDoNotShareContext(t *testing.T) {
	r := NewRepository()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_ = r.Save(ctx, Record{ConsistID: "old"})
	if err := r.Save(context.Background(), Record{ConsistID: "new"}); err != nil {
		t.Fatalf("fresh repository call poisoned: %v", err)
	}
}
func TestClearanceRetryStopsAfterCancellation(t *testing.T) {
	var calls atomic.Int32
	v := VerifierFunc(func(ctx context.Context, id string) error { calls.Add(1); return errors.New("busy") })
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_ = verifyWithRetry(ctx, v, "C1", 3)
	if calls.Load() != 0 {
		t.Fatalf("calls after cancel=%d", calls.Load())
	}
}
func TestClearanceSessionUsesCallerContext(t *testing.T) {
	var calls atomic.Int32
	p := NewPipeline(NewRepository(), VerifierFunc(func(context.Context, string) error { calls.Add(1); return nil }))
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := RunSession(ctx, p, "C9")
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("err=%v", err)
	}
	if calls.Load() != 0 {
		t.Fatalf("verifier called %d times", calls.Load())
	}
}
