package handlers

import (
	"context"
	"github.com/andreyromancev/belt/src"
)

type Future interface {
	Future(context.Context) belt.Handler
}

type Past interface {
	Future(context.Context) belt.Handler
}
