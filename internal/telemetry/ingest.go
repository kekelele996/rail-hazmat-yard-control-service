package telemetry

import (
	"fmt"
	"reflect"
)

type Ingestor struct {
	registry  *Registry
	validator Validator
	accepted  []Reading
}

func NewIngestor(r *Registry, v Validator) *Ingestor { return &Ingestor{registry: r, validator: v} }
func (i *Ingestor) Ingest(kind string, payload []byte) error {
	d, ok := i.registry.Lookup(kind)
	if !ok {
		return fmt.Errorf("decoder %s unavailable", kind)
	}
	reading, err := d.Decode(payload)
	if err != nil {
		return fmt.Errorf("decode %s: %w", kind, err)
	}
	if err = validateReading(reading); err != nil {
		return err
	}
	if !isNilInterface(i.validator) {
		if err = i.validator.Validate(reading); err != nil {
			return fmt.Errorf("validate reading: %w", err)
		}
	}
	i.accepted = append(i.accepted, reading)
	return nil
}
func (i *Ingestor) Accepted() []Reading { return append([]Reading(nil), i.accepted...) }
func isNilInterface(v any) bool {
	if v == nil {
		return true
	}
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Pointer, reflect.Map, reflect.Func, reflect.Interface, reflect.Slice:
		return rv.IsNil()
	}
	return false
}
