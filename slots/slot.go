package slots

import (
	"sync"

	"github.com/andreyromancev/belt"
)

type Slot struct {
	middleware belt.Middleware
	iLock      sync.RWMutex
	items      []belt.Item
}

func NewSlot(m belt.Middleware) *Slot {
	return &Slot{middleware: m}
}

func (s *Slot) AddItem(i belt.Item) error {
	s.iLock.Lock()
	s.items = append(s.items, i)
	s.iLock.Unlock()
	return nil
}

func (s *Slot) Middleware() belt.Middleware {
	return s.middleware
}

func (s *Slot) Reset(state belt.Middleware) {
	s.iLock.RLock()
	for _, i := range s.items {
		if c, ok := i.(belt.Canceler); ok {
			c.Cancel()
		}
	}
	s.iLock.RUnlock()
	s.middleware = state
}
