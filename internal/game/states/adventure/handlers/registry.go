package handlers

import (
	"fisherevans.com/project/f/internal/game/events"
	"github.com/rs/zerolog/log"
)

var registry map[string]func(map[string]any) events.EventHandler

func Register(name string, factory func(map[string]any) events.EventHandler) {
	if registry == nil {
		registry = make(map[string]func(map[string]any) events.EventHandler)
	}
	if _, exists := registry[name]; exists {
		log.Fatal().Msgf("Handler %s already exists", name)
	}
	registry[name] = factory
}

func Get(name string, props map[string]any) events.EventHandler {
	if registry == nil {
		return nil
	}
	if props == nil {
		props = map[string]any{}
	}
	factory, ok := registry[name]
	if !ok {
		return nil
	}
	return factory(props)
}

type None struct{}
