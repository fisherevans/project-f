package adventure

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
		if err := effect.FillDefaultsAndValidate(); err != nil {
			log.Err(err).
				Interface("effect", effect).
				Str("source", r.ctx.EntityId()).
				Msg("Effect validation failed, ignoring")
			continue
		}
		effects = append(effects, DispatchedEffect{
			Source: r.ctx,
			Effect: effect,
		})
	}
	return effects
}

type Dispatcher struct {
	gameState          GameState
	effectDispatcher   EffectDispatcher
	registeredHandlers map[string]*registeredEventHandler
}

func NewDispatcher(gameState GameState, effectDispatcher EffectDispatcher) *Dispatcher {
	return &Dispatcher{
		gameState:          gameState,
		effectDispatcher:   effectDispatcher,
		registeredHandlers: make(map[string]*registeredEventHandler),
	}
}

func (d *Dispatcher) Register(ctx EntityContext, handler EventHandler) {
	if ctx == nil {
		log.Fatal().Msg("Event handler context is nil")
	}
	if handler == nil {
		log.Fatal().Msgf("Event handler for %s is nil", ctx.EntityId())
	}
	if _, exists := d.registeredHandlers[ctx.EntityId()]; exists {
		log.Fatal().Msgf("Event handler for %s is already registered", ctx.EntityId())
	}
	d.registeredHandlers[ctx.EntityId()] = &registeredEventHandler{
		ctx:     ctx,
		Handler: handler,
		State:   nil,
	}
}

func (d *Dispatcher) GetHandler(ctx EntityContext) (EventHandler, bool) {
	rh, ok := d.registeredHandlers[ctx.EntityId()]
	if ok {
		return rh.Handler, true
	}
	return nil, false
}

func (d *Dispatcher) Unregister(ctx EntityContext) {
	delete(d.registeredHandlers, ctx.EntityId())
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
			output := registeredHandler.Handler.HandleEvent(registeredHandler.ctx, d.gameState, registeredHandler.State, eventPtr)
			d.handleOutput(registeredHandler, output)
		}
	}
}

func (d *Dispatcher) Init() {
	for _, registeredHandler := range d.registeredHandlers {
		output := registeredHandler.Handler.Init(registeredHandler.ctx, d.gameState, registeredHandler.State)
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
		if err := effect.FillDefaultsAndValidate(); err != nil {
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
	Effect Effect // Effect is now an interface
}
