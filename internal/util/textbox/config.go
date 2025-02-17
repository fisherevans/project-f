package textbox

import (
	"fisherevans.com/project/f/internal/resources"
	"github.com/gopxl/pixel/v2"
	"github.com/gopxl/pixel/v2/ext/text"
)

type Config struct {
	maxWidth   int
	expandMode ExpandMode
	alignment  Alignment

	foreground pixel.RGBA

	linesPerPage int
	lineByLine   bool

	origin Origin

	scrollTimePerLine float64
}

func NewConfig(maxWidth int) Config {
	c := Config{
		maxWidth:          maxWidth,
		alignment:         AlignLeft,
		expandMode:        ExpandFull,
		foreground:        pixel.RGB(0, 0, 0),
		linesPerPage:      0,
		scrollTimePerLine: 0.2,
	}
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

func (c Config) Foreground(foreground pixel.RGBA) Config {
	c.foreground = foreground
	return c
}

func (c Config) Paging(linesPerPage int, lineByLine bool) Config {
	c.linesPerPage = linesPerPage
	c.lineByLine = lineByLine
	return c
}

func (c Config) RenderFrom(origin Origin) Config {
	c.origin = origin
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
	lineSpacing:  3,
	tailHeight:   2,
}

var FontLarge = Font{
	atlas:        resources.Fonts.M7.Atlas,
	letterHeight: 7,
	lineSpacing:  3,
	tailHeight:   2,
}

var FontLargeSpaced = Font{
	atlas:        resources.Fonts.M7.Atlas,
	letterHeight: 10,
	lineSpacing:  3,
	tailHeight:   2,
}

func (f Font) GetAtlas() *text.Atlas {
	return f.atlas
}

type Origin int

const (
	BottomLeft Origin = iota
	TopLeft
)
