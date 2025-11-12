package adventure

import (
	"reflect"

	"fisherevans.com/project/f/internal/game/rpg"
	"github.com/rs/zerolog/log"
)

type registeredEventHandler struct {
	entity  EntityReader
	Handler EventHandler
	State   any
}

type Dispatcher struct {
	globals            rpg.GlobalsReader
	effectDispatcher   EffectDispatcher
	registeredHandlers map[string]*registeredEventHandler
}

func NewDispatcher(globals rpg.GlobalsReader, effectDispatcher EffectDispatcher) *Dispatcher {
	return &Dispatcher{
		globals:            globals,
		effectDispatcher:   effectDispatcher,
		registeredHandlers: make(map[string]*registeredEventHandler),
	}
}

func (d *Dispatcher) Register(entity EntityReader, handler EventHandler) {
	if entity == nil {
		log.Fatal().Msg("Event handler context is nil")
	}
	if handler == nil {
		log.Fatal().Msgf("Event handler for %s is nil", entity.GetId())
	}
	if _, exists := d.registeredHandlers[entity.GetId()]; exists {
		log.Fatal().Msgf("Event handler for %s is already registered", entity.GetId())
	}
	d.registeredHandlers[entity.GetId()] = &registeredEventHandler{
		entity:  entity,
		Handler: handler,
		State:   nil,
	}
}

func (d *Dispatcher) GetHandler(entity EntityReader) (EventHandler, bool) {
	rh, ok := d.registeredHandlers[entity.GetId()]
	if ok {
		return rh.Handler, true
	}
	return nil, false
}

func (d *Dispatcher) Unregister(entity EntityReader) {
	delete(d.registeredHandlers, entity.GetId())
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
			output := registeredHandler.Handler.HandleEvent(registeredHandler.entity, d.globals, registeredHandler.State, eventPtr)
			d.handleOutput(registeredHandler, output)
		}
	}
}

func (d *Dispatcher) Init() {
	for _, registeredHandler := range d.registeredHandlers {
		output := registeredHandler.Handler.Init(registeredHandler.entity, d.globals, registeredHandler.State)
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
			Source: handler.entity,
			Effect: effect,
		})
	}
}

type DispatchedEffect struct {
	Source EntityReader
	Effect Effect // Effect is now an interface
}
