package belts

import (
	"context"

	"github.com/andreyromancev/belt"
)

type Belt struct {
	worker belt.Worker
}

func NewBelt(w belt.Worker) *Belt {
	return &Belt{
		worker: w,
	}
}

func (b *Belt) Run(ctx context.Context, events <-chan belt.Event) error {
	return b.worker.Work(ctx, events)
}
