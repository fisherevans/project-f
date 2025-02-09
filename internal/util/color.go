package util

import (
	"github.com/gopxl/pixel/v2"
	"hash/fnv"
	"image/color"
	"log"
	"math"
	"strconv"
	"strings"
)

const hexErrMsg = "failed to parse color hex value"

// HexColor converts #RGB, #RGBA, #RRGGBB, and #RRGGBBAA hex codes to colors (with or withou leading #)
func HexColor(hex string) color.Color {
	hex = strings.TrimPrefix(hex, "#")

	var r, g, b, a uint64
	var err error

	a = 255
	l := len(hex)
	if l == 3 || l == 4 {
		r, err = strconv.ParseUint(string(hex[0])+string(hex[0]), 16, 8)
		if err != nil {
			log.Fatal(hexErrMsg, hex, err)
		}
		g, err = strconv.ParseUint(string(hex[1])+string(hex[1]), 16, 8)
		if err != nil {
			log.Fatal(hexErrMsg, hex, err)
		}
		b, err = strconv.ParseUint(string(hex[2])+string(hex[2]), 16, 8)
		if err != nil {
			log.Fatal(hexErrMsg, hex, err)
		}
		if l == 4 {
			a, err = strconv.ParseUint(string(hex[3])+string(hex[3]), 16, 8)
			if err != nil {
				log.Fatal(hexErrMsg, hex, err)
			}
		}
	} else if l == 6 || l == 8 {
		r, err = strconv.ParseUint(hex[0:2], 16, 8)
		if err != nil {
			log.Fatal(hexErrMsg, hex, err)
		}
		g, err = strconv.ParseUint(hex[2:4], 16, 8)
		if err != nil {
			log.Fatal(hexErrMsg, hex, err)
		}
		b, err = strconv.ParseUint(hex[4:6], 16, 8)
		if err != nil {
			log.Fatal(hexErrMsg, hex, err)
		}
		if l == 8 {
			a, err = strconv.ParseUint(hex[6:8], 16, 8)
			if err != nil {
				log.Fatal(hexErrMsg, hex, err)
			}
		}
	} else {
		log.Fatal(hexErrMsg, hex)
	}

	return color.RGBA{uint8(r), uint8(g), uint8(b), uint8(a)}
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
