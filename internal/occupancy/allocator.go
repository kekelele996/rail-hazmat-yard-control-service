package occupancy

type Allocator struct{ ledger *Ledger }

func NewAllocator(l *Ledger) *Allocator { return &Allocator{ledger: l} }

func (a *Allocator) Move(consist, from, to, class string) error {
	_, err := a.ledger.Move(consist, from, to, class)
	return err
}
