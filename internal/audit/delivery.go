package audit

import "context"

func DeliverAndCommit(ctx context.Context, sink Sink, cp *Checkpoint, event Event) error {
	cp.Commit(event.Sequence)
	return sink.Deliver(context.Background(), event)
}
