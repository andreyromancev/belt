package workers

import (
	"context"
	"github.com/andreyromancev/belt"
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
	slot, item, _ := w.sorter.Sort(i)
	item.SetContext(ctx)
	// TODO: log errors
	_ = handle(slot, item)
}

func handle(slot belt.Slot, item belt.Item) error {
	m := slot.Middleware()
	results, err := m.Handle(item.Context(), item.Handler())
	if err != nil {
		return err
	}
	for _, h := range results {
		err := handle(slot, item.MakeChild(h))
		if err != nil {
			return err
		}
	}

	return nil
}
