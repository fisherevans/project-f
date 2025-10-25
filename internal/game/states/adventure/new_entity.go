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

	positions         map[string]*EntityPosition
	occupiedLocations map[MapLocation]map[string]struct{}
}

func NewEntitySystem(state *State) *EntitySystem {
	return &EntitySystem{
		state: state,

		entityIds: map[string]struct{}{},

		presences: map[string]EntityPresence{},
		renderers: map[string]EntityRenderer{},
		behaviors: map[string]EntityBehavior{},
		states:    map[string]EntityState{},

		positions:         map[string]*EntityPosition{},
		occupiedLocations: map[MapLocation]map[string]struct{}{},
	}
}

func (es *EntitySystem) RegisterEntity(id string, location MapLocation, presence EntityPresence, behavior EntityBehavior, renderer EntityRenderer, state EntityState) *EntityPosition {
	if _, exists := es.entityIds[id]; exists {
		log.Fatal().Str("id", id).Msg("entity already exists")
	}
	es.entityIds[id] = struct{}{}
	es.positions[id] = &EntityPosition{
		System:   es,
		Id:       id,
		Location: location,
	}
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
	es.occupy(id, location)
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

func (es *EntitySystem) occupy(id string, loc MapLocation) {
	if _, exists := es.occupiedLocations[loc]; !exists {
		es.occupiedLocations[loc] = map[string]struct{}{}
	}
	es.occupiedLocations[loc][id] = struct{}{}
}

func (es *EntitySystem) vacateAndEmitEvent(id string, loc MapLocation) {
	if _, exists := es.occupiedLocations[loc]; !exists {
		return
	}
	if _, exists := es.occupiedLocations[loc][id]; !exists {
		return
	}
	delete(es.occupiedLocations[loc], id)
	es.emitEntityLocationExitEvents(id, loc)
}

func (es *EntitySystem) occupyingEntityIds(loc MapLocation) map[string]struct{} {
	if _, exists := es.occupiedLocations[loc]; !exists {
		return map[string]struct{}{}
	}
	return es.occupiedLocations[loc]
}

func (es *EntitySystem) SetEntityLocation(id string, loc MapLocation) {
	warnLog := log.Log().Str("id", id).Any("location", loc)
	position, exists := es.positions[id]
	if !exists {
		warnLog.Msgf("no position for entity")
		return
	}
	if position.Location == loc {
		warnLog.Msgf("attempted movement to same location")
		return
	}
	if position.IsMoving() {
		position.CancelMovement()
	}
	es.vacateAndEmitEvent(id, position.Location)
	position.Location = loc
	es.occupy(id, position.Location)
	es.emitEntityLocationEnterEvents(id, position.Location)
}

var validAttemptMovementStates = []types.MoveState{types.MoveStateWalking, types.MoveStateRunning, types.MoveStateDashing}

func (es *EntitySystem) AttemptMovement(id string, targetLocation MapLocation, movementState types.MoveState) bool {
	warnLog := log.Log().Str("id", id).Any("state", movementState).Any("target", targetLocation)
	position, exists := es.positions[id]
	if !exists {
		warnLog.Msgf("no position for entity")
		return false
	}
	if position.IsMoving() {
		return false
	}
	if position.Location == targetLocation {
		warnLog.Msgf("attempted movement to same location")
		return true
	}
	if !slices.Contains(validAttemptMovementStates, movementState) {
		warnLog.Msgf("invalid attempted movement state")
		return false
	}
	movementDirection := position.Location.DirectionTowards(targetLocation)
	if !es.isValidTransition(id, targetLocation, movementDirection, true) {
		return false
	}
	if !es.isValidTransition(id, position.Location, movementDirection, false) {
		return false
	}
	es.occupy(id, position.Location) // emit enter events within movement update - don't "enter" until primary is updated
	position.MovementTargetLocation = targetLocation
	position.MovementState = movementState
	return true
}

func (es *EntitySystem) isValidTransition(id string, loc MapLocation, movementDirection input.Direction, isIngress bool) bool {
	for occupiedById, _ := range es.occupyingEntityIds(loc) {
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

func (es *EntitySystem) emitEntityLocationExitEvents(id string, from MapLocation) {
	for _, zoneId := range es.state.zones.ZonesAt(from) {
		es.state.eventDispatcher.Dispatch(events.EventEntityZoneActivity{
			EntityId:   id,
			ZoneId:     zoneId,
			IsEntering: false,
		})
	}
}

func (es *EntitySystem) emitEntityLocationEnterEvents(id string, to MapLocation) {
	for _, zoneId := range es.state.zones.ZonesAt(to) {
		es.state.eventDispatcher.Dispatch(events.EventEntityZoneActivity{
			EntityId:   id,
			ZoneId:     zoneId,
			IsEntering: true,
		})
	}
}

func (es *EntitySystem) Update(timeDelta float64) {
	dispatcher := &stateDispatcher{
		s: es.state,
	}
	for id, _ := range es.entityIds {
		position, hasPosition := es.positions[id]
		behavior, hasBehavior := es.behaviors[id]
		renderer, hasRenderer := es.renderers[id]
		if hasPosition {
			remaining := timeDelta
			for remaining > 0 {
				remaining = position.ProgressMovement(timeDelta)
				if remaining > 0 && hasBehavior {
					behavior.MovementComplete(dispatcher)
				}
			}
		}
		if hasBehavior {
			if hasPosition {
				behavior.Update(timeDelta, position, dispatcher)
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
