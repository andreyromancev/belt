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

func (m *Chain) Handle(ctx context.Context, i belt.Item) (results []belt.Handler, err error) {
	err = m.handle(ctx, &results, 0, i, i.Handler())
	return
}

func (m *Chain) handle(ctx context.Context, total *[]belt.Handler, mIndex int, item belt.Item, handler belt.Handler) error {
	if mIndex >= len(m.chain) {
		*total = append(*total, handler)
		return nil
	}

	state := m.chain[mIndex]
	results, err := state.Handle(ctx, item)
	if err != nil {
		return err
	}
	for _, res := range results {
		err := m.handle(ctx, total, mIndex+1, item, res)
		if err != nil {
			return err
		}
	}

	return nil
}
