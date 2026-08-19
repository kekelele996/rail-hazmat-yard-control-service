package permit

import (
	"context"
	"errors"
	"fmt"
	"rail-hazmat-yard-control-service/internal/platform"
	"testing"
)

func TestPermitPreservesSentinelChain(t *testing.T) {
	g := GatewayFunc(func(context.Context, Request) (string, error) { return "", ErrPermitDenied })
	_, err := NewService(g, 3).Issue(context.Background(), Request{ConsistID: "C1"})
	if !errors.Is(err, ErrPermitDenied) {
		t.Fatalf("lost denial: %v", err)
	}
}
func TestPermitDoesNotRetryUnauthorized(t *testing.T) {
	calls := 0
	g := GatewayFunc(func(context.Context, Request) (string, error) { calls++; return "", platform.ErrUnauthorized })
	_, _ = NewService(g, 4).Issue(context.Background(), Request{ConsistID: "C1"})
	if calls != 1 {
		t.Fatalf("unauthorized retried %d times", calls)
	}
}
func TestPermitMapsBusyDependency(t *testing.T) {
	g := GatewayFunc(func(context.Context, Request) (string, error) { return "", ErrAuthorityBusy })
	_, err := NewService(g, 1).Issue(context.Background(), Request{ConsistID: "C1"})
	r := MapError(err)
	if r.Status != 503 || !r.Retryable {
		t.Fatalf("mapping=%+v err=%v", r, err)
	}
}

func TestPermitClassifierHandlesWrappedBusy(t *testing.T) {
	wrapped := fmt.Errorf("authority response: %w", ErrAuthorityBusy)
	if !ShouldRetry(wrapped) {
		t.Fatal("wrapped busy error classified as terminal")
	}
	if ShouldRetry(fmt.Errorf("denied: %w", ErrPermitDenied)) {
		t.Fatal("denial classified as retryable")
	}
}
