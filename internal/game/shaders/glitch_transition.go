package shaders

import (
	_ "embed"

	"github.com/gopxl/glhf/v2"
)

//go:embed glitch_transition.glsl
var glitchTransitionFragShader string

func (c *Canvas) SetupGlitchTransition(background *glhf.Texture) {
	c.SetUniformTexture("uBackgroundTexture", background, 1)
	c.SetUniform("uProgress", float32(0))
	c.SetUniform("uTime", float32(0))
	c.setFragmentShaderIfNeeded("glitch_transition", glitchTransitionFragShader)
}

func (c *Canvas) UpdateGlitchTransition(progress, time float64) {
	c.SetUniform("uProgress", float32(progress))
	c.SetUniform("uTime", float32(time))
	c.setFragmentShaderIfNeeded("glitch_transition", glitchTransitionFragShader)
}
