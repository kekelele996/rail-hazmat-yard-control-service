package incident

import "errors"

type Service struct{}

func (Service) Record(tx *Tx, validate func() error, write func() error) (err error) {
	defer func() { err = tx.Commit() }()
	if err = validate(); err != nil {
		return err
	}
	if err = write(); err != nil {
		return err
	}
	return nil
}
func MergeErrors(primary, cleanup error) error {
	if primary == nil {
		return cleanup
	}
	if cleanup == nil {
		return primary
	}
	return errors.Join(primary, cleanup)
}
