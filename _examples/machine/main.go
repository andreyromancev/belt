package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

type Handle func(Controller)

type Controller interface {
	YieldFirst(migrate Handle, first Adapter, rest ...Adapter)
	YieldAll(migrate Handle, first Adapter, rest ...Adapter)
}

type controller struct {
	cancel    <-chan struct{}
	migrateTo Handle
	adapters  map[Adapter]chan struct{}
}

type cancelPanic struct{}

func (c *controller) YieldFirst(migrate Handle, first Adapter, rest ...Adapter) {
	panic("implement me")
}

func (c *controller) YieldAll(migrate Handle, first Adapter, rest ...Adapter) {
	c.migrateTo = migrate

	all := append(rest, first)
	var wg sync.WaitGroup
	wg.Add(len(all))
	for _, a := range all {
		a := a
		if d, ok := c.adapters[a]; ok {
			<-d
			wg.Done()
			return
		}

		done := make(chan struct{})
		c.adapters[a] = done
		go func() {
			a.Adapt()
			close(done)
			wg.Done()
		}()
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-c.cancel:
		panic(cancelPanic{})
	case <-done:
	}
}

type Adapter interface {
	Adapt()
}

type adapter struct {
	request string
	result  string
}

func (a *adapter) Adapt() {
	<-time.After(1000 * time.Millisecond)
	a.result = "[adapted]" + a.request
}

type State struct {
	keep *adapter
}

func (s *State) Future(c Controller) {}

func (s *State) Present(c Controller) {
	fmt.Println("started")
	s.keep = &adapter{request: "keep"}
	discard := &adapter{request: "discard"}
	c.YieldAll(s.MigrateToPast, s.keep, discard)
	fmt.Println("present result: ", s.keep.result, " ", discard.result)
}

func (s *State) Past(c Controller) {}

func (s *State) MigrateToPast(c Controller) {
	fmt.Println("migrated to past")
	c.YieldAll(nil, s.keep)
	fmt.Println("past result: ", s.keep.result)
}

func worker(cancel <-chan struct{}) {
	c := &controller{cancel: cancel, adapters: map[Adapter]chan struct{}{}}
	init := &State{}
	// Switch by pulse.
	handle(init.Present, c)
}

func handle(h Handle, c *controller) {
	defer func() {
		if r := recover(); r != nil && c.migrateTo != nil {
			if _, ok := r.(cancelPanic); ok {
				handle(c.migrateTo, c)
			} else {
				panic(r)
			}
		}
	}()
	h(c)
}

func main() {
	for {
		cancel := make(chan struct{})
		go worker(cancel)
		<-time.After(time.Duration(rand.Int()%1700) * time.Millisecond)
		close(cancel)
		<-time.After(time.Duration(rand.Int()%3000) * time.Millisecond)
	}
}
