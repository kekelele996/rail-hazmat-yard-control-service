package dispatch

type Worker struct{ service *Service }

func NewWorker(s *Service) *Worker { return &Worker{service: s} }
func (w *Worker) Recover(id string) error {
	job, ok := w.service.Get(id)
	if !ok {
		return nil
	}
	path, err := RecoveryPath(job.State)
	if err != nil {
		return err
	}
	return w.service.MovePath(id, path)
}
