package platform

import (
	"fmt"
	"strings"
	"sync/atomic"
)

type IDFactory struct {
	prefix string
	seq    atomic.Uint64
}

func NewIDFactory(prefix string) *IDFactory {
	return &IDFactory{prefix: strings.ToUpper(strings.TrimSpace(prefix))}
}
func (f *IDFactory) Next() string { n := f.seq.Add(1); return fmt.Sprintf("%s-%08d", f.prefix, n) }
