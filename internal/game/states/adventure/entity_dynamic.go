package adventure

import (
	"fisherevans.com/project/f/internal/game/anim"
	"fisherevans.com/project/f/internal/game/events"
	"github.com/gopxl/pixel/v2"
)

type DynamicEntity struct {
	InnateEntity
	lights     map[string][]*Light
	animations map[string][]*anim.AnimatedSprite

	lastMode string
}

func NewDynamicEntity(id EntityId, location MapLocation) *DynamicEntity {
	return &DynamicEntity{
		InnateEntity: InnateEntity{
			BaseEntity: BaseEntity{
				Id:       id,
				Passable: false,
			},
			MapLocation: location,
		},
	}
}

func forEach[T any](key string, valueMap map[string][]T, do func(T)) {
	values, ok := valueMap[key]
	if !ok {
		return
	}
	for _, value := range values {
		do(value)
	}
}

func (e *DynamicEntity) RenderScene(target pixel.Target, matrix pixel.Matrix) {
	forEach(e.GetMode(), e.animations, func(a *anim.AnimatedSprite) {
		a.Sprite().Draw(target, matrix)
	})
}

func (e *DynamicEntity) RenderLight(target pixel.Target, matrix pixel.Matrix) {
	forEach(e.GetMode(), e.lights, func(l *Light) {
		l.Render(target, matrix)
	})
}

func (e *DynamicEntity) Update(adv *State, timeDelta float64) {
	doReset := e.lastMode != e.GetMode()
	forEach(e.GetMode(), e.animations, func(a *anim.AnimatedSprite) {
		if doReset {
			a.Reset()
		}
		a.Update(timeDelta)
	})
	forEach(e.GetMode(), e.lights, func(l *Light) {
		l.Update(timeDelta)
	})
	e.lastMode = e.GetMode()
}

func (e *DynamicEntity) SetDynamicAnimations(modeAnimations map[string][]events.DynamicAnimationReference) {
	for mode, animationRefs := range modeAnimations {
		var animations []*anim.AnimatedSprite
		for _, animationRef := range animationRefs {
			animations = append(animations, anim.Load(atlas, animationRef.Tilesheet, animationRef.Name))
		}
		e.animations[mode] = animations
	}
}

func (e *DynamicEntity) SetDynamicLights(lights map[string][]events.LightConfig) {
	for mode, lightConfigs := range lights {
		var lights []*Light
		for _, lightConfig := range lightConfigs {
			lights = append(lights, NewDynamicLight(lightConfig.Color, lightConfig.Size, lightConfig.Modifier))
		}
		e.lights[mode] = lights
	}
}
