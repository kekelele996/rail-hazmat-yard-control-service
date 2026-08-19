package manifest

import (
	"fmt"
	"sync"
	"testing"
)

func TestManifestSnapshotIsolation(t *testing.T) {
	s := NewStore([]Car{{ID: "A", UNNumber: "1203", Seal: "S1"}})
	snap, _ := s.Snapshot()
	if !s.ReplaceSeal("A", "S2") {
		t.Fatal("replace failed")
	}
	if snap[0].Seal != "S1" {
		t.Fatalf("historical snapshot changed: %s", snap[0].Seal)
	}
}

func TestManifestConcurrentViewStable(t *testing.T) {
	s := NewStore([]Car{{ID: "A", UNNumber: "1203", Seal: "S1"}})
	snap, _ := s.Snapshot()
	start := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		<-start
		for i := 0; i < 1000; i++ {
			_ = snap[0].Seal
		}
	}()
	go func() {
		defer wg.Done()
		<-start
		for i := 0; i < 1000; i++ {
			s.ReplaceSeal("A", fmt.Sprintf("S%d", i))
		}
	}()
	close(start)
	wg.Wait()
	if snap[0].Seal != "S1" {
		t.Fatalf("concurrent writer changed published snapshot: %s", snap[0].Seal)
	}
}

func TestManifestRevisionMatchesPublishedView(t *testing.T) {
	s := NewStore(nil)
	svc := NewService(s)
	for i := 0; i < 25; i++ {
		v := svc.AddAndRefresh(Car{ID: fmt.Sprint(i), UNNumber: "1017"})
		if v.Revision != len(v.Cars)+1 {
			t.Fatalf("revision %d cars %d", v.Revision, len(v.Cars))
		}
	}
}

func TestManifestPublisherDetachesView(t *testing.T) {
	p := &Publisher{}
	v := NewView([]Car{{ID: "A", UNNumber: "1203", Seal: "S1"}}, 7)
	p.Publish(v)
	v.Cars[0].Seal = "input"
	first := p.Snapshot()
	first.Cars[0].Seal = "output"
	again := p.Snapshot()
	if again.Cars[0].Seal != "S1" {
		t.Fatalf("publisher view polluted: %s", again.Cars[0].Seal)
	}
}
