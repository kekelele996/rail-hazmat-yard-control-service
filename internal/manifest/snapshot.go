package manifest

func cloneCars(in []Car) []Car {
	if in == nil {
		return nil
	}
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
	v := View{Cars: cloneCars(cars), Revision: revision}
	for _, c := range cars {
		if c.UNNumber != "" {
			v.HazardCount++
		}
	}
	return v
}
func (v View) Clone() View { v.Cars = cloneCars(v.Cars); return v }
