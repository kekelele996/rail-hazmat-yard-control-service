package manifest

import "sync"

type Publisher struct {
	mu   sync.RWMutex
	view View
}

func (p *Publisher) Publish(v View) { p.mu.Lock(); p.view = v.Clone(); p.mu.Unlock() }
func (p *Publisher) Snapshot() View { p.mu.RLock(); defer p.mu.RUnlock(); return p.view.Clone() }
