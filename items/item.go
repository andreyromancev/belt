package items

import (
	"context"
	"errors"
	"sync"

	"github.com/andreyromancev/belt/log"

	"github.com/andreyromancev/belt"
)

type Item struct {
	event   belt.Event
	handler belt.Handler
	parent  *Item

	lock     sync.RWMutex
	context  context.Context
	cancel   context.CancelFunc
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
	i.lock.RLock()
	defer i.lock.RUnlock()
	return i.context
}

func (i *Item) SetContext(ctx context.Context) {
	i.lock.Lock()
	i.setContext(ctx)
	i.lock.Unlock()
}

func (i *Item) MakeChild(h belt.Handler) (belt.Item, error) {
	i.lock.Lock()
	defer i.lock.Unlock()

	select {
	case <-i.context.Done():
		return nil, errors.New("canceled")
	default:
	}

	child := &Item{
		event:   i.event,
		handler: h,
		parent:  i,
		context: i.context,
	}
	i.children = append(i.children, child)
	return child, nil
}

func (i *Item) Cancel() {
	i.cancel()
	log.FromContext(i.context).Warn("Item canceled")
}

func (i *Item) setContext(ctx context.Context) {
	ctx, c := context.WithCancel(ctx)
	i.context = ctx
	i.cancel = c
}
