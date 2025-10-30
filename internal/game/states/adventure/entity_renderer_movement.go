package adventure

import (
	"fisherevans.com/project/f/internal/game/input"
	"fisherevans.com/project/f/internal/game/states/adventure/types"
	"github.com/gopxl/pixel/v2"
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
		r.renderers[state] = NewBasicEntityRenderer()
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
}

func (r *MovementBasedEntityRenderer) RenderToLightMap(target pixel.Target, matrix pixel.Matrix) {
	r.getRenderer(r.getRenderState()).RenderToLightMap(target, matrix)
}
