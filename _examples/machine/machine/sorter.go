package machine

import (
	"context"
	"fmt"
	"sync"

	"github.com/andreyromancev/belt/log"

	"github.com/andreyromancev/belt/mware"

	"github.com/andreyromancev/belt/items"

	"github.com/andreyromancev/belt"
	"github.com/andreyromancev/belt/slots"
	"github.com/pkg/errors"
)

type Sorter struct {
	sLock       sync.RWMutex
	currentTime int
	slots       map[int]belt.Slot
}

func NewSorter(time int) *Sorter {
	return &Sorter{
		currentTime: time,
		slots: map[int]belt.Slot{
			time - 1: slots.NewSlot(PastMiddleware),
			time:     slots.NewSlot(PresentMiddleware),
			time + 1: slots.NewSlot(FutureMiddleware),
		},
	}
}

func (s *Sorter) Sort(ctx context.Context, e belt.Event) (slot belt.Slot, item belt.Item, err error) {
	switch event := e.(type) {
	case Message:
		return s.handleMessage(ctx, event)
	case TimeChange:
		err = s.changeTime()
		return
	default:
		err = errors.New("unknown event type")
		return
	}
}

func (s *Sorter) handleMessage(ctx context.Context, msg Message) (slot belt.Slot, item belt.Item, err error) {
	ctx = context.WithValue(ctx, "Time", s.currentTime)
	ctx = log.WithLogger(ctx, log.FromContext(ctx).WithField("time", fmt.Sprintf("%d", s.currentTime)))

	slot, err = s.slot(msg.Time)
	if err != nil {
		return
	}
	var handler belt.Handler
	switch msg.Time {
	case s.currentTime - 1:
		handler, err = PastInit(msg)
	case s.currentTime:
		handler, err = PresentInit(msg)
	case s.currentTime + 1:
		handler, err = FutureInit(msg)
	default:
		err = errors.New("wrong Time")
	}
	if err != nil {
		return
	}

	item = items.NewItem(ctx, msg, handler)
	return
}

func (s *Sorter) changeTime() error {
	s.sLock.Lock()
	defer s.sLock.Unlock()

	// Deactivate past.
	if past, ok := s.slots[s.currentTime-1]; ok {
		inactive := mware.Func(func(c context.Context, handler belt.Item) ([]belt.Handler, error) {
			return nil, errors.New("inactive")
		})
		past.Reset(inactive)
		delete(s.slots, s.currentTime-1)
	}

	// Move present to past.
	if present, ok := s.slots[s.currentTime]; ok {
		present.Reset(PastMiddleware)
	}

	// Move future to present.
	if future, ok := s.slots[s.currentTime+1]; ok {
		future.Reset(PresentMiddleware)
	}

	// Update Time.
	s.currentTime += 1

	// Create future.
	slot := slots.NewSlot(FutureMiddleware)
	s.slots[s.currentTime+1] = slot

	return nil
}

func (s *Sorter) slot(time int) (slot belt.Slot, err error) {
	s.sLock.RLock()
	slot, ok := s.slots[time]
	if !ok {
		err = errors.New("no slot for message")
	}
	s.sLock.RUnlock()
	return
}

func (s *Sorter) Time() int {
	return s.currentTime
}
