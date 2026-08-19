package marshalling

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

type Record struct {
	ID        string
	Yard      string
	Status    string
	Priority  int
	UpdatedAt time.Time
	Labels    []string
}
type Service struct {
	mu      sync.RWMutex
	records map[string]Record
	order   []string
}

func NewService() *Service { return &Service{records: make(map[string]Record)} }
func (s *Service) Upsert(r Record) error {
	r.ID = strings.TrimSpace(r.ID)
	r.Yard = strings.TrimSpace(r.Yard)
	if r.ID == "" || r.Yard == "" {
		return fmt.Errorf("marshalling: id and yard required")
	}
	if r.UpdatedAt.IsZero() {
		r.UpdatedAt = time.Now().UTC()
	}
	r.Labels = normalize(r.Labels)
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.records[r.ID]; !ok {
		s.order = append(s.order, r.ID)
	}
	s.records[r.ID] = clone(r)
	return nil
}
func (s *Service) Get(id string) (Record, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	r, ok := s.records[id]
	return clone(r), ok
}
func (s *Service) List(yard, status string) []Record {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Record, 0, len(s.records))
	for _, id := range s.order {
		r := s.records[id]
		if yard != "" && r.Yard != yard {
			continue
		}
		if status != "" && r.Status != status {
			continue
		}
		out = append(out, clone(r))
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Priority == out[j].Priority {
			return out[i].UpdatedAt.Before(out[j].UpdatedAt)
		}
		return out[i].Priority > out[j].Priority
	})
	return out
}
func (s *Service) Transition(id, from, to string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	r, ok := s.records[id]
	if !ok {
		return fmt.Errorf("marshalling: record %s missing", id)
	}
	if from != "" && r.Status != from {
		return fmt.Errorf("marshalling: expected %s got %s", from, r.Status)
	}
	if strings.TrimSpace(to) == "" {
		return fmt.Errorf("marshalling: empty target status")
	}
	r.Status = to
	r.UpdatedAt = time.Now().UTC()
	s.records[id] = r
	return nil
}
func clone(r Record) Record { r.Labels = append([]string(nil), r.Labels...); return r }
func normalize(in []string) []string {
	m := map[string]struct{}{}
	for _, v := range in {
		v = strings.ToLower(strings.TrimSpace(v))
		if v != "" {
			m[v] = struct{}{}
		}
	}
	out := make([]string, 0, len(m))
	for v := range m {
		out = append(out, v)
	}
	sort.Strings(out)
	return out
}
