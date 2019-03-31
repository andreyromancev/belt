package timed

import (
	"fmt"
)

type Message struct {
	Time    int
	Kind    string
	Payload string
}

func (m Message) String() string {
	return m.Payload
}

type TimeChange struct {
	Time int
}

func (t TimeChange) String() string {
	return fmt.Sprintf("TimeChange{Time: %d}", t.Time)
}
