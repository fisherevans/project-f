package util

import (
	"image/color"
	"log"
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
