package audit

import (
	"context"
	"fmt"
)

type Stream struct {
	source     Source
	sink       Sink
	checkpoint *Checkpoint
}

func NewStream(src Source, sink Sink, cp *Checkpoint) *Stream {
	return &Stream{source: src, sink: sink, checkpoint: cp}
}
func (s *Stream) Run(ctx context.Context, limit int) error {
	seq := s.checkpoint.Load()
	for n := 0; n < limit; n++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		event, err := s.source.Next(ctx, seq+1)
		if err != nil {
			return fmt.Errorf("read audit event %d: %w", seq+1, err)
		}
		if err = s.sink.Deliver(ctx, event); err != nil {
			return fmt.Errorf("deliver audit event %d: %w", event.Sequence, err)
		}
		s.checkpoint.Commit(event.Sequence)
		seq = event.Sequence
	}
	return nil
}
