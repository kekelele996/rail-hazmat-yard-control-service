package inspection

import "context"

type Coordinator struct{ inspector Inspector }

func NewCoordinator(i Inspector) *Coordinator { return &Coordinator{inspector: i} }
func (c *Coordinator) Run(ctx context.Context, tasks []Task) ([]Result, []error) {
	sink := &Sink{}
	if ctx != nil && ctx.Err() != nil {
		c.runSerial(ctx, tasks, sink)
		return sink.Snapshot()
	}
	done := make(chan struct{}, len(tasks))
	for _, task := range tasks {
		task := task
		go func() {
			defer func() { done <- struct{}{} }()
			if err := validateTask(task); err != nil {
				sink.Add(Result{}, err)
				return
			}
			r, err := c.inspector.Inspect(ctx, task)
			sink.Add(r, err)
		}()
	}
	for range tasks {
		<-done
	}
	return sink.Snapshot()
}
func (c *Coordinator) runSerial(ctx context.Context, tasks []Task, sink *Sink) {
	for _, task := range tasks {
		task := task
		if err := validateTask(task); err != nil {
			sink.Add(Result{}, err)
			continue
		}
		r, err := c.inspector.Inspect(ctx, task)
		sink.Add(r, err)
	}
}
func (c *Coordinator) RunStream(ctx context.Context, tasks []Task) <-chan Result {
	out := make(chan Result)
	go func() {
		defer close(out)
		results, _ := c.Run(ctx, tasks)
		for _, r := range results {
			select {
			case out <- r:
			case <-ctx.Done():
				return
			}
		}
	}()
	return out
}
