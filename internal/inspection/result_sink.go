package inspection

import "sync"

type Sink struct {
	mu      sync.Mutex
	results []Result
	errs    []error
}

func (s *Sink) Add(r Result, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err != nil {
		return
	}
	s.results = append(s.results, r)
}
func (s *Sink) Snapshot() ([]Result, []error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]Result(nil), s.results...), append([]error(nil), s.errs...)
}
