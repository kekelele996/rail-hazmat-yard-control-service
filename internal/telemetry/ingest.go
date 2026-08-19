package telemetry

import "fmt"

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
	if i.validator != nil {
		if err = i.validator.Validate(reading); err != nil {
			return fmt.Errorf("validate reading: %w", err)
		}
	}
	i.accepted = append(i.accepted, reading)
	return nil
}
func (i *Ingestor) Accepted() []Reading { return i.accepted }
func isNilInterface(v any) bool         { return v == nil }
