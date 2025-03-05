package textbox

import (
	"github.com/gopxl/pixel/v2"
)

type Config struct {
	maxWidth   int
	expandMode ExpandMode
	alignment  Alignment

	foreground pixel.RGBA

	linesPerPage int
	lineByLine   bool

	extraLineSpacing int

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

func (c Config) ExtraLineSpacing(amount int) Config {
	c.extraLineSpacing = amount
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

type Origin int

const (
	BottomLeft Origin = iota
	TopLeft
)
