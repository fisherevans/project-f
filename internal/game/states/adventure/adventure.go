package adventure

import (
	"fisherevans.com/project/f/internal/game"
	resources2 "fisherevans.com/project/f/internal/resources"
	"fisherevans.com/project/f/internal/util"
	"github.com/gopxl/pixel/v2"
	"math"
	"sort"
)

const (
	cameraLagSpeed = 6.0
	characterSpeed = 5.0
)

var (
	cameraRenderDistanceX = int(math.Ceil(float64(game.GameWidth) / resources2.TileSizeF64 / 2.0))
	cameraRenderDistanceY = int(math.Ceil(float64(game.GameHeight) / resources2.TileSizeF64 / 2.0))
)

var _ game.State = &State{}

type State struct {
	mapWidth, mapHeight int
	baseRenderLayers    []renderLayer
	overlayRenderLayers []renderLayer

	targetCameraLocation  pixel.Vec
	currentCameraLocation pixel.Vec

	player *Player

	entities             map[EntityReference]Entity
	occupiedLocations    map[MapLocation]EntityReference
	movementRestrictions map[MapLocation]MovementRestriction

	batch *pixel.Batch
}

func New(mapName string) game.State {
	m := resources2.Maps[mapName]
	a := &State{
		entities:             make(map[EntityReference]Entity),
		occupiedLocations:    make(map[MapLocation]EntityReference),
		movementRestrictions: make(map[MapLocation]MovementRestriction),
		batch:                pixel.NewBatch(&pixel.TrianglesData{}, resources2.SpriteAtlas),
	}
	initializeMap(a, m)
	return a
}

func (s *State) OnTick(ctx *game.Context, target pixel.Target, targetBounds pixel.Rect, timeDelta float64) {
	ctx.DebugTL("delta: %.3f", timeDelta)

	for _, entity := range s.entities {
		remaining := timeDelta
		for remaining > 0 {
			nextRemaining := entity.Move(s, remaining)
			elapsed := remaining - nextRemaining
			entity.Update(ctx, s, elapsed)
			remaining = nextRemaining
		}
	}

	cameraDelta := s.player.RenderMapLocation().Sub(s.currentCameraLocation)
	s.currentCameraLocation = s.currentCameraLocation.Add(cameraDelta.Scaled(math.Min(timeDelta*cameraLagSpeed, 1.0)))
	s.currentCameraLocation = s.player.RenderMapLocation()

	cameraMapX := int(math.Round(s.currentCameraLocation.X))
	cameraMapY := int(math.Round(s.currentCameraLocation.Y))
	fromLocation := MapLocation{
		X: util.MaxInt(0, cameraMapX-cameraRenderDistanceX),
		Y: util.MaxInt(0, cameraMapY-cameraRenderDistanceY),
	}
	toLocation := MapLocation{
		X: util.MinInt(s.mapWidth-1, cameraMapX+cameraRenderDistanceX),
		Y: util.MinInt(s.mapHeight-1, cameraMapY+cameraRenderDistanceY),
	}

	cameraMatrix := pixel.IM.
		Moved(pixel.V(-s.currentCameraLocation.X*resources2.TileSizeF64, -s.currentCameraLocation.Y*resources2.TileSizeF64)).
		Moved(targetBounds.Center())

	for _, thisRenderLayer := range s.baseRenderLayers {
		thisRenderLayer.Render(s.batch, cameraMatrix, fromLocation, toLocation)
	}

	for _, entity := range s.locationSortedEntities() {
		renderLocation := entity.RenderMapLocation().Scaled(resources2.TileSizeF64)
		entity.Sprite().Draw(s.batch, cameraMatrix.Moved(renderLocation))
	}

	for _, thisRenderLayer := range s.overlayRenderLayers {
		thisRenderLayer.Render(s.batch, cameraMatrix, fromLocation, toLocation)
	}

	s.batch.Draw(target)
	s.batch.Clear()

	ctx.DebugTR("location: %d, %d", s.player.CurrentLocation.X, s.player.CurrentLocation.Y)
}

func (s *State) locationSortedEntities() []Entity {
	sortedEntities := make([]Entity, 0, len(s.entities))
	for _, ent := range s.entities {
		sortedEntities = append(sortedEntities, ent)
	}
	sort.Slice(sortedEntities, func(i, j int) bool {
		iL, jL := sortedEntities[i].Location(), sortedEntities[j].Location()
		if iL.Y != jL.Y {
			return iL.Y < jL.Y
		}
		return iL.X < jL.X
	})
	return sortedEntities
}
