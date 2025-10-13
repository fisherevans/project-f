package shaders

import (
	_ "embed"

	"github.com/gopxl/glhf/v2"
	"github.com/rs/zerolog/log"
)

//go:embed screen_transition.glsl
var screenTransitionShader string

// SetupScreenTransition initializes the screen transition shader with the given configuration
func (c *Canvas) SetupScreenTransition(background *glhf.Texture, transitionDuration float32) {
	if background == nil {
		log.Fatal().Msg("background canvas is required for screen transition")
		return
	}

	if transitionDuration <= 0 {
		log.Fatal().Msg("transition duration must be positive, defaulting to 0.5")
	}

	// Set the background texture (old content) on texture unit 1
	// Unit 0 is reserved for the primary texture (uTexture)
	c.SetUniformTexture("uBackgroundTexture", background, 1)

	// Set the transition duration
	c.SetUniform("uTransitionDuration", transitionDuration)

	// Initialize time values to 0
	c.SetUniform("uTimeElapsed", float32(0))
	c.SetUniform("uStableTime", float32(0))

	// Set default transition direction (1.0 = expand from center)
	c.SetUniform("uTransitionDirection", float32(1.0))

	// Set default effect intensities (tunable)
	c.SetUniform("uScanlineIntensity", float32(0.045))
	c.SetUniform("uVignetteIntensity", float32(0.333)) // Reduced vignette

	// Set the shader if not already active
	c.setFragmentShaderIfNeeded("screen_transition", screenTransitionShader)
}

// UpdateScreenTransition updates the time elapsed for the transition animation
func (c *Canvas) UpdateScreenTransition(totalTimeElapsed, transitionTimeElapsed float32, middleOut bool, background *glhf.Texture) {
	var direction float32
	if middleOut {
		direction = 1.0
	} else {
		direction = -1.0
	}
	c.SetUniform("uStableTime", totalTimeElapsed)
	c.SetUniform("uTimeElapsed", transitionTimeElapsed)
	c.SetUniform("uTransitionDirection", direction)
	c.SetUniformTexture("uBackgroundTexture", background, 1) // Use texture unit 1 for background
}
