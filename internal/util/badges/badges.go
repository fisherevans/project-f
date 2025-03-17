package badges

import (
	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/resources"
	"fisherevans.com/project/f/internal/util/gfx"
	"github.com/gopxl/pixel/v2"
)

type Instance interface {
	gfx.Bounded
	Render(ctx *game.Context, target pixel.Target, matrix pixel.Matrix, origin gfx.OriginLocation)
}

type Builder struct {
	atlas *resources.Atlas
}

func Using(atlas *resources.Atlas) *Builder {
	return &Builder{
		atlas: atlas,
	}
}

func alignFromBottomLeft(matrix pixel.Matrix, origin gfx.OriginLocation, width, height int) pixel.Matrix {
	matrix = matrix.Moved(gfx.IVec(-width/2, -height/2)) // move from bottom left to center first
	return matrix.Moved(origin.AlignInt(width, height))
}
