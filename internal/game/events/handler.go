package events

import (
	"fmt"
)

type EventType string

const (
	EventTypeInteract = "interact"
)

type EventHandler interface {
	Init(self, world, state ReadableObject) *HandlerOutput
	OnInteract(self, world, state, event ReadableObject) *HandlerOutput
}

func dispatchEvent(handler EventHandler, self ReadableObject, world ReadableObject, handlerState MutableObject, event Event) (*HandlerOutput, error) {
	switch event.Type {
	case EventTypeInteract:
		return handler.OnInteract(self, world, handlerState, event.Data), nil
	default:
		return nil, fmt.Errorf("unknown event type: %s", event.Type)
	}
}

type Event struct {
	Type EventType
	Data ReadableObject
}

type HandlerOutput struct {
	State   MutableObject
	Effects []Effect
}
