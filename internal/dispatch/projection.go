package dispatch

type Projection struct {
	State      State
	InProgress bool
	Attempts   int
}

func Project(j Job) Projection {
	return Projection{State: j.State, InProgress: IsInProgress(j.State), Attempts: j.Attempts}
}
