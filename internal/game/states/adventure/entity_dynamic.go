package adventure

import (
	"fisherevans.com/project/f/internal/game/anim"
	"fisherevans.com/project/f/internal/game/events"
	"fisherevans.com/project/f/internal/game/input"
	"fisherevans.com/project/f/internal/util/colors"
	"github.com/gopxl/pixel/v2"
)

type DynamicEntity struct {
	InnateEntity

	lastRenderMode string

	defaultIsPassable *blockIngressPassable
	modes             map[string]*DynamicEntityMode
}

type DynamicEntityMode struct {
	Passable
	Lights                []*Light
	LightOriginOffset     pixel.Vec
	Animations            []*anim.AnimatedSprite
	AnimationOriginOffset pixel.Vec
}

func NewDynamicEntity(id EntityId, location MapLocation) *DynamicEntity {
	return &DynamicEntity{
		InnateEntity: InnateEntity{
			BaseEntity: BaseEntity{
				id:             id,
				isInteractable: true,
			},
			MapLocation: location,
		},
		defaultIsPassable: newPassablePreventIngress(true),
		modes:             map[string]*DynamicEntityMode{},
	}
}

func (e *DynamicEntity) getDynamicEntityMode(mode string) *DynamicEntityMode {
	if _, ok := e.modes[mode]; !ok {
		e.modes[mode] = &DynamicEntityMode{}
	}
	return e.modes[mode]
}

func (e *DynamicEntity) WithMode(mode string) *DynamicEntity {
	e.mode = mode
	return e
}

func (e *DynamicEntity) WithLights(mode string, lights ...*Light) *DynamicEntity {
	e.getDynamicEntityMode(mode).Lights = lights
	return e
}

func (e *DynamicEntity) WithLightOriginOffset(mode string, offset pixel.Vec) *DynamicEntity {
	e.getDynamicEntityMode(mode).LightOriginOffset = offset
	return e
}

func (e *DynamicEntity) WithAnimations(mode string, animations ...*anim.AnimatedSprite) *DynamicEntity {
	e.getDynamicEntityMode(mode).Animations = animations
	return e
}

func (e *DynamicEntity) WithAnimationOriginOffset(mode string, offset pixel.Vec) *DynamicEntity {
	e.getDynamicEntityMode(mode).AnimationOriginOffset = offset
	return e
}

func (e *DynamicEntity) SetDefaultIsPassable(isPassable bool) {
	e.defaultIsPassable.doBlockIngress = !isPassable
}

func (e *DynamicEntity) getPassable() Passable {
	dm := e.getDynamicEntityMode(e.GetMode())
	if dm.Passable == nil {
		return e.defaultIsPassable
	}
	return dm.Passable
}

func (e *DynamicEntity) CanEgress(side input.Direction, id EntityId) bool {
	return e.getPassable().CanEgress(side, id)
}

func (e *DynamicEntity) CanIngress(side input.Direction, id EntityId) bool {
	return e.getPassable().CanIngress(side, id)
}

func (e *DynamicEntity) RenderScene(target pixel.Target, matrix pixel.Matrix) {
	doReset := e.lastRenderMode != e.GetMode()
	dm := e.getDynamicEntityMode(e.GetMode())
	for _, a := range dm.Animations {
		if doReset {
			a.Reset()
		}
		a.Sprite().Draw(target, matrix.Moved(dm.AnimationOriginOffset))
	}
	e.lastRenderMode = e.GetMode()
}

func (e *DynamicEntity) RenderLight(target pixel.Target, matrix pixel.Matrix) {
	dm := e.getDynamicEntityMode(e.GetMode())
	for _, l := range dm.Lights {
		l.Render(target, matrix.Moved(dm.LightOriginOffset))
	}
}

func (e *DynamicEntity) Update(adv *State, timeDelta float64) {
	dm := e.getDynamicEntityMode(e.GetMode())
	for _, a := range dm.Animations {
		a.Update(timeDelta)
	}
	for _, l := range dm.Lights {
		l.Update(timeDelta)
	}
}

func (e *DynamicEntity) SetDynamicAnimations(modeAnimations map[string][]events.DynamicAnimationReference) {
	for mode, animationRefs := range modeAnimations {
		dm := e.getDynamicEntityMode(mode)
		dm.Animations = nil
		for _, animationRef := range animationRefs {
			dm.Animations = append(dm.Animations, anim.Load(atlas, animationRef.Tilesheet, animationRef.Name))
		}
	}
}

func (e *DynamicEntity) SetDynamicLights(lights map[string][]events.LightConfig) {
	for mode, lightConfigs := range lights {
		dm := e.getDynamicEntityMode(mode)
		dm.Lights = nil
		for _, lightConfig := range lightConfigs {
			dm.Lights = append(dm.Lights, NewDynamicLight(colors.FromString(lightConfig.Color), lightConfig.Size, lightConfig.Modifier))
		}
	}
}
