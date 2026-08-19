package consist

import "sync"

type ReservationBook struct {
	mu    sync.Mutex
	plans map[string]Plan
}

func NewReservationBook() *ReservationBook        { return &ReservationBook{plans: make(map[string]Plan)} }
func (b *ReservationBook) Save(id string, p Plan) { b.mu.Lock(); b.plans[id] = p; b.mu.Unlock() }
func (b *ReservationBook) Load(id string) (Plan, bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	p, ok := b.plans[id]
	return p, ok
}
