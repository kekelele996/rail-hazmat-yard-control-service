package permit

import (
	"context"
	"errors"
	"rail-hazmat-yard-control-service/internal/platform"
	"time"
)

type Service struct {
	gateway  Gateway
	attempts int
	delay    time.Duration
}

func NewService(g Gateway, attempts int) *Service {
	if attempts < 1 {
		attempts = 1
	}
	return &Service{gateway: g, attempts: attempts, delay: time.Millisecond}
}
func (s *Service) Issue(ctx context.Context, r Request) (string, error) {
	var err error
	for i := 0; i < s.attempts; i++ {
		var id string
		id, err = callGateway(ctx, s.gateway, r)
		if err == nil {
			return id, nil
		}
		if errors.Is(err, ErrPermitDenied) || errors.Is(err, platform.ErrUnauthorized) {
			return "", err
		}
		if !errors.Is(err, ErrAuthorityBusy) && !errors.Is(err, platform.ErrUnavailable) {
			return "", err
		}
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-time.After(s.delay):
		}
	}
	return "", err
}
