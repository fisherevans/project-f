package resources

import (
	"github.com/gopxl/pixel/v2"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

var (
	frames = map[string]*SpriteFrame{}

	resourceFrames = LocalResource{
		FileRoot:        "sprites",
		RequiredTags:    []string{"frame"},
		FileExtension:   "yaml",
		FileLoader:      unmarshaler(&frames, yaml.Unmarshal),
		ResourceEncoder: jsonEncoder,
	}
)

func GetFrame(id string) *SpriteFrame {
	frame, exists := frames[id]
	if !exists {
		log.Error().Msgf("missing frame: %s", id)
	}
	return frame
}

type FrameSide string

const (
	FrameTop         FrameSide = "top"
	FrameTopLeft     FrameSide = "top-left"
	FrameLeft        FrameSide = "left"
	FrameBottomLeft  FrameSide = "bottom-left"
	FrameBottom      FrameSide = "bottom"
	FrameBottomRight FrameSide = "bottom-right"
	FrameRight       FrameSide = "right"
	FrameTopRight    FrameSide = "top-right"
	FrameMiddle      FrameSide = "middle"
)

type SpriteFrame struct {
	CutMargin  map[FrameSide]int       `yaml:"cutMargin"`
	Padding    map[FrameSide]int       `yaml:"padding"`
	FrameModes map[FrameSide]FrameMode `yaml:"frameModes"`
	Defaults   SpriteFrameDefaults     `yaml:"defaults"`

	Sprites map[FrameSide]*SpriteReference `yaml:"-"`
}

type SpriteFrameDefaults struct {
	CutMargin int       `yaml:"cutMargin"`
	Padding   int       `yaml:"padding"`
	FrameMode FrameMode `yaml:"frameMode"`
}

type FrameMode string

const (
	FrameModeRepeat  FrameMode = "repeat"
	FrameModeStretch FrameMode = "stretch"
)

func fillDefaults[T any](ref *map[FrameSide]T, defaultValue T, sides ...FrameSide) {
	if *ref == nil {
		*ref = map[FrameSide]T{}
	}
	m := *ref
	for _, side := range sides {
		if _, ok := m[side]; !ok {
			m[side] = defaultValue
		}
	}
}

func processFrames() {
	for spriteId, frame := range frames {
		sprite, exists := sprites[spriteId]
		if !exists {
			log.Fatal().Msgf("missing sprite for frame: %s", spriteId)
		}

		fillDefaults(&frame.CutMargin, frame.Defaults.CutMargin, FrameTop, FrameLeft, FrameBottom, FrameRight)

		fillDefaults(&frame.Padding, frame.Defaults.Padding, FrameTop, FrameLeft, FrameBottom, FrameRight)

		fillDefaults(&frame.FrameModes, frame.Defaults.FrameMode, FrameTop, FrameLeft, FrameBottom, FrameRight, FrameMiddle)

		frame.splitRectUsingMargins(sprite.Bounds, func(rect pixel.Rect) *SpriteReference {
			return &SpriteReference{
				Source: sprite.Source,
				Bounds: rect,
				Sprite: pixel.NewSprite(sprite.Source, rect),
			}
		})
	}

}

func (sf *SpriteFrame) splitRectUsingMargins(rect pixel.Rect, converter func(pixel.Rect) *SpriteReference) {
	top := float64(sf.CutMargin[FrameTop])
	left := float64(sf.CutMargin[FrameLeft])
	bottom := float64(sf.CutMargin[FrameBottom])
	right := float64(sf.CutMargin[FrameRight])
	sf.Sprites = map[FrameSide]*SpriteReference{
		FrameTopLeft:     converter(pixel.R(rect.Min.X, rect.Max.Y-top, rect.Min.X+left, rect.Max.Y)),
		FrameTop:         converter(pixel.R(rect.Min.X+left, rect.Max.Y-top, rect.Max.X-right, rect.Max.Y)),
		FrameTopRight:    converter(pixel.R(rect.Max.X-right, rect.Max.Y-top, rect.Max.X, rect.Max.Y)),
		FrameLeft:        converter(pixel.R(rect.Min.X, rect.Min.Y+bottom, rect.Min.X+left, rect.Max.Y-top)),
		FrameMiddle:      converter(pixel.R(rect.Min.X+left, rect.Min.Y+bottom, rect.Max.X-right, rect.Max.Y-top)),
		FrameRight:       converter(pixel.R(rect.Max.X-right, rect.Min.Y+bottom, rect.Max.X, rect.Max.Y-top)),
		FrameBottomLeft:  converter(pixel.R(rect.Min.X, rect.Min.Y, rect.Min.X+left, rect.Min.Y+bottom)),
		FrameBottom:      converter(pixel.R(rect.Min.X+left, rect.Min.Y, rect.Max.X-right, rect.Min.Y+bottom)),
		FrameBottomRight: converter(pixel.R(rect.Max.X-right, rect.Min.Y, rect.Max.X, rect.Min.Y+bottom)),
	}
}

func (sf *SpriteFrame) HorizontalPadding() int {
	return sf.RightPadding() + sf.LeftPadding()
}

func (sf *SpriteFrame) VerticalPadding() int {
	return sf.BottomPadding() + sf.TopPadding()
}

func (sf *SpriteFrame) BottomPadding() int {
	return sf.Padding[FrameBottom]
}

func (sf *SpriteFrame) TopPadding() int {
	return sf.Padding[FrameTop]
}

func (sf *SpriteFrame) RightPadding() int {
	return sf.Padding[FrameRight]
}

func (sf *SpriteFrame) LeftPadding() int {
	return sf.Padding[FrameLeft]
}
