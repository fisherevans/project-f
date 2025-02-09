package textbox

import (
	"fisherevans.com/project/f/internal/resources"
	"github.com/gopxl/pixel/v2"
	"github.com/gopxl/pixel/v2/ext/text"
	"image/color"
)

type Config struct {
	maxWidth   int
	alignment  Alignment
	expandMode ExpandMode

	padding Padding

	foreground color.Color
	background color.Color

	lineCount int
}

func NewConfig(maxWidth int) Config {
	c := Config{
		maxWidth:   maxWidth,
		alignment:  AlignLeft,
		expandMode: ExpandFull,
		foreground: pixel.RGB(.1, .2, .2),
		background: pixel.RGB(.6, .9, .9),
		lineCount:  0,
	}
	c = c.PaddingNormal()
	return c
}

func (c Config) Aligned(alignment Alignment) Config {
	c.alignment = alignment
	return c
}

func (c Config) ExpandMode(mode ExpandMode) Config {
	c.expandMode = mode
	return c
}

func (c Config) Padded(padded Padding) Config {
	c.padding = padded
	return c
}

func (c Config) PaddingNarrow() Config {
	c.padding = Padding{
		x: 4,
		y: 2,
	}
	return c
}

func (c Config) PaddingNormal() Config {
	c.padding = Padding{
		x: 5,
		y: 3,
	}
	return c
}

func (c Config) Foreground(foreground color.Color) Config {
	c.foreground = foreground
	return c
}

func (c Config) Background(background color.Color) Config {
	c.background = background
	return c
}

func (c Config) LineCount(count int) Config {
	c.lineCount = count
	return c
}

type Alignment int

const (
	AlignLeft Alignment = iota
	AlignCenter
	AlignRight
)

type ExpandMode int

const (
	ExpandFull ExpandMode = iota
	ExpandFit
)

type Font struct {
	atlas        *text.Atlas
	letterHeight int
	lineSpacing  int
	tailHeight   int
}

var FontSmall = Font{
	atlas:        resources.Fonts.M6.Atlas,
	letterHeight: 6,
	lineSpacing:  2,
	tailHeight:   2,
}

var FontLarge = Font{
	atlas:        resources.Fonts.M7.Atlas,
	letterHeight: 7,
	lineSpacing:  3,
	tailHeight:   2,
}

type Padding struct {
	x int
	y int
}
