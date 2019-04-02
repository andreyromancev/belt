package mware

import (
	"context"

	"github.com/andreyromancev/belt"
)

type Func func(context.Context, belt.Item) ([]belt.Handler, error)

func (m Func) Handle(c context.Context, i belt.Item) ([]belt.Handler, error) {
	return m(c, i)
}
