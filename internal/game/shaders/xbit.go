package shaders

import (
	_ "embed"
	"fmt"
	"sort"

	"github.com/gopxl/pixel/v2"
	"github.com/rs/zerolog/log"
)

//go:embed xbit.glsl
var xbitShader string

type ColorThreshold struct {
	Threshold float32
	Color     pixel.RGBA
}

func (c *Canvas) SetXbitShaderWithThresholds(black pixel.RGBA, colors []ColorThreshold, epsilon float32) {
	if len(colors) == 0 {
		log.Warn().Msg("no color thresholds provided")
		return
	}

	// Sort colors by threshold to ensure they're in order
	sortedColors := make([]ColorThreshold, len(colors))
	copy(sortedColors, colors)
	sort.Slice(sortedColors, func(i, j int) bool {
		return sortedColors[i].Threshold < sortedColors[j].Threshold
	})

	// Build palette and thresholds
	palette := make([]pixel.RGBA, len(sortedColors)+1)
	thresholds := make([]float32, len(sortedColors))

	// First color is always black (or the provided base color)
	palette[0] = black

	// Add each color and its threshold
	for i, ct := range sortedColors {
		palette[i+1] = ct.Color
		thresholds[i] = ct.Threshold
	}

	c.SetXbitShader(palette, thresholds, epsilon)
}

func (c *Canvas) SetXbitShaderThresholds(thresholds []float32) {
	for i := 0; i < len(thresholds); i++ {
		c.SetUniform(fmt.Sprintf("uThresholds[%d]", i), thresholds[i])
	}
}

func (c *Canvas) SetXbitShader(palette []pixel.RGBA, thresholds []float32, epsilon float32) {
	if len(palette) < 1 || len(palette) > 16 {
		log.Warn().Msgf("palette must have 1-16 colors, got %d", len(palette))
		return
	}
	if len(thresholds) != len(palette)-1 {
		log.Warn().Msgf("expected %d thresholds for %d colors, got %d",
			len(palette)-1, len(palette), len(thresholds))
		return
	}

	// Set palette colors
	for i := 0; i < 16; i++ {
		var color pixel.RGBA
		if i < len(palette) {
			color = palette[i]
		} else {
			// Fill remaining palette slots with last color
			color = palette[len(palette)-1]
		}
		// Convert pixel.RGBA's float64 to [3]float32 for the shader
		c.SetUniform(fmt.Sprintf("uPalette[%d]", i), RGBAtoVec3(color))
	}

	// Set thresholds (always set all 15 to avoid shader validation issues)
	for i := 0; i < 15; i++ {
		var thresh float32 = 1.0
		if i < len(thresholds) {
			thresh = thresholds[i]
		} else if i < len(palette)-1 {
			// If thresholds were provided but not enough, extend with last threshold
			thresh = thresholds[len(thresholds)-1]
		}
		c.SetUniform(fmt.Sprintf("uThresholds[%d]", i), thresh)
	}

	c.SetUniform("uThresholdCount", int32(len(palette)))
	c.SetUniform("uPaletteEpsilon", epsilon)

	// Set the shader if not already active
	c.setFragmentShaderIfNeeded("xbit", xbitShader)
}
