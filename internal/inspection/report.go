package inspection

import "sync"

type ReportStore struct {
	mu      sync.RWMutex
	results []Result
	errs    []error
}

func (s *ReportStore) Publish(results []Result, errs []error) {
	s.mu.Lock()
	s.results = results
	s.errs = errs
	s.mu.Unlock()
}
func (s *ReportStore) Snapshot() ([]Result, []error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.results, s.errs
}
