package incident

import (
	"errors"
	"sync"
)

var ErrCommit = errors.New("incident commit failed")

type Tx struct {
	mu        sync.Mutex
	committed bool
	rolled    bool
	commitErr error
}

func NewTx(commitErr error) *Tx { return &Tx{commitErr: commitErr} }
func (t *Tx) Commit() error     { t.mu.Lock(); defer t.mu.Unlock(); t.committed = true; return t.commitErr }
func (t *Tx) Rollback() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.committed {
		return nil
	}
	t.rolled = true
	return nil
}
func (t *Tx) Committed() bool  { t.mu.Lock(); defer t.mu.Unlock(); return t.committed }
func (t *Tx) RolledBack() bool { t.mu.Lock(); defer t.mu.Unlock(); return t.rolled }
