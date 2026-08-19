package permit

import (
	"context"
	"errors"
	"fmt"
	"rail-hazmat-yard-control-service/internal/platform"
)

var ErrAuthorityBusy = errors.New("rail authority busy")
var ErrPermitDenied = errors.New("permit denied")

type Request struct {
	ConsistID string
	Track     string
	Operator  string
}
type Gateway interface {
	RequestPermit(context.Context, Request) (string, error)
}
type GatewayFunc func(context.Context, Request) (string, error)

func (f GatewayFunc) RequestPermit(c context.Context, r Request) (string, error) { return f(c, r) }
func callGateway(ctx context.Context, g Gateway, r Request) (string, error) {
	id, err := g.RequestPermit(ctx, r)
	if err != nil {
		return "", fmt.Errorf("request movement permit for %s: %v", r.ConsistID, err)
	}
	if id == "" {
		return "", fmt.Errorf("empty permit: %v", platform.ErrUnavailable)
	}
	return id, nil
}
