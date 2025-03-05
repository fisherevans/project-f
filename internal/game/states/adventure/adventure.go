package adventure

import (
	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/game/rpg"
	"fisherevans.com/project/f/internal/resources"
	"fisherevans.com/project/f/internal/util/colors"
	"github.com/gopxl/pixel/v2"
	"image/color"
	"math"
	"sort"
)

const (
	characterSpeed = 3.0
)

var (
	cameraRenderDistanceX = int(math.Ceil(float64(game.GameWidth) / resources.MapTileSize.Float() / 2.0))
	cameraRenderDistanceY = int(math.Ceil(float64(game.GameHeight) / resources.MapTileSize.Float() / 2.0))
	atlas                 = resources.CreateAtlas(resources.AtlasFilter{
		FontNames: []string{
			resources.FontNameM5x7,
			resources.FontNameM3x6,
		},
	})
)

func init() {
	atlas.Dump("temp", "adventure")
}

var _ game.State = &State{}

type inputMode int

const (
	inputModePlayerMovement inputMode = iota
	inputModeDialogue
)

type State struct {
	game.BaseState

	animech *rpg.DeployedAnimech

	mapWidth, mapHeight int
	baseRenderLayers    []renderLayer
	overlayRenderLayers []renderLayer

	camera Camera
	player *Player

	entities             map[EntityId]Entity
	occupiedLocations    map[MapLocation]EntityId
	movementRestrictions map[MapLocation]MovementRestriction

	actions   *ActionQueue
	chatters  *ChatterSystem
	dialogues *DialogueSystem

	batch *pixel.Batch
}

func New(mapName string, save *rpg.GameSave) game.State {
	m := resources.GetMap(mapName)
	a := &State{
		entities:             make(map[EntityId]Entity),
		occupiedLocations:    make(map[MapLocation]EntityId),
		movementRestrictions: make(map[MapLocation]MovementRestriction),
		camera:               NewStaticCamera(pixel.Vec{}),
		actions:              NewActionQueue(),
		chatters:             NewChatterSystem(),
		dialogues:            NewDialogueSystem(),

		batch: atlas.NewBatch(),
	}
	initializeMap(a, m)
	a.animech = save.NewDeployment()
	return a
}

var clearColor = colors.HexColor("#181425")

func (s *State) ClearColor() color.Color {
	return clearColor
}

func (s *State) OnTick(ctx *game.Context, target pixel.Target, targetBounds pixel.Rect, timeDelta float64) {
	s.batch.Clear()

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
		renderLocation := entity.RenderMapLocation().Scaled(resources.MapTileSize.Float())
		entity.Render(s.batch, cameraMatrix.Moved(renderLocation))
	}

	for _, thisRenderLayer := range s.overlayRenderLayers {
		thisRenderLayer.Render(s.batch, cameraMatrix, renderBounds)
	}

	s.chatters.OnTick(ctx, s, s.batch, cameraMatrix, renderBounds, timeDelta)
	s.dialogues.OnTick(ctx, s, s.batch, renderBounds, timeDelta)

	ctx.DebugTR("location: %d, %d", s.player.CurrentLocation.X, s.player.CurrentLocation.Y)

	s.batch.Draw(target)
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

func (s *State) inputMode() inputMode {
	if s.dialogues.HasPriority() {
		return inputModeDialogue
	}
	return inputModePlayerMovement
}
