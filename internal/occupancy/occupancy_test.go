package occupancy

import (
	"fmt"
	"sync"
	"testing"
)

func TestOccupancySnapshotDetached(t *testing.T) {
	l := NewLedger()
	l.Reserve(Slot{Track: "T1", Consist: "C1", HazardClass: "3"})
	s, _ := l.Snapshot()
	delete(s, "T1")
	again, _ := l.Snapshot()
	if _, ok := again["T1"]; !ok {
		t.Fatal("external map mutation changed ledger")
	}
}

func TestOccupancyConcurrentViewsAreConsistent(t *testing.T) {
	l := NewLedger()
	start := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		<-start
		for i := 0; i < 300; i++ {
			track := fmt.Sprintf("T%d", i%8)
			l.Reserve(Slot{Track: track, Consist: fmt.Sprint(i), HazardClass: "3"})
		}
	}()
	go func() {
		defer wg.Done()
		<-start
		for i := 0; i < 300; i++ {
			v := Observe(l)
			if v.Hazardous != len(v.Slots) {
				t.Errorf("torn view hazardous=%d slots=%d", v.Hazardous, len(v.Slots))
				return
			}
		}
	}()
	close(start)
	wg.Wait()
}

func TestOccupancyMovePublishesSingleRevision(t *testing.T) {
	l := NewLedger()
	l.Reserve(Slot{Track: "T1", Consist: "C1", HazardClass: "3"})
	before := Observe(l)
	if err := NewAllocator(l).Move("C1", "T1", "T2", "3"); err != nil {
		t.Fatal(err)
	}
	after := Observe(l)
	if after.Revision != before.Revision+1 {
		t.Fatalf("move published %d revisions", after.Revision-before.Revision)
	}
	if _, ok := after.Slots["T1"]; ok {
		t.Fatal("source track still occupied")
	}
	if after.Slots["T2"].Consist != "C1" {
		t.Fatal("target track missing")
	}
}

func TestOccupancyCacheDoesNotExposePublishedMap(t *testing.T) {
	c := &Cache{}
	v := View{Slots: map[string]Slot{"T1": {Track: "T1", Consist: "C1"}}, Revision: 3}
	c.Publish(v)
	delete(v.Slots, "T1")
	first := c.Snapshot()
	delete(first.Slots, "T1")
	again := c.Snapshot()
	if _, ok := again.Slots["T1"]; !ok {
		t.Fatal("cache view map was externally mutated")
	}
}
