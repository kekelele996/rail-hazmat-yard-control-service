package telemetry

import "sync"

type Decoder interface{ Decode([]byte) (Reading, error) }
type Registry struct {
	mu       sync.RWMutex
	decoders map[string]Decoder
}

func NewRegistry() *Registry { return &Registry{decoders: make(map[string]Decoder)} }
func (r *Registry) Register(kind string, d Decoder) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.decoders == nil {
		r.decoders = make(map[string]Decoder)
	}
	r.decoders[kind] = d
}
func (r *Registry) Lookup(kind string) (Decoder, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	d, ok := r.decoders[kind]
	return d, ok && !isNilDecoder(d)
}
