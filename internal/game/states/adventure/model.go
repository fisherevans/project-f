package adventure

import (
	"github.com/gopxl/pixel/v2"
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
