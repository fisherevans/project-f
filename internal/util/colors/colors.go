package colors

import (
	"github.com/gopxl/pixel/v2"
	"github.com/rs/zerolog/log"
)

var (
	Red    = HexColor("#FF0000")
	Purple = HexColor("#FF00FF")
	Blue   = HexColor("#0000FF")
	Green  = HexColor("#00FF00")
	Yellow = HexColor("#FFFF00")
	Orange = HexColor("#FF7F00")

	White     = HexColor("#FFFFFF")
	LightGrey = HexColor("#C0C0C0")
	Grey      = HexColor("#808080")
	DarkGrey  = HexColor("#404040")
	Black     = HexColor("#000000")
)

const (
	NameRed       = "red"
	NamePurple    = "purple"
	NameBlue      = "blue"
	NameGreen     = "green"
	NameYellow    = "yellow"
	NameOrange    = "orange"
	NameWhite     = "white"
	NameLightGrey = "lightgrey"
	NameGrey      = "grey"
	NameDarkGrey  = "darkgrey"
	NameBlack     = "black"
)

// ColorFromName returns a color from a name
func ColorFromName(name string) pixel.RGBA {
	switch name {
	case NameRed:
		return Red
	case NamePurple:
		return Purple
	case NameBlue:
		return Blue
	case NameGreen:
		return Green
	case NameYellow:
		return Yellow
	case NameOrange:
		return Orange
	case NameWhite:
		return White
	case NameLightGrey:
		return LightGrey
	case NameGrey:
		return Grey
	case NameDarkGrey:
		return DarkGrey
	case NameBlack:
		return Black
	default:
		log.Fatal().Msgf("unknown color name: %s", name)
	}
	return pixel.RGBA{}
}
