package main

import (
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"
)

var Done uint64
var Migrate uint64

type Handle func(Controller)

type Controller interface {
	YieldFirst(migrate Handle, first Adapter, rest ...Adapter)
	YieldAll(migrate Handle, first Adapter, rest ...Adapter) error
	Wait(migrate Handle)
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

func (c *controller) Wait(migrate Handle) {
	c.migrateTo = migrate
	<-c.cancel
	panic(cancelPanic{})
}

func (c *controller) YieldAll(migrate Handle, first Adapter, rest ...Adapter) error {
	c.migrateTo = migrate

	all := append(rest, first)
	var wg sync.WaitGroup
	wg.Add(len(all))
	for _, a := range all {
		a := a
		if d, ok := c.adapters[a]; ok {
			<-d
			wg.Done()
			return nil
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
		return errors.New("interrupted")
	case <-done:
	}
	return nil
}

type Adapter interface {
	Adapt()
}

type adapter struct {
	request string
	result  string
}

func (a *adapter) Adapt() {
	<-time.After(time.Duration(rand.Int()%2000) * time.Millisecond)
	a.result = "[adapted]" + a.request
}

type State struct {
	stash []*adapter
}

func (s *State) Future(c Controller) {
	c.Wait(s.Present)
}

func (s *State) Present(c Controller) {
	// fmt.Println("started")
	fatStack := make([]byte, 100*1000)
	var nostash []*adapter
	s.stash = append(s.stash, &adapter{request: "test"})
	s.stash = append(s.stash, &adapter{request: "test"})

	nostash = append(nostash, &adapter{request: "test"})
	nostash = append(nostash, &adapter{request: "test"})
	nostash = append(nostash, &adapter{request: "test"})
	nostash = append(nostash, &adapter{request: "test"})
	fatAdapter := &adapter{request: string(fatStack)}
	err := c.YieldAll(s.MigrateToPast, s.stash[0], nostash[2])
	if err != nil {
		return
	}

	if rand.Int()%2 == 0 {
		err = c.YieldAll(s.MigrateToPast, s.stash[1])
		if err != nil {
			return
		}
	} else {
		err = c.YieldAll(s.Past, fatAdapter)
		if err != nil {
			return
		}
		// c.Wait(s.Past)
	}
	// fmt.Println("present result: ", s.keep.result, " ", discard.result)

	s.Done()
}

func (s *State) Past(c Controller) {
	err := c.YieldAll(s.MigrateToPast, s.stash[0])
	if err != nil {
		s.Done()
		return
	}
	s.Done()
}

func (s *State) MigrateToPast(c Controller) {
	// fmt.Println("migrated to past")
	err := c.YieldAll(nil, s.stash[1])
	if err != nil {
		s.Done()
		return
	}
	s.Done()
	// fmt.Println("past result: ", s.keep.result)
}

func (s *State) Done() {
	atomic.AddUint64(&Done, 1)
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
				atomic.AddUint64(&Migrate, 1)
				handle(c.migrateTo, c)
			} else {
				panic(r)
			}
		}
	}()
	h(c)
}

func main() {
	const (
		pulseTime  = 10
		pulseCount = 5
	)
	pulseCounter := pulseCount
	pulseTicker := time.NewTicker(pulseTime * time.Second)
	cancel := make(chan struct{})
	go func() {
		for range pulseTicker.C {
			if pulseCounter < 0 {
				return
			}
			close(cancel)
			cancel = make(chan struct{})
			pulseCounter--
			fmt.Println(pulseCounter)
		}
	}()

	start := time.Now()
	for pulseCounter > 0 {
		time.After(10 * time.Millisecond)
		go worker(cancel)
	}
	t := time.Now()
	elapsed := t.Sub(start)

	fmt.Printf(
		"In %d seconds were processed %d requests (%d per second). Migrated: %d requests (%d per second)",
		int(elapsed.Seconds()),
		Done,
		int(float64(Done)/elapsed.Seconds()),
		Migrate,
		int(float64(Migrate)/elapsed.Seconds()),
	)
}
