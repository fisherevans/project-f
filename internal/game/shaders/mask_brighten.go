package shaders

import (
	_ "embed"

	"github.com/gopxl/glhf/v2"
)

//go:embed mask_brighten.glsl
var maskBrightenShader string

func (c *Canvas) SetMaskBrightenShader(
	mask *glhf.Texture,
) {
	c.SetUniformTexture("uMask", mask, 1)
	c.setFragmentShaderIfNeeded("mask_brighten", maskBrightenShader)
}
