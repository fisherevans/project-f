package adventure

import (
	"github.com/gopxl/pixel/v2"
	"github.com/gopxl/pixel/v2/backends/opengl"
	"github.com/gopxl/pixel/v2/ext/imdraw"

	"fisherevans.com/project/f/internal/resources"
	"fisherevans.com/project/f/internal/util/colors"
)

// DrawAmbient renders all ambient zones as solid rects with outer glows.
func (s *State) DrawAmbient(target *opengl.Canvas, cameraDelta pixel.Vec, renderBounds MapBounds, imd *imdraw.IMDraw) {
	target.SetComposeMethod(pixel.ComposeOver)
	target.Clear(s.lightClear)
	imd.Clear()

	imd.SetMatrix(pixel.IM.Moved(cameraDelta))

	for _, zone := range s.ambientLightAreas {
		// Calculate glow size in tiles for bounds checking
		glowTiles := int(zone.GlowSize)

		// Check if zone (including glow) is within render bounds
		zoneMinX := zone.X - glowTiles
		zoneMinY := zone.Y - glowTiles
		zoneMaxX := zone.X + zone.W + glowTiles
		zoneMaxY := zone.Y + zone.H + glowTiles

		// Skip if completely outside render bounds
		if zoneMaxX < renderBounds.MinX || zoneMinX > renderBounds.MaxX ||
			zoneMaxY < renderBounds.MinY || zoneMinY > renderBounds.MaxY {
			continue
		}

		x := float64(zone.X) * resources.MapTileSize.Float()
		y := float64(zone.Y) * resources.MapTileSize.Float()
		w := float64(zone.W) * resources.MapTileSize.Float()
		h := float64(zone.H) * resources.MapTileSize.Float()

		// Calculate glow size in pixels
		glowSize := zone.GlowSize * resources.MapTileSize.Float()

		// Draw outer glow gradient (from transparent to zone color)
		if glowSize > 0 {
			transparentColor := colors.WithAlpha(zone.Color, 0)

			// Draw glow as a series of quads from outer edge to inner edge
			// Outer rectangle (transparent)
			outerX1, outerY1 := x-glowSize, y-glowSize
			outerX2, outerY2 := x+w+glowSize, y+h+glowSize

			// Inner rectangle (zone color)
			innerX1, innerY1 := x, y
			innerX2, innerY2 := x+w, y+h

			// Bottom edge glow
			imd.Color = transparentColor
			imd.Push(pixel.V(outerX1, outerY1))
			imd.Color = zone.Color
			imd.Push(pixel.V(innerX1, innerY1))
			imd.Color = zone.Color
			imd.Push(pixel.V(innerX2, innerY1))
			imd.Color = transparentColor
			imd.Push(pixel.V(outerX2, outerY1))
			imd.Polygon(0)

			// Top edge glow
			imd.Color = zone.Color
			imd.Push(pixel.V(innerX1, innerY2))
			imd.Color = transparentColor
			imd.Push(pixel.V(outerX1, outerY2))
			imd.Color = transparentColor
			imd.Push(pixel.V(outerX2, outerY2))
			imd.Color = zone.Color
			imd.Push(pixel.V(innerX2, innerY2))
			imd.Polygon(0)

			// Left edge glow
			imd.Color = transparentColor
			imd.Push(pixel.V(outerX1, outerY1))
			imd.Color = transparentColor
			imd.Push(pixel.V(outerX1, outerY2))
			imd.Color = zone.Color
			imd.Push(pixel.V(innerX1, innerY2))
			imd.Color = zone.Color
			imd.Push(pixel.V(innerX1, innerY1))
			imd.Polygon(0)

			// Right edge glow
			imd.Color = zone.Color
			imd.Push(pixel.V(innerX2, innerY1))
			imd.Color = zone.Color
			imd.Push(pixel.V(innerX2, innerY2))
			imd.Color = transparentColor
			imd.Push(pixel.V(outerX2, outerY2))
			imd.Color = transparentColor
			imd.Push(pixel.V(outerX2, outerY1))
			imd.Polygon(0)
		}

		// Draw filled rectangle for the ambient zone core
		imd.Color = zone.Color
		imd.Push(pixel.V(x, y), pixel.V(x+w, y+h))
		imd.Rectangle(0) // 0 = filled rectangle
	}

	imd.Draw(target)
}
