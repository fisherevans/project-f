// AUTO-GENERATED - DO NOT EDIT
// Generated at 2026-02-08T13:39:10-05:00 by go generate
// Source: internal/game/adventure/events_handler.go

package adventure

// BasicHandlerBuilder provides a simple way to build event handlers with type-safe state
type BasicHandlerBuilder[T any] struct {
	DefaultState func() T

	Init func(thisEntity EntityReader, globals StateGlobalsReader, state T) *HandlerOutput

	OnInteract func(thisEntity EntityReader, globals StateGlobalsReader, state T, event *EventOnInteract) *HandlerOutput
	DialogueComplete func(thisEntity EntityReader, globals StateGlobalsReader, state T, event *EventDialogueComplete) *HandlerOutput
	ChatterComplete func(thisEntity EntityReader, globals StateGlobalsReader, state T, event *EventChatterComplete) *HandlerOutput
	TimerComplete func(thisEntity EntityReader, globals StateGlobalsReader, state T, event *EventTimerComplete) *HandlerOutput
	EntityZoneActivity func(thisEntity EntityReader, globals StateGlobalsReader, state T, event *EventEntityZoneActivity) *HandlerOutput
	GlobalVariableUpdated func(thisEntity EntityReader, globals StateGlobalsReader, state T, event *EventGlobalVariableUpdated) *HandlerOutput
	CombatComplete func(thisEntity EntityReader, globals StateGlobalsReader, state T, event *EventCombatComplete) *HandlerOutput
	ScriptedMotionComplete func(thisEntity EntityReader, globals StateGlobalsReader, state T, event *EventScriptedMotionComplete) *HandlerOutput
	OnStateEnter func(thisEntity EntityReader, globals StateGlobalsReader, state T, event *EventOnStateEnter) *HandlerOutput
	Broadcast func(thisEntity EntityReader, globals StateGlobalsReader, state T, event *EventBroadcast) *HandlerOutput
}

func NewBasicHandler[T any](defaultState T) *BasicHandlerBuilder[T] {
	return &BasicHandlerBuilder[T]{
		DefaultState: func() T {
			return defaultState
		},
	}
}

func (b *BasicHandlerBuilder[T]) WithInit(init func(thisEntity EntityReader, globals StateGlobalsReader, state T) *HandlerOutput) *BasicHandlerBuilder[T] {
	b.Init = init
	return b
}

func (b *BasicHandlerBuilder[T]) WithOnInteract(onInteract func(thisEntity EntityReader, globals StateGlobalsReader, state T, event *EventOnInteract) *HandlerOutput) *BasicHandlerBuilder[T] {
	b.OnInteract = onInteract
	return b
}

func (b *BasicHandlerBuilder[T]) WithDialogueComplete(dialogueComplete func(thisEntity EntityReader, globals StateGlobalsReader, state T, event *EventDialogueComplete) *HandlerOutput) *BasicHandlerBuilder[T] {
	b.DialogueComplete = dialogueComplete
	return b
}

func (b *BasicHandlerBuilder[T]) WithChatterComplete(chatterComplete func(thisEntity EntityReader, globals StateGlobalsReader, state T, event *EventChatterComplete) *HandlerOutput) *BasicHandlerBuilder[T] {
	b.ChatterComplete = chatterComplete
	return b
}

func (b *BasicHandlerBuilder[T]) WithTimerComplete(timerComplete func(thisEntity EntityReader, globals StateGlobalsReader, state T, event *EventTimerComplete) *HandlerOutput) *BasicHandlerBuilder[T] {
	b.TimerComplete = timerComplete
	return b
}

func (b *BasicHandlerBuilder[T]) WithEntityZoneActivity(entityZoneActivity func(thisEntity EntityReader, globals StateGlobalsReader, state T, event *EventEntityZoneActivity) *HandlerOutput) *BasicHandlerBuilder[T] {
	b.EntityZoneActivity = entityZoneActivity
	return b
}

func (b *BasicHandlerBuilder[T]) WithGlobalVariableUpdated(globalVariableUpdated func(thisEntity EntityReader, globals StateGlobalsReader, state T, event *EventGlobalVariableUpdated) *HandlerOutput) *BasicHandlerBuilder[T] {
	b.GlobalVariableUpdated = globalVariableUpdated
	return b
}

func (b *BasicHandlerBuilder[T]) WithCombatComplete(combatComplete func(thisEntity EntityReader, globals StateGlobalsReader, state T, event *EventCombatComplete) *HandlerOutput) *BasicHandlerBuilder[T] {
	b.CombatComplete = combatComplete
	return b
}

func (b *BasicHandlerBuilder[T]) WithScriptedMotionComplete(scriptedMotionComplete func(thisEntity EntityReader, globals StateGlobalsReader, state T, event *EventScriptedMotionComplete) *HandlerOutput) *BasicHandlerBuilder[T] {
	b.ScriptedMotionComplete = scriptedMotionComplete
	return b
}

