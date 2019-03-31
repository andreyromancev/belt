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
	return fmt.Sprintf("TimeChange{Time: %d, Kind: %s, Payload: %s}", m.Time, m.Kind, m.Payload)
}

type TimeChange struct {
	Time int
}

func (t TimeChange) String() string {
	return fmt.Sprintf("TimeChange{Time: %d}", t.Time)
}
