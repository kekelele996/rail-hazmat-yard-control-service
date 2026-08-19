package manifest

func cloneCars(in []Car) []Car {
	out := make([]Car, len(in))
	copy(out, in)
	return out
}

type View struct {
	Cars        []Car
	Revision    int
	HazardCount int
}

func NewView(cars []Car, revision int) View {
	v := View{Cars: cars, Revision: revision}
	for _, c := range cars {
		if c.UNNumber != "" {
			v.HazardCount++
		}
	}
	return v
}
func (v View) Clone() View {
	return View{
		Cars:        cloneCars(v.Cars),
		Revision:    v.Revision,
		HazardCount: v.HazardCount,
	}
}
