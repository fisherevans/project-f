package adventure

import (
	"math"
	"sort"

	"fisherevans.com/project/f/internal/game/input"
	"fisherevans.com/project/f/internal/resources"
	"github.com/gopxl/pixel/v2"
	"github.com/rs/zerolog/log"
)

type EntitySystem struct {
	state *State

	// exist for all entities

	movements   map[string]*EntityMovement
	occupations *Positions

	// optional traits

	presences map[string]EntityPresence
	renderers map[string]EntityRenderer

	behaviors         map[string][]EntityBehavior
	disabledBehaviors map[string]map[string]struct{}

	soundProviders map[string][]EntitySoundProvider
	disabledSounds map[string]map[string]struct{}

	metadata map[string]*EntityMetadata
}

func NewEntitySystem(state *State) *EntitySystem {
	return &EntitySystem{
		state: state,

		movements:   map[string]*EntityMovement{},
		occupations: NewPositions(state),

		presences: map[string]EntityPresence{},

		renderers: map[string]EntityRenderer{},

		behaviors:         map[string][]EntityBehavior{},
		disabledBehaviors: map[string]map[string]struct{}{},

		soundProviders: map[string][]EntitySoundProvider{},
		disabledSounds: map[string]map[string]struct{}{},

		metadata: map[string]*EntityMetadata{},
	}
}

func (es *EntitySystem) RegisterEntity(id string, location MapLocation) Entity {
	if _, exists := es.movements[id]; exists {
		log.Fatal().Str("id", id).Msg("entity already exists")
	}
	es.movements[id] = NewEntityMovement(id, es, location)
	es.occupations.Occupy(id, location)
	entity, _ := es.GetEntity(id)
	return entity
}

func (es *EntitySystem) DeleteEntity(id string) {
	es.occupations.RemoveEntity(id)
	delete(es.movements, id)
	delete(es.presences, id)
	delete(es.renderers, id)
	delete(es.behaviors, id)
	delete(es.disabledBehaviors, id)
	delete(es.soundProviders, id)
	delete(es.disabledSounds, id)
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

var validAttemptMovementStates = []MoveState{MoveStateWalking, MoveStateRunning, MoveStateDashing}

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
	if validEgress, _ := es.isValidTransition(entity, targetLocation, movementDirection, true); !validEgress {
		debugLog.Msgf("movement ingress not valid")
		return false, movementDirection
	}
	if validIngress, _ := es.isValidTransition(entity, entity.GetLocation(), movementDirection, false); !validIngress {
		debugLog.Msgf("movement egress not valid")
		return false, movementDirection
	}
	return true, movementDirection
}

func (es *EntitySystem) isValidTransition(entity Entity, location MapLocation, movementDirection input.Direction, isIngress bool) (bool, float64) {
	totalImpedanceWeight := 0.
	isValid := true
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
			isValid = false
			totalImpedanceWeight += presence.PathfindingImpedanceWeight(entity.GetId())
		}
	}
	if totalImpedanceWeight == 0 {
		totalImpedanceWeight = ImpedanceBase // a tile with no impediments == base
	}
	return isValid, totalImpedanceWeight
}

func (es *EntitySystem) Update(timeDelta float64) {
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
				behavior.MovementComplete()
			}
			lastRemaining = remaining
		}
		if hasBehavior && entity.IsBehaviorEnabled() {
			behavior.Update(timeDelta)
		}
		if hasRenderer {
			renderer.Update(timeDelta)
		}
		if entity.IsSoundEnabled() {
			cameraDistance := es.state.camera.CurrentLocation().Sub(entity.GetPreciseLocation()).Len()
			for _, soundProvider := range es.soundProviders[id] {
				soundProvider.Update(timeDelta, cameraDistance)
			}
		}
	}
}

func (es *EntitySystem) Render(sceneTarget, lightMapTarget pixel.Target, cameraDelta pixel.Vec) {
	entities := es.locationSortedRenderers()
	for _, entity := range entities {
		renderer, _ := entity.GetRenderer()
		moveDelta := entity.GetPreciseLocation().Scaled(resources.MapTileSize.Float())
		renderMatrix := pixel.IM.Moved(cameraDelta).Moved(moveDelta)
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

func (es *EntitySystem) loadGenericMetadata(id string, metadata any) {
	entity, exists := es.GetEntity(id)
	if !exists {
		log.Error().Str("entityId", id).Msg("entity not found when loading metadata")
		return
	}

	TalkerConfigMetadataKey.loadMetadata(entity, metadata)
}

func (es *EntitySystem) pauseAllSounds() {
	es.forEachSoundProvider(func(provider EntitySoundProvider, cameraDistance float64) {
		provider.Pause(cameraDistance)
	})
}

func (es *EntitySystem) resumeAllSounds() {
	es.forEachSoundProvider(func(provider EntitySoundProvider, cameraDistance float64) {
		provider.Resume(cameraDistance)
	})
}

func (es *EntitySystem) forEachSoundProvider(apply func(provicer EntitySoundProvider, cameraDistance float64)) {
	for entityId, providers := range es.soundProviders {
		entity, exists := es.GetEntity(entityId)
		var cameraDistance float64
		if exists {
			cameraDistance = entity.GetPreciseLocation().Sub(es.state.camera.CurrentLocation()).Len()
		} else {
			cameraDistance = math.MaxFloat64
		}
		for _, provider := range providers {
			apply(provider, cameraDistance)
		}
	}
}
