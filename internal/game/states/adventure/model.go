package adventure

import (
	"fisherevans.com/project/f/internal/game/input"
	"github.com/gopxl/pixel/v2"
	"math"
)

type MapLocation struct {
	X, Y int
}

func (l MapLocation) ToVec() pixel.Vec {
	return pixel.V(float64(l.X), float64(l.Y))
}

type MapBounds struct {
	MinX, MinY, MaxX, MaxY int
}

func DirectionTowards(from, to pixel.Vec) input.Direction {
	delta := to.Sub(from)
	if math.Abs(delta.X) >= math.Abs(delta.Y) {
		if delta.X > 0 {
			return input.Right
		}
		return input.Left
	} else {
		if delta.Y > 0 {
			return input.Up
		}
		return input.Down
	}
}
