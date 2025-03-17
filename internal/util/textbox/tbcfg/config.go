package tbcfg

import (
	"fisherevans.com/project/f/internal/util/gfx"
	"github.com/gopxl/pixel/v2"
)

type Config struct {
	BoxWidth   int
	BoxHeight  int
	ExpandMode ExpandMode
	HAlignment HAlignment
	VAlignment VAlignment

	Foreground pixel.RGBA

	LinesPerPage int
	LineByLine   bool

	ExtraLineSpacing int

	Origin gfx.OriginLocation

	ScrollTimePerLine float64
}

func NewConfig(boxWidth, boxHeight int, opts ...ConfigOpt) Config {
	c := Config{
		BoxWidth:          boxWidth,
		BoxHeight:         boxHeight,
		HAlignment:        AlignLeft,
		VAlignment:        AlignBottom,
		Origin:            gfx.BottomLeft,
		ExpandMode:        ExpandFull,
		Foreground:        pixel.RGB(0, 0, 0),
		LinesPerPage:      0,
		ScrollTimePerLine: 0.2,
	}
	for _, opt := range opts {
		opt(&c)
	}
	return c
}

type ConfigOpt func(c *Config)

func HAligned(alignment HAlignment) func(c *Config) {
	return func(c *Config) {
		c.HAlignment = alignment
	}
}

func VAligned(alignment VAlignment) func(c *Config) {
	return func(c *Config) {
		c.VAlignment = alignment
	}
}

func WithExpandMode(mode ExpandMode) func(c *Config) {
	return func(c *Config) {
		c.ExpandMode = mode
	}
}

func Foreground(foreground pixel.RGBA) func(c *Config) {
	return func(c *Config) {
		c.Foreground = foreground
	}
}

func Paging(linesPerPage int, lineByLine bool) func(c *Config) {
	return func(c *Config) {
		c.LinesPerPage = linesPerPage
		c.LineByLine = lineByLine
	}
}

func RenderFrom(origin gfx.OriginLocation) func(c *Config) {
	return func(c *Config) {
		c.Origin = origin
	}
}

func ExtraLineSpacing(amount int) func(c *Config) {
	return func(c *Config) {
		c.ExtraLineSpacing = amount
	}
}

type HAlignment int

func (a HAlignment) Name() string {
	switch a {
	case AlignLeft:
		return "AlignLeft"
	case AlignCenter:
		return "AlignCenter"
	case AlignRight:
		return "AlignRight"
	default:
		panic("unknown alignment")
	}
}

const (
	AlignLeft HAlignment = iota
	AlignCenter
	AlignRight
)

type VAlignment int

func (a VAlignment) Name() string {
	switch a {
	case AlignTop:
		return "AlignTop"
	case AlignMiddle:
		return "AlignMiddle"
	case AlignBottom:
		return "AlignBottom"
	default:
		panic("unknown alignment")
	}
}

const (
	AlignTop VAlignment = iota
	AlignMiddle
	AlignBottom
)

type ExpandMode int

func (m ExpandMode) Name() string {
	switch m {
	case ExpandFull:
		return "ExpandFull"
	case ExpandFit:
		return "ExpandFit"
	default:
		panic("unknown expand mode")
	}
}

const (
	ExpandFull ExpandMode = iota
	ExpandFit
)
