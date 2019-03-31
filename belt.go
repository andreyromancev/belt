package belt

import (
	"context"
)

type Event interface {}

type Item interface {
	Event() Event
	Handler() Handler
	MakeChild(Handler) Item
	Context() context.Context
	SetContext(ctx context.Context)
}

type Canceler interface {
	Cancel()
}

type Handler interface {
	Handle(context.Context) ([]Handler, error)
}

type Middleware interface {
	Handle(context.Context, Handler) ([]Handler, error)
}

type Slot interface {
	Middleware() Middleware
	AddItem(Item) error
}

type Sorter interface {
	Sort(Event) (Slot, Item, error)
}

type Worker interface {
	Work(ctx context.Context, items <-chan Event) error
}
