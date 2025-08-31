package shaders

import (
	_ "embed"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/gopxl/glhf/v2"
	"github.com/gopxl/pixel/v2"
	"github.com/gopxl/pixel/v2/backends/opengl"
)

type Options interface {
	SetSwirlShader(
		center mgl32.Vec2,
		radius, swirl, falloff, progress float32,
	)
	Reset()
}

type Canvas struct {
	*opengl.Canvas
}

func NewCanvas(width, height int) *Canvas {
	c := &Canvas{
		Canvas: opengl.NewCanvas(pixel.R(0, 0, float64(width), float64(height))),
	}
	return c
}

func (c *Canvas) Reset() {
	c.SetFragmentShader(opengl.BaseCanvasFragmentShader)
}

/*
Example:
setup
	// configure the custom fragment shader
	//a.litSceneCanvas.SetComposeMethod(pixel.ComposeCopy)
	//a.litSceneCanvas.SetUniformTexture(shaders.UniformLightMap, a.lightMapCanvas.Texture(), 1)
	//a.litSceneCanvas.SetUniform(shaders.UniformLightThreshold, float32(0.8))
	//a.litSceneCanvas.SetUniform(shaders.UniformDarkStrength, float32(0.9))
	//a.litSceneCanvas.SetUniform(shaders.UniformBrightStrength, float32(0.5))
	//a.litSceneCanvas.SetFragmentShader(shaders.CustomFragShader)
	//a.lightMapCanvas.SetComposeMethod(pixel.ComposeOver)

in render
	//s.litSceneCanvas.SetUniformTexture(shaders.UniformLightMap, s.lightMapCanvas.Texture(), 1)
	//s.litSceneCanvas.SetFragmentShader(shaders.CustomFragShader)
	//s.sceneCanvas.Draw(s.litSceneCanvas, pixel.IM.Moved(targetBounds.Center()))
	//s.litSceneCanvas.Draw(target, pixel.IM.Moved(targetBounds.Center()))
*/

func (c *Canvas) SetLightShader(lightMap *glhf.Texture, threshold, darkStrength, brightStrength float32) {
	c.SetUniformTexture(UniformLightMap, lightMap, 1)
	c.SetUniform(UniformLightThreshold, threshold)
	c.SetUniform(UniformDarkStrength, darkStrength)
	c.SetUniform(UniformBrightStrength, brightStrength)
	c.SetFragmentShader(lightingFragShader)
}

// Shader uniform names for lighting.glsl
const (
	// UniformLightMap is the sampler2D for the colored light map.
	// It should be rendered at the same resolution and mapping as the scene.
	UniformLightMap = "uLightMap"

	// UniformLightThreshold is the threshold at which the shader switches
	// from darkening to brightening per channel.
	// Values below this darken the scene; values above this brighten.
	// Typical range: 0.4–0.6, default 0.5.
	UniformLightThreshold = "uLightThreshold"

	// UniformDarkStrength controls how strongly the darkening effect is applied
	// when the light map channel is below the threshold.
	// Range: 0.0–1.0 (0 disables darkening, 1 applies full darkening).
	UniformDarkStrength = "uDarkStrength"

	// UniformBrightStrength controls how strongly the brightening effect is applied
	// when the light map channel is above the threshold.
	// Range: 0.0–1.0 (0 disables brightening, 1 applies full brightening).
	UniformBrightStrength = "uBrightStrength"
)

//go:embed lighting.glsl
var lightingFragShader string

// Shader uniform names for squiggle.glsl
const (
	// Time in seconds (monotonic). Drives wave animation.
	UniformTime = "uTime"
	// Transition progress 0..1.
	UniformProgress = "uProgress"
	// 0 = normal (use p), 1 = reverse (use 1-p) for the incoming scene.
	UniformReverse = "uReverse"
	// Easing exponent (>=1). 2–4 gives a strong finish.
	UniformCurve = "uCurve"

	// Max strengths reached when ramp==1 (scaled by the eased progress).
	UniformAmpMax   = "uAmpMax"   // ~0.02..0.08
	UniformSwirlMax = "uSwirlMax" // ~1.5..4.0
	UniformCABMax   = "uCABMax"   // ~0..0.003

	// Wave parameters.
	UniformFreq1 = "uFreq1" // mgl32.Vec2, e.g. (8,11)
	UniformFreq2 = "uFreq2" // mgl32.Vec2, e.g. (15,9)
	UniformSpeed = "uSpeed" // float32, e.g. 3.0
)

//go:embed squiggle.glsl
var squiggleFragShader string

// SetSquiggleShader sets all uniforms and attaches the squiggle fragment shader.
// Use mgl32.Vec2 for vec2 uniforms so your getAttrType picks them up.
func (c *Canvas) SetSquiggleShader(
	timeSec, progress float32,
	reverse bool,
	curve float32,
	ampMax, swirlMax, cabMax float32,
	freq1, freq2 mgl32.Vec2,
	speed float32,
) {
	c.SetUniform(UniformTime, timeSec)
	c.SetUniform(UniformProgress, progress)
	var rev int32
	if reverse {
		rev = 1
	}
	c.SetUniform(UniformReverse, rev)
	c.SetUniform(UniformCurve, curve)

	c.SetUniform(UniformAmpMax, ampMax)
	c.SetUniform(UniformSwirlMax, swirlMax)
	c.SetUniform(UniformCABMax, cabMax)

	c.SetUniform(UniformFreq1, freq1)
	c.SetUniform(UniformFreq2, freq2)
	c.SetUniform(UniformSpeed, speed)

	c.SetFragmentShader(squiggleFragShader)
}

// Shader uniform names for swirl.glsl
const (
	// UniformCenter is the swirl center in normalized atlas UV (0..1), e.g. (0.5, 0.5).
	UniformCenter = "uCenter"
	// UniformRadius is the fraction (0..1) of the maximum safe radius that fits inside the quad.
	// 1.0 guarantees no sampling outside the texture.
	UniformRadius = "uRadius"
	// UniformSwirl is the rotation in radians at the center; falls off toward the edge.
	// Typical 1.5–4.0.
	UniformSwirl = "uSwirl"
	// UniformFalloff controls softness of the boundary (0 = hard, 1 = very soft).
	UniformFalloff = "uFalloff"
)

//go:embed swirl.glsl
var swirlFragShader string

// SetSwirlShader attaches the swirl shader and sets all uniforms.
// Use mgl32.Vec2 for vec2 so your getAttrType switch handles it.
func (c *Canvas) SetSwirlShader(
	center mgl32.Vec2,
	radius, swirl, falloff, progress float32,
) {
	c.SetUniform(UniformCenter, center)
	c.SetUniform(UniformRadius, radius)
	c.SetUniform(UniformSwirl, swirl)
	c.SetUniform(UniformFalloff, falloff)
	c.SetUniform(UniformProgress, progress)
	c.SetFragmentShader(swirlFragShader)
}
