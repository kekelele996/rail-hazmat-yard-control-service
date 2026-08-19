package manifest

import "sync"

type Car struct {
	ID       string
	UNNumber string
	Seal     string
	Revision int
}
type Store struct {
	mu       sync.RWMutex
	cars     []Car
	revision int
}

func NewStore(seed []Car) *Store { return &Store{cars: seed, revision: 1} }
func (s *Store) Snapshot() ([]Car, int) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.cars, s.revision
}
func (s *Store) Append(car Car) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.revision++
	car.Revision = s.revision
	s.cars = append(s.cars, car)
	return s.revision
}
func (s *Store) ReplaceSeal(id, seal string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.cars {
		if s.cars[i].ID == id {
			s.revision++
			s.cars[i].Seal = seal
			s.cars[i].Revision = s.revision
			return true
		}
	}
	return false
}
