package adventure

import (
	"image/color"
	"math"
	"sort"

	"github.com/gopxl/pixel/v2"
	"github.com/gopxl/pixel/v2/backends/opengl"

	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/game/rpg"
	"fisherevans.com/project/f/internal/game/states/menu"
	"fisherevans.com/project/f/internal/resources"
	"fisherevans.com/project/f/internal/util/colors"
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
	inputModeBlocked inputMode = iota
	inputModePlayerMovement
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
	teleports            map[TeleportReference]Teleport
	lights               *LightSystem
	actions              *ActionQueue
	chatters             *ChatterSystem
	dialogues            *DialogueSystem
	overlays             *OverlaySystem

	blockInput bool

	underBatch *pixel.Batch
	overBatch  *pixel.Batch

	lightBatch  *pixel.Batch
	lightCanvas *opengl.Canvas
}

func New(mapName string, save *rpg.GameSave) game.State {
	m := resources.GetMap(mapName)
	a := &State{
		entities:             make(map[EntityId]Entity),
		occupiedLocations:    make(map[MapLocation]EntityId),
		movementRestrictions: make(map[MapLocation]MovementRestriction),
		teleports:            make(map[TeleportReference]Teleport),
		camera:               NewStaticCamera(pixel.Vec{}),
		actions:              NewActionQueue(),
		chatters:             NewChatterSystem(),
		dialogues:            NewDialogueSystem(),
		lights:               NewLightSystem(),
		overlays:             NewOverlaySystem(),

		underBatch: atlas.NewBatch(),
		lightBatch: atlas.NewBatch(),
		overBatch:  atlas.NewBatch(),
	}

	a.lightCanvas = opengl.NewCanvas(pixel.R(0, 0, game.GameWidth, game.GameHeight))
	//a.lightCanvas.SetFragmentShader(lightShader)

	initializeMap(a, m)
	a.animech = save.NewDeployment()
	return a
}

var clearColor = colors.HexColor("#181425")

func (s *State) ClearColor() color.Color {
	return clearColor
}

func (s *State) OnTick(ctx *game.Context, target *opengl.Canvas, targetBounds pixel.Rect, timeDelta float64) {
	s.underBatch.Clear()
	s.lightBatch.Clear()
	s.overBatch.Clear()

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
		thisRenderLayer.Render(s.underBatch, cameraMatrix, renderBounds)
	}

	for _, entity := range s.locationSortedEntities() {
		renderLocation := entity.RenderMapLocation().Scaled(resources.MapTileSize.Float())
		entity.Render(s.underBatch, cameraMatrix.Moved(renderLocation))
	}

	for _, thisRenderLayer := range s.overlayRenderLayers {
		thisRenderLayer.Render(s.underBatch, cameraMatrix, renderBounds)
	}

	s.underBatch.Draw(target)

	s.lightCanvas.Clear(colors.Blurple3.RGBA)
	s.lights.Render(s.lightBatch, cameraMatrix)
	s.lightBatch.Draw(s.lightCanvas)
	if !ctx.DebugToggles.F1().Pressed() {
		target.SetComposeMethod(pixel.ComposeMultiply)
	}
	s.lightCanvas.Draw(target, pixel.IM.Moved(targetBounds.Center()))
	target.SetComposeMethod(pixel.ComposeOver)

	s.chatters.OnTick(ctx, s, s.overBatch, cameraMatrix, renderBounds, timeDelta)
	s.overlays.OnTick(ctx, s, s.overBatch, timeDelta)
	s.dialogues.OnTick(ctx, s, s.overBatch, renderBounds, timeDelta)

	s.overBatch.Draw(target)

	ctx.DebugTR("location: %d, %d", s.player.CurrentLocation.X, s.player.CurrentLocation.Y)

	if ctx.Controls.ButtonStart().JustPressed() {
		ctx.SwapActiveState(menu.New(s))
	}
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
	if s.blockInput {
		return inputModeBlocked
	}
	if s.dialogues.HasPriority() {
		return inputModeDialogue
	}
	return inputModePlayerMovement
}
