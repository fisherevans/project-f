package resources

import (
	"github.com/rs/zerolog/log"

	"fisherevans.com/project/f/internal/schema"
)

type FrameSide = schema.FrameSide

const (
	FrameTop         = schema.FrameTop
	FrameTopLeft     = schema.FrameTopLeft
	FrameLeft        = schema.FrameLeft
	FrameBottomLeft  = schema.FrameBottomLeft
	FrameBottom      = schema.FrameBottom
	FrameBottomRight = schema.FrameBottomRight
	FrameRight       = schema.FrameRight
	FrameTopRight    = schema.FrameTopRight
	FrameMiddle      = schema.FrameMiddle
)

type FrameMode = schema.FrameMode

const (
	FrameModeRepeat  = schema.FrameModeRepeat
	FrameModeStretch = schema.FrameModeStretch
)

type SpriteFrame = schema.SpriteFrame
type SpriteFrameDefaults = schema.SpriteFrameDefaults
type FrameSpriteId = schema.FrameSpriteId

var (
	frames = map[string]*SpriteFrame{}
)

func GetFrameNames() []string {
	names := make([]string, 0, len(frames))
	for name := range frames {
		names = append(names, name)
	}
	return names
}

func GetFrame(id string) *SpriteFrame {
	frame, exists := frames[id]
	if !exists {
		log.Error().Msgf("missing frame: %s", id)
	}
	return frame
}
