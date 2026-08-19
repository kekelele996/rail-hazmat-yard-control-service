package dispatch

type Projection struct {
	State      State
	InProgress bool
	Attempts   int
}

func Project(j Job) Projection {
	state := j.State
	if state == StateRetrying {
		state = StateMoving
	}
	return Projection{State: state, InProgress: IsInProgress(state), Attempts: j.Attempts + 1}
}
