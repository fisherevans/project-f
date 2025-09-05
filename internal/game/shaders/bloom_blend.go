package shaders

import (
	_ "embed"

	"github.com/gopxl/glhf/v2"
)

//go:embed bloom_blend.glsl
var bloomBlendShader string

func (c *Canvas) SetBloomBlendShader(
	intensity float32,
	bloomBias float32,
	sceneBias float32,
	base *glhf.Texture,
) {
	c.SetUniform("uIntensity", intensity)
	c.SetUniform("uBloomBias", bloomBias)
	c.SetUniform("uSceneBias", sceneBias)
	c.SetUniformTexture("uBase", base, 0)
	c.setFragmentShaderIfNeeded("blend", bloomBlendShader)
}
