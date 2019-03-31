package items

import (
	"context"
	"github.com/andreyromancev/belt"
	"sync"
)

type Item struct {
	event belt.Event
	handler belt.Handler
	parent *Item

	ctxLock sync.RWMutex
	context context.Context
	cancel context.CancelFunc

	chLock sync.RWMutex
	children []*Item
}

func NewItem(e belt.Event, h belt.Handler) *Item {
	return &Item{
		event: e,
		handler: h,
	}
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
	ctx, c := context.WithCancel(ctx)
	i.ctxLock.Lock()
	i.context = ctx
	i.cancel = c
	i.ctxLock.Unlock()
}

func (i *Item) MakeChild(h belt.Handler) belt.Item {
	child := &Item{
		event: i.event,
		handler: h,
		context: i.context,
		parent: i,
	}
	i.chLock.Lock()
	i.children = append(i.children, child)
	i.chLock.Unlock()
	return child
}

func (i *Item) Cancel() {
	i.ctxLock.RLock()
	i.cancel()
	i.ctxLock.RUnlock()
}
