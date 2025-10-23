package adventure

import (
	"math/rand"

	"fisherevans.com/project/f/internal/game/anim"
	"fisherevans.com/project/f/internal/game/events"
	"fisherevans.com/project/f/internal/game/input"
	"fisherevans.com/project/f/internal/resources"
	"fisherevans.com/project/f/internal/util/colors"
	"fisherevans.com/project/f/internal/util/tiles"
	"github.com/dop251/goja"
	"github.com/gopxl/pixel/v2"
	"github.com/rs/zerolog/log"
)

func init() {
	registerDynamicEntity().byTile(tiles.Elythium).register(constructDynamicEntity)
	registerDynamicEntity().byClass("NPC").register(constructNPCEntity)
	registerDynamicEntity().byClass("Script").register(constructScriptEntity)
}

func constructDynamicEntity(entityId EntityId, location MapLocation, mapEntity *resources.Entity) (Entity, events.EventHandler) {
	e := NewDynamicEntity(entityId, location)
	e.lights = map[string][]*Light{
		"": {
			{
				RenderDetails: LightRenderDetails{
					SizeScale: 1.5,
					ColorMask: colors.HexString("#f06"),
				},
				Modifiers: []LightModifier{
					&LightModifierPulse{
						PeriodSeconds:       4,
						SizeIntensity:       0.1,
						BrightnessIntensity: 0.4,
					},
				},
			},
		},
	}
	e.animations = map[string][]*anim.AnimatedSprite{
		"": {
			anim.Load(atlas, "adventure/entities/elythium/crystals", "default"),
			anim.Load(atlas, "adventure/entities/elythium/crystals_sparkle", "default"),
		},
		"mined": {
			anim.Load(atlas, "adventure/entities/elythium/crystals_rock", "default"),
		},
	}
	return e, newReferenceEventHandler(entityId, "test")
}

func constructNPCEntity(entityId EntityId, location MapLocation, mapEntity *resources.Entity) (Entity, events.EventHandler) {
	e := &NPC{
		AnimatedMoveableEntity: AnimatedMoveableEntity{
			MoveableEntity: MoveableEntity{
				BaseEntity: BaseEntity{
					Id:           entityId,
					Interactable: true,
				},
				CurrentLocation: location,
				MoveSpeeds: map[MoveState]float64{
					MoveStateWalking: 2,
				},
				Passable: newPassablePreventIngress(true),
			},
			Animations: map[MoveState]map[input.Direction]*anim.AnimatedSprite{
				MoveStateIdle:    anim.AshaIdle(atlas),
				MoveStateWalking: anim.AshaWalk(atlas),
				MoveStateRunning: anim.AshaRun(atlas),
			},
			ColorMask: pixel.RGB(rand.Float64(), rand.Float64(), rand.Float64()),
		},
		DoesMove:        true,
		IdleChance:      0.05,
		MaxIdleDuration: 6,
	}
	// TODO movement & speed
	return e, nil
}

func constructScriptEntity(entityId EntityId, location MapLocation, mapEntity *resources.Entity) (Entity, events.EventHandler) {
	e := NewDynamicEntity(entityId, location)
	return e, nil
}

type entityConstructor func(EntityId, MapLocation, *resources.Entity) (Entity, events.EventHandler)

var dynamicEntitiesByClass = map[string]entityConstructor{}
var dynamicEntitiesByTile = map[resources.TilesheetSpriteId]entityConstructor{}

type dynamicEntityRegistrar struct {
	class *string
	tile  *resources.TilesheetSpriteId
}

func registerDynamicEntity() dynamicEntityRegistrar {
	return dynamicEntityRegistrar{}
}

func (r dynamicEntityRegistrar) byClass(class string) dynamicEntityRegistrar {
	r.class = &class
	return r
}

func (r dynamicEntityRegistrar) byTile(tile resources.TilesheetSpriteId) dynamicEntityRegistrar {
	r.tile = &tile
	return r
}

func (r dynamicEntityRegistrar) register(constructor entityConstructor) {
	if r.class != nil {
		dynamicEntitiesByClass[*r.class] = constructor
	}
	if r.tile != nil {
		dynamicEntitiesByTile[*r.tile] = constructor
	}
}

func (s *State) registerDynamicEntity(entityId EntityId, location MapLocation, mapEntity *resources.Entity) bool {
	if string(entityId) == "tiled-174" {
		log.Info().Msgf("Registering tiled entity with id %s", entityId)
	}
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

	entity, eventHandler := constructor(entityId, location, mapEntity)

	if scriptRef := mapEntity.GetStringMetadata("script_ref", ""); scriptRef != "" {
		if eventHandler != nil {
			log.Fatal().Str("entityId", string(entityId)).Msgf("entity has more than one handler configured!")
		}
		eventHandler = newReferenceEventHandler(entityId, scriptRef)
	}
	if scriptSource := mapEntity.GetStringMetadata("script_src", ""); scriptSource != "" {
		if eventHandler != nil {
			log.Fatal().Str("entityId", string(entityId)).Msgf("entity has more than one handler configured!")
		}
		eventHandler = newSourceEventHandler(entityId, scriptSource)
	}
	if eventHandler != nil {
		s.eventDispatcher.Register(newEventEntityWrapper(entity), eventHandler)
		log.Info().Msgf("Registered event handler %s", entityId)
	}

	s.AddEntity(entity)
	return true
}

func newSourceEventHandler(id EntityId, source string) events.EventHandler {
	program, err := goja.Compile(string(id)+"-inline", source, true) // strict=true
	if err != nil {
		log.Fatal().Err(err).Msgf("Failed to compile script source for entity %s", id)
	}
	eventHandler, err := events.NewGojaEventHandler(string(id), program)
	if err != nil {
		log.Fatal().Msgf("Unable to create event handler from source for %s: %s", string(id), err)
	}
	return eventHandler
}

func newReferenceEventHandler(id EntityId, name string) events.EventHandler {
	program := resources.GetScript(name)
	if program == nil {
		log.Fatal().Msgf("Unable to find script: %s", name)
	}
	eventHandler, err := events.NewGojaEventHandler(string(id), program)
	if err != nil {
		log.Fatal().Msgf("Unable to create event handler from reference for %s (ref: %s): %s", string(id), name, err)
	}
	return eventHandler
}
