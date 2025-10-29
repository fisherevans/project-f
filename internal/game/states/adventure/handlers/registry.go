package handlers

import (
	"fisherevans.com/project/f/internal/game/events"
	"fisherevans.com/project/f/internal/util"
	"github.com/rs/zerolog/log"
)

var registry map[string]func(properties *util.Properties) events.EventHandler

func Register(name string, factory func(*util.Properties) events.EventHandler) {
	if registry == nil {
		registry = make(map[string]func(properties *util.Properties) events.EventHandler)
	}
	if _, exists := registry[name]; exists {
		log.Fatal().Msgf("Handler %s already exists", name)
	}
	registry[name] = factory
}

func Get(name string, props *util.Properties) events.EventHandler {
	if registry == nil {
		return nil
	}
	factory, ok := registry[name]
	if !ok {
		return nil
	}
	return factory(props)
}

type None struct{}
