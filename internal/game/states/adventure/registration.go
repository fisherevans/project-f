package adventure

import (
	"fisherevans.com/project/f/internal/resources"
	"fisherevans.com/project/f/internal/util"
	"github.com/rs/zerolog/log"
)

type None struct{}

type entityRegistrar func(NewEntityParams, *EntitySystem) (Entity, EventHandler)

var registrarsByClass = map[string]entityRegistrar{}
var registrarsEntitiesByTile = map[resources.TilesheetSpriteId]entityRegistrar{}

type entityRegistrarTargets struct {
	classes []string
	tiles   []resources.TilesheetSpriteId
}

func newRegistrarBuilder() entityRegistrarTargets {
	return entityRegistrarTargets{}
}

func (r entityRegistrarTargets) byClass(classes ...string) entityRegistrarTargets {
	r.classes = append(r.classes, classes...)
	return r
}

func (r entityRegistrarTargets) byTile(tiles ...resources.TilesheetSpriteId) entityRegistrarTargets {
	r.tiles = append(r.tiles, tiles...)
	return r
}

func (r entityRegistrarTargets) registrar(registrar entityRegistrar) {
	for _, class := range r.classes {
		if _, exists := registrarsByClass[class]; exists {
			log.Fatal().Str("class", class).Msg("duplicate entity class")
		}
		registrarsByClass[class] = registrar
	}
	for _, tile := range r.tiles {
		if _, exists := registrarsEntitiesByTile[tile]; exists {
			log.Fatal().Interface("tile", tile).Msg("duplicate entity tile")
		}
		registrarsEntitiesByTile[tile] = registrar
	}
}

var (
	entityPropertyTemplates = map[string]map[string]any{}
)

func registerPropertyTemplate(key string, props map[string]any) {
	if _, ok := entityPropertyTemplates[key]; ok {
		log.Fatal().Str("key", key).Msg("duplicate property template")
	}
	entityPropertyTemplates[key] = props
}

type NewEntityParams struct {
	EntityId   string
	Class      string
	SpriteId   *resources.TilesheetSpriteId
	Location   MapLocation
	Properties *util.Properties
}

func (s *State) registerParameterizedEntity(params NewEntityParams) bool {
	var registerer entityRegistrar
	var exists bool

	if params.SpriteId != nil {
		registerer, exists = registrarsEntitiesByTile[*params.SpriteId]
	}
	if !exists {
		registerer, exists = registrarsByClass[params.Class]
	}
	if registerer == nil {
		return false
	}

	entity, eventHandler := registerer(params, s.entities)
	if entity == nil {
		if eventHandler != nil {
			log.Fatal().Str("entityId", string(params.EntityId)).Msgf("entity was not created but an event handler was")
		}
		return false
	}
	if s.entities.GetDebugType(params.EntityId) == "" && params.Class != "" {
		s.entities.SetDebugType(params.EntityId, params.Class)
	}
	log.Debug().Msgf("Registered entity %s", params.EntityId)

	metadata, ok := params.Properties.Get("metadata")
	if ok {
		s.entities.loadGenericMetadata(entity.GetId(), metadata)
	}

	if scriptRef := params.Properties.GetString("script_ref", ""); scriptRef != "" {
		if eventHandler != nil {
			log.Fatal().Str("entityId", string(params.EntityId)).Msgf("entity has more than one handler configured!")
		}
		eventHandler = getEventHandler(scriptRef, params.Properties)
		if eventHandler == nil {
			log.Fatal().Str("entityId", string(params.EntityId)).Str("script_ref", scriptRef).Msg("script_ref does not match any registered handler")
		}
		s.entities.SetHandlerRef(params.EntityId, scriptRef)
	}

	if eventHandler != nil {
		s.eventDispatcher.Register(entity, eventHandler)
		log.Debug().Msgf("Registered event handler for %s", params.EntityId)
	}

	return true
}

var eventHandlerRegistry map[string]func(properties *util.Properties) EventHandler

func registerEventHandler(name string, factory func(*util.Properties) EventHandler) {
	if eventHandlerRegistry == nil {
		eventHandlerRegistry = make(map[string]func(properties *util.Properties) EventHandler)
	}
	if _, exists := eventHandlerRegistry[name]; exists {
		log.Fatal().Msgf("Handler %s already exists", name)
	}
	eventHandlerRegistry[name] = factory
}

func getEventHandler(name string, props *util.Properties) EventHandler {
	if eventHandlerRegistry == nil {
		return nil
	}
	factory, ok := eventHandlerRegistry[name]
	if !ok {
		return nil
	}
	return factory(props)
}
