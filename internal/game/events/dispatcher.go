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
				Str("source", r.ctx.EntityId()).
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
	worldState         WorldState
	effectDispatcher   EffectDispatcher
	registeredHandlers []*registeredEventHandler
}

func NewDispatcher(worldState WorldState, effectDispatcher EffectDispatcher) *Dispatcher {
	return &Dispatcher{
		worldState:       worldState,
		effectDispatcher: effectDispatcher,
	}
}

func (d *Dispatcher) Register(ctx EntityContext, handler EventHandler) {
	if ctx == nil {
		log.Fatal().Msg("Event handler context is nil")
	}
	if handler == nil {
		log.Fatal().Msgf("Event handler for %s is nil", ctx.EntityId())
	}
	d.registeredHandlers = append(d.registeredHandlers, &registeredEventHandler{
		ctx:     ctx,
		Handler: handler,
		State:   nil,
	})
}

func (d *Dispatcher) Dispatch(events ...any) {
	for _, event := range events {
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
			output := registeredHandler.Handler.HandleEvent(registeredHandler.ctx, d.worldState, registeredHandler.State, eventPtr)
			d.handleOutput(registeredHandler, output)
		}
	}
}

func (d *Dispatcher) Init(worldState WorldStateReader) {
	for _, registeredHandler := range d.registeredHandlers {
		output := registeredHandler.Handler.Init(registeredHandler.ctx, worldState, registeredHandler.State)
		d.handleOutput(registeredHandler, output)
	}
}

func (d *Dispatcher) handleOutput(handler *registeredEventHandler, output *HandlerOutput) {
	if output == nil {
		return
	}
	if output.State != nil {
		handler.State = output.State
	}
	for _, effect := range output.Effects {
		if err := effect.Validate(); err != nil {
			log.Error().Interface("effect", effect).Msg("Invalid effect, ignoring")
			continue
		}
		d.effectDispatcher(DispatchedEffect{
			Source: handler.ctx,
			Effect: effect,
		})
	}
}

type DispatchedEffect struct {
	Source EntityContext
	Effect
}
