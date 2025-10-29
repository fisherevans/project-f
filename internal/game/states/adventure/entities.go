package adventure

import (
	"slices"
	"sort"

	"fisherevans.com/project/f/internal/game/input"
	"fisherevans.com/project/f/internal/game/states/adventure/types"
	"fisherevans.com/project/f/internal/resources"
	"github.com/gopxl/pixel/v2"
	"github.com/rs/zerolog/log"
)

type Entity struct {
	Id string
	EntityPresence
	EntityBehavior
	EntityRenderer
	EntityState
	*EntityPosition
}

type EntitySystem struct {
	state *State

	entityIds map[string]struct{}

	presences map[string]EntityPresence
	renderers map[string]EntityRenderer
	behaviors map[string]EntityBehavior
	states    map[string]EntityState

	positions   map[string]*EntityPosition
	occupations *Occupation
}

func NewEntitySystem(state *State) *EntitySystem {
	return &EntitySystem{
		state: state,

		entityIds: map[string]struct{}{},

		presences: map[string]EntityPresence{},
		renderers: map[string]EntityRenderer{},
		behaviors: map[string]EntityBehavior{},
		states:    map[string]EntityState{},

		positions:   map[string]*EntityPosition{},
		occupations: NewOccupation(state),
	}
}

func (es *EntitySystem) RegisterEntity(id string, location MapLocation, presence EntityPresence, behavior EntityBehavior, renderer EntityRenderer, state EntityState) *EntityPosition {
	if _, exists := es.entityIds[id]; exists {
		log.Fatal().Str("id", id).Msg("entity already exists")
	}
	es.entityIds[id] = struct{}{}
	es.positions[id] = NewEntityPosition(id, es, location)
	if presence != nil {
		es.presences[id] = presence
	}
	if behavior != nil {
		es.behaviors[id] = behavior
	}
	if renderer != nil {
		es.renderers[id] = renderer
	}
	if state != nil {
		es.states[id] = state
	}
	es.occupations.Occupy(id, location)
	return es.positions[id]
}

