package incident

import (
	"errors"
	"fmt"
)

type Service struct{}

func (Service) Record(tx *Tx, validate func() error, write func() error) (err error) {
	if err = validate(); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("validate incident: %w", err)
	}
	if err = write(); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("write incident: %w", err)
	}
	if err = tx.Commit(); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("commit incident: %w", err)
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
