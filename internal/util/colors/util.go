package colors

import (
	"math"
	"strconv"
	"strings"

	"github.com/gopxl/pixel/v2"
	"github.com/rs/zerolog/log"
)

type ColorName string
type NamedColor struct {
	RGBA pixel.RGBA
	Name ColorName
}

var colorsByName = map[ColorName]NamedColor{}
var colorsByRGBA = map[pixel.RGBA]NamedColor{}

func registerNamedColor(rgba pixel.RGBA, name ColorName) NamedColor {
	namedColor := NamedColor{RGBA: rgba, Name: name}
	// only allow unique names
	if _, exists := colorsByName[name]; exists {
		log.Fatal().Msgf("color name already exists: %s", name)
	}
	// only register the first RGBA lookup
	if _, exists := colorsByRGBA[rgba]; !exists {
		colorsByRGBA[rgba] = namedColor
	}
	colorsByName[name] = namedColor
	return namedColor
}

func FromString(s string) pixel.RGBA {
	if strings.HasPrefix(s, "#") {
		return HexString(s)
	} else {
		return ColorFromName(ColorName(s)).RGBA
	}
}

func ColorFromName(name ColorName) NamedColor {
	if color, exists := colorsByName[name]; exists {
		return color
	}
	log.Error().Msgf("color name not found: %s", name)
	return Black
}

func Hex(u uint32) pixel.RGBA {
	if u < 0xfff {
		r := (u >> 8) & 0xF
		g := (u >> 4) & 0xF
		b := u & 0xF
		r = r * 17 // duplicate nibble
		g = g * 17
		b = b * 17
		return pixel.RGBA{
			R: float64(r) / 255,
			G: float64(g) / 255,
			B: float64(b) / 255,
			A: 1,
		}
	}
	return pixel.RGBA{
		R: float64((u>>16)&0xFF) / 255,
		G: float64((u>>8)&0xFF) / 255,
		B: float64(u&0xFF) / 255,
		A: 1,
	}
}

const hexErrMsg = "failed to parse color hex value"

// HexString converts #RGB, #RGBA, #RRGGBB, and #RRGGBBAA hex codes to colors (with or withou leading #)
func HexString(originalHex string) pixel.RGBA {
	hex := strings.TrimPrefix(originalHex, "#")
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
		log.Fatal().Msgf("%s: %s", hexErrMsg, originalHex)

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

func ToHex(c pixel.RGBA) string {
	return "#" + toHexByte(c.R) + toHexByte(c.G) + toHexByte(c.B) + toHexByte(c.A)
}

func toHexByte(r float64) string {
	return strconv.FormatInt(int64(r*255), 16)
}

// HSLToRGBA converts HSL color values (all in range [0, 1]) to RGBA
func HSLToRGBA(h, s, l float64) pixel.RGBA {
	// Scale hue from [0,1] to [0,360] degrees
	hue := h * 360.0

	c := (1 - math.Abs(2*l-1)) * s
	x := c * (1 - math.Abs(math.Mod(hue/60, 2)-1))
	m := l - c/2

	var r, g, b float64
	switch {
	case hue < 60:
		r, g, b = c, x, 0
	case hue < 120:
		r, g, b = x, c, 0
	case hue < 180:
		r, g, b = 0, c, x
	case hue < 240:
		r, g, b = 0, x, c
	case hue < 300:
		r, g, b = x, 0, c
	default:
		r, g, b = c, 0, x
	}

	// Ensure values are clamped to [0,1] before conversion
	clamp := func(v float64) float64 {
		if v < 0 {
			return 0
		}
		if v > 1 {
			return 1
		}
		return v
	}

	return pixel.RGB(clamp(r+m), clamp(g+m), clamp(b+m))
}

func ScaleColor(c pixel.RGBA, v float64) pixel.RGBA {
	return pixel.RGBA{
		R: c.R * v,
		G: c.G * v,
		B: c.B * v,
		A: c.A,
	}
}

func MixColor(a, b pixel.RGBA) pixel.RGBA {
	return pixel.RGBA{
		R: a.R * b.R,
		G: a.G * b.G,
		B: a.B * b.B,
		A: a.A * b.A,
	}
}

func WithAlpha(c pixel.RGBA, a float64) pixel.RGBA {
	unalpha := 1.0 / c.A
	return pixel.RGBA{
		R: c.R * unalpha * a,
		G: c.G * unalpha * a,
		B: c.B * unalpha * a,
		A: a,
	}
}

func LayerAlpha(c pixel.RGBA, a float64) pixel.RGBA {
	return pixel.RGBA{
		R: c.R * a,
		G: c.G * a,
		B: c.B * a,
		A: c.A * a,
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
func Lerp(from, to pixel.RGBA, t float64) pixel.RGBA {
	if t < 0 {
		t = 0
	} else if t > 1 {
		t = 1
	}
	return pixel.RGBA{
		R: from.R*(1-t) + to.R*t,
		G: from.G*(1-t) + to.G*t,
		B: from.B*(1-t) + to.B*t,
		A: from.A*(1-t) + to.A*t,
	}
}

func GammaLerp(from, to pixel.RGBA, t float64) pixel.RGBA {
	// Clamp progress to [0,1]
	if t < 0 {
		t = 0
	} else if t > 1 {
		t = 1
	}

	// sRGB <-> linear helpers
	srgbToLinear := func(c float64) float64 {
		if c <= 0.04045 {
			return c / 12.92
		}
		return math.Pow((c+0.055)/1.055, 2.4)
	}
	linearToSrgb := func(c float64) float64 {
		if c <= 0.0031308 {
			return 12.92 * c
		}
		return 1.055*math.Pow(c, 1.0/2.4) - 0.055
	}
	lerp := func(a, b, tt float64) float64 {
		return a*(1-tt) + b*tt
	}
	clamp01 := func(c float64) float64 {
		if c < 0 {
			return 0
		}
		if c > 1 {
			return 1
		}
		return c
	}

	// Convert to linear, interpolate RGB in linear space
	rLin := lerp(srgbToLinear(from.R), srgbToLinear(to.R), t)
	gLin := lerp(srgbToLinear(from.G), srgbToLinear(to.G), t)
	bLin := lerp(srgbToLinear(from.B), srgbToLinear(to.B), t)

	// Convert back to sRGB and lerp alpha linearly
	return pixel.RGBA{
		R: clamp01(linearToSrgb(rLin)),
		G: clamp01(linearToSrgb(gLin)),
		B: clamp01(linearToSrgb(bLin)),
		A: clamp01(lerp(from.A, to.A, t)),
	}
}
