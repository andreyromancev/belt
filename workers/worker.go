package workers

import (
	"context"
	"fmt"
	"sync"

	"github.com/andreyromancev/belt"
	"github.com/andreyromancev/belt/log"
)

type Worker struct {
	sorter belt.Sorter
}

func NewWorker(s belt.Sorter) *Worker {
	return &Worker{
		sorter: s,
	}
}

func (w *Worker) Work(ctx context.Context, items <-chan belt.Event) error {
	var wg sync.WaitGroup
	for i := range items {
		wg.Add(1)
		go func() {
			w.process(ctx, i)
			wg.Done()
		}()
	}

	wg.Wait()
	return nil
}

func (w *Worker) process(ctx context.Context, i belt.Event) {
	logger := log.FromContext(ctx).WithField("event", fmt.Sprint(i))
	ctx = log.WithLogger(ctx, logger)
	logger.Infof("Received event")
	slot, item, err := w.sorter.Sort(ctx, i)
	if err != nil {
		logger.Error("Sorting failed:", err)
		return
	}
	if slot == nil || item == nil {
		logger.Info("Filtered by sorter")
		return
	}
	err = slot.AddItem(item)
	if err != nil {
		logger.Error("Failed to add item:", err)
	}
	w.handle(slot, item)
	logger.Info("Finished event")
}

func (w *Worker) handle(slot belt.Slot, item belt.Item) {
	m := slot.Middleware()
	ctx := item.Context()
	logger := log.FromContext(ctx).WithField("handler", fmt.Sprint(item.Handler()))
	ctx = log.WithLogger(item.Context(), logger)

	logger.Info("Handler started")
	results, err := m.Handle(item.Context(), item.Handler())
	if err != nil {
		logger.Error("Handler failed:", err)
	}
	defer func() {
		logger.Info("Handler finished")
	}()

	if len(results) == 0 {
		return
	}
	var wg sync.WaitGroup
	wg.Add(len(results))
	for _, h := range results {
		child, err := item.MakeChild(h)
		if err != nil {
			logger.Error("Failed to create child:", err)
			break
		}
		go func() {
			w.handle(slot, child)
			wg.Done()
		}()
	}
	wg.Wait()
}
