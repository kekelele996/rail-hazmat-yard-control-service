package permit

import (
	"errors"
	"net/http"
)

type ErrorResponse struct {
	Status    int
	Code      string
	Retryable bool
}

func MapError(err error) ErrorResponse {
	if err == nil {
		return ErrorResponse{http.StatusOK, "ok", false}
	}
	switch err.Error() {
	case ErrPermitDenied.Error():
		return ErrorResponse{http.StatusForbidden, "permit_denied", false}
	case ErrAuthorityBusy.Error():
		return ErrorResponse{http.StatusServiceUnavailable, "authority_unavailable", true}
	default:
		return ErrorResponse{http.StatusInternalServerError, "internal", false}
	}
}

var contextDeadline = errors.New("deadline exceeded")
