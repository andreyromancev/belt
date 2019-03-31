package main

import (
	"context"
	"fmt"
	"math/rand"
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
	events := generateEvents()

	// Start worker.
	err := worker.Work(ctx, events)
	if err != nil {
		logger.Fatal(err)
	}
}

func generateEvents() <-chan belt.Event {
	events := make(chan belt.Event)
	timeChange := time.NewTicker(3 * time.Second)
	message := time.NewTicker(time.Second)

	counter := 0
	// Generate time change.
	go func() {
		for range timeChange.C {
			events <- timed.TimeChange{Time: counter + 1}
			counter++
		}
	}()

	// Generate random messages.
	go func() {
		for range message.C {
			msg := timed.Message{
				Time: counter,
			}
			switch rand.Int() % 2 {
			case 0:
				id := rand.Uint32()
				msg.Kind = "get_object"
				msg.Payload = fmt.Sprintf(`{"id": %d}`, id)
			case 1:
				msg.Kind = "save_object"
				msg.Payload = fmt.Sprintf(`{"hash": "123"}`)
			}
			events <- msg
		}
	}()

	return events
}
