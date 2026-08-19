package clearance

import "context"

func RunSession(ctx context.Context, p *Pipeline, id string) error {
	return p.Clear(context.Background(), id)
}
