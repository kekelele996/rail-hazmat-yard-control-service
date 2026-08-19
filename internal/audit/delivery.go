package audit

import "context"

func DeliverAndCommit(ctx context.Context, sink Sink, cp *Checkpoint, event Event) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := sink.Deliver(ctx, event); err != nil {
		return err
	}
	cp.Commit(event.Sequence)
	return nil
}
