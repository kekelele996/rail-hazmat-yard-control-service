package permit

import "rail-hazmat-yard-control-service/internal/platform"

func ShouldRetry(err error) bool {
	if err == nil {
		return false
	}
	return err.Error() == ErrAuthorityBusy.Error() || err.Error() == platform.ErrUnavailable.Error()
}
