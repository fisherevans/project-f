package adventure

import (
	"fisherevans.com/project/f/internal/game/input"
	"fmt"
	"github.com/gopxl/pixel/v2"
	"math"
)

type MapLocation struct {
	X, Y int
}

func (l MapLocation) String() string {
	return fmt.Sprintf("(%d, %d)", l.X, l.Y)
}

func (l MapLocation) ToVec() pixel.Vec {
	return pixel.V(float64(l.X), float64(l.Y))
}

func (l MapLocation) Moved(dx int, dy int) MapLocation {
	return MapLocation{
		X: l.X + dx,
		Y: l.Y + dy,
	}
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
