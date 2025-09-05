package shaders

import (
	_ "embed"
	"fmt"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/rs/zerolog/log"
)

//go:embed brightness_threshold.glsl
var brightnessThresholdShader string

func (c *Canvas) SetBrightnessThresholdShader(threshold, thresholdIntensity, knee, desaturate, highlightEpsilon, highlightIntensity float32, highlightColors []mgl32.Vec3) {
	c.SetUniform("uThreshold", threshold)
	c.SetUniform("uThresholdIntensity", thresholdIntensity)
	c.SetUniform("uKnee", knee)
	c.SetUniform("uDesaturate", desaturate)
	if len(highlightColors) > 10 {
		log.Warn().Msgf("too many highlights, ignoring some")
		highlightColors = highlightColors[:10]
	}
	c.SetUniform("uAlwaysCount", int32(len(highlightColors)))
	for i, color := range highlightColors {
		c.SetUniform(fmt.Sprintf("uAlwaysColors[%d]", i), color)
	}
	c.SetUniform("uAlwaysEpsilon", highlightEpsilon)
	c.SetUniform("uAlwaysIntensity", highlightIntensity)
	c.setFragmentShaderIfNeeded("brightness_threshold", brightnessThresholdShader)
}
