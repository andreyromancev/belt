package workers

import (
	"context"
	"fmt"

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
	for i := range items {
		go w.process(ctx, i)
	}
	return nil
}

func (w *Worker) process(ctx context.Context, i belt.Event) {
	logger := log.FromContext(ctx).WithField("event", fmt.Sprint(i))
	ctx = log.WithLogger(ctx, logger)
	logger.Infof("Received event")
	slot, item, err := w.sorter.Sort(ctx, i)
	if err != nil {
		logger.Error("Sorting failed: ", err)
		return
	}
	if slot == nil || item == nil {
		logger.Info("Filtered by sorter")
		return
	}
	handle(slot, item)
}

func handle(slot belt.Slot, item belt.Item) {
	m := slot.Middleware()
	results, err := m.Handle(item.Context(), item.Handler())
	if err != nil {
		log.FromContext(item.Context()).Error("Handling failed: ", err)
	}
	for _, h := range results {
		go handle(slot, item.MakeChild(h))
	}
}
