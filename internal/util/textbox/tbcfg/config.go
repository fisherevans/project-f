package tbcfg

import (
	"fisherevans.com/project/f/internal/util/gfx"
	"github.com/gopxl/pixel/v2"
)

type Config struct {
	MaxWidth   int
	ExpandMode ExpandMode
	Alignment  Alignment

	Foreground pixel.RGBA

	LinesPerPage int
	LineByLine   bool

	ExtraLineSpacing int

	Origin gfx.OriginLocation

	ScrollTimePerLine float64
}

func NewConfig(maxWidth int, opts ...ConfigOpt) Config {
	c := Config{
		MaxWidth:          maxWidth,
		Alignment:         AlignLeft,
		ExpandMode:        ExpandFull,
		Origin:            gfx.BottomLeft,
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

func Aligned(alignment Alignment) func(c *Config) {
	return func(c *Config) {
		c.Alignment = alignment
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

type Alignment int

func (a Alignment) Name() string {
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
	AlignLeft Alignment = iota
	AlignCenter
	AlignRight
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
