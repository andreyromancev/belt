package timed

import (
	"context"

	"github.com/andreyromancev/belt/log"

	"github.com/andreyromancev/belt/mware"

	"github.com/pkg/errors"

	"github.com/andreyromancev/belt"
)

type FutureHandler interface {
	Future(ctx context.Context) ([]belt.Handler, error)
}

type PastHandler interface {
	Past(ctx context.Context) ([]belt.Handler, error)
}

var PresentMiddleware mware.Func = func(ctx context.Context, h belt.Handler) ([]belt.Handler, error) {
	return h.Handle(ctx)
}

var FutureMiddleware mware.Func = func(ctx context.Context, h belt.Handler) ([]belt.Handler, error) {
	if f, ok := h.(FutureHandler); ok {
		return f.Future(ctx)
	}

	log.FromContext(ctx).Info("Waiting for present")
	// Future waits for reset by default.
	<-ctx.Done()
	return []belt.Handler{h}, nil
}

var PastMiddleware mware.Func = func(ctx context.Context, h belt.Handler) ([]belt.Handler, error) {
	if f, ok := h.(PastHandler); ok {
		return f.Past(ctx)
	}

	// Past fails by default.
	return nil, errors.New("no past handler")
}
