package telemetry

import (
	"errors"
	"testing"
	"time"
)

type decoder struct{}

func (*decoder) Decode([]byte) (Reading, error) {
	return Reading{WagonID: "W1", Kind: "temp", Value: 12, At: time.Now()}, nil
}

type nilDecoder struct{}

func (*nilDecoder) Decode([]byte) (Reading, error) { panic("typed nil decoder called") }

type nilValidator struct{}

func (*nilValidator) Validate(Reading) error { panic("typed nil validator called") }
func TestTelemetryZeroRegistryCanRegister(t *testing.T) {
	var r Registry
	r.Register("temp", &decoder{})
	if _, ok := r.Lookup("temp"); !ok {
		t.Fatal("decoder missing")
	}
}
func TestTelemetryRejectsTypedNilDecoder(t *testing.T) {
	r := NewRegistry()
	var d *nilDecoder
	r.Register("temp", d)
	if _, ok := r.Lookup("temp"); ok {
		t.Fatal("typed nil decoder exposed")
	}
}
func TestTelemetryValidationCannotBeBypassed(t *testing.T) {
	r := NewRegistry()
	r.Register("temp", &decoder{})
	want := errors.New("temperature blocked")
	i := NewIngestor(r, ValidatorFunc(func(Reading) error { return want }))
	if err := i.Ingest("temp", nil); !errors.Is(err, want) {
		t.Fatalf("validation lost: %v", err)
	}
	if len(i.Accepted()) != 0 {
		t.Fatal("invalid reading accepted")
	}
}

func TestTelemetryProbeRejectsTypedNil(t *testing.T) {
	var d *nilDecoder
	if DecoderAvailable(d) {
		t.Fatal("probe accepted typed nil decoder")
	}
}
