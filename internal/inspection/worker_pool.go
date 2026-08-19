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

func inspectionContext(ctx context.Context) (context.Context, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return ctx, err
	}
	return ctx, nil
}
func (f InspectorFunc) Inspect(ctx context.Context, t Task) (Result, error) {
	child, err := inspectionContext(ctx)
	if err != nil {
		return Result{}, err
	}
	return f(child, t)
}
func validateTask(t Task) error {
	if t.CarID == "" || t.Gate == "" {
		return fmt.Errorf("invalid inspection task")
	}
	return nil
}
