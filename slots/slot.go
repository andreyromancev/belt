package slots

import (
	"sync"

	"github.com/andreyromancev/belt"
)

type Slot struct {
	middleware belt.Middleware
	iLock      sync.RWMutex
	items      map[belt.Item]struct{}
}

func NewSlot(m belt.Middleware) *Slot {
	return &Slot{
		middleware: m,
		items:      make(map[belt.Item]struct{}),
	}
}

func (s *Slot) AddItem(i belt.Item) error {
	s.iLock.Lock()
	s.items[i] = struct{}{}
	s.iLock.Unlock()
	return nil
}

func (s *Slot) RemoveItem(i belt.Item) error {
	s.iLock.Lock()
	delete(s.items, i)
	s.iLock.Unlock()
	return nil
}

func (s *Slot) Middleware() belt.Middleware {
	return s.middleware
}

func (s *Slot) Reset(state belt.Middleware) {
	s.iLock.RLock()
	for i := range s.items {
		if c, ok := i.(belt.Canceler); ok {
			c.Cancel()
		}
	}
	s.iLock.RUnlock()
	s.middleware = state
}
