package map_editor

import (
	"fisherevans.com/project/f/internal/game"
	resources2 "fisherevans.com/project/f/internal/resources"
	"github.com/gopxl/pixel/v2"
	"github.com/gopxl/pixel/v2/backends/opengl"
	"image/color"
	"math"
)

var keyboardMoveRate = resources2.TileSizeF64 * 5
var mouseScrollRate = 3.0
var cameraLagSpeed = 5.0

type MapEditor struct {
	cameraLocation pixel.Vec
	cameraLag      pixel.Vec
	cameraMatrix   pixel.Matrix

	swatch *Swatch

	selectedMap     string
	selectedLayer   resources2.MapLayerName
	layerRenderMode layerRenderMode

	win *opengl.Window
}

type layerRenderMode string

const (
	layerRenderAll         layerRenderMode = "all"
	layerRenderSelected    layerRenderMode = "selected"
	layerRenderMix         layerRenderMode = "mix"
	layerRenderBelow       layerRenderMode = "below"
	layerRenderTransparent layerRenderMode = "transparent"
)

type Context struct {
	*game.Context
	MouseMapLocation Location
	MouseInCanvas    bool
}

type Location struct {
	X, Y int
}

func New(window *opengl.Window) *MapEditor {
	return &MapEditor{
		cameraMatrix:    pixel.IM,
		swatch:          newSwatch(),
		selectedMap:     "dummy",
		selectedLayer:   resources2.LayerBase,
		layerRenderMode: layerRenderMix,
		win:             window,
	}
}

func (m *MapEditor) OnTick(gameCtx *game.Context, target pixel.Target, targetBounds pixel.Rect, timeDelta float64) {
	ctx := &Context{
		Context: gameCtx,
	}

	mouseMapPosition := m.cameraMatrix.Unproject(ctx.CanvasMousePosition).Scaled(1 / resources2.TileSizeF64)
	ctx.MouseMapLocation.X, ctx.MouseMapLocation.Y = int(math.Round(mouseMapPosition.X)), int(math.Round(mouseMapPosition.Y))
	ctx.MouseInCanvas = targetBounds.Contains(ctx.CanvasMousePosition)
	if ctx.MouseInCanvas {
		ctx.DebugTL("location: %d, %d", ctx.MouseMapLocation.X, ctx.MouseMapLocation.Y)
	} else {
		ctx.DebugTL("location: -")
	}

	m.readInputs(ctx, timeDelta)

	m.cameraLag = m.cameraLag.Add(m.cameraLocation.Sub(m.cameraLag).Scaled(math.Min(2*timeDelta, 1.0) * cameraLagSpeed))
	cameraLagRounded := m.cameraLag //pixel.V(math.Round(m.cameraLag.X), math.Round(m.cameraLag.Y))
	m.cameraMatrix = pixel.IM.
		Moved(pixel.V(-cameraLagRounded.X, -cameraLagRounded.Y)).
		Moved(targetBounds.Center())

	hitSelected := false
	for _, layerName := range resources2.MapLayerOrder {
		layer, exists := resources2.Maps[m.selectedMap].Layers[layerName]
		if !exists {
			continue
		}
		var transparency float64
		switch m.layerRenderMode {
		case layerRenderAll:
			transparency = 1
		case layerRenderSelected:
			transparency = 0
		case layerRenderMix:
			transparency = 0.5
		case layerRenderBelow:
			if hitSelected {
				transparency = 0
			} else {
				transparency = 1
			}
		case layerRenderTransparent:
			transparency = 0.5
		}
		if layerName == m.selectedLayer {
			hitSelected = true
			if m.layerRenderMode != layerRenderTransparent {
				transparency = 1
			}
		}
		alpha := uint8(255 * transparency)
		mask := color.RGBA{alpha, alpha, alpha, alpha}
		for _, tile := range layer.Tiles {
			spriteRef := resources2.Sprites[tile.SpriteId]
			spriteRef.Sprite.DrawColorMask(target, m.cameraMatrix.Moved(pixel.V(float64(tile.X*resources2.TileSize), float64(tile.Y*resources2.TileSize))), mask)
		}
	}

	m.swatch.DrawCanvasOverlay(ctx, m.win, target, m.cameraMatrix)
	m.swatch.DrawSwatch(ctx, m.win)

	ctx.DebugBR("map: %s", m.selectedMap)
	ctx.DebugBR("layer: %s", m.selectedLayer)
	ctx.DebugBR("swatch: %s", m.swatch.SelectedSwatch)
}

func (m *MapEditor) mouseTile(ctx *Context, mapName string, layerName resources2.MapLayerName) (*resources2.Tile, int, bool) {
	for index, tile := range resources2.Maps[mapName].Layers[layerName].Tiles {
		if tile.X == ctx.MouseMapLocation.X && tile.Y == ctx.MouseMapLocation.Y {
			return tile, index, true
		}
	}
	return nil, -1, false
}
