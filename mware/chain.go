package mware

import (
	"context"
	"github.com/andreyromancev/belt"
)

type Chain struct {
	chain []belt.Middleware
}

func NewChain() *Chain {
	return &Chain{}
}

func (m *Chain) AddMiddleware(s belt.Middleware) {
	m.chain = append(m.chain, s)
}

func (m *Chain) Handle(ctx context.Context, h belt.Handler) (results []belt.Handler, err error) {
	err = m.handle(ctx, &results, 0, h)
	return
}

func (m *Chain) handle(ctx context.Context, total *[]belt.Handler, mIndex int, handler belt.Handler) error {
	if mIndex >= len(m.chain) {
		*total = append(*total, handler)
		return nil
	}

	state := m.chain[mIndex]
	results, err := state.Handle(ctx, handler)
	if err != nil {
		return err
	}
	for _, res := range results {
		err := m.handle(ctx, total, mIndex + 1, res)
		if err != nil {
			return err
		}
	}

	return nil
}
