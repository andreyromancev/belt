package belts

import (
	"context"
	"github.com/andreyromancev/belt"
)

type Belt struct {
	worker belt.Worker
	slots []belt.Slot
}

func NewBelt(w belt.Worker) *Belt {
	return &Belt{
		worker: w,
	}
}

func (b *Belt) AddSlot(slot belt.Slot) error {
	b.slots = append(b.slots, slot)
	return nil
}

func (b *Belt) Run(ctx context.Context, events <-chan belt.Event) error {
	return b.worker.Work(ctx, events)
}
