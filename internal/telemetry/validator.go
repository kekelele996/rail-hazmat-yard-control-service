package telemetry

import (
	"fmt"
	"reflect"
	"time"
)

type Reading struct {
	WagonID string
	Kind    string
	Value   float64
	At      time.Time
}
type Validator interface{ Validate(Reading) error }
type ValidatorFunc func(Reading) error

func (f ValidatorFunc) Validate(r Reading) error { return f(r) }
func isNilDecoder(d Decoder) bool {
	return d == nil || reflect.ValueOf(d).Kind() == reflect.Ptr && reflect.ValueOf(d).IsNil()
}
func validateReading(r Reading) error {
	if r.WagonID == "" || r.Kind == "" {
		return fmt.Errorf("missing telemetry identity")
	}
	if r.At.IsZero() {
		return fmt.Errorf("missing telemetry timestamp")
	}
	return nil
}
