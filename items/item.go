package items

import (
	"context"
	"sync"

	"github.com/andreyromancev/belt"
)

type Item struct {
	event   belt.Event
	handler belt.Handler
	parent  *Item

	ctxLock sync.RWMutex
	context context.Context
	cancel  context.CancelFunc

	chLock   sync.RWMutex
	children []*Item
}

func NewItem(ctx context.Context, e belt.Event, h belt.Handler) *Item {
	i := &Item{
		event:   e,
		handler: h,
	}
	i.SetContext(ctx)
	return i
}

func (i *Item) Event() belt.Event {
	return i.event
}

func (i *Item) Handler() belt.Handler {
	return i.handler
}

func (i *Item) Context() context.Context {
	return i.context
}

func (i *Item) SetContext(ctx context.Context) {
	i.ctxLock.Lock()
	i.setContext(ctx)
	i.ctxLock.Unlock()
}

func (i *Item) MakeChild(h belt.Handler) belt.Item {
	child := &Item{
		event:   i.event,
		handler: h,
		parent:  i,
	}
	child.setContext(i.context)

	i.chLock.Lock()
	i.children = append(i.children, child)
	i.chLock.Unlock()
	return child
}

func (i *Item) Cancel() {
	i.cancel()
}

func (i *Item) setContext(ctx context.Context) {
	ctx, c := context.WithCancel(ctx)
	i.context = ctx
	i.cancel = c
}
