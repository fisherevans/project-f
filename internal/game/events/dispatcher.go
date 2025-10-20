package events

import (
	"github.com/dop251/goja"
	"github.com/rs/zerolog/log"
)

type registeredEventHandler struct {
	Id      string
	self    MutableObject
	Handler EventHandler
	State   MutableObject
}

type Dispatcher struct {
	registeredHandlers []*registeredEventHandler
	queuedEvents       []Event
}

func NewDispatcher() *Dispatcher {
	return &Dispatcher{}
}

func (d *Dispatcher) NewGojaEventHandler(program *goja.Program) (EventHandler, error) {
	return NewGojaEventHandler(program)
}

func (d *Dispatcher) Register(id string, handler EventHandler, state MutableObject) {
	if state == nil {
		state = d.NewObject()
	}
	if handler == nil {
		log.Fatal().Msgf("Event handler for %s is nil", id)
	}
	self := d.NewObject()
	self.Set("id", id)
	d.registeredHandlers = append(d.registeredHandlers, &registeredEventHandler{
		Id:      id,
		self:    self,
		Handler: handler,
		State:   state,
	})
}

func (d *Dispatcher) RegisterAndInit(id string, handler EventHandler, worldState ReadableObject) {
	if handler == nil {
		log.Fatal().Msgf("Event handler for %s is nil", id)
	}
	
	self := d.NewObject()
	self.Set("id", id)
	
	// Call Init with nil state so JS can initialize it
	output := handler.Init(self, worldState, nil)
	
	state := d.NewObject()
	if output != nil && output.State != nil {
		state = output.State
	}
	
	d.registeredHandlers = append(d.registeredHandlers, &registeredEventHandler{
		Id:      id,
		self:    self,
		Handler: handler,
		State:   state,
	})
}

func (d *Dispatcher) Dispatch(event Event) {
	d.queuedEvents = append(d.queuedEvents, event)
}

func (d *Dispatcher) Flush(worldState ReadableObject) []Effect {
	var effects []Effect
	for _, event := range d.queuedEvents {
		for _, registeredHandler := range d.registeredHandlers {
			output, err := dispatchEvent(
				registeredHandler.Handler,
				registeredHandler.self,
				worldState,
				registeredHandler.State,
				event)
			if err != nil {
				log.Error().Err(err).Msgf("failed to dispatch event %s to %s", event.Type, registeredHandler.Id)
				continue
			}
			if output == nil {
				continue
			}
			// Update state if returned, otherwise keep existing state
			if output.State != nil {
				registeredHandler.State = output.State
			}
			effects = append(effects, output.Effects...)
		}
	}
	d.queuedEvents = nil
	return effects
}
