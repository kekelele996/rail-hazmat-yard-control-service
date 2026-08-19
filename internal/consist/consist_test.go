package consist

import (
	"reflect"
	"testing"
)

func TestConsistFilteringDoesNotMutateInput(t *testing.T) {
	in := []Wagon{{"A", false, ""}, {"B", true, ""}, {"C", false, ""}}
	before := Clone(in)
	_ = HazardOnly(in)
	if !reflect.DeepEqual(in, before) {
		t.Fatalf("input changed: %#v", in)
	}
}
func TestConsistPlanSlicesAreIndependent(t *testing.T) {
	p := BuildPlan([]Wagon{{"A", true, ""}, {"B", true, ""}}, "T9")
	p.Routed[0].ID = "changed"
	if p.Hazard[0].ID == "changed" || p.Original[0].ID == "changed" {
		t.Fatal("plan slices share storage")
	}
}
func TestConsistReservationReturnsDetachedPlan(t *testing.T) {
	b := NewReservationBook()
	b.Save("P", BuildPlan([]Wagon{{"A", true, ""}}, "T1"))
	p, _ := b.Load("P")
	p.Routed[0].Track = "BAD"
	again, _ := b.Load("P")
	if again.Routed[0].Track != "T1" {
		t.Fatalf("stored plan polluted: %#v", again)
	}
}

func TestConsistExportDoesNotExposePlanStorage(t *testing.T) {
	plan := BuildPlan([]Wagon{{"A", true, ""}}, "T4")
	exported := ExportPlan(plan)
	exported.Routed[0].Track = "BAD"
	if plan.Routed[0].Track != "T4" {
		t.Fatalf("export mutation leaked into plan: %#v", plan)
	}
}
