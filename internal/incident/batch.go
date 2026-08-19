package incident

import (
	"fmt"
	"sync"
)

type ResourceTracker struct {
	mu   sync.Mutex
	open int
	max  int
}
type Resource struct {
	t      *ResourceTracker
	closed bool
}

func NewResourceTracker(max int) *ResourceTracker { return &ResourceTracker{max: max} }
func (t *ResourceTracker) Open() (*Resource, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.open >= t.max {
		return nil, fmt.Errorf("resource limit %d reached", t.max)
	}
	t.open++
	return &Resource{t: t}, nil
}
func (r *Resource) Close() error {
	r.t.mu.Lock()
	defer r.t.mu.Unlock()
	if !r.closed {
		r.closed = true
		r.t.open--
	}
	return nil
}
func (t *ResourceTracker) OpenCount() int { t.mu.Lock(); defer t.mu.Unlock(); return t.open }
func ProcessBatch(items []string, t *ResourceTracker, handle func(string) error) error {
	for _, item := range items {
		if err := processOne(item, t, handle); err != nil {
			return err
		}
	}
	return nil
}
func processOne(item string, t *ResourceTracker, handle func(string) error) (err error) {
	r, err := t.Open()
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := r.Close(); err == nil {
			err = closeErr
		}
	}()
	return handle(item)
}
