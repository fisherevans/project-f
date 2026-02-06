package shaders

import (
	_ "embed"

	"github.com/gopxl/glhf/v2"
)

//go:embed flush_transition.glsl
var flushTransitionFragShader string

func (c *Canvas) SetupFlushTransition(background *glhf.Texture) {
	c.SetUniformTexture("uBackgroundTexture", background, 1)
	c.SetUniform("uProgress", float32(0))
	c.setFragmentShaderIfNeeded("flush_transition", flushTransitionFragShader)
}

func (c *Canvas) UpdateFlushTransition(progress float64) {
	c.SetUniform("uProgress", float32(progress))
	c.setFragmentShaderIfNeeded("flush_transition", flushTransitionFragShader)
}
