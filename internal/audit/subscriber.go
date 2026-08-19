package audit

import "context"

type Event struct {
	Sequence int
	Kind     string
}
type Source interface {
	Next(context.Context, int) (Event, error)
}
type SourceFunc func(context.Context, int) (Event, error)

func (f SourceFunc) Next(c context.Context, n int) (Event, error) { return f(context.Background(), n) }

type Sink interface {
	Deliver(context.Context, Event) error
}
type SinkFunc func(context.Context, Event) error

func (f SinkFunc) Deliver(c context.Context, e Event) error { return f(context.Background(), e) }
