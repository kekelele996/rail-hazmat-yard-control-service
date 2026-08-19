package clearance

import (
	"context"
	"fmt"
)

type Pipeline struct {
	repo     *Repository
	verifier Verifier
}

func NewPipeline(r *Repository, v Verifier) *Pipeline { return &Pipeline{repo: r, verifier: v} }
func (p *Pipeline) Clear(ctx context.Context, id string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := verifyWithRetry(ctx, p.verifier, id, 3); err != nil {
		return fmt.Errorf("verify clearance %s: %w", id, err)
	}
	if err := p.repo.Save(ctx, Record{ConsistID: id, Cleared: true}); err != nil {
		return fmt.Errorf("save clearance %s: %w", id, err)
	}
	return nil
}
func (p *Pipeline) Status(ctx context.Context, id string) (bool, error) {
	rec, ok, err := p.repo.Get(ctx, id)
	if err != nil {
		return false, err
	}
	return ok && rec.Cleared, nil
}
