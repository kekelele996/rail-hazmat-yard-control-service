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

func (l *Ledger) Move(consist, from, to, class string) (int, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	current, ok := l.slots[from]
	if !ok || current.Consist != consist {
		return l.revision, fmt.Errorf("consist %s not on %s", consist, from)
	}
	if _, busy := l.slots[to]; busy {
		return l.revision, fmt.Errorf("track %s occupied", to)
	}
	l.revision++
	delete(l.slots, from)
	l.slots[to] = Slot{Track: to, Consist: consist, HazardClass: class, Revision: l.revision}
	return l.revision, nil
}

func (l *Ledger) Snapshot() (map[string]Slot, int) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	out := make(map[string]Slot, len(l.slots))
	for k, v := range l.slots {
		out[k] = v
	}
	return out, l.revision
}
