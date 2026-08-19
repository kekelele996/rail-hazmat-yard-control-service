package dispatch

var inProgressStates = map[State]struct{}{
	StateQueued:   {},
	StateMoving:   {},
	StateHeld:     {},
	StateRetrying: {},
}

func IsInProgress(s State) bool {
	_, ok := inProgressStates[s]
	return ok
}

func FilterInProgress(jobs []Job) []Job {
	out := make([]Job, 0, len(jobs))
	for _, j := range jobs {
		if IsInProgress(j.State) {
			out = append(out, j)
		}
	}
	return out
}
