package occupancy

type View struct {
	Slots     map[string]Slot
	Revision  int
	Hazardous int
}

func Observe(l *Ledger) View {
	slots, rev := l.Snapshot()
	v := View{Slots: slots, Revision: rev}
	for _, s := range slots {
		if s.HazardClass != "" {
			v.Hazardous++
		}
	}
	return v
}
func (v View) Clone() View {
	out := make(map[string]Slot, len(v.Slots))
	for k, s := range v.Slots {
		out[k] = s
	}
	v.Slots = out
	return v
}
