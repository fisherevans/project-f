package adventure

import (
	"fisherevans.com/project/f/internal/game/anim"
	"fisherevans.com/project/f/internal/util"
	"fisherevans.com/project/f/internal/util/colors"
	"github.com/gopxl/pixel/v2"
)

type ModeBaseRenderConfig struct {
	Mode       *string                         `yaml:"mode"`
	Animations map[string][]AnimationReference `yaml:"animations"`
	Lights     map[string][]LightConfig        `yaml:"lights"`
}

type LightConfig struct {
	Color    string  `yaml:"color"`
	Size     float64 `yaml:"size"`
	Modifier *string `yaml:"modifier"`
}

type AnimationReference struct {
	Name      string  `yaml:"name"`
	ColorMask *string `yaml:"colorMask"`
}

type ModeBasedEntityRenderer struct {
	entity Entity

	currentMode    string
	lastRenderMode string
	modes          map[string]*BasicEntityRenderer
}

func AttachModeBasedEntityRenderer(entity Entity) *ModeBasedEntityRenderer {
	renderer := &ModeBasedEntityRenderer{
		entity: entity,
		modes:  map[string]*BasicEntityRenderer{},
	}
	entity.SetRenderer(renderer)
	return renderer
}

func (r *ModeBasedEntityRenderer) WithMode(mode string) *ModeBasedEntityRenderer {
	r.currentMode = mode
	return r
}

func (r *ModeBasedEntityRenderer) getBasicEntityRenderer(mode string) *BasicEntityRenderer {
	if _, ok := r.modes[mode]; !ok {
		r.modes[mode] = NewBasicEntityRenderer(r.entity)
	}
	return r.modes[mode]
}

func (r *ModeBasedEntityRenderer) WithModeRenderer(mode string, renderer *BasicEntityRenderer) *ModeBasedEntityRenderer {
	r.modes[mode] = renderer
	return r
}

func (r *ModeBasedEntityRenderer) SetModeAnimations(modeAnimations map[string][]AnimationReference) {
	for mode, animationRefs := range modeAnimations {
		var animations []ColorMaskAnimation
		for _, animationRef := range animationRefs {
			cma := ColorMaskAnimation{
				Animation: anim.Load(atlas, animationRef.Name),
			}
			if animationRef.ColorMask != nil {
				cma.ColorMask = util.Ptr(colors.FromString(*animationRef.ColorMask))
			}
			animations = append(animations, cma)
		}
		r.getBasicEntityRenderer(mode).WithColorMaskAnimations(animations...)
	}
}

func (r *ModeBasedEntityRenderer) SetModeLights(modeLights map[string][]LightConfig) {
	for mode, lightConfigs := range modeLights {
		var lights []*Light
		for _, lightConfig := range lightConfigs {
			var l *Light
			c := colors.FromString(lightConfig.Color)
			if lightConfig.Modifier == nil {
				l = NewLight(c, lightConfig.Size)
			} else {
				l = NewLightWithModifier(c, lightConfig.Size, *lightConfig.Modifier)
			}
			lights = append(lights, l)
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

func (r *ModeBasedEntityRenderer) WithPropConfigurations(props *util.Properties) *ModeBasedEntityRenderer {
	config := &ModeBaseRenderConfig{}
	if !props.LoadStructFromKey("render_config", config) {
		return r
	}
	return r.WithConfig(config)
}

func (r *ModeBasedEntityRenderer) WithConfig(config *ModeBaseRenderConfig) *ModeBasedEntityRenderer {
	if config == nil {
		return r
	}
	if config.Mode != nil {
		r.currentMode = *config.Mode
	}
	r.SetModeAnimations(config.Animations)
	r.SetModeLights(config.Lights)
	return r
}
