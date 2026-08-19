package inspection

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

func TestInspectionWaitsForEveryWorker(t *testing.T) {
	started := make(chan string, 3)
	releaseFirst := make(chan struct{})
	releaseRest := make(chan struct{})
	i := InspectorFunc(func(ctx context.Context, task Task) (Result, error) {
		started <- task.CarID
		if task.CarID == "A" {
			<-releaseFirst
		} else {
			<-releaseRest
		}
		return Result{CarID: task.CarID, Passed: true}, nil
	})
	c := NewCoordinator(i)
	done := make(chan []Result, 1)
	go func() { r, _ := c.Run(context.Background(), []Task{{"A", "G1"}, {"B", "G1"}, {"C", "G2"}}); done <- r }()
	for i := 0; i < 3; i++ {
		<-started
	}
	close(releaseFirst)
	select {
	case r := <-done:
		t.Fatalf("coordinator returned after first worker: %d results", len(r))
	case <-time.After(25 * time.Millisecond):
	}
	close(releaseRest)
	r := <-done
	if len(r) != 3 {
		t.Fatalf("got %d results", len(r))
	}
}

func TestInspectionCollectsWorkerErrors(t *testing.T) {
	boom := errors.New("sensor offline")
	var gate sync.WaitGroup
	gate.Add(2)
	c := NewCoordinator(InspectorFunc(func(context.Context, Task) (Result, error) { gate.Done(); gate.Wait(); return Result{}, boom }))
	r, e := c.Run(context.Background(), []Task{{"A", "G1"}, {"B", "G1"}})
	if len(r) != 0 || len(e) != 2 {
		t.Fatalf("results=%d errors=%d", len(r), len(e))
	}
}

func TestInspectionStreamClosesAfterResults(t *testing.T) {
	c := NewCoordinator(InspectorFunc(func(_ context.Context, task Task) (Result, error) {
		if task.CarID == "B" {
			time.Sleep(20 * time.Millisecond)
		}
		return Result{CarID: task.CarID, Passed: true}, nil
	}))
	n := 0
	for range c.RunStream(context.Background(), []Task{{"A", "G1"}, {"B", "G2"}}) {
		n++
	}
	if n != 2 {
		t.Fatalf("streamed %d", n)
	}
}

func TestInspectionReportPublicationDetached(t *testing.T) {
	store := &ReportStore{}
	results := []Result{{CarID: "A", Passed: true}}
	errs := []error{errors.New("gate warning")}
	store.Publish(results, errs)
	start := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(2)
	go func() { defer wg.Done(); <-start; results[0].CarID = "input"; errs[0] = nil }()
	go func() {
		defer wg.Done()
		<-start
		got, gotErrs := store.Snapshot()
		got[0].CarID = "output"
		gotErrs[0] = nil
	}()
	close(start)
	wg.Wait()
	again, againErrs := store.Snapshot()
	if again[0].CarID != "A" || againErrs[0] == nil {
		t.Fatalf("published report was mutated: results=%v errors=%v", again, againErrs)
	}
}
