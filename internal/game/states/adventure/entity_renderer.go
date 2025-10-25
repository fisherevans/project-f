package adventure

import (
	"fisherevans.com/project/f/internal/game/anim"
	"fisherevans.com/project/f/internal/game/events"
	"fisherevans.com/project/f/internal/game/input"
	"fisherevans.com/project/f/internal/game/states/adventure/types"
	"fisherevans.com/project/f/internal/resources"
	"fisherevans.com/project/f/internal/util/colors"
	"github.com/gopxl/pixel/v2"
	"github.com/rs/zerolog/log"
)

type EntityRenderer interface {
	Update(timeDelta float64)
	RenderToScene(target pixel.Target, matrix pixel.Matrix)
	RenderToLightMap(target pixel.Target, matrix pixel.Matrix)
	ZPriority() int
}

type AnimationSpeedScaler interface {
	GetAnimationSpeedScale() float64
}

type MoveAnimationSpeedScaler struct {
	system *EntitySystem
	id     string
}

func NewMoveAnimationSpeedScaler(system *EntitySystem, id string) *MoveAnimationSpeedScaler {
	return &MoveAnimationSpeedScaler{
		system: system,
		id:     id,
	}
}

func (s *MoveAnimationSpeedScaler) GetAnimationSpeedScale() float64 {
	return s.system.positions[s.id].MovementSpeeds[s.system.positions[s.id].MovementState]
}

type BasicEntityRenderer struct {
	zPriority             int
	lights                []*Light
	lightOriginOffset     pixel.Vec
	animations            []*anim.AnimatedSprite
	animationOriginOffset pixel.Vec
	animationColorMask    pixel.RGBA
	animationSpeedScaler  AnimationSpeedScaler
}

func NewBasicEntityRenderer() *BasicEntityRenderer {
	return &BasicEntityRenderer{
		animationColorMask: colors.White.RGBA,
	}
}

func (r *BasicEntityRenderer) WithAnimationColorMask(colorMask pixel.RGBA) *BasicEntityRenderer {
	r.animationColorMask = colorMask
	return r
}

func (r *BasicEntityRenderer) WithZPriority(zPriority int) *BasicEntityRenderer {
	r.zPriority = zPriority
	return r
}

func (r *BasicEntityRenderer) WithLights(lights ...*Light) *BasicEntityRenderer {
	r.lights = lights
	return r
}

func (r *BasicEntityRenderer) WithLightOriginOffset(offset pixel.Vec) *BasicEntityRenderer {
	r.lightOriginOffset = offset
	return r
}

func (r *BasicEntityRenderer) WithAnimations(animations ...*anim.AnimatedSprite) *BasicEntityRenderer {
	r.animations = animations
	return r
}

func (r *BasicEntityRenderer) WithAnimationOriginOffset(offset pixel.Vec) *BasicEntityRenderer {
	r.animationOriginOffset = offset
	return r
}

func (r *BasicEntityRenderer) WithAnimationSpeedScaler(speedScaler AnimationSpeedScaler) *BasicEntityRenderer {
	r.animationSpeedScaler = speedScaler
	return r
}

func (r *BasicEntityRenderer) Reset() {
	for _, a := range r.animations {
		a.Reset()
	}
}

func (r *BasicEntityRenderer) ZPriority() int {
	return r.zPriority
}

func (r *BasicEntityRenderer) Update(timeDelta float64) {
	animationSpeedScale := 1.0
	if r.animationSpeedScaler != nil {
		animationSpeedScale = r.animationSpeedScaler.GetAnimationSpeedScale()
	}
	for _, a := range r.animations {
		a.Update(timeDelta * animationSpeedScale)
	}
	for _, l := range r.lights {
		l.Update(timeDelta)
	}
}

func (r *BasicEntityRenderer) RenderToScene(target pixel.Target, matrix pixel.Matrix) {
	for _, a := range r.animations {
		a.Sprite().DrawColorMask(target, matrix.Moved(r.animationOriginOffset.Scaled(resources.MapTileSize.Float())), r.animationColorMask)
	}
}

func (r *BasicEntityRenderer) RenderToLightMap(target pixel.Target, matrix pixel.Matrix) {
	for _, l := range r.lights {
		l.Render(target, matrix.Moved(r.lightOriginOffset))
	}
}

type ModeBasedEntityRenderer struct {
	id     string
	system *EntitySystem

	currentMode    string
	lastRenderMode string
	modes          map[string]*BasicEntityRenderer
}

func NewModeBasedEntityRenderer(id string, system *EntitySystem) *ModeBasedEntityRenderer {
	return &ModeBasedEntityRenderer{
		id:     id,
		system: system,
		modes:  map[string]*BasicEntityRenderer{},
	}
}

func (r *ModeBasedEntityRenderer) WithMode(mode string) *ModeBasedEntityRenderer {
	r.currentMode = mode
	return r
}

func (r *ModeBasedEntityRenderer) GenerateEntityContext() *events.BasicEntityContext {
	return events.NewBasicEntityContext(r.id).WithMetadata(types.MetadataKeyMode, func() any {
		return r.currentMode
	})
}

func (r *ModeBasedEntityRenderer) getBasicEntityRenderer(mode string) *BasicEntityRenderer {
	if _, ok := r.modes[mode]; !ok {
		r.modes[mode] = NewBasicEntityRenderer()
	}
	return r.modes[mode]
}

func (r *ModeBasedEntityRenderer) WithModeRenderer(mode string, renderer *BasicEntityRenderer) *ModeBasedEntityRenderer {
	r.modes[mode] = renderer
	return r
}

func (r *ModeBasedEntityRenderer) SetModeAnimations(modeAnimations map[string][]events.AnimationReference) {
	for mode, animationRefs := range modeAnimations {
		var animations []*anim.AnimatedSprite
		for _, animationRef := range animationRefs {
			animations = append(animations, anim.Load(atlas, animationRef.Tilesheet, animationRef.Name))
		}
		r.getBasicEntityRenderer(mode).WithAnimations(animations...)
	}
}

func (r *ModeBasedEntityRenderer) SetModeLights(modeLights map[string][]events.LightConfig) {
	for mode, lightConfigs := range modeLights {
		var lights []*Light
		for _, lightConfig := range lightConfigs {
			lights = append(lights, NewDynamicLight(colors.FromString(lightConfig.Color), lightConfig.Size, lightConfig.Modifier))
		}
		r.getBasicEntityRenderer(mode).WithLights(lights...)
	}
}

func (r *ModeBasedEntityRenderer) ZPriority() int {
	return r.getBasicEntityRenderer(r.currentMode).ZPriority()
}

func (r *ModeBasedEntityRenderer) Update(timeDelta float64) {
	r.getBasicEntityRenderer(r.currentMode).Update(timeDelta)
}

func (r *ModeBasedEntityRenderer) RenderToScene(target pixel.Target, matrix pixel.Matrix) {
	renderer := r.getBasicEntityRenderer(r.currentMode)
	if r.lastRenderMode != r.currentMode {
		renderer.Reset()
	}
	r.lastRenderMode = r.currentMode
	renderer.RenderToScene(target, matrix)
}

func (r *ModeBasedEntityRenderer) RenderToLightMap(target pixel.Target, matrix pixel.Matrix) {
	r.getBasicEntityRenderer(r.currentMode).RenderToLightMap(target, matrix)
}

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
