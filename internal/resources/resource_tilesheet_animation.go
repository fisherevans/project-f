package resources

import "fisherevans.com/project/f/internal/schema"

type SpriteTilesheetAnimation = schema.SpriteTilesheetAnimation
type SpriteTilesheetAnimationHSequence = schema.SpriteTilesheetAnimationHSequence
type SpriteTilesheetAnimationVSequence = schema.SpriteTilesheetAnimationVSequence
type SpriteTilesheetAnimationTile = schema.SpriteTilesheetAnimationTile

var (
	tilesheetAnimations = map[tilesheetAnimationKey]*SpriteTilesheetAnimation{}
)

type tilesheetAnimationKey struct {
	tilesheet string
	name      string
}

func GetTilesheetAnimation(tilesheet string, name string) *SpriteTilesheetAnimation {
	return tilesheetAnimations[tilesheetAnimationKey{tilesheet, name}]
}
