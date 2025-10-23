package events

import (
	"reflect"

	"github.com/rs/zerolog/log"
)

type registeredEventHandler struct {
	ctx     EntityContext
	Handler EventHandler
	State   any
}

func (r *registeredEventHandler) handleOutput(output *HandlerOutput) []DispatchedEffect {
	if output == nil {
		return nil
	}
	if output.State != nil {
		r.State = output.State
	}
	var effects []DispatchedEffect
	for _, effect := range output.Effects {
		if err := effect.Validate(); err != nil {
			log.Err(err).
				Interface("effect", effect).
				Str("source", r.ctx.Id()).
				Msg("Invalid effect")
		} else {
			effects = append(effects, DispatchedEffect{
				Source: r.ctx,
				Effect: effect,
			})
		}
	}
	return effects
}

type Dispatcher struct {
	registeredHandlers []*registeredEventHandler
	queuedEvents       []any
}

func NewDispatcher() *Dispatcher {
	return &Dispatcher{}
}

func (d *Dispatcher) Register(ctx EntityContext, handler EventHandler) {
	if ctx == nil {
		log.Fatal().Msg("Event handler context is nil")
	}
	if handler == nil {
		log.Fatal().Msgf("Event handler for %s is nil", ctx.Id())
	}
	d.registeredHandlers = append(d.registeredHandlers, &registeredEventHandler{
		ctx:     ctx,
		Handler: handler,
		State:   nil,
	})
}

func (d *Dispatcher) Dispatch(event any) {
	d.queuedEvents = append(d.queuedEvents, event)
}

func (d *Dispatcher) Flush(worldState WorldStateReader) []DispatchedEffect {
	var effects []DispatchedEffect
	for _, event := range d.queuedEvents {
		log.Debug().Interface("event", event).Type("type", event).Msg("dispatching event")
		eventPtr := event
		if reflect.TypeOf(event).Kind() != reflect.Ptr {
			// Create a pointer to the value
			v := reflect.ValueOf(event)
			ptr := reflect.New(v.Type())
			ptr.Elem().Set(v)
			eventPtr = ptr.Interface()
		}
		for _, registeredHandler := range d.registeredHandlers {
			output := registeredHandler.Handler.HandleEvent(registeredHandler.ctx, worldState, registeredHandler.State, eventPtr)
			effects = append(effects, registeredHandler.handleOutput(output)...)
		}
	}
	d.queuedEvents = nil
	return effects
}

func (d *Dispatcher) Init(worldState WorldStateReader) []DispatchedEffect {
	var effects []DispatchedEffect
	for _, registeredHandler := range d.registeredHandlers {
		output := registeredHandler.Handler.Init(registeredHandler.ctx, worldState, registeredHandler.State)
		effects = append(effects, registeredHandler.handleOutput(output)...)
	}
	return effects
}

type DispatchedEffect struct {
	Source EntityContext
	Effect
}
