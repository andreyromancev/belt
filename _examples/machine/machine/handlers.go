package machine

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/andreyromancev/belt/log"

	"github.com/pkg/errors"

	"github.com/andreyromancev/belt"
)

func FutureInit(msg Message) (belt.Handler, error) {
	// Same as present. Handlers define future functions themselves.
	return PresentInit(msg)
}

func PresentInit(msg Message) (belt.Handler, error) {
	switch msg.Kind {
	case "get_object":
		return &CheckUser{Next: &GetObject{Message: msg}}, nil
	case "save_object":
		return &CheckUser{Next: &SaveObject{Message: msg}}, nil
	default:
		return nil, errors.New("no handler for present")
	}
}

func PastInit(msg Message) (belt.Handler, error) {
	switch msg.Kind {
	case "get_object":
		return &CheckUser{Next: &GetObject{Message: msg}}, nil
	case "save_object":
		msg := Message{
			Kind:    "error",
			Payload: "save in the past is forbidden",
			Time:    msg.Time,
		}
		return &SendReply{Message: msg}, nil
	default:
		return nil, errors.New("no handler for present")
	}
}

type CheckUser struct {
	Next belt.Handler
}

func (CheckUser) String() string {
	return "CheckUser"
}

func (h *CheckUser) Handle(ctx context.Context) ([]belt.Handler, error) {
	select {
	case <-ctx.Done():
		return nil, nil
	case <-simulateWork():
		return []belt.Handler{h.Next}, nil
	}
}

// CheckUser can be done in the future to save Time for further handling.
func (h *CheckUser) Future(ctx context.Context) ([]belt.Handler, error) {
	return h.Handle(ctx)
}

type GetObject struct {
	Message Message
}

func (GetObject) String() string {
	return "GetObject"
}

func (h *GetObject) Handle(ctx context.Context) ([]belt.Handler, error) {
	select {
	case <-ctx.Done():
		return nil, nil
	case <-simulateWork(): // Find object in db.
		if simulateCondition() { // We have the object.
			reply := Message{
				Kind:    "object",
				Payload: fmt.Sprintf("reply for: %s", h.Message.Payload),
				Time:    ctx.Value("Time").(int),
			}
			return []belt.Handler{&SendReply{Message: reply}}, nil
		} else {
			return []belt.Handler{&Redirect{Address: "somewhere else"}}, nil
		}
	}
}

type SaveObject struct {
	Message Message
}

func (SaveObject) String() string {
	return "SaveObject"
}

func (h *SaveObject) Handle(ctx context.Context) ([]belt.Handler, error) {
	select {
	case <-ctx.Done():
		return nil, nil
	case <-simulateWork(): // Save object to db.
		reply := Message{
			Kind:    "object_id",
			Payload: fmt.Sprintf("reply for: %s", h.Message.Payload),
			Time:    ctx.Value("Time").(int),
		}
		return []belt.Handler{&SendReply{Message: reply}}, nil
	}
}

type SendReply struct {
	Message Message
}

func (SendReply) String() string {
	return "SendReply"
}

func (h *SendReply) Handle(ctx context.Context) ([]belt.Handler, error) {
	select {
	case <-simulateTimeout():
		return nil, errors.New("timeout")
	case <-simulateWork(): // Send reply over network.
		log.FromContext(ctx).Info("Replied")
		return nil, nil
	}
}

type Redirect struct {
	Address string
}

func (Redirect) String() string {
	return "Redirect"
}

func (h *Redirect) Handle(ctx context.Context) ([]belt.Handler, error) {
	select {
	case <-simulateTimeout():
		return nil, errors.New("timeout")
	case <-simulateWork(): // Redirect the requester.
		log.FromContext(ctx).Info("Redirected")
		return nil, nil
	}
}

func simulateWork() <-chan time.Time {
	factor := rand.Uint32() % 10
	return time.After(time.Duration(100*factor) * time.Millisecond)
}

func simulateTimeout() <-chan time.Time {
	factor := rand.Uint32() % 5
	return time.After(time.Duration(100*factor) * time.Second)
}

func simulateCondition() bool {
	return rand.Int()%2 == 0
}
