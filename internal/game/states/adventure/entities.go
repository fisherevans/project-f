package adventure

import (
	"slices"
	"sort"

	"fisherevans.com/project/f/internal/game/events"
	"fisherevans.com/project/f/internal/game/input"
	"fisherevans.com/project/f/internal/game/states/adventure/types"
	"fisherevans.com/project/f/internal/resources"
	"github.com/gopxl/pixel/v2"
	"github.com/rs/zerolog/log"
)

type EntitySystem struct {
	state *State

	// exist for all entities

	movements   map[string]*EntityMovement
	occupations *Positions
	contexts    map[string]*events.BasicEntityContext

	// optional traits

	presences         map[string]EntityPresence
	renderers         map[string]EntityRenderer
	behaviors         map[string][]EntityBehavior
	disabledBehaviors map[string]map[string]struct{}

	states map[string]EntityState
}

func NewEntitySystem(state *State) *EntitySystem {
	return &EntitySystem{
		state: state,

		movements:   map[string]*EntityMovement{},
		occupations: NewOccupation(state),
		contexts:    map[string]*events.BasicEntityContext{},

		presences:         map[string]EntityPresence{},
		renderers:         map[string]EntityRenderer{},
		behaviors:         map[string][]EntityBehavior{},
		disabledBehaviors: map[string]map[string]struct{}{},
		states:            map[string]EntityState{},
	}
}

func (es *EntitySystem) RegisterEntity(id string, location MapLocation) Entity {
	if _, exists := es.movements[id]; exists {
		log.Fatal().Str("id", id).Msg("entity already exists")
	}
	es.contexts[id] = events.NewBasicEntityContext(id)
	es.movements[id] = NewEntityMovement(id, es, location)
	es.occupations.Occupy(id, location)
	entity, _ := es.GetEntity(id)
	initializeMetadata(es.contexts[id], entity)
	return entity
}

func initializeMetadata(context *events.BasicEntityContext, entity Entity) {
	context.WithMetadata(types.MetadataKeyIsTalking, func() any {
		behavior, ok := entity.GetBehavior()
		if !ok {
			return nil
		}
		npc, ok := behavior.(*NPCBehavior)
		if !ok {
			return nil
		}
		return npc.talkingTowards != ""
	})
	context.WithMetadata(types.MetadataKeyMode, func() any {
		renderer, ok := entity.GetRenderer()
		if !ok {
			return nil
		}
		modeBasedRenderer, ok := renderer.(*ModeBasedEntityRenderer)
		if !ok {
			return nil
		}
		return modeBasedRenderer.currentMode
	})
}

func (es *EntitySystem) GetEntity(id string) (Entity, bool) {
	if _, exists := es.movements[id]; !exists {
		return nil, false
	}
	return &entityReference{
		system: es,
		id:     id,
	}, true
}

func (es *EntitySystem) interactableEntityIds(loc MapLocation) map[string]struct{} {
	out := map[string]struct{}{}
	es.occupations.ForEachOccupyingEntity(loc, func(entityId string) {
		presence, exists := es.presences[entityId]
		if !exists {
			return
		}
		if presence.IsInteractable() {
			out[entityId] = struct{}{}
		}
	})
	return out
}

var validAttemptMovementStates = []types.MoveState{types.MoveStateWalking, types.MoveStateRunning, types.MoveStateDashing}

// todo move?
func (es *EntitySystem) AttemptMovement(id string, targetLocation MapLocation, movementState types.MoveState) bool {
	debugLog := log.Debug().Str("id", id).Any("state", movementState).Any("target", targetLocation)
	if !slices.Contains(validAttemptMovementStates, movementState) {
		debugLog.Msgf("movement state not valid")
		return false
	}
	entity, ok := es.GetEntity(id)
	if !ok {
		log.Warn().Str("entityId", id).Msg("no entity for id, skipping")
		return false
	}
	isValid, movementDirection := es.isMovementValid(entity, targetLocation)
	if !isValid {
		debugLog.Msgf("movement not valid")
		return false
	}
	es.occupations.Occupy(id, targetLocation) // emit enter events within movement update - don't "enter" until primary is updated
	// todo messy
	movement := es.movements[entity.GetId()]
	movement.TargetLocation = targetLocation
	movement.MovementState = movementState
	movement.FacingDirection = movementDirection
	distance := entity.GetLocation().DistanceTo(targetLocation)
	movement.ProgressionScale = 1.0 / distance
	return true
}

