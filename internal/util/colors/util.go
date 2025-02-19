package colors

import (
	"github.com/gopxl/pixel/v2"
	"github.com/rs/zerolog/log"
	"hash/fnv"
	"math"
	"strconv"
	"strings"
)

type NamedColor pixel.RGBA

var colorsByName = map[string]NamedColor{}

func (c NamedColor) register(name string) pixel.RGBA {
	if _, exists := colorsByName[name]; exists {
		log.Fatal().Msgf("color name already exists: %s", name)
	}
	colorsByName[name] = c
	return pixel.RGBA(c)
}

func ColorFromName(name string) pixel.RGBA {
	if color, exists := colorsByName[name]; exists {
		return pixel.RGBA(color)
	}
	log.Error().Msgf("color name not found: %s", name)
	return Black
}

const hexErrMsg = "failed to parse color hex value"

// HexColor converts #RGB, #RGBA, #RRGGBB, and #RRGGBBAA hex codes to colors (with or withou leading #)
func HexColor(hex string) pixel.RGBA {
	hex = strings.TrimPrefix(hex, "#")
	var r, g, b, a uint8 = 0, 0, 0, 255 // Default alpha to 255 (fully opaque)

	switch len(hex) {
	case 3: // #RGB
		r = parseHexDigit(hex[0]) * 17
		g = parseHexDigit(hex[1]) * 17
		b = parseHexDigit(hex[2]) * 17
	case 4: // #RGBA
		r = parseHexDigit(hex[0]) * 17
		g = parseHexDigit(hex[1]) * 17
		b = parseHexDigit(hex[2]) * 17
		a = parseHexDigit(hex[3]) * 17
	case 6: // #RRGGBB
		r = parseHexByte(hex[0:2])
		g = parseHexByte(hex[2:4])
		b = parseHexByte(hex[4:6])
	case 8: // #RRGGBBAA
		r = parseHexByte(hex[0:2])
		g = parseHexByte(hex[2:4])
		b = parseHexByte(hex[4:6])
		a = parseHexByte(hex[6:8])
	default:
		log.Fatal().Msgf("%s: %s", hexErrMsg, hex)

	}

	// Convert uint8 values to pixel.RGBA (normalized to 0-1 range)
	return pixel.RGBA{
		R: float64(r) / 255,
		G: float64(g) / 255,
		B: float64(b) / 255,
		A: float64(a) / 255,
	}
}

func parseHexDigit(digit byte) uint8 {
	val, err := strconv.ParseUint(string(digit), 16, 8)
	if err != nil {
		panic("invalid hex digit")
	}
	return uint8(val)
}

func parseHexByte(hexStr string) uint8 {
	val, err := strconv.ParseUint(hexStr, 16, 8)
	if err != nil {
		panic("invalid hex byte")
	}
	return uint8(val)
}

func HSLToRGB(h, s, l float64) (float64, float64, float64) {
	c := (1 - math.Abs(2*l-1)) * s
	x := c * (1 - math.Abs(math.Mod(h/60, 2)-1))
	m := l - c/2

	var r, g, b float64
	switch {
	case h >= 0 && h < 60:
		r, g, b = c, x, 0
	case h >= 60 && h < 120:
		r, g, b = x, c, 0
	case h >= 120 && h < 180:
		r, g, b = 0, c, x
	case h >= 180 && h < 240:
		r, g, b = 0, x, c
	case h >= 240 && h < 300:
		r, g, b = x, 0, c
	case h >= 300 && h < 360:
		r, g, b = c, 0, x
	default:
		r, g, b = 0, 0, 0 // Fallback for unexpected input
	}

	return r + m, g + m, b + m
}

// StringToColor generates a stable color from a string with a fixed saturation and brightness.
func StringToColor(input string, saturation, lightness float64) pixel.RGBA {
	hasher := fnv.New32a()
	hasher.Write([]byte(input))
	hash := hasher.Sum32()

	// Map the hash to a hue value (0-360 degrees)
	hue := float64(hash % 360)
	r, g, b := HSLToRGB(hue, saturation, lightness)
	return pixel.RGB(r, g, b)
}

func ScaleColor(c pixel.RGBA, v float64) pixel.RGBA {
	return pixel.RGBA{
		R: c.R * v,
		G: c.G * v,
		B: c.B * v,
		A: c.A,
	}
}

func WithAlpha(c pixel.RGBA, a float64) pixel.RGBA {
	return pixel.RGBA{
		R: c.R,
		G: c.G,
		B: c.B,
		A: a,
	}
}

func Alpha(alpha float64) pixel.RGBA {
	return pixel.RGBA{
		R: alpha,
		G: alpha,
		B: alpha,
		A: alpha,
	}
}
