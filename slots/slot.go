package slots

import (
	"github.com/andreyromancev/belt"
	"sync"
)

type Slot struct {
	state belt.Middleware
	iLock sync.RWMutex
	items []belt.Item
}

func NewSlot() *Slot {
	return &Slot{}
}

func (s *Slot) AddItem(i belt.Item) error {
	s.iLock.Lock()
	s.items = append(s.items, i)
	s.iLock.Unlock()
	return nil
}

func (s *Slot) Middleware() belt.Middleware {
	return s.state
}

func (s *Slot) SetMiddleware(state belt.Middleware) error {
	s.iLock.RLock()
	for _, i := range s.items {
		if c, ok := i.(belt.Canceler); ok {
			c.Cancel()
		}
	}
	s.iLock.RUnlock()
	s.state = state
	return nil
}
