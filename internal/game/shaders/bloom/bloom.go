package bloom

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/gopxl/pixel/v2"
	"github.com/gopxl/pixel/v2/backends/opengl"

	"fisherevans.com/project/f/internal/game/shaders"
)

type BlendConfig struct {
	Intensity float32
	BloomBias float32
	SceneBias float32
}

func DefaultBlendConfig() BlendConfig {
	return BlendConfig{
		Intensity: 4,
		BloomBias: 1.75,
		SceneBias: 1.0,
	}
}

type BrightnessConfig struct {
	Threshold          float32
	ThresholdIntensity float32
	Knee               float32
	Desaturate         float32
	HighlightEpsilon   float32
	HighlightIntensity float32
	HighlightColors    []mgl32.Vec3
}

func DefaultBrightnessConfig() BrightnessConfig {
	return BrightnessConfig{
		Threshold:          0.9,
		ThresholdIntensity: 0.5,
		Knee:               0.15,
		Desaturate:         0.0,
		HighlightEpsilon:   0.05,
		HighlightIntensity: 0.9,
	}
}

type BlurConfig struct {
	Passes      int
	Radius      int32
	SigmaFactor float32
}

func DefaultBlurConfig() BlurConfig {
	return BlurConfig{
		Passes:      5,
		Radius:      7,
		SigmaFactor: 0.1,
	}
}

type Helper struct {
	BrightnessConfig
	BlurConfig
	BlendConfig

	m          pixel.Matrix
	brightness *shaders.Canvas
	blurH      *shaders.Canvas
	blurV      *shaders.Canvas
	blend      *shaders.Canvas
}

func NewHelper(width, height int, brightnessCfg BrightnessConfig, blurCfg BlurConfig, blendCfg BlendConfig) *Helper {
	b := &Helper{
		BrightnessConfig: brightnessCfg,
		BlurConfig:       blurCfg,
		BlendConfig:      blendCfg,

		m:          pixel.IM.Moved(pixel.V(float64(width)/2.0, float64(height)/2.0)),
		brightness: shaders.NewCanvas(width, height),
		blurH:      shaders.NewCanvas(width, height),
		blurV:      shaders.NewCanvas(width, height),
		blend:      shaders.NewCanvas(width, height),
	}
	return b
}

func (b *Helper) GenerateBloomCanvas(scene *opengl.Canvas) *shaders.Canvas {
	// draw colors to brightness threshold canvas - i.e. get the highlights only
	b.brightness.Clear(pixel.RGBA{})
	b.brightness.SetBrightnessThresholdShader(
		b.Threshold,
		b.ThresholdIntensity,
		b.Knee,
		b.Desaturate,
		b.HighlightEpsilon,
		b.HighlightIntensity,
		b.HighlightColors,
	)
	scene.Draw(b.brightness, b.m)

	// blur the highlights
	b.blurH.Clear(pixel.RGBA{})
	b.blurH.SetBlurShader(
		true,
		b.Radius,
		b.SigmaFactor,
	)
	b.blurV.Clear(pixel.RGBA{})
	b.blurV.SetBlurShader(
		false,
		b.Radius,
		b.SigmaFactor,
	)
	blurred := b.brightness
	for i := 0; i < b.Passes; i++ {
		b.blurH.Clear(pixel.RGBA{})
		blurred.Draw(b.blurH, b.m)

		b.blurV.Clear(pixel.RGBA{})
		b.blurH.Draw(b.blurV, b.m)

		blurred = b.blurV
	}

	return blurred
}

func (b *Helper) ApplyBloom(scene *opengl.Canvas) *shaders.Canvas {
	blurred := b.GenerateBloomCanvas(scene)

	// blend the blurred highlights with the scene
	b.blend.Clear(pixel.RGBA{})
	b.blend.SetBloomBlendShader(b.Intensity, b.BloomBias, b.SceneBias, scene.Texture())
	blurred.Draw(b.blend, b.m)

	return b.blend
}
