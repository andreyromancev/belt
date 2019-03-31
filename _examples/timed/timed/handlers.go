package timed

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/pkg/errors"

	"github.com/andreyromancev/belt"
)

func FutureInit(msg *Message) (belt.Handler, error) {
	// Same as present. Handlers define future functions themselves.
	return PresentInit(msg)
}

func PresentInit(msg *Message) (belt.Handler, error) {
	switch msg.kind {
	case "get_object":
		return &CheckUser{Next: &GetObject{Message: msg}}, nil
	case "save_object":
		return &CheckUser{Next: &SaveObject{Message: msg}}, nil
	default:
		return nil, errors.New("no handler for present")
	}
}

func PastInit(msg *Message) (belt.Handler, error) {
	switch msg.kind {
	case "get_object":
		return &CheckUser{Next: &GetObject{Message: msg}}, nil
	case "save_object":
		msg := Message{
			kind:    "error",
			payload: "save in the past is forbidden",
			time:    msg.time,
		}
		return &SendReply{Message: msg}, nil
	default:
		return nil, errors.New("no handler for present")
	}
}

type CheckUser struct {
	Next belt.Handler
}

func (h *CheckUser) Handle(ctx context.Context) ([]belt.Handler, error) {
	select {
	case <-ctx.Done():
		return nil, nil
	case <-simulateWork(100):
		return []belt.Handler{h.Next}, nil
	}
}

// CheckUser can be done in the future to save time for further handling.
func (h *CheckUser) Future(ctx context.Context) ([]belt.Handler, error) {
	return h.Handle(ctx)
}

type GetObject struct {
	Message *Message
}

func (h *GetObject) Handle(ctx context.Context) ([]belt.Handler, error) {
	select {
	case <-ctx.Done():
		return nil, nil
	case <-simulateWork(100): // Find object in db.
		if simulateCondition() { // We have the object.
			reply := Message{
				kind:    "object",
				payload: fmt.Sprintf("reply for: %s", h.Message.payload),
				time:    ctx.Value("time").(int),
			}
			return []belt.Handler{&SendReply{Message: reply}}, nil
		} else {
			return []belt.Handler{&Redirect{Address: "somewhere else"}}, nil
		}
	}
}

type SaveObject struct {
	Message *Message
}

func (h *SaveObject) Handle(ctx context.Context) ([]belt.Handler, error) {
	select {
	case <-ctx.Done():
		return nil, nil
	case <-simulateWork(100): // Save object to db.
		reply := Message{
			kind:    "object_id",
			payload: fmt.Sprintf("reply for: %s", h.Message.payload),
			time:    ctx.Value("time").(int),
		}
		return []belt.Handler{&SendReply{Message: reply}}, nil
	}
}

type SendReply struct {
	Message Message
}

func (h *SendReply) Handle(context.Context) ([]belt.Handler, error) {
	select {
	case <-simulateTimeout(3):
		return nil, errors.New("timeout")
	case <-simulateWork(100): // Send reply over network.
		return nil, nil
	}
}

type Redirect struct {
	Address string
}

func (h *Redirect) Handle(context.Context) ([]belt.Handler, error) {
	select {
	case <-simulateTimeout(3):
		return nil, errors.New("timeout")
	case <-simulateWork(100): // Redirect the requester.
		return nil, nil
	}
}

func simulateWork(mSec time.Duration) <-chan time.Time {
	return time.After(time.Millisecond * mSec)
}

func simulateTimeout(sec time.Duration) <-chan time.Time {
	return time.After(time.Second * sec)
}

func simulateCondition() bool {
	return rand.Int()%2 == 0
}
