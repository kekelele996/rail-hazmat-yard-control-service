package clearance

import "context"

func RunSession(ctx context.Context, p *Pipeline, id string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return p.Clear(ctx, id)
}
