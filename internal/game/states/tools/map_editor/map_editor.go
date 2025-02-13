package map_editor

import (
	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/resources"
	"github.com/gopxl/pixel/v2"
	"github.com/gopxl/pixel/v2/backends/opengl"
	"image/color"
	"math"
)

var keyboardMoveRate = resources.TileSizeF64 * 5
var mouseScrollRate = 3.0
var cameraLagSpeed = 5.0

type editMode int

const (
	editModeLayers editMode = iota
	editModeEntities
)

type MapEditor struct {
	game.BaseState

	cameraLocation pixel.Vec
	cameraMatrix   pixel.Matrix

	win      *opengl.Window
	editMode editMode

	// layers stuff
	swatch *Swatch

	selectedMap     string
	selectedLayer   resources.MapLayerName
	layerRenderMode layerRenderMode

	// entity stuff
	lastDeletedEntity *entityReference
}

type entityReference struct {
	id     string
	entity *resources.Entity
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
		selectedLayer:   resources.LayerBase,
		layerRenderMode: layerRenderMix,
		win:             window,
	}
}

func (m *MapEditor) OnTick(gameCtx *game.Context, target pixel.Target, targetBounds pixel.Rect, timeDelta float64) {
	ctx := &Context{
		Context: gameCtx,
	}

	mouseMapPosition := m.cameraMatrix.Unproject(ctx.CanvasMousePosition).Scaled(1 / resources.TileSizeF64)
	ctx.MouseMapLocation.X, ctx.MouseMapLocation.Y = int(math.Round(mouseMapPosition.X)), int(math.Round(mouseMapPosition.Y))
	ctx.MouseInCanvas = targetBounds.Contains(ctx.CanvasMousePosition)
	if ctx.MouseInCanvas {
		ctx.DebugTL("location: %d, %d", ctx.MouseMapLocation.X, ctx.MouseMapLocation.Y)
	} else {
		ctx.DebugTL("location: -")
	}

	m.readInputs(ctx, timeDelta)

	m.cameraMatrix = pixel.IM.
		Moved(pixel.V(-m.cameraLocation.X, -m.cameraLocation.Y)).
		Moved(targetBounds.Center())

	hitSelected := false
	for _, layerName := range resources.MapLayerOrder {
		layer, exists := m.getSelectedMap().Layers[layerName]
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
			spriteRef := resources.TilesheetSprites[tile.SpriteId]
			spriteRef.Sprite.DrawColorMask(target, m.cameraMatrix.Moved(pixel.V(float64(tile.X*resources.TileSize), float64(tile.Y*resources.TileSize))), mask)
		}
	}

	for _, entity := range m.getSelectedMap().Entities {
		alpha := uint8(128)
		mask := color.RGBA{alpha, alpha, alpha, alpha}
		entitySpriteExists.DrawColorMask(target, m.cameraMatrix.Moved(pixel.V(float64(entity.X*resources.TileSize), float64(entity.Y*resources.TileSize))), mask)
	}

	switch m.editMode {
	case editModeLayers:
		m.swatch.DrawCanvasOverlay(ctx, m.win, target, m.cameraMatrix)
		m.swatch.DrawSwatch(ctx, m.win)
		ctx.DebugBR("map: %s", m.selectedMap)
		ctx.DebugBR("layer: %s", m.selectedLayer)
		ctx.DebugBR("swatch: %s", m.swatch.SelectedSwatch)
	case editModeEntities:
		m.DrawEntityOverlay(ctx, m.win, target, m.cameraMatrix)
		ctx.DebugTR("layer entities")
	}

}

func (m *MapEditor) mouseTile(ctx *Context, layerName resources.MapLayerName) (*resources.Tile, int, bool) {
	for index, tile := range m.getSelectedMap().Layers[layerName].Tiles {
		if tile.X == ctx.MouseMapLocation.X && tile.Y == ctx.MouseMapLocation.Y {
			return tile, index, true
		}
	}
	return nil, -1, false
}

func (m *MapEditor) mouseEntity(ctx *Context) (string, *resources.Entity) {
	for id, entity := range m.getSelectedMap().Entities {
		if entity.X == ctx.MouseMapLocation.X && entity.Y == ctx.MouseMapLocation.Y {
			return id, entity
		}
	}
	return "", nil
}

func (m *MapEditor) getSelectedMap() *resources.Map {
	return resources.Maps[m.selectedMap]
}
