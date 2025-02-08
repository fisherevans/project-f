package util

import (
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

// HSLToRGB converts HSL color values to RGB.
func HSLToRGB(h, s, l float64) color.Color {
	c := (1 - math.Abs(2*l-1)) * s
	x := c * (1 - math.Abs(math.Mod(h/60, 2)-1))
	m := l - c/2

	var r, g, b float64

	switch {
	case 0 <= h && h < 60:
		r, g, b = c, x, 0
	case 60 <= h && h < 120:
		r, g, b = x, c, 0
	case 120 <= h && h < 180:
		r, g, b = 0, c, x
	case 180 <= h && h < 240:
		r, g, b = 0, x, c
	case 240 <= h && h < 300:
		r, g, b = x, 0, c
	case 300 <= h && h < 360:
		r, g, b = c, 0, x
	}

	return color.RGBA{
		R: uint8((r + m) * 255),
		G: uint8((g + m) * 255),
		B: uint8((b + m) * 255),
		A: 255,
	}
}

// StringToColor generates a stable color from a string with a fixed saturation and brightness.
func StringToColor(input string, saturation, lightness float64) color.Color {
	hasher := fnv.New32a()
	hasher.Write([]byte(input))
	hash := hasher.Sum32()

	// Map the hash to a hue value (0-360 degrees)
	hue := float64(hash % 360)

	return HSLToRGB(hue, saturation, lightness)
}
