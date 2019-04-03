package main

import (
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"
)

var Done uint64
var Migrate uint64

var MapCounter uint64
var MapLock sync.RWMutex
var Map map[uint64]*State

const enableMap bool = true

type Handle func(Controller)

type Controller interface {
	YieldFirst(migrate Handle, first Adapter, rest ...Adapter)
	YieldAll(migrate Handle, first Adapter, rest ...Adapter)
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
	<-time.After(time.Duration(rand.Int()%2000) * time.Millisecond)
	a.result = "[adapted]" + a.request
}

type State struct {
	id       uint64
	stash    []*adapter
	fatState []byte
}

func (s *State) Future(c Controller) {
	c.Wait(s.Present)
}

func (s *State) Present(c Controller) {
	// fmt.Println("started")
	fatStack := make([]byte, 10*1000)
	s.fatState = make([]byte, 10*1000)
	var nostash []*adapter
	s.stash = append(s.stash, &adapter{request: "test"})
	s.stash = append(s.stash, &adapter{request: "test"})

	nostash = append(nostash, &adapter{request: "test"})
	nostash = append(nostash, &adapter{request: "test"})
	nostash = append(nostash, &adapter{request: "test"})
	nostash = append(nostash, &adapter{request: "test"})
	fatAdapter := &adapter{request: string(fatStack)}
	c.YieldAll(s.MigrateToPast, s.stash[0], nostash[2])

	if rand.Int()%2 == 0 {
		c.YieldAll(s.MigrateToPast, fatAdapter)
	} else {
		c.Wait(s.Past)
	}
	// fmt.Println("present result: ", s.keep.result, " ", discard.result)

	s.Done()
}

func (s *State) Past(c Controller) {
	c.YieldAll(s.MigrateToPast, s.stash[0])
	s.Done()
}

func (s *State) MigrateToPast(c Controller) {
	// fmt.Println("migrated to past")
	c.YieldAll(nil, s.stash[1])
	// fmt.Println("past result: ", s.keep.result)
	s.Done()
}

func (s *State) Done() {
	if enableMap {
		MapLock.RLock()
		me, ok := Map[s.id]
		_ = me
		_ = ok
		MapLock.RUnlock()
	}
	atomic.AddUint64(&Done, 1)
}

func worker(cancel <-chan struct{}) {
	init := &State{}
	if enableMap {
		MapLock.Lock()
		MapCounter++
		init.id = MapCounter
		Map[MapCounter] = init
		MapLock.Unlock()
	}

	c := &controller{cancel: cancel, adapters: map[Adapter]chan struct{}{}}
	// Switch by pulse.
	handle(init.Present, c)

	if enableMap {
		MapLock.Lock()
		delete(Map, init.id)
		MapLock.Unlock()
	}
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
	Map = map[uint64]*State{}
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
		<-time.After(5 * time.Microsecond)
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
