package adventure

import (
	"fisherevans.com/project/f/internal/game/events"
	"fisherevans.com/project/f/internal/game/states/adventure/handlers"
	"fisherevans.com/project/f/internal/resources"
	"github.com/rs/zerolog/log"
)

type None struct{}

func init() {
	// todo rename class to ModeBasedEntity
	targetRegistration().byClass("Script").registrar(registerModeBasedEntity)
}

func registerModeBasedEntity(entityId string, location MapLocation, mapEntity *resources.Entity, system *EntitySystem) events.EventHandler {
	state := NewStringModeEntityState("")
	system.RegisterEntity(entityId, location, nil, nil, NewModeBasedEntityRenderer(entityId, system), state)
	return nil
}

type entityRegistrar func(string, MapLocation, *resources.Entity, *EntitySystem) events.EventHandler

var registrarsByClass = map[string]entityRegistrar{}
var registrarsEntitiesByTile = map[resources.TilesheetSpriteId]entityRegistrar{}

type entityRegistrarTargets struct {
	classes []string
	tiles   []resources.TilesheetSpriteId
}

func targetRegistration() entityRegistrarTargets {
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

func (s *State) registerParameterizedEntity(entityId string, location MapLocation, mapEntity *resources.Entity, system *EntitySystem) bool {
	var registerer entityRegistrar
	var exists bool

	if mapEntity.SpriteId != nil {
		registerer, exists = registrarsEntitiesByTile[*mapEntity.SpriteId]
	}
	if !exists {
		registerer, exists = registrarsByClass[mapEntity.Class]
	}
	if registerer == nil {
		return false
	}

	eventHandler := registerer(entityId, location, mapEntity, system)

	if scriptRef := mapEntity.GetStringMetadata("script_ref", ""); scriptRef != "" {
		if eventHandler != nil {
			log.Fatal().Str("entityId", string(entityId)).Msgf("entity has more than one handler configured!")
		}
		eventHandler = handlers.Get(scriptRef)
	}
	if eventHandler != nil {
		ctx := &systemEntityContext{
			id:     entityId,
			system: system,
		}
		s.eventDispatcher.Register(ctx, eventHandler)
		log.Info().Msgf("Registered event handler %s", entityId)
	}

	return true
}

type systemEntityContext struct {
	id     string
	system *EntitySystem
}

func (s *systemEntityContext) Id() string {
	return s.id
}

func (s *systemEntityContext) Mode() string {
	// todo, clean up - maybe move to entity state that I removed?
	state, ok := s.system.states[s.id]
	if !ok {
		return ""
	}
	stringMode, ok := state.(EntityStateStringMode)
	if !ok {
		return ""
	}
	return stringMode.mode
}
