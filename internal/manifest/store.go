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

func NewStore(seed []Car) *Store { return &Store{cars: cloneCars(seed), revision: 1} }
func (s *Store) Snapshot() ([]Car, int) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return cloneCars(s.cars), s.revision
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
	next := cloneCars(s.cars)
	for i := range next {
		if next[i].ID == id {
			s.revision++
			next[i].Seal = seal
			next[i].Revision = s.revision
			s.cars = next
			return true
		}
	}
	return false
}
