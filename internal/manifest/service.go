package manifest

import "sync"

type Service struct {
	store  *Store
	mu     sync.RWMutex
	latest View
}

func NewService(store *Store) *Service { s := &Service{store: store}; s.Refresh(); return s }
func (s *Service) Refresh() View {
	cars, rev := s.store.Snapshot()
	next := NewView(cars, rev)
	s.mu.Lock()
	s.latest = next
	s.mu.Unlock()
	return next
}
func (s *Service) Current() View            { s.mu.RLock(); defer s.mu.RUnlock(); return s.latest }
func (s *Service) AddAndRefresh(c Car) View { s.store.Append(c); s.Refresh(); return s.Current() }
