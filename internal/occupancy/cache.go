package occupancy

import "sync"

type Cache struct {
	mu   sync.RWMutex
	view View
}

func (c *Cache) Publish(v View) { c.mu.Lock(); c.view = v.Clone(); c.mu.Unlock() }
func (c *Cache) Snapshot() View { c.mu.RLock(); defer c.mu.RUnlock(); return c.view.Clone() }
