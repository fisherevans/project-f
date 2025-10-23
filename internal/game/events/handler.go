package events

import (
	"reflect"

	"fisherevans.com/project/f/internal/game"
)

func init() {
	registerGojaEventHandler[EventOnInteract]("OnInteract")
	registerGojaEventHandler[EventDialogueComplete]("OnDialogueComplete")
	registerGojaEventHandler[EventChatterComplete]("OnChatterComplete")
	registerGojaEventHandler[EventTimerComplete]("OnTimerComplete")
	registerGojaEventHandler[EventEntityZoneActivity]("OnEntityZoneActivity")
	registerGojaEventHandler[EventWorldStateUpdated]("OnWorldStateUpdated")
	registerGojaEventHandler[EventCombatComplete]("OnCombatComplete")
}

type EventOnInteract struct {
	TargetId string
}

type EventDialogueComplete struct {
	DialogueId string
}

type EventChatterComplete struct {
	ChatterId string
	EntityId  string
}

type EventTimerComplete struct {
	CreatedBy       string
	TimerId         string
	DurationSeconds float64
	TriggerCount    int
}

type EventEntityZoneActivity struct {
	EntityId   string
	ZoneId     string
	IsEntering bool
}

type EventWorldStateUpdated struct {
	Key      string
	NewValue any
	OldValue any
	SetBy    string
}

type EventCombatComplete struct {
	CombatId string
	Result   game.CombatIntentResult
}

type HandlerOutput struct {
	State   any
	Effects []Effect
}

type EventHandler interface {
	Init(ctx EntityContext, world WorldStateReader, state any) *HandlerOutput
	HandleEvent(ctx EntityContext, world WorldStateReader, state any, event any) *HandlerOutput
}

type EventHandlerFunc func(ctx EntityContext, world WorldStateReader, state any, event any) *HandlerOutput

type eventRegistry struct {
	handlers map[reflect.Type]string // event type -> JS function name
}

var registry = &eventRegistry{
	handlers: make(map[reflect.Type]string),
}

func registerGojaEventHandler[T any](jsFunctionName string) {
	var zero T
	eventType := reflect.TypeOf(zero)
	registry.handlers[eventType] = jsFunctionName
}
