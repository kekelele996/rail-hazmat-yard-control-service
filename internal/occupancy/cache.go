package occupancy

import "sync"

type Cache struct {
	mu   sync.RWMutex
	view View
}

// Publish stores a private copy of the view so the caller retains no alias
// into the cache's internal map.
func (c *Cache) Publish(v View) {
	c.mu.Lock()
	c.view = v.Clone()
	c.mu.Unlock()
}

// Snapshot returns a private copy so callers cannot mutate the cache's map
// through the returned view.
func (c *Cache) Snapshot() View {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.view.Clone()
}
