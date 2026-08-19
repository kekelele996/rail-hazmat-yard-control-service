package incident

import (
	"errors"
	"fmt"
	"testing"
)

func TestIncidentBatchReleasesEachResource(t *testing.T) {
	tracker := NewResourceTracker(2)
	items := make([]string, 20)
	for i := range items {
		items[i] = fmt.Sprint(i)
	}
	if err := ProcessBatch(items, tracker, func(string) error { return nil }); err != nil {
		t.Fatal(err)
	}
	if tracker.OpenCount() != 0 {
		t.Fatalf("open=%d", tracker.OpenCount())
	}
}
func TestIncidentBusinessErrorSurvivesCleanup(t *testing.T) {
	business := errors.New("invalid placard")
	tx := NewTx(nil)
	err := Service{}.Record(tx, func() error { return business }, func() error { return nil })
	if !errors.Is(err, business) {
		t.Fatalf("business error lost: %v", err)
	}
	if !tx.RolledBack() || tx.Committed() {
		t.Fatalf("tx state committed=%v rolled=%v", tx.Committed(), tx.RolledBack())
	}
}
func TestIncidentCommitFailureRollsBack(t *testing.T) {
	tx := NewTx(ErrCommit)
	err := Service{}.Record(tx, func() error { return nil }, func() error { return nil })
	if !errors.Is(err, ErrCommit) {
		t.Fatalf("commit error lost: %v", err)
	}
	if !tx.RolledBack() {
		t.Fatal("commit failure did not rollback")
	}
}

func TestIncidentCleanupKeepsPrimaryError(t *testing.T) {
	primary := errors.New("placard mismatch")
	cleanup := errors.New("close failed")
	err := Finalize(primary, func() error { return cleanup })
	if !errors.Is(err, primary) || !errors.Is(err, cleanup) {
		t.Fatalf("combined error lost component: %v", err)
	}
}