func (es *EntitySystem) GetEntity(id string) Entity {
	return Entity{
		Id:             id,
		EntityPresence: es.presences[id],
		EntityBehavior: es.behaviors[id],
		EntityRenderer: es.renderers[id],
		EntityState:    es.states[id],
		EntityPosition: es.positions[id],
	}
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

func (es *EntitySystem) TeleportEntity(id string, toLocation MapLocation) {
	warnLog := log.Log().Str("id", id).Any("location", toLocation)
	position, exists := es.positions[id]
	if !exists {
		warnLog.Msgf("no position for entity")
		return
	}
	if position.GetPrimaryLocation() == toLocation {
		warnLog.Msgf("attempted movement to same location")
		return
	}
	if position.IsMoving() {
		position.CancelMovement()
	}
	es.occupations.Vacate(id, position.GetPrimaryLocation())
	position.SetPrimaryLocation(toLocation, true)
}

var validAttemptMovementStates = []types.MoveState{types.MoveStateWalking, types.MoveStateRunning, types.MoveStateDashing}

func (es *EntitySystem) AttemptMovement(id string, targetLocation MapLocation, movementState types.MoveState) bool {
	log := log.Debug().Str("id", id).Any("state", movementState).Any("target", targetLocation)
	if !slices.Contains(validAttemptMovementStates, movementState) {
		log.Msgf("movement state not valid")
		return false
	}
	isValid, position, movementDirection := es.isMovementValid(id, targetLocation)
	if !isValid {
		log.Msgf("movement not valid")
		return false
	}
	es.occupations.Occupy(id, targetLocation) // emit enter events within movement update - don't "enter" until primary is updated
	position.MovementTargetLocation = targetLocation
	position.MovementState = movementState
	position.FacingDirection = movementDirection
	distance := position.GetPrimaryLocation().DistanceTo(targetLocation)
	position.MovementProgressionScale = 1.0 / distance
	return true
}
func (es *EntitySystem) isMovementValid(id string, targetLocation MapLocation) (bool, *EntityPosition, input.Direction) {
	log := log.Debug().Str("id", id)
	position, exists := es.positions[id]
	if !exists {
		log.Msgf("no position for entity")
		return false, position, input.NotPressed
	}
	movementDirection := position.GetPrimaryLocation().DirectionTowards(targetLocation)
	if position.IsMoving() {
		log.Msgf("entity is already moving")
		return false, position, movementDirection
	}
	if position.GetPrimaryLocation() == targetLocation {
		log.Msgf("attempted movement to same location")
		return false, position, movementDirection
	}
	if !es.isValidTransition(id, targetLocation, movementDirection, true) {
		log.Msgf("movement ingress not valid")
		return false, position, movementDirection
	}
	if !es.isValidTransition(id, position.GetPrimaryLocation(), movementDirection, false) {
		log.Msgf("movement egress not valid")
		return false, position, movementDirection
	}
	return true, position, movementDirection
}

func (es *EntitySystem) isValidTransition(id string, loc MapLocation, movementDirection input.Direction, isIngress bool) bool {
	for _, occupiedById := range es.occupations.OccupyingEntityList(loc) {
		if occupiedById == id {
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
		if !doesAllow(side, id) {
			return false
		}
	}
	return true
}

func (es *EntitySystem) Update(timeDelta float64) {
	dispatcher := &stateDispatcher{
		s: es.state,
	}
	for id, _ := range es.entityIds { // todo consider tracking just update-able entities (basic collisions are included here)
		position, hasPosition := es.positions[id]
		behavior, hasBehavior := es.behaviors[id]
		renderer, hasRenderer := es.renderers[id]
		if hasPosition {
			remaining := timeDelta
			var lastRemaining float64 // prevent infinite loops if movement is not making progress
			for remaining > 0 && remaining != lastRemaining {
				remaining = position.ProgressMovement(timeDelta)
				if remaining > 0 && hasBehavior && behavior.IsEnabled() {
					behavior.MovementComplete(dispatcher)
				}
				lastRemaining = remaining
			}
		}
		if hasBehavior {
			if hasPosition {
				if behavior.IsEnabled() {
					behavior.Update(timeDelta, position, dispatcher)
				}
			} else {
				log.Warn().Str("entityId", id).Msg("no position for entity with behavior, skipping")
			}
		}
		if hasRenderer {
			renderer.Update(timeDelta)
		}
	}
}

func (es *EntitySystem) Render(sceneTarget, lightMapTarget pixel.Target, cameraMatrix pixel.Matrix) {
	entities := es.locationSortedRenderers()
	for _, entity := range entities {
		renderMatrix := cameraMatrix.Moved(entity.position.PreciseLocation().Scaled(resources.MapTileSize.Float()))
		entity.renderer.RenderToScene(sceneTarget, renderMatrix)
		entity.renderer.RenderToLightMap(lightMapTarget, renderMatrix)
	}
}

type sortedRenderableEntity struct {
	id       string
	renderer EntityRenderer
	presence EntityPresence
	position *EntityPosition
}

func (es *EntitySystem) locationSortedRenderers() []sortedRenderableEntity {
	sorted := make([]sortedRenderableEntity, 0, len(es.renderers))
	for entityId, renderer := range es.renderers {
		position, hasPosition := es.positions[entityId]
		if !hasPosition {
			log.Warn().Str("entityId", entityId).Msg("no position for entity with renderer, skipping")
			continue
		}
		sorted = append(sorted, sortedRenderableEntity{
			id:       entityId,
			renderer: renderer,
			position: position,
		})
	}
	sort.Slice(sorted, func(i, j int) bool {
		iZ, jZ := sorted[i].renderer.ZPriority(), sorted[j].renderer.ZPriority()
		if iZ != jZ {
			return iZ < jZ
		}
		// todo do we need to account for offsets here?
		iL, jL := sorted[i].position.PreciseLocation(), sorted[j].position.PreciseLocation()
		if iL.Y != jL.Y {
			return iL.Y > jL.Y
		}
		if iL.X != jL.X {
			return iL.X < jL.X
		}
		return i < j
	})
	return sorted
}

func (es *EntitySystem) OverrideBehavior(id string, behavior EntityBehavior) {
	existing, exists := es.behaviors[id]
	if !exists {
		es.behaviors[id] = behavior
		log.Warn().Str("entityId", id).Msg("no behavior for entity, just setting")
		return
	}
	es.behaviors[id] = NewEntityBehaviorOverride(existing, behavior)
}

func (es *EntitySystem) PopOverrideBehavior(id string) {
	existing, exists := es.behaviors[id]
	if !exists {
		log.Warn().Str("entityId", id).Msg("no behavior for entity, can't pop")
		return
	}
	override, ok := existing.(*EntityBehaviorOverride)
	if !ok {
		log.Warn().Str("entityId", id).Msg("behavior isn't overridden, just removing")
		delete(es.behaviors, id)
	}
	es.behaviors[id] = override.replacedBehavior
}
