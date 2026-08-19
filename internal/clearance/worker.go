package clearance

import (
	"context"
	"time"
)

type Verifier interface {
	Verify(context.Context, string) error
}
type VerifierFunc func(context.Context, string) error

func (f VerifierFunc) Verify(c context.Context, id string) error { return f(c, id) }
func verifyWithRetry(ctx context.Context, v Verifier, id string, attempts int) error {
	var err error
	for n := 0; n < attempts; n++ {
		if err = ctx.Err(); err != nil {
			return err
		}
		if err = v.Verify(ctx, id); err == nil {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Millisecond):
		}
	}
	return err
}
