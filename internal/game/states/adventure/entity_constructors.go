package adventure

import (
	"fisherevans.com/project/f/internal/game/events"
	"fisherevans.com/project/f/internal/game/states/adventure/handlers"
	"fisherevans.com/project/f/internal/resources"
	"github.com/rs/zerolog/log"
)

type None struct{}

func init() {
	registerDynamicEntity().byClass("Script").register(constructScriptEntity)
}

func constructScriptEntity(s *State, entityId EntityId, location MapLocation, mapEntity *resources.Entity) (Entity, events.EventHandler) {
	e := NewDynamicEntity(entityId, location)
	return e, nil
}

type entityConstructor func(s *State, entityId EntityId, location MapLocation, mapEntity *resources.Entity) (Entity, events.EventHandler)
var dynamicEntitiesByClass = map[string]entityConstructor{}
var dynamicEntitiesByTile = map[resources.TilesheetSpriteId]entityConstructor{}

type dynamicEntityRegistrar struct {
	classes []string
	tiles   []resources.TilesheetSpriteId
}

func registerDynamicEntity() dynamicEntityRegistrar {
	return dynamicEntityRegistrar{}
}

func (r dynamicEntityRegistrar) byClass(classes ...string) dynamicEntityRegistrar {
	r.classes = append(r.classes, classes...)
	return r
}

func (r dynamicEntityRegistrar) byTile(tiles ...resources.TilesheetSpriteId) dynamicEntityRegistrar {
	r.tiles = append(r.tiles, tiles...)
	return r
}

func (r dynamicEntityRegistrar) register(constructor entityConstructor) {
	for _, class := range r.classes {
		if _, exists := dynamicEntitiesByClass[class]; exists {
			log.Fatal().Str("class", class).Msg("duplicate entity class")
		}
		dynamicEntitiesByClass[class] = constructor
	}
	for _, tile := range r.tiles {
		if _, exists := dynamicEntitiesByTile[tile]; exists {
			log.Fatal().Interface("tile", tile).Msg("duplicate entity tile")
		}
		dynamicEntitiesByTile[tile] = constructor
	}
}

func (s *State) registerParameterizedEntity(entityId EntityId, location MapLocation, mapEntity *resources.Entity) bool {
	var constructor entityConstructor
	var exists bool

	if mapEntity.SpriteId != nil {
		constructor, exists = dynamicEntitiesByTile[*mapEntity.SpriteId]
	}
	if !exists {
		constructor, exists = dynamicEntitiesByClass[mapEntity.Class]
	}
	if constructor == nil {
		return false
	}

	entity, eventHandler := constructor(s, entityId, location, mapEntity)

	if scriptRef := mapEntity.GetStringMetadata("script_ref", ""); scriptRef != "" {
		if eventHandler != nil {
			log.Fatal().Str("entityId", string(entityId)).Msgf("entity has more than one handler configured!")
		}
		eventHandler = handlers.Get(scriptRef)
	}
	if eventHandler != nil {
		s.eventDispatcher.Register(entity, eventHandler)
		log.Info().Msgf("Registered event handler %s", entityId)
	}

	s.AddEntity(entity)
	return true
}
