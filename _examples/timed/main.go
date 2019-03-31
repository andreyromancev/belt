package main

import (
	"context"
	"log"

	"github.com/andreyromancev/belt/_examples/timed/timed"

	"github.com/andreyromancev/belt/workers"
)

func main() {
	sorter := timed.NewSorter(0)
	worker := workers.NewWorker(sorter)
	err := worker.Work(context.Background(), nil)
	if err != nil {
		log.Fatal(err)
	}
}
