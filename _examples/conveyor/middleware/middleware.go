package middleware

import (
	"context"
	"github.com/andreyromancev/belt"
	"github.com/andreyromancev/belt/_examples/conveyor/handlers"
)


type Future struct {}

func (Future) Handle(ctx context.Context, handler belt.Handler) belt.Handler {
	if h, ok := handler.(handlers.Future); ok {
		return h.Future(ctx)
	}

	return handler.Handle(ctx)
}

type Past struct {}

func (Past) Handle(ctx context.Context, handler belt.Handler) belt.Handler {
	if h, ok := handler.(handlers.Past); ok {
		return h.Future(ctx)
	}

	return handler.Handle(ctx)
}
