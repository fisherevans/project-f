package schema

import (
    "fmt"
    "image"
)

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

type FrameMode string

const (
    FrameModeRepeat  FrameMode = "repeat"
    FrameModeStretch FrameMode = "stretch"
)

type SpriteFrame struct {
    CutMargin  map[FrameSide]int       `yaml:"cutMargin"`
    Padding    map[FrameSide]int       `yaml:"padding"`
    FrameModes map[FrameSide]FrameMode `yaml:"frameModes"`
    Defaults   SpriteFrameDefaults     `yaml:"defaults"`
}

func FillDefaults[T any](ref *map[FrameSide]T, defaultValue T, sides ...FrameSide) {
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

func (sf *SpriteFrame) Init(image.Image) {
    FillDefaults(&sf.CutMargin, sf.Defaults.CutMargin, FrameTop, FrameLeft, FrameBottom, FrameRight)
    FillDefaults(&sf.Padding, sf.Defaults.Padding, FrameTop, FrameLeft, FrameBottom, FrameRight)
    FillDefaults(&sf.FrameModes, sf.Defaults.FrameMode, FrameTop, FrameLeft, FrameBottom, FrameRight, FrameMiddle)
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

type SpriteFrameDefaults struct {
    CutMargin int       `yaml:"cutMargin"`
    Padding   int       `yaml:"padding"`
    FrameMode FrameMode `yaml:"frameMode"`
}

type FrameSpriteId struct {
    Frame string    `json:"frame"`
    Side  FrameSide `json:"side"`
}

func (s FrameSpriteId) String() string {
    return fmt.Sprintf("f:%s,s:%s", s.Frame, s.Side)
}
