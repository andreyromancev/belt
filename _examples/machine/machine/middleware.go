package machine

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

var PresentMiddleware mware.Func = func(ctx context.Context, i belt.Item) ([]belt.Handler, error) {
	return i.Handler().Handle(ctx)
}

var FutureMiddleware mware.Func = func(ctx context.Context, i belt.Item) ([]belt.Handler, error) {
	if f, ok := i.Handler().(FutureHandler); ok {
		return f.Future(ctx)
	}

	log.FromContext(ctx).Info("Waiting for present")
	// Future waits for reset by default.
	<-ctx.Done()
	return []belt.Handler{i.Handler()}, nil
}

var PastMiddleware mware.Func = func(ctx context.Context, i belt.Item) ([]belt.Handler, error) {
	if f, ok := i.Handler().(PastHandler); ok {
		return f.Past(ctx)
	}

	// Past fails by default.
	return nil, errors.New("no past handler")
}
