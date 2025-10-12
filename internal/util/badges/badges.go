package badges

import (
	"github.com/gopxl/pixel/v2"

	"fisherevans.com/project/f/internal/resources"
	"fisherevans.com/project/f/internal/util/gfx"
)

type Instance interface {
	gfx.Bounded
	Render(target pixel.Target, matrix pixel.Matrix, origin gfx.OriginLocation)
	RenderMask(target pixel.Target, matrix pixel.Matrix, origin gfx.OriginLocation, fg, bg pixel.RGBA)
}

type Builder struct {
	atlas *resources.Atlas
}

func Using(atlas *resources.Atlas) *Builder {
	return &Builder{
		atlas: atlas,
	}
}
