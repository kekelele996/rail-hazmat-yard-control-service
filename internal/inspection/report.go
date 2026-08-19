package inspection

import "sync"

type ReportStore struct {
	mu      sync.RWMutex
	results []Result
	errs    []error
}

func (s *ReportStore) Publish(results []Result, errs []error) {
	s.mu.Lock()
	s.results = append([]Result(nil), results...)
	s.errs = append([]error(nil), errs...)
	s.mu.Unlock()
}
func (s *ReportStore) Snapshot() ([]Result, []error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]Result(nil), s.results...), append([]error(nil), s.errs...)
}
