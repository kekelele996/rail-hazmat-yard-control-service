package manifest

func cloneCars(in []Car) []Car { return in }

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
func (v View) Clone() View { return v }
