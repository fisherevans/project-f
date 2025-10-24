package handlers

import (
	"fisherevans.com/project/f/internal/game/events"
	"github.com/rs/zerolog/log"
)

var registry map[string]func() events.EventHandler

func Register(name string, factory func() events.EventHandler) {
	if registry == nil {
		registry = make(map[string]func() events.EventHandler)
	}
	if _, exists := registry[name]; exists {
		log.Fatal().Msgf("Handler %s already exists", name)
	}
	registry[name] = factory
}

func Get(name string) events.EventHandler {
	if registry == nil {
		return nil
	}
	factory, ok := registry[name]
	if !ok {
		return nil
	}
	return factory()
}

type None struct{}
