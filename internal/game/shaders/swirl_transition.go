package shaders

import (
	_ "embed"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/gopxl/glhf/v2"
)

//go:embed swirl_transition.glsl
var swirlTransitionFragShader string

func (c *Canvas) SetupSwirlTransition(background *glhf.Texture, center mgl32.Vec2, radius, swirl, falloff float32) {
	c.SetUniformTexture("uBackgroundTexture", background, 1)
	c.SetUniform("uCenter", center)
	c.SetUniform("uRadius", radius)
	c.SetUniform("uSwirl", swirl)
	c.SetUniform("uFalloff", falloff)
	c.SetUniform("uProgress", float32(0))
	c.setFragmentShaderIfNeeded("swirl_transition", swirlTransitionFragShader)
}

func (c *Canvas) UpdateSwirlTransition(progress float64) {
	c.SetUniform("uProgress", float32(progress))
	c.setFragmentShaderIfNeeded("swirl_transition", swirlTransitionFragShader)
}
