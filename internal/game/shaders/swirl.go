package shaders

import (
	_ "embed"

	"github.com/go-gl/mathgl/mgl32"
)

//go:embed swirl.glsl
var swirlFragShader string

func (c *Canvas) SetSwirlShader(center mgl32.Vec2, radius, swirl, falloff, progress float32) {
	c.SetUniform("uCenter", center)
	c.SetUniform("uRadius", radius)
	c.SetUniform("uSwirl", swirl)
	c.SetUniform("uFalloff", falloff)
	c.SetUniform("uProgress", progress)
	c.setFragmentShaderIfNeeded("swirl", pixelGridOverlayFrag)
}
