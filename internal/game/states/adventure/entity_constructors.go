package adventure

import (
	"fisherevans.com/project/f/internal/game/events"
	"fisherevans.com/project/f/internal/game/states/adventure/handlers"
	"fisherevans.com/project/f/internal/resources"
	"github.com/rs/zerolog/log"
)

type None struct{}

func init() {
	targetRegistration().byClass("ModeBasedEntity").registrar(registerModeBasedEntity)
}

func registerModeBasedEntity(entityId string, location MapLocation, mapEntity *resources.Entity, system *EntitySystem) (events.EntityContext, events.EventHandler) {
	entity := system.RegisterEntity(entityId, location)
	AttachPresenceFromConfig(entity, mapEntity.Properties)
	AttachModeBasedEntityRenderer(entity).
		WithPropConfigurations(mapEntity.Properties)
	return entity.GetEntityContext(), nil
}

type entityRegistrar func(string, MapLocation, *resources.Entity, *EntitySystem) (events.EntityContext, events.EventHandler)

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

	entityContext, eventHandler := registerer(entityId, location, mapEntity, system)
	if entityContext == nil {
		entityContext = events.NewBasicEntityContext(entityId)
	}

	if scriptRef := mapEntity.Properties.GetString("script_ref", ""); scriptRef != "" {
		if eventHandler != nil {
			log.Fatal().Str("entityId", string(entityId)).Msgf("entity has more than one handler configured!")
		}
		eventHandler = handlers.Get(scriptRef, mapEntity.Properties)
	}

	if eventHandler != nil {
		s.eventDispatcher.Register(entityContext, eventHandler)
		log.Debug().Msgf("Registered event handler for %s", entityId)
	}

	return true
}
