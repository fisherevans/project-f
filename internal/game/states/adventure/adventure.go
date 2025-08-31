package adventure

import (
	"image/color"
	"math"
	"sort"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/gopxl/pixel/v2"
	"github.com/gopxl/pixel/v2/backends/opengl"
	"golang.org/x/image/colornames"

	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/game/rpg"
	"fisherevans.com/project/f/internal/game/shaders"
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
	actions              *ActionQueue
	chatters             *ChatterSystem
	dialogues            *DialogueSystem
	overlays             *OverlaySystem

	blockInput bool

	sceneBatch  *pixel.Batch
	sceneCanvas *opengl.Canvas

	lightMapBatch  *pixel.Batch
	lightMapCanvas *opengl.Canvas

	litSceneCanvas *opengl.Canvas

	hudBatch *pixel.Batch
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
		overlays:             NewOverlaySystem(),

		sceneBatch:  atlas.NewBatch(),
		sceneCanvas: opengl.NewCanvas(pixel.R(0, 0, game.GameWidth, game.GameHeight)),

		lightMapBatch:  atlas.NewBatch(),
		lightMapCanvas: opengl.NewCanvas(pixel.R(0, 0, game.GameWidth, game.GameHeight)),

		litSceneCanvas: opengl.NewCanvas(pixel.R(0, 0, game.GameWidth, game.GameHeight)),

		hudBatch: atlas.NewBatch(),
	}

	initializeMap(a, m)
	a.animech = save.NewDeployment()
	return a
}

var clearColor = colors.HexColor("#181425")

func (s *State) ClearColor() color.Color {
	return clearColor
}

var (
	f1 = mgl32.Vec2{8, 11}
	f2 = mgl32.Vec2{15, 9}
)

func (s *State) OnTick(ctx *game.Context, target *shaders.Canvas, targetBounds pixel.Rect, timeDelta float64) {
	s.sceneBatch.Clear()
	s.sceneCanvas.Clear(colornames.Black)

	s.lightMapBatch.Clear()
	s.lightMapCanvas.Clear(colors.HSLToRGBA(260, 0.35, 0.25))

	s.litSceneCanvas.Clear(colornames.Black)

	s.hudBatch.Clear()

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
		thisRenderLayer.Render(s.sceneBatch, cameraMatrix, renderBounds)
	}

	for _, entity := range s.locationSortedEntities() {
		renderLocation := entity.RenderMapLocation().Scaled(resources.MapTileSize.Float())
		entity.RenderScene(s.sceneBatch, cameraMatrix.Moved(renderLocation))
	}

	for _, thisRenderLayer := range s.overlayRenderLayers {
		thisRenderLayer.Render(s.sceneBatch, cameraMatrix, renderBounds)
	}

	s.sceneBatch.Draw(s.sceneCanvas)

	for _, entity := range s.locationSortedEntities() {
		renderLocation := entity.RenderMapLocation().Scaled(resources.MapTileSize.Float())
		entity.RenderLight(s.lightMapBatch, cameraMatrix.Moved(renderLocation))
	}
	s.lightMapCanvas.SetComposeMethod(pixel.ComposeScreen)
	s.lightMapBatch.Draw(s.lightMapCanvas)

	s.litSceneCanvas.SetComposeMethod(pixel.ComposeOver)
	s.sceneCanvas.Draw(s.litSceneCanvas, pixel.IM.Moved(targetBounds.Center()))
	s.litSceneCanvas.SetComposeMethod(pixel.ComposeMultiply)
	s.lightMapCanvas.Draw(s.litSceneCanvas, pixel.IM.Moved(targetBounds.Center()))

	s.litSceneCanvas.Draw(target, pixel.IM.Moved(targetBounds.Center()))

	if ctx.DebugToggles.F1().ToggleState() {
		s.lightMapCanvas.Draw(target, pixel.IM.Moved(targetBounds.Center()))
	}

	s.chatters.OnTick(ctx, s, s.hudBatch, cameraMatrix, renderBounds, timeDelta)
	s.overlays.OnTick(ctx, s, target, s.hudBatch, timeDelta)
	s.dialogues.OnTick(ctx, s, s.hudBatch, renderBounds, timeDelta)
	s.hudBatch.Draw(target)

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
		iZ, jZ := sortedEntities[i].GetRenderZPriority(), sortedEntities[j].GetRenderZPriority()
		if iZ != jZ {
			return iZ < jZ
		}
		iL, jL := sortedEntities[i].RenderMapLocation(), sortedEntities[j].RenderMapLocation()
		if iL.Y != jL.Y {
			return iL.Y > jL.Y
		}
		iP, jP := sortedEntities[i].IsPassable(), sortedEntities[j].IsPassable()
		if iP != jP {
			return jP
		}
		if iL.X != jL.X {
			return iL.X < jL.X
		}
		return i < j
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
