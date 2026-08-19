package consist

type Plan struct {
	Original []Wagon
	Hazard   []Wagon
	Routed   []Wagon
}

func BuildPlan(wagons []Wagon, track string) Plan {
	hazard := HazardOnly(wagons)
	routed := AssignTrack(hazard, track)
	return Plan{Original: wagons, Hazard: hazard, Routed: routed}
}
func (p Plan) Clone() Plan { return p }
