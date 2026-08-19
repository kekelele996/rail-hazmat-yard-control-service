package occupancy

import "fmt"

type Allocator struct{ ledger *Ledger }

func NewAllocator(l *Ledger) *Allocator { return &Allocator{ledger: l} }
func (a *Allocator) Move(consist, from, to, class string) error {
	slots, _ := a.ledger.Snapshot()
	if current, ok := slots[from]; !ok || current.Consist != consist {
		return fmt.Errorf("consist %s not on %s", consist, from)
	}
	if _, busy := slots[to]; busy {
		return fmt.Errorf("track %s occupied", to)
	}
	a.ledger.Release(from)
	a.ledger.Reserve(Slot{Track: to, Consist: consist, HazardClass: class})
	return nil
}
