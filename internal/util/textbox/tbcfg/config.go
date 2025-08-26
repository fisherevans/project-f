package tbcfg

import (
	"github.com/gopxl/pixel/v2"

	"fisherevans.com/project/f/internal/util/gfx"
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
		HAlignment:        HAlignLeft,
		VAlignment:        VAlignBottom,
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
	case HAlignLeft:
		return "HAlignLeft"
	case HAlignCenter:
		return "HAlignCenter"
	case HAlignRight:
		return "HAlignRight"
	default:
		panic("unknown alignment")
	}
}

const (
	HAlignLeft HAlignment = iota
	HAlignCenter
	HAlignRight
)

type VAlignment int

func (a VAlignment) Name() string {
	switch a {
	case VAlignTop:
		return "VAlignTop"
	case VAlignMiddle:
		return "VAlignMiddle"
	case VAlignBottom:
		return "VAlignBottom"
	default:
		panic("unknown alignment")
	}
}

const (
	VAlignTop VAlignment = iota
	VAlignMiddle
	VAlignBottom
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
