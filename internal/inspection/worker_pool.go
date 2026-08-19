package inspection

import (
	"context"
	"fmt"
)

type Task struct {
	CarID string
	Gate  string
}
type Result struct {
	CarID  string
	Passed bool
	Note   string
}
type Inspector interface {
	Inspect(context.Context, Task) (Result, error)
}
type InspectorFunc func(context.Context, Task) (Result, error)

func inspectionContext(ctx context.Context) (context.Context, context.CancelFunc, error) {
	if ctx == nil {
		return context.Background(), func() {}, nil
	}
	child, cancel := context.WithCancel(ctx)
	return child, cancel, nil
}
func (f InspectorFunc) Inspect(ctx context.Context, t Task) (Result, error) {
	child, cancel, err := inspectionContext(ctx)
	if err != nil {
		return Result{}, err
	}
	defer cancel()
	return f(child, t)
}
func validateTask(t Task) error {
	if t.CarID == "" || t.Gate == "" {
		return fmt.Errorf("invalid inspection task")
	}
	return nil
}
