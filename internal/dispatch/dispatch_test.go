package dispatch

import "testing"

func TestDispatchRetryCanComplete(t *testing.T) {
	s := NewService()
	s.Put(Job{ID: "J", State: StateHeld})
	if err := NewWorker(s).Recover("J"); err != nil {
		t.Fatal(err)
	}
	j, _ := s.Get("J")
	if j.State != StateCompleted {
		t.Fatalf("state=%s", j.State)
	}
}
func TestDispatchRetryingIsVisibleInProgress(t *testing.T) {
	got := FilterInProgress([]Job{{ID: "A", State: StateRetrying}, {ID: "B", State: StateCompleted}})
	if len(got) != 1 || got[0].ID != "A" {
		t.Fatalf("got=%v", got)
	}
}
func TestDispatchRecoveryIncrementsAttemptsOnce(t *testing.T) {
	s := NewService()
	s.Put(Job{ID: "J", State: StateHeld})
	if err := NewWorker(s).Recover("J"); err != nil {
		t.Fatal(err)
	}
	j, _ := s.Get("J")
	if j.Attempts != 1 {
		t.Fatalf("attempts=%d", j.Attempts)
	}
}

func TestDispatchProjectionPreservesRetrying(t *testing.T) {
	p := Project(Job{ID: "J", State: StateRetrying, Attempts: 2})
	if p.State != StateRetrying || !p.InProgress || p.Attempts != 2 {
		t.Fatalf("projection=%+v", p)
	}
}
