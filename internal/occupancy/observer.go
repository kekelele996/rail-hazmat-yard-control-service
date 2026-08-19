package occupancy

type View struct {
	Slots     map[string]Slot
	Revision  int
	Hazardous int
}

// Observe builds a self-consistent view: the slots map is cloned under the
// ledger's read lock and Hazardous is derived from that same frozen copy, so
// a concurrent Reserve/Release can never mutate the map a reader is iterating
// or produce a torn Hazardous != len(Slots) view.
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

// Clone returns a view whose slots map is independent of the receiver's.
func (v View) Clone() View {
	if v.Slots == nil {
		return v
	}
	clone := make(map[string]Slot, len(v.Slots))
	for k, s := range v.Slots {
		clone[k] = s
	}
	return View{Slots: clone, Revision: v.Revision, Hazardous: v.Hazardous}
}
