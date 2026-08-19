package consist

type Plan struct {
	Original []Wagon
	Hazard   []Wagon
	Routed   []Wagon
}

func BuildPlan(wagons []Wagon, track string) Plan {
	original := Clone(wagons)
	hazard := HazardOnly(original)
	routed := AssignTrack(hazard, track)
	return Plan{Original: original, Hazard: hazard, Routed: routed}
}
func (p Plan) Clone() Plan {
	return Plan{
		Original: Clone(p.Original),
		Hazard:   Clone(p.Hazard),
		Routed:   Clone(p.Routed),
	}
}
