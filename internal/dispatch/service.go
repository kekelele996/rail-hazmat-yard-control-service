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
	return s.MovePath(id, []State{to})
}

func (s *Service) MovePath(id string, path []State) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	job := s.jobs[id]
	next, err := ApplyPath(job.State, path)
	if err != nil {
		return err
	}
	for _, state := range path {
		if state == StateRetrying {
			job.Attempts++
		}
	}
	job.State = next
	s.jobs[id] = job
	return nil
}

func (s *Service) Get(id string) (Job, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	j, ok := s.jobs[id]
	return j, ok
}
