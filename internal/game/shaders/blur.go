package shaders

import (
	_ "embed"

	"github.com/go-gl/mathgl/mgl32"
)

//go:embed blur.glsl
var blurShader string

func (c *Canvas) SetBlurShader(
	horizontal bool,
	radius int32,
	sigmaFactor float32,
) {
	if horizontal {
		c.SetUniform("uDirection", mgl32.Vec2{1, 0})
	} else {
		c.SetUniform("uDirection", mgl32.Vec2{0, 1})
	}
	c.SetUniform("uRadius", radius)
	sigma := float32(radius) * sigmaFactor
	if sigma < 1.0 {
		sigma = 1.0
	}
	c.SetUniform("uSigma", sigma)
	c.setFragmentShaderIfNeeded("blur", blurShader)
}
