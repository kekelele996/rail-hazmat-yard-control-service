package occupancy

type Allocator struct{ ledger *Ledger }

func NewAllocator(l *Ledger) *Allocator { return &Allocator{ledger: l} }

// Move is a thin wrapper over Ledger.Move, which performs the relocate as a
// single atomic revision. Earlier versions read a snapshot, released, and
// reserved in separate locked steps, which both published two revisions per
// move and left a check-then-act window for concurrent allocators.
func (a *Allocator) Move(consist, from, to, class string) error {
	return a.ledger.Move(consist, from, to, class)
}
