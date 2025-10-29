package adventure

import (
	"fisherevans.com/project/f/internal/game/input"
	"fisherevans.com/project/f/internal/game/states/adventure/types"
	"github.com/gopxl/pixel/v2"
	"github.com/rs/zerolog/log"
)

type MovementBasedEntityRenderer struct {
	id     string
	system *EntitySystem

	lastRenderMode movementRenderState
	renderers      map[movementRenderState]*BasicEntityRenderer
}

type movementRenderState struct {
	direction input.Direction
	moveState types.MoveState
}

func NewMovementBasedEntityRenderer(id string, system *EntitySystem) *MovementBasedEntityRenderer {
	return &MovementBasedEntityRenderer{
		id:        id,
		system:    system,
		renderers: map[movementRenderState]*BasicEntityRenderer{},
	}
}

func (r *MovementBasedEntityRenderer) getRenderState() movementRenderState {
	p, ok := r.system.positions[r.id]
	if !ok {
		log.Warn().Str("id", r.id).Msg("failed to find entity position in movement renderer")
		return movementRenderState{}
	}
	return movementRenderState{
		direction: p.FacingDirection,
		moveState: p.MovementState,
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
