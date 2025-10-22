package events

import (
	"github.com/dop251/goja"
	"github.com/rs/zerolog/log"
)

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

type HandlerOutput struct {
	State   any
	Effects []Effect
}

type EventHandler interface {
	Init(ctx EntityContext, world WorldStateReader, state any) *HandlerOutput
	OnInteract(ctx EntityContext, world WorldStateReader, state any, event *EventOnInteract) *HandlerOutput
	OnDialogueComplete(ctx EntityContext, world WorldStateReader, state any, event *EventDialogueComplete) *HandlerOutput
	OnChatterComplete(ctx EntityContext, world WorldStateReader, state any, event *EventChatterComplete) *HandlerOutput
	OnTimerComplete(ctx EntityContext, world WorldStateReader, state any, event *EventTimerComplete) *HandlerOutput
	OnEntityZoneActivity(ctx EntityContext, world WorldStateReader, state any, event *EventEntityZoneActivity) *HandlerOutput
}

func dispatchEvent(handler EventHandler, ctx EntityContext, world WorldStateReader, handlerState any, event any) *HandlerOutput {
	switch e := event.(type) {
	case *EventOnInteract:
		return handler.OnInteract(ctx, world, handlerState, e)
	case *EventDialogueComplete:
		return handler.OnDialogueComplete(ctx, world, handlerState, e)
	case *EventChatterComplete:
		return handler.OnChatterComplete(ctx, world, handlerState, e)
	case *EventTimerComplete:
		return handler.OnTimerComplete(ctx, world, handlerState, e)
	case *EventEntityZoneActivity:
		return handler.OnEntityZoneActivity(ctx, world, handlerState, e)
	default:
		log.Error().Msgf("unknown event type: %T", event)
		return nil
	}
}

var gojaHandlerMapping = map[string]func(*gojaEventHandler, goja.Callable){
	"OnInteract": func(handler *gojaEventHandler, fn goja.Callable) {
		handler.fnOnInteract = wrapHandlerFn[EventOnInteract](handler, fn)
	},
	"OnDialogueComplete": func(handler *gojaEventHandler, fn goja.Callable) {
		handler.fnOnDialogueComplete = wrapHandlerFn[EventDialogueComplete](handler, fn)
	},
	"OnChatterComplete": func(handler *gojaEventHandler, fn goja.Callable) {
		handler.fnOnChatterComplete = wrapHandlerFn[EventChatterComplete](handler, fn)
	},
	"OnTimerComplete": func(handler *gojaEventHandler, fn goja.Callable) {
		handler.fnOnTimerComplete = wrapHandlerFn[EventTimerComplete](handler, fn)
	},
	"OnEntityZoneActivity": func(handler *gojaEventHandler, fn goja.Callable) {
		handler.fnOnEntityZoneActivity = wrapHandlerFn[EventEntityZoneActivity](handler, fn)
	},
}

type BaseEventHandler struct{}

func (b BaseEventHandler) Init(EntityContext, WorldStateReader, any) *HandlerOutput {
	return nil
}

func (b BaseEventHandler) OnInteract(EntityContext, WorldStateReader, any, *EventOnInteract) *HandlerOutput {
	return nil
}

func (b BaseEventHandler) OnDialogueComplete(EntityContext, WorldStateReader, any, *EventDialogueComplete) *HandlerOutput {
	return nil
}

func (b BaseEventHandler) OnChatterComplete(EntityContext, WorldStateReader, any, *EventChatterComplete) *HandlerOutput {
	return nil
}

func (b BaseEventHandler) OnTimerComplete(EntityContext, WorldStateReader, any, *EventTimerComplete) *HandlerOutput {
	return nil
}

func (b BaseEventHandler) OnEntityZoneActivity(EntityContext, WorldStateReader, any, *EventEntityZoneActivity) *HandlerOutput {
	return nil
}
