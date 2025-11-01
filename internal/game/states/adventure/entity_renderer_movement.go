package adventure

import (
	"fisherevans.com/project/f/internal/game/input"
	"fisherevans.com/project/f/internal/game/states/adventure/types"
	"fisherevans.com/project/f/internal/resources"
	"github.com/gopxl/pixel/v2"
)

var (
	renderPathFinderFuture = false
)

type MovementBasedEntityRenderer struct {
	entity Entity

	lastRenderMode movementRenderState
	renderers      map[movementRenderState]*BasicEntityRenderer
}

type movementRenderState struct {
	direction input.Direction
	moveState types.MoveState
}

func AttachMovementBasedEntityRenderer(entity Entity) *MovementBasedEntityRenderer {
	renderer := &MovementBasedEntityRenderer{
		entity:    entity,
		renderers: map[movementRenderState]*BasicEntityRenderer{},
	}
	entity.SetRenderer(renderer)
	return renderer
}

func (r *MovementBasedEntityRenderer) getRenderState() movementRenderState {
	return movementRenderState{
		direction: r.entity.GetFacingDirection(),
		moveState: r.entity.GetMovementState(),
	}
}

func (r *MovementBasedEntityRenderer) getRenderer(state movementRenderState) *BasicEntityRenderer {
	_, ok := r.renderers[state]
	if !ok {
		r.renderers[state] = NewBasicEntityRenderer(r.entity)
	}
	return r.renderers[state]
}

func (r *MovementBasedEntityRenderer) WithMovementStateRenderer(moveState types.MoveState, direction input.Direction, renderer *BasicEntityRenderer) *MovementBasedEntityRenderer {
	r.renderers[movementRenderState{
		moveState: moveState,
		direction: direction,
	}] = renderer
	return r
}

func (r *MovementBasedEntityRenderer) ZPriority() int {
	return r.getRenderer(r.getRenderState()).ZPriority()
}

func (r *MovementBasedEntityRenderer) Update(timeDelta float64) {
	r.getRenderer(r.getRenderState()).Update(timeDelta)
}

func (r *MovementBasedEntityRenderer) RenderToScene(target pixel.Target, matrix pixel.Matrix) {
	renderState := r.getRenderState()
	renderer := r.getRenderer(renderState)
	if r.lastRenderMode != renderState {
		renderer.Reset()
	}
	r.lastRenderMode = renderState
	renderer.RenderToScene(target, matrix)

	if !renderPathFinderFuture {
		return
	}
	b, ok := r.entity.GetBehavior()
	if !ok {
		return
	}
	sm, ok := b.(*ScriptedMotionBehavior)
	if !ok || sm.target == nil {
		return
	}
	pm, ok := sm.target.(*PathfindingMotion)
	if !ok || pm.path == nil || len(pm.path.Tiles) <= 1 {
		return
	}
	l := r.entity.GetPreciseLocation()
	maxTilesAhead := 12
	maxAlpha := 0.5
	for id, t := range pm.path.Tiles[1:] {
		if id > maxTilesAhead {
			break
		}
		alpha := maxAlpha * float64(maxTilesAhead-id) / float64(maxTilesAhead)
		delta := pixel.V(
			float64(t.X)-l.X,
			float64(t.Y)-l.Y,
		).Scaled(resources.MapTileSize.Float())

		renderer.RenderToSceneAlpha(target, matrix.Moved(delta), alpha)
	}
}

func (r *MovementBasedEntityRenderer) RenderToLightMap(target pixel.Target, matrix pixel.Matrix) {
	r.getRenderer(r.getRenderState()).RenderToLightMap(target, matrix)
}
