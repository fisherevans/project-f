package adventure

import (
	"fisherevans.com/project/f/internal/util"
	"github.com/rs/zerolog/log"
)

var registry map[string]func(properties *util.Properties) EventHandler

func registerHandlerReference(name string, factory func(*util.Properties) EventHandler) {
	if registry == nil {
		registry = make(map[string]func(properties *util.Properties) EventHandler)
	}
	if _, exists := registry[name]; exists {
		log.Fatal().Msgf("Handler %s already exists", name)
	}
	registry[name] = factory
}

func Get(name string, props *util.Properties) EventHandler {
	if registry == nil {
		return nil
	}
	factory, ok := registry[name]
	if !ok {
		return nil
	}
	return factory(props)
}
