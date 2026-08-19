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
	if d == nil {
		return true
	}
	v := reflect.ValueOf(d)
	switch v.Kind() {
	case reflect.Pointer, reflect.Map, reflect.Func, reflect.Interface, reflect.Slice:
		return v.IsNil()
	}
	return false
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
