package items

import (
	"context"
	"sync"

	"github.com/andreyromancev/belt/log"

	"github.com/pkg/errors"

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

func (i *Item) MakeChild(h belt.Handler) (belt.Item, error) {
	child := &Item{
		event:   i.event,
		handler: h,
		parent:  i,
	}
	child.setContext(i.context)

	i.chLock.Lock()
	defer i.chLock.Unlock()
	select {
	case <-i.context.Done():
		return nil, errors.New("canceled")
	default:
	}
	i.children = append(i.children, child)
	return child, nil
}

func (i *Item) Cancel() {
	i.chLock.RLock()
	for _, ch := range i.children {
		ch.Cancel()
		log.FromContext(i.context).Warn("Item canceled")
	}
	i.chLock.RUnlock()
	i.cancel()
}

func (i *Item) setContext(ctx context.Context) {
	ctx, c := context.WithCancel(ctx)
	i.context = ctx
	i.cancel = c
}
