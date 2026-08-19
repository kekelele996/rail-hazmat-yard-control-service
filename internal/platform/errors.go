package platform

import "errors"

var ErrNotFound = errors.New("not found")
var ErrConflict = errors.New("state conflict")
var ErrUnavailable = errors.New("dependency unavailable")
var ErrUnauthorized = errors.New("unauthorized operation")
var ErrCancelled = errors.New("operation cancelled")
