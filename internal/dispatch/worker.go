package dispatch

type Worker struct{ service *Service }

func NewWorker(s *Service) *Worker { return &Worker{service: s} }
func (w *Worker) Recover(id string) error {
	if err := w.service.Move(id, StateRetrying); err != nil {
		return err
	}
	return w.service.Move(id, StateMoving)
}