func (b *BasicHandlerBuilder[T]) WithOnStateEnter(onStateEnter func(thisEntity EntityReader, globals StateGlobalsReader, state T, event *EventOnStateEnter) *HandlerOutput) *BasicHandlerBuilder[T] {
	b.OnStateEnter = onStateEnter
	return b
}

func (b *BasicHandlerBuilder[T]) WithBroadcast(broadcast func(thisEntity EntityReader, globals StateGlobalsReader, state T, event *EventBroadcast) *HandlerOutput) *BasicHandlerBuilder[T] {
	b.Broadcast = broadcast
	return b
}

// CreateHandler creates an EventHandler from the builder
func (b BasicHandlerBuilder[T]) CreateHandler() EventHandler {
	return &basicHandler[T]{
		builder: b,
	}
}

type basicHandler[T any] struct {
	builder BasicHandlerBuilder[T]
}

func (h *basicHandler[T]) convertState(original any) T {
	if original == nil {
		if h.builder.DefaultState != nil {
			return h.builder.DefaultState()
		}
		var zero T
		return zero
	}
	v, ok := original.(T)
	if !ok {
		var zero T
		return zero
	}
	return v
}

func (h *basicHandler[T]) Init(thisEntity EntityReader, globals StateGlobalsReader, state any) *HandlerOutput {
	if h.builder.Init == nil {
		return nil
	}
	return h.builder.Init(thisEntity, globals, h.convertState(state))
}

func (h *basicHandler[T]) HandleEvent(thisEntity EntityReader, globals StateGlobalsReader, state any, event any) *HandlerOutput {
	convertedState := h.convertState(state)
	switch e := event.(type) {
	case *EventOnInteract:
		if h.builder.OnInteract == nil {
			return nil
		}
		return h.builder.OnInteract(thisEntity, globals, convertedState, e)
	case *EventDialogueComplete:
		if h.builder.DialogueComplete == nil {
			return nil
		}
		return h.builder.DialogueComplete(thisEntity, globals, convertedState, e)
	case *EventChatterComplete:
		if h.builder.ChatterComplete == nil {
			return nil
		}
		return h.builder.ChatterComplete(thisEntity, globals, convertedState, e)
	case *EventTimerComplete:
		if h.builder.TimerComplete == nil {
			return nil
		}
		return h.builder.TimerComplete(thisEntity, globals, convertedState, e)
	case *EventEntityZoneActivity:
		if h.builder.EntityZoneActivity == nil {
			return nil
		}
		return h.builder.EntityZoneActivity(thisEntity, globals, convertedState, e)
	case *EventGlobalVariableUpdated:
		if h.builder.GlobalVariableUpdated == nil {
			return nil
		}
		return h.builder.GlobalVariableUpdated(thisEntity, globals, convertedState, e)
	case *EventCombatComplete:
		if h.builder.CombatComplete == nil {
			return nil
		}
		return h.builder.CombatComplete(thisEntity, globals, convertedState, e)
	case *EventScriptedMotionComplete:
		if h.builder.ScriptedMotionComplete == nil {
			return nil
		}
		return h.builder.ScriptedMotionComplete(thisEntity, globals, convertedState, e)
	case *EventOnStateEnter:
		if h.builder.OnStateEnter == nil {
			return nil
		}
		return h.builder.OnStateEnter(thisEntity, globals, convertedState, e)
	case *EventBroadcast:
		if h.builder.Broadcast == nil {
			return nil
		}
		return h.builder.Broadcast(thisEntity, globals, convertedState, e)
	}
	return nil
}

// TypedEventHandler defines the typed interface for event handlers
// This interface is primarily for IDE autocomplete and documentation.
// When you embed BaseHandler[T] and start typing a method name, your IDE will
// suggest the correct signature from this interface.
//
// Note: Override TypedInit (not Init) and the typed event methods.
// BaseHandler will automatically call your TypedInit from Init.
type TypedEventHandler[T any] interface {
	TypedInit(thisEntity EntityReader, globals StateGlobalsReader, state T) *HandlerOutput
	OnInteract(thisEntity EntityReader, globals StateGlobalsReader, state T, event *EventOnInteract) *HandlerOutput
	DialogueComplete(thisEntity EntityReader, globals StateGlobalsReader, state T, event *EventDialogueComplete) *HandlerOutput
	ChatterComplete(thisEntity EntityReader, globals StateGlobalsReader, state T, event *EventChatterComplete) *HandlerOutput
	TimerComplete(thisEntity EntityReader, globals StateGlobalsReader, state T, event *EventTimerComplete) *HandlerOutput
	EntityZoneActivity(thisEntity EntityReader, globals StateGlobalsReader, state T, event *EventEntityZoneActivity) *HandlerOutput
	GlobalVariableUpdated(thisEntity EntityReader, globals StateGlobalsReader, state T, event *EventGlobalVariableUpdated) *HandlerOutput
	CombatComplete(thisEntity EntityReader, globals StateGlobalsReader, state T, event *EventCombatComplete) *HandlerOutput
	ScriptedMotionComplete(thisEntity EntityReader, globals StateGlobalsReader, state T, event *EventScriptedMotionComplete) *HandlerOutput
	OnStateEnter(thisEntity EntityReader, globals StateGlobalsReader, state T, event *EventOnStateEnter) *HandlerOutput
	Broadcast(thisEntity EntityReader, globals StateGlobalsReader, state T, event *EventBroadcast) *HandlerOutput
}

