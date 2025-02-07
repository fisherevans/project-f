package adventure

import (
	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/resources"
	"github.com/gopxl/pixel/v2"
	"math"
	"sort"
)

const (
	characterSpeed = 5.0
)

var (
	cameraRenderDistanceX = int(math.Ceil(float64(game.GameWidth) / resources.TileSizeF64 / 2.0))
	cameraRenderDistanceY = int(math.Ceil(float64(game.GameHeight) / resources.TileSizeF64 / 2.0))
)

var _ game.State = &State{}

type inputMode int

const (
	inputModePlayerMovement inputMode = iota
	inputModeDialogue
)

type State struct {
	mapWidth, mapHeight int
	baseRenderLayers    []renderLayer
	overlayRenderLayers []renderLayer

	camera    Camera
	inputMode inputMode

	player *Player

	entities             map[EntityId]Entity
	occupiedLocations    map[MapLocation]EntityId
	movementRestrictions map[MapLocation]MovementRestriction

	actions  *ActionQueue
	chatters *ChatterSystem

	batch *pixel.Batch
}

func New(mapName string) game.State {
	m := resources.Maps[mapName]
	a := &State{
		entities:             make(map[EntityId]Entity),
		occupiedLocations:    make(map[MapLocation]EntityId),
		movementRestrictions: make(map[MapLocation]MovementRestriction),
		batch:                pixel.NewBatch(&pixel.TrianglesData{}, resources.SpriteAtlas),
		camera:               NewStaticCamera(pixel.Vec{}),
		actions:              NewActionQueue(),
		chatters:             NewChatterSystem(),
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

	s.actions.ExecuteActions(ctx, s, timeDelta)

	s.camera.Update(ctx, s, timeDelta)
	renderBounds, cameraMatrix := s.camera.ComputeRenderDetails(ctx, s, targetBounds)

	for _, thisRenderLayer := range s.baseRenderLayers {
		thisRenderLayer.Render(s.batch, cameraMatrix, renderBounds)
	}

	for _, entity := range s.locationSortedEntities() {
		renderLocation := entity.RenderMapLocation().Scaled(resources.TileSizeF64)
		entity.Render(s.batch, cameraMatrix.Moved(renderLocation))
	}

	for _, thisRenderLayer := range s.overlayRenderLayers {
		thisRenderLayer.Render(s.batch, cameraMatrix, renderBounds)
	}

	s.batch.Draw(target)
	s.batch.Clear()

	s.chatters.OnTick(ctx, s, target, cameraMatrix, renderBounds, timeDelta)

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