func (es *EntitySystem) isMovementValid(entity Entity, targetLocation MapLocation) (bool, input.Direction) {
	debugLog := log.Debug().Str("id", entity.GetId())
	movementDirection := entity.GetLocation().DirectionTowards(targetLocation)
	if entity.IsMoving() {
		debugLog.Msgf("entity is already moving")
		return false, movementDirection
	}
	if entity.GetLocation() == targetLocation {
		debugLog.Msgf("attempted movement to same location")
		return false, movementDirection
	}
	if !es.isValidTransition(entity, targetLocation, movementDirection, true) {
		debugLog.Msgf("movement ingress not valid")
		return false, movementDirection
	}
	if !es.isValidTransition(entity, entity.GetLocation(), movementDirection, false) {
		debugLog.Msgf("movement egress not valid")
		return false, movementDirection
	}
	return true, movementDirection
}

func (es *EntitySystem) isValidTransition(entity Entity, location MapLocation, movementDirection input.Direction, isIngress bool) bool {
	for _, occupiedById := range es.occupations.OccupyingEntityList(location) {
		if occupiedById == entity.GetId() {
			continue
		}
		presence, exists := es.presences[occupiedById]
		if !exists {
			continue
		}
		doesAllow := presence.AllowsEgress
		side := movementDirection
		if isIngress {
			doesAllow = presence.AllowsIngress
			side = movementDirection.Opposite()
		}
		if !doesAllow(side, entity.GetId()) {
			return false
		}
	}
	return true
}

func (es *EntitySystem) Update(timeDelta float64) {
	dispatcher := &stateDispatcher{
		s: es.state,
	}
	for id, _ := range es.movements { // todo consider tracking just update-able entities (basic collisions are included here)
		entity, _ := es.GetEntity(id)
		movement := es.movements[id]
		renderer, hasRenderer := es.renderers[id]
		behaviors, _ := es.behaviors[id]
		var behavior EntityBehavior
		if len(behaviors) > 0 {
			behavior = behaviors[len(behaviors)-1]
		}
		hasBehavior := behavior != nil
		remaining := timeDelta
		var lastRemaining float64 // prevent infinite loops if movement is not making progress
		for remaining > 0 && remaining != lastRemaining {
			remaining = movement.ProgressMovement(timeDelta)
			if remaining > 0 && hasBehavior && entity.IsBehaviorEnabled() {
				behavior.MovementComplete(dispatcher)
			}
			lastRemaining = remaining
		}
		if hasBehavior && entity.IsBehaviorEnabled() {
			behavior.Update(timeDelta, dispatcher)
		}
		if hasRenderer {
			renderer.Update(timeDelta)
		}
	}
}

func (es *EntitySystem) Render(sceneTarget, lightMapTarget pixel.Target, cameraMatrix pixel.Matrix) {
	entities := es.locationSortedRenderers()
	for _, entity := range entities {
		renderer, _ := entity.GetRenderer()
		renderMatrix := cameraMatrix.Moved(entity.GetPreciseLocation().Scaled(resources.MapTileSize.Float()))
		renderer.RenderToScene(sceneTarget, renderMatrix)
		renderer.RenderToLightMap(lightMapTarget, renderMatrix)
	}
}

func (es *EntitySystem) locationSortedRenderers() []Entity {
	sorted := make([]Entity, 0, len(es.renderers))
	for entityId, _ := range es.renderers {
		entity, ok := es.GetEntity(entityId)
		if !ok {
			log.Error().Str("entityId", entityId).Msg("entity with renderer does not exist, skipping")
			continue
		}
		sorted = append(sorted, entity)
	}
	sort.Slice(sorted, func(iId, jId int) bool {
		i, j := sorted[iId], sorted[jId]
		iRenderer, _ := es.renderers[i.GetId()]
		jRenderer, _ := es.renderers[j.GetId()]
		iZ, jZ := iRenderer.ZPriority(), jRenderer.ZPriority()
		if iZ != jZ {
			return iZ < jZ
		}
		// todo do we need to account for offsets here?
		iL, jL := i.GetPreciseLocation(), j.GetPreciseLocation()
		if iL.Y != jL.Y {
			return iL.Y > jL.Y
		}
		if iL.X != jL.X {
			return iL.X < jL.X
		}
		return i.GetId() < j.GetId()
	})
	return sorted
}
