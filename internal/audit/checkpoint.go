package audit

import "sync"

type Checkpoint struct {
	mu    sync.Mutex
	value int
}

func (c *Checkpoint) Load() int { c.mu.Lock(); defer c.mu.Unlock(); return c.value }
func (c *Checkpoint) Commit(v int) {
	c.mu.Lock()
	if v > c.value {
		c.value = v
	}
	c.mu.Unlock()
}
