package adventure

import (
	"fisherevans.com/project/f/internal/resources"
	"fisherevans.com/project/f/internal/util"
	"github.com/rs/zerolog/log"
)

type None struct{}

func init() {
	newRegistrarBuilder().byClass("ModeBasedEntity").registrar(registerModeBasedEntity)
}

func registerModeBasedEntity(params NewEntityParams, system *EntitySystem) (Entity, EventHandler) {
	entity := system.RegisterEntity(params.EntityId, params.Location)
	AttachPresenceFromConfig(entity, params.Properties)
	AttachModeBasedEntityRenderer(entity).
		WithPropConfigurations(params.Properties)
	return entity, nil
}

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
	log.Debug().Msgf("Registered entity %s", params.EntityId)

	if scriptRef := params.Properties.GetString("script_ref", ""); scriptRef != "" {
		if eventHandler != nil {
			log.Fatal().Str("entityId", string(params.EntityId)).Msgf("entity has more than one handler configured!")
		}
		eventHandler = Get(scriptRef, params.Properties)
	}

	if eventHandler != nil {
		s.eventDispatcher.Register(entity.GetEntityContext(), eventHandler)
		log.Debug().Msgf("Registered event handler for %s", params.EntityId)
	}

	return true
}

type NewEntityParams struct {
	EntityId   string
	Class      string
	SpriteId   *resources.TilesheetSpriteId
	Location   MapLocation
	Properties *util.Properties
}
