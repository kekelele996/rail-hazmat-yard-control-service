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
func (v View) Clone() View { return v }