// BaseHandler provides a base implementation with no-op methods for all events
// Embed this in your custom handler and override only the methods you need
type BaseHandler[T any] struct {
	DefaultState func() T
}

func (h *BaseHandler[T]) convertState(original any) T {
	if original == nil {
		if h.DefaultState != nil {
			return h.DefaultState()
		}
		var zero T
		return zero
	}
	v, ok := original.(T)
	if !ok {
		var zero T
		return zero
	}
	return v
}

func (h *BaseHandler[T]) Init(thisEntity EntityReader, globals StateGlobalsReader, state any) *HandlerOutput {
	return h.TypedInit(thisEntity, globals, h.convertState(state))
}

func (h *BaseHandler[T]) TypedInit(thisEntity EntityReader, globals StateGlobalsReader, state T) *HandlerOutput {
	return nil
}

func (h *BaseHandler[T]) HandleEvent(thisEntity EntityReader, globals StateGlobalsReader, state any, event any) *HandlerOutput {
	convertedState := h.convertState(state)
	switch e := event.(type) {
	case *EventOnInteract:
		return h.OnInteract(thisEntity, globals, convertedState, e)
	case *EventDialogueComplete:
		return h.DialogueComplete(thisEntity, globals, convertedState, e)
	case *EventChatterComplete:
		return h.ChatterComplete(thisEntity, globals, convertedState, e)
	case *EventTimerComplete:
		return h.TimerComplete(thisEntity, globals, convertedState, e)
	case *EventEntityZoneActivity:
		return h.EntityZoneActivity(thisEntity, globals, convertedState, e)
	case *EventGlobalVariableUpdated:
		return h.GlobalVariableUpdated(thisEntity, globals, convertedState, e)
	case *EventCombatComplete:
		return h.CombatComplete(thisEntity, globals, convertedState, e)
	case *EventScriptedMotionComplete:
		return h.ScriptedMotionComplete(thisEntity, globals, convertedState, e)
	case *EventOnStateEnter:
		return h.OnStateEnter(thisEntity, globals, convertedState, e)
	case *EventBroadcast:
		return h.Broadcast(thisEntity, globals, convertedState, e)
	}
	return nil
}

func (h *BaseHandler[T]) OnInteract(thisEntity EntityReader, globals StateGlobalsReader, state T, event *EventOnInteract) *HandlerOutput {
	return nil
}

func (h *BaseHandler[T]) DialogueComplete(thisEntity EntityReader, globals StateGlobalsReader, state T, event *EventDialogueComplete) *HandlerOutput {
	return nil
}

func (h *BaseHandler[T]) ChatterComplete(thisEntity EntityReader, globals StateGlobalsReader, state T, event *EventChatterComplete) *HandlerOutput {
	return nil
}

func (h *BaseHandler[T]) TimerComplete(thisEntity EntityReader, globals StateGlobalsReader, state T, event *EventTimerComplete) *HandlerOutput {
	return nil
}

func (h *BaseHandler[T]) EntityZoneActivity(thisEntity EntityReader, globals StateGlobalsReader, state T, event *EventEntityZoneActivity) *HandlerOutput {
	return nil
}

func (h *BaseHandler[T]) GlobalVariableUpdated(thisEntity EntityReader, globals StateGlobalsReader, state T, event *EventGlobalVariableUpdated) *HandlerOutput {
	return nil
}

func (h *BaseHandler[T]) CombatComplete(thisEntity EntityReader, globals StateGlobalsReader, state T, event *EventCombatComplete) *HandlerOutput {
	return nil
}

func (h *BaseHandler[T]) ScriptedMotionComplete(thisEntity EntityReader, globals StateGlobalsReader, state T, event *EventScriptedMotionComplete) *HandlerOutput {
	return nil
}

func (h *BaseHandler[T]) OnStateEnter(thisEntity EntityReader, globals StateGlobalsReader, state T, event *EventOnStateEnter) *HandlerOutput {
	return nil
}

func (h *BaseHandler[T]) Broadcast(thisEntity EntityReader, globals StateGlobalsReader, state T, event *EventBroadcast) *HandlerOutput {
	return nil
}

