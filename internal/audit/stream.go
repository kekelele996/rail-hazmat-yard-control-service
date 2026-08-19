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
		event, err := s.source.Next(context.Background(), seq+1)
		if err != nil {
			return fmt.Errorf("read audit event %d: %w", seq+1, err)
		}
		s.checkpoint.Commit(event.Sequence)
		seq = event.Sequence
		if err = s.sink.Deliver(context.Background(), event); err != nil {
			return fmt.Errorf("deliver audit event %d: %w", event.Sequence, err)
		}
	}
	return nil
}
