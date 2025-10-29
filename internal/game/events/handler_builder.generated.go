// AUTO-GENERATED - DO NOT EDIT
// Generated at 2025-10-28T22:29:10-04:00 by go generate
// Source: internal/game/events/handler.go

package events

// BasicHandlerBuilder provides a simple way to build event handlers with type-safe state
type BasicHandlerBuilder[T any] struct {
	DefaultState func() T

	Init func(ctx EntityContext, world WorldStateReader, state T) *HandlerOutput

	OnInteract func(ctx EntityContext, world WorldStateReader, state T, event *EventOnInteract) *HandlerOutput
	DialogueComplete func(ctx EntityContext, world WorldStateReader, state T, event *EventDialogueComplete) *HandlerOutput
	ChatterComplete func(ctx EntityContext, world WorldStateReader, state T, event *EventChatterComplete) *HandlerOutput
	TimerComplete func(ctx EntityContext, world WorldStateReader, state T, event *EventTimerComplete) *HandlerOutput
	EntityZoneActivity func(ctx EntityContext, world WorldStateReader, state T, event *EventEntityZoneActivity) *HandlerOutput
	WorldStateUpdated func(ctx EntityContext, world WorldStateReader, state T, event *EventWorldStateUpdated) *HandlerOutput
	CombatComplete func(ctx EntityContext, world WorldStateReader, state T, event *EventCombatComplete) *HandlerOutput
	ScriptedMotionComplete func(ctx EntityContext, world WorldStateReader, state T, event *EventScriptedMotionComplete) *HandlerOutput
}

func NewBasicHandler[T any](defaultState T) *BasicHandlerBuilder[T] {
	return &BasicHandlerBuilder[T]{
		DefaultState: func() T {
			return defaultState
		},
	}
}

func (b *BasicHandlerBuilder[T]) WithInit(init func(ctx EntityContext, world WorldStateReader, state T) *HandlerOutput) *BasicHandlerBuilder[T] {
	b.Init = init
	return b
}

func (b *BasicHandlerBuilder[T]) WithOnInteract(onInteract func(ctx EntityContext, world WorldStateReader, state T, event *EventOnInteract) *HandlerOutput) *BasicHandlerBuilder[T] {
	b.OnInteract = onInteract
	return b
}

func (b *BasicHandlerBuilder[T]) WithDialogueComplete(dialogueComplete func(ctx EntityContext, world WorldStateReader, state T, event *EventDialogueComplete) *HandlerOutput) *BasicHandlerBuilder[T] {
	b.DialogueComplete = dialogueComplete
	return b
}

func (b *BasicHandlerBuilder[T]) WithChatterComplete(chatterComplete func(ctx EntityContext, world WorldStateReader, state T, event *EventChatterComplete) *HandlerOutput) *BasicHandlerBuilder[T] {
	b.ChatterComplete = chatterComplete
	return b
}

func (b *BasicHandlerBuilder[T]) WithTimerComplete(timerComplete func(ctx EntityContext, world WorldStateReader, state T, event *EventTimerComplete) *HandlerOutput) *BasicHandlerBuilder[T] {
	b.TimerComplete = timerComplete
	return b
}

func (b *BasicHandlerBuilder[T]) WithEntityZoneActivity(entityZoneActivity func(ctx EntityContext, world WorldStateReader, state T, event *EventEntityZoneActivity) *HandlerOutput) *BasicHandlerBuilder[T] {
	b.EntityZoneActivity = entityZoneActivity
	return b
}

func (b *BasicHandlerBuilder[T]) WithWorldStateUpdated(worldStateUpdated func(ctx EntityContext, world WorldStateReader, state T, event *EventWorldStateUpdated) *HandlerOutput) *BasicHandlerBuilder[T] {
	b.WorldStateUpdated = worldStateUpdated
	return b
}

func (b *BasicHandlerBuilder[T]) WithCombatComplete(combatComplete func(ctx EntityContext, world WorldStateReader, state T, event *EventCombatComplete) *HandlerOutput) *BasicHandlerBuilder[T] {
	b.CombatComplete = combatComplete
	return b
}

func (b *BasicHandlerBuilder[T]) WithScriptedMotionComplete(scriptedMotionComplete func(ctx EntityContext, world WorldStateReader, state T, event *EventScriptedMotionComplete) *HandlerOutput) *BasicHandlerBuilder[T] {
	b.ScriptedMotionComplete = scriptedMotionComplete
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

func (h *basicHandler[T]) Init(ctx EntityContext, world WorldStateReader, state any) *HandlerOutput {
	if h.builder.Init == nil {
		return nil
	}
	return h.builder.Init(ctx, world, h.convertState(state))
}

func (h *basicHandler[T]) HandleEvent(ctx EntityContext, world WorldStateReader, state any, event any) *HandlerOutput {
	convertedState := h.convertState(state)
	switch e := event.(type) {
	case *EventOnInteract:
		if h.builder.OnInteract == nil {
			return nil
		}
		return h.builder.OnInteract(ctx, world, convertedState, e)
	case *EventDialogueComplete:
		if h.builder.DialogueComplete == nil {
			return nil
		}
		return h.builder.DialogueComplete(ctx, world, convertedState, e)
	case *EventChatterComplete:
		if h.builder.ChatterComplete == nil {
			return nil
		}
		return h.builder.ChatterComplete(ctx, world, convertedState, e)
	case *EventTimerComplete:
		if h.builder.TimerComplete == nil {
			return nil
		}
		return h.builder.TimerComplete(ctx, world, convertedState, e)
	case *EventEntityZoneActivity:
		if h.builder.EntityZoneActivity == nil {
			return nil
		}
		return h.builder.EntityZoneActivity(ctx, world, convertedState, e)
	case *EventWorldStateUpdated:
		if h.builder.WorldStateUpdated == nil {
			return nil
		}
		return h.builder.WorldStateUpdated(ctx, world, convertedState, e)
	case *EventCombatComplete:
		if h.builder.CombatComplete == nil {
			return nil
		}
		return h.builder.CombatComplete(ctx, world, convertedState, e)
	case *EventScriptedMotionComplete:
		if h.builder.ScriptedMotionComplete == nil {
			return nil
		}
		return h.builder.ScriptedMotionComplete(ctx, world, convertedState, e)
	}
	return nil
}
