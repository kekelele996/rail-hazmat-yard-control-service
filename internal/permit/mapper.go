package permit

import (
	"errors"
	"net/http"
	"rail-hazmat-yard-control-service/internal/platform"
)

type ErrorResponse struct {
	Status    int
	Code      string
	Retryable bool
}

func MapError(err error) ErrorResponse {
	switch {
	case errors.Is(err, ErrPermitDenied), errors.Is(err, platform.ErrUnauthorized):
		return ErrorResponse{http.StatusForbidden, "permit_denied", false}
	case errors.Is(err, ErrAuthorityBusy), errors.Is(err, platform.ErrUnavailable):
		return ErrorResponse{http.StatusServiceUnavailable, "authority_unavailable", true}
	case errors.Is(err, contextDeadline):
		return ErrorResponse{http.StatusGatewayTimeout, "deadline", true}
	default:
		return ErrorResponse{http.StatusInternalServerError, "internal", false}
	}
}

var contextDeadline = errors.New("deadline exceeded")
