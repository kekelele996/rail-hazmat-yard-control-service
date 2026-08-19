package dispatch

func IsInProgress(s State) bool { return s == StateQueued || s == StateMoving || s == StateHeld }
func FilterInProgress(jobs []Job) []Job {
	out := make([]Job, 0, len(jobs))
	for _, j := range jobs {
		if IsInProgress(j.State) {
			out = append(out, j)
		}
	}
	return out
}
