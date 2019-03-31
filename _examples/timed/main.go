package main

import (
	"context"
	"time"

	"github.com/andreyromancev/belt"
	"github.com/andreyromancev/belt/_examples/timed/timed"
	"github.com/andreyromancev/belt/log"
	"github.com/andreyromancev/belt/workers"
)

func main() {
	// Init components.
	logger := log.NewConsoleLogger()
	ctx := log.WithLogger(context.Background(), logger)
	sorter := timed.NewSorter(0)
	worker := workers.NewWorker(sorter)

	// Generate events.
	events, done := generateEvents(5 * time.Second)

	// Start worker.
	err := worker.Work(ctx, events)
	if err != nil {
		logger.Fatal(err)
	}

	<-done
}

func generateEvents(duration time.Duration) (<-chan belt.Event, <-chan struct{}) {
	done := make(chan struct{})
	events := make(chan belt.Event)
	ticker := time.NewTicker(1 * time.Second)

	// Periodically generate time change.
	go func() {
		counter := 1
		for range ticker.C {
			events <- timed.TimeChange{Time: counter}
			counter++
		}
	}()

	// Stop generating after duration.
	go func() {
		<-time.After(duration)
		ticker.Stop()
		close(done)
	}()

	return events, done
}
