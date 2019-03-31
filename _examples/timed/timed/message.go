package timed

type Message struct {
	time    int
	kind    string
	payload string
}

type TimeChange struct {
	time int
}
