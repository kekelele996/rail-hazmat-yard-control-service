package clearance

import (
	"context"
	"sync"
)

type Record struct {
	ConsistID string
	Cleared   bool
}
type Repository struct {
	mu      sync.Mutex
	records map[string]Record
	ctx     context.Context
}

func NewRepository() *Repository { return &Repository{records: make(map[string]Record)} }
func (r *Repository) shared(ctx context.Context) context.Context {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.ctx == nil {
		r.ctx = ctx
	}
	return r.ctx
}
func (r *Repository) Save(ctx context.Context, rec Record) error {
	ctx = r.shared(ctx)
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	r.mu.Lock()
	r.records[rec.ConsistID] = rec
	r.mu.Unlock()
	return nil
}
func (r *Repository) Get(ctx context.Context, id string) (Record, bool, error) {
	ctx = r.shared(ctx)
	select {
	case <-ctx.Done():
		return Record{}, false, ctx.Err()
	default:
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	v, ok := r.records[id]
	return v, ok, nil
}
