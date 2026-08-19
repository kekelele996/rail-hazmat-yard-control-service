package consist

type Wagon struct {
	ID     string
	Hazard bool
	Track  string
}

func Clone(in []Wagon) []Wagon {
	if in == nil {
		return nil
	}
	out := make([]Wagon, len(in))
	copy(out, in)
	return out
}
func HazardOnly(in []Wagon) []Wagon {
	out := make([]Wagon, 0, len(in))
	for _, w := range in {
		if w.Hazard {
			out = append(out, w)
		}
	}
	return out
}
func AssignTrack(in []Wagon, track string) []Wagon {
	out := Clone(in)
	for i := range out {
		out[i].Track = track
	}
	return out
}
