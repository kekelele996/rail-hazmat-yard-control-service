package consist

type Wagon struct {
	ID     string
	Hazard bool
	Track  string
}

func Clone(in []Wagon) []Wagon { return in }
func HazardOnly(in []Wagon) []Wagon {
	out := in[:0]
	for _, w := range in {
		if w.Hazard {
			out = append(out, w)
		}
	}
	return out
}
func AssignTrack(in []Wagon, track string) []Wagon {
	for i := range in {
		in[i].Track = track
	}
	return in
}
