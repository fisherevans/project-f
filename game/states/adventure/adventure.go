package adventure

import (
	"fisherevans.com/project/f/game"
	"fisherevans.com/project/f/game/input"
	"fisherevans.com/project/f/resources"
	"fisherevans.com/project/f/util"
	"github.com/gopxl/pixel/v2"
	"github.com/gopxl/pixel/v2/backends/opengl"
	"math"
)

const (
	cameraLagSpeed = 6.0
	characterSpeed = 4.0

	cameraRenderDistanceX = (game.GameWidth / resources.TileSize / 2) + 1
	cameraRenderDistanceY = (game.GameHeight / resources.TileSize / 2)
)

var _ game.State = &AdventureState{}

type AdventureState struct {
	mapWidth, mapHeight int
	baseRenderLayers    []renderLayer
	overlayRenderLayers []renderLayer

	targetCameraLocation  pixel.Vec
	currentCameraLocation pixel.Vec

	controls *input.Controls
	player   *Player

	entities          map[EntityReference]Entity
	occupiedLocations map[MapLocation]EntityReference
}

func New(mapName string) game.State {
	m := resources.Maps[mapName]
	a := &AdventureState{
		entities:          make(map[EntityReference]Entity),
		occupiedLocations: make(map[MapLocation]EntityReference),
	}
	initializeMap(a, m)
	a.controls = input.NewControls()
	return a
}

func (a *AdventureState) OnTick(ctx *game.Context, win *opengl.Window, canvas *opengl.Canvas, timeDelta float64) {
	a.controls.Update(ctx, win)

	ctx.DebugTL("delta: %.3f", timeDelta)

	for _, entity := range a.entities {
		remaining := timeDelta
		for remaining > 0 {
			nextRemaining := entity.Move(a, remaining)
			elapsed := remaining - nextRemaining
			entity.Update(ctx, a, elapsed)
			remaining = nextRemaining
		}
	}

	cameraDelta := a.player.RenderMapLocation().Sub(a.currentCameraLocation)
	a.currentCameraLocation = a.currentCameraLocation.Add(cameraDelta.Scaled(math.Min(timeDelta*cameraLagSpeed, 1.0)))

	cameraMapX := int(math.Round(a.currentCameraLocation.X))
	cameraMapY := int(math.Round(a.currentCameraLocation.Y))
	fromLocation := MapLocation{
		X: util.MaxInt(0, cameraMapX-cameraRenderDistanceX),
		Y: util.MaxInt(0, cameraMapY-cameraRenderDistanceY),
	}
	toLocation := MapLocation{
		X: util.MinInt(a.mapWidth-1, cameraMapX+cameraRenderDistanceX),
		Y: util.MinInt(a.mapHeight-1, cameraMapY+cameraRenderDistanceY),
	}

	cameraMatrix := pixel.IM.
		Moved(pixel.V(-a.currentCameraLocation.X*resources.TileSizeF64, -a.currentCameraLocation.Y*resources.TileSizeF64)).
		Moved(canvas.Bounds().Center())

	for _, thisRenderLayer := range a.baseRenderLayers {
		thisRenderLayer.Render(canvas, cameraMatrix, fromLocation, toLocation)
	}

	for _, entity := range a.entities {
		renderLocation := entity.RenderMapLocation().Scaled(resources.TileSizeF64)
		entity.Sprite().Draw(canvas, cameraMatrix.Moved(renderLocation))
	}

	for _, thisRenderLayer := range a.overlayRenderLayers {
		thisRenderLayer.Render(canvas, cameraMatrix, fromLocation, toLocation)
	}

	ctx.DebugTR("location: %d, %d", a.player.CurrentLocation.X, a.player.CurrentLocation.Y)
}
