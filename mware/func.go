package mware

import (
	"context"

	"github.com/andreyromancev/belt"
)

type Func func(context.Context, belt.Handler) ([]belt.Handler, error)

func (m Func) Handle(c context.Context, h belt.Handler) ([]belt.Handler, error) {
	return m(c, h)
}
