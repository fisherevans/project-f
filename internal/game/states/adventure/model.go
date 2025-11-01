package adventure

import (
	"fmt"
	"math"

	"fisherevans.com/project/f/internal/util/gfx"
	"github.com/gopxl/pixel/v2"

	"fisherevans.com/project/f/internal/game/input"
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

func (l MapLocation) MovedDelta(dx int, dy int) MapLocation {
	return MapLocation{
		X: l.X + dx,
		Y: l.Y + dy,
	}
}

func (l MapLocation) Moved(dir input.Direction) MapLocation {
	return l.MovedDelta(dir.GetVector())
}

func (l MapLocation) DirectionTowards(other MapLocation) input.Direction {
	return DirectionTowards(gfx.IVec(l.X, l.Y), gfx.IVec(other.X, other.Y))
}

func (l MapLocation) DistanceTo(location MapLocation) float64 {
	return l.ToVec().Sub(location.ToVec()).Len()
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
