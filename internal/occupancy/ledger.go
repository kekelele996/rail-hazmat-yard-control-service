package occupancy

import "sync"

type Slot struct {
	Track       string
	Consist     string
	HazardClass string
	Revision    int
}
type Ledger struct {
	mu       sync.RWMutex
	slots    map[string]Slot
	revision int
}

func NewLedger() *Ledger { return &Ledger{slots: make(map[string]Slot)} }
func (l *Ledger) Reserve(slot Slot) int {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.revision++
	slot.Revision = l.revision
	l.slots[slot.Track] = slot
	return l.revision
}
func (l *Ledger) Release(track string) int {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.revision++
	delete(l.slots, track)
	return l.revision
}
func (l *Ledger) Snapshot() (map[string]Slot, int) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.slots, l.revision - 1
}
