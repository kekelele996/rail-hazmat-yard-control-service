package audit

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

func TestAuditCancellationReachesSource(t *testing.T) {
	missing := errors.New("source context not cancelled")
	src := SourceFunc(func(ctx context.Context, n int) (Event, error) {
		select {
		case <-ctx.Done():
			return Event{}, ctx.Err()
		case <-time.After(80 * time.Millisecond):
			return Event{}, missing
		}
	})
	s := NewStream(src, SinkFunc(func(context.Context, Event) error { return nil }), &Checkpoint{})
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	err := s.Run(ctx, 1)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("err=%v", err)
	}
}

func TestAuditFailedDeliveryDoesNotAdvanceCheckpoint(t *testing.T) {
	cp := &Checkpoint{}
	src := SourceFunc(func(context.Context, int) (Event, error) { return Event{Sequence: 1, Kind: "move"}, nil })
	s := NewStream(src, SinkFunc(func(context.Context, Event) error { return errors.New("down") }), cp)
	_ = s.Run(context.Background(), 1)
	if cp.Load() != 0 {
		t.Fatalf("checkpoint=%d", cp.Load())
	}
}

func TestAuditStopsReadingAfterCancel(t *testing.T) {
	var reads atomic.Int32
	ctx, cancel := context.WithCancel(context.Background())
	src := SourceFunc(func(context.Context, int) (Event, error) {
		n := reads.Add(1)
		if n == 1 {
			cancel()
		}
		return Event{Sequence: int(n)}, nil
	})
	s := NewStream(src, SinkFunc(func(context.Context, Event) error { return nil }), &Checkpoint{})
	_ = s.Run(ctx, 10)
	if reads.Load() != 1 {
		t.Fatalf("reads=%d", reads.Load())
	}
}

func TestAuditDeliveryHelperCommitsAfterSuccess(t *testing.T) {
	cp := &Checkpoint{}
	err := DeliverAndCommit(context.Background(), SinkFunc(func(context.Context, Event) error { return errors.New("offline") }), cp, Event{Sequence: 9})
	if err == nil {
		t.Fatal("expected delivery error")
	}
	if cp.Load() != 0 {
		t.Fatalf("checkpoint advanced to %d", cp.Load())
	}
}
