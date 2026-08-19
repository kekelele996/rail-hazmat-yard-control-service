package inspection

import "context"

type Coordinator struct{ inspector Inspector }

func NewCoordinator(i Inspector) *Coordinator { return &Coordinator{inspector: i} }
func completionTarget(tasks []Task) int       { return len(tasks) }
func (c *Coordinator) Run(ctx context.Context, tasks []Task) ([]Result, []error) {
	sink := &Sink{}
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
	for i := 0; i < completionTarget(tasks); i++ {
		<-done
	}
	return sink.Snapshot()
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
