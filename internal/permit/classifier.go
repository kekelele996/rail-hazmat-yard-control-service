package permit

import (
	"errors"
	"rail-hazmat-yard-control-service/internal/platform"
)

func ShouldRetry(err error) bool {
	return errors.Is(err, ErrAuthorityBusy) || errors.Is(err, platform.ErrUnavailable)
}
