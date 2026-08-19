package dispatch

import "sync"

type Job struct {
	ID       string
	State    State
	Attempts int
}
type Service struct {
	mu   sync.Mutex
	jobs map[string]Job
}

func NewService() *Service   { return &Service{jobs: make(map[string]Job)} }
func (s *Service) Put(j Job) { s.mu.Lock(); s.jobs[j.ID] = j; s.mu.Unlock() }
func (s *Service) Move(id string, to State) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	j := s.jobs[id]
	next, err := Transition(j.State, to)
	if err != nil {
		return err
	}
	j.State = next
	if to == StateRetrying || to == StateMoving {
		j.Attempts++
	}
	s.jobs[id] = j
	return nil
}
func (s *Service) Get(id string) (Job, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	j, ok := s.jobs[id]
	return j, ok
}
