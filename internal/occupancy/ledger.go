package occupancy

import (
	"fmt"
	"sync"
)

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

// Move relocates a consist from one track to another as a single atomic
// revision. Validation, release, and reserve all happen under the write lock
// so that no concurrent mutation can interleave and the ledger advances by
// exactly one revision per move.
func (l *Ledger) Move(consist, from, to, class string) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	current, occupied := l.slots[from]
	if !occupied || current.Consist != consist {
		return fmt.Errorf("consist %s not on %s", consist, from)
	}
	if _, busy := l.slots[to]; busy {
		return fmt.Errorf("track %s occupied", to)
	}

	l.revision++
	delete(l.slots, from)
	l.slots[to] = Slot{Track: to, Consist: consist, HazardClass: class, Revision: l.revision}
	return nil
}

// Snapshot returns a defensive copy of the current slots together with the
// current revision. Callers may freely mutate the returned map without
// affecting the ledger.
func (l *Ledger) Snapshot() (map[string]Slot, int) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	clone := make(map[string]Slot, len(l.slots))
	for k, v := range l.slots {
		clone[k] = v
	}
	return clone, l.revision
}
