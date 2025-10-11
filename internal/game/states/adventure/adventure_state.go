package adventure

import (
	"image/color"
	"math"
	"sort"

	"github.com/gopxl/pixel/v2"
	"github.com/gopxl/pixel/v2/backends/opengl"
	"golang.org/x/image/colornames"

	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/game/shaders"
	"fisherevans.com/project/f/internal/game/shaders/bloom"
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
			resources.FontNameAddStandard,
			resources.FontNameFF,
			resources.FontName3x5,
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

	hud *Hud

	blockInput     bool
	enteringCombat bool

	sceneBatch  *pixel.Batch
	sceneCanvas *opengl.Canvas

	bloom *bloom.Helper

	lightMapBatch  *pixel.Batch
	lightMapCanvas *opengl.Canvas

	litSceneCanvas *opengl.Canvas

	hudBatch *pixel.Batch
	mobs     []*ShadowMob
}

func New(i game.AdventureIntent) game.State {
	m := resources.GetMap(i.MapName)
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

		hud: NewHud(),

		sceneBatch:  atlas.NewBatch(),
		sceneCanvas: opengl.NewCanvas(pixel.R(0, 0, game.GameWidth, game.GameHeight)),

		bloom: bloom.NewHelper(
			game.GameWidth,
			game.GameHeight,
			bloom.DefaultBrightnessConfig(),
			bloom.DefaultBlurConfig(),
			bloom.DefaultBlendConfig()),

		lightMapBatch:  atlas.NewBatch(),
		lightMapCanvas: opengl.NewCanvas(pixel.R(0, 0, game.GameWidth, game.GameHeight)),

		litSceneCanvas: opengl.NewCanvas(pixel.R(0, 0, game.GameWidth, game.GameHeight)),

		hudBatch: atlas.NewBatch(),
	}

	a.bloom.Threshold = 1.0
	a.bloom.HighlightColors = shaders.RGBAtoVec3(
		colors.HexString("#f9324c"), // red coin
		colors.HexString("#feae34"), // torch yellow
		colors.HexString("#f77622"), // torch orange
		colors.HexString("#0069aa"), // water

		colors.HexString("#ed3579"), // red led
		colors.HexString("#4CC9F0"), // blue led
	)

	initializeMap(a, m)
	return a
}

var clearColor = colors.HexString("#1a2d3b")
var lightMapClear = colors.HexString("#7e899e")

func (s *State) ClearColor() color.Color {
	return clearColor
}

func (s *State) OnTick(target *shaders.Canvas, targetBounds pixel.Rect, timeDelta float64) {
	game.DebugTL("delta: %.3f", timeDelta)

	for _, entity := range s.entities {
		remaining := timeDelta
		for remaining > 0 {
			nextRemaining := entity.Move(s, remaining)
			elapsed := remaining - nextRemaining
			entity.Update(s, elapsed)
			remaining = nextRemaining
		}
	}

	for _, mob := range s.mobs {
		mob.Update(s, timeDelta)
	}

	s.actions.ExecuteActions(s, timeDelta)

	s.camera.Update(s, timeDelta)
	renderBounds, cameraMatrix := s.camera.ComputeRenderDetails(s, targetBounds)

	// SCENE

	game.DebugBR("player moving: %.1f", s.player.ConstantMovement)

	s.sceneBatch.Clear()
	s.sceneCanvas.Clear(s.ClearColor())

	for _, thisRenderLayer := range s.baseRenderLayers {
		thisRenderLayer.Render(s.sceneBatch, cameraMatrix, renderBounds)
	}

	for _, mob := range s.mobs {
		renderLocation := mob.Location.Scaled(resources.MapTileSize.Float())
		mob.Render(s.sceneBatch, cameraMatrix.Moved(renderLocation))
	}

	for _, entity := range s.locationSortedEntities() {
		renderLocation := entity.RenderMapLocation().Scaled(resources.MapTileSize.Float())
		entity.RenderScene(s.sceneBatch, cameraMatrix.Moved(renderLocation))
	}

	for _, thisRenderLayer := range s.overlayRenderLayers {
		thisRenderLayer.Render(s.sceneBatch, cameraMatrix, renderBounds)
	}

	s.sceneBatch.Draw(s.sceneCanvas)

	// LIGHTING

	s.lightMapBatch.Clear()
	s.lightMapCanvas.Clear(lightMapClear)
	s.litSceneCanvas.Clear(colornames.Black)
	for _, entity := range s.locationSortedEntities() {
		renderLocation := entity.RenderMapLocation().Scaled(resources.MapTileSize.Float())
		entity.RenderLight(s.lightMapBatch, cameraMatrix.Moved(renderLocation))
	}
	s.lightMapCanvas.SetComposeMethod(pixel.ComposeScreen)
	s.lightMapBatch.Draw(s.lightMapCanvas)

	switch game.DebugToggles().F2().Presses() % 4 {
	case 0, 1:
		s.litSceneCanvas.SetComposeMethod(pixel.ComposeOver)
		s.sceneCanvas.Draw(s.litSceneCanvas, pixel.IM.Moved(targetBounds.Center()))
		s.litSceneCanvas.SetComposeMethod(pixel.ComposeMultiply)
		if game.DebugToggles().F2().Presses()%4 == 1 {
			s.lightMapCanvas.Clear(lightMapClear)
			game.DebugBL("lighting mode: on (ambient only)")
		} else {
			game.DebugBL("lighting mode: on")
		}
		s.lightMapCanvas.Draw(s.litSceneCanvas, pixel.IM.Moved(targetBounds.Center()))
		s.litSceneCanvas.Draw(target, pixel.IM.Moved(targetBounds.Center()))
	case 2:
		s.sceneCanvas.Draw(target, pixel.IM.Moved(targetBounds.Center()))
		game.DebugBL("lighting mode: off")
	case 3:
		s.lightMapCanvas.Draw(target, pixel.IM.Moved(targetBounds.Center()))
		game.DebugBL("lighting mode: debug")
	}

	// BLOOM

	switch game.DebugToggles().F1().Presses() % 4 {
	case 0:
		bloomed := s.bloom.ApplyBloom(s.sceneCanvas)
		bloomed.Draw(target, pixel.IM.Moved(targetBounds.Center()))
		game.DebugBL("bloom mode: on")
	case 1:
		game.DebugBL("bloom mode: off")
	case 2, 3:
		oldPasses := s.bloom.Passes
		if game.DebugToggles().F1().Presses()%4 == 3 {
			s.bloom.Passes = 0
			game.DebugBL("bloom mode: debug - no blur")
		} else {
			game.DebugBL("bloom mode: debug")
		}
		bloomed := s.bloom.GenerateBloomCanvas(s.sceneCanvas)
		s.bloom.Passes = oldPasses
		target.Clear(pixel.RGBA{})
		bloomed.Draw(target, pixel.IM.Moved(targetBounds.Center()))
	}

	if game.Window().Pressed(pixel.KeyI) {
		s.bloom.Intensity -= 0.01
	}
	if game.Window().Pressed(pixel.KeyO) {
		s.bloom.Intensity += 0.01
	}
	if s.bloom.Intensity < 0 {
		s.bloom.Intensity = 0
	}

	if game.Window().Pressed(pixel.KeyK) {
		s.bloom.BloomBias -= 0.01
	}
	if game.Window().Pressed(pixel.KeyL) {
		s.bloom.BloomBias += 0.01
	}
	if s.bloom.BloomBias < 0 {
		s.bloom.BloomBias = 0
	}

	if game.Window().Pressed(pixel.KeyComma) {
		s.bloom.SceneBias -= 0.01
	}
	if game.Window().Pressed(pixel.KeyPeriod) {
		s.bloom.SceneBias += 0.01
	}
	if s.bloom.SceneBias < 0 {
		s.bloom.SceneBias = 0
	}
	game.DebugBL("blend intensity: %.2f", s.bloom.Intensity)
	game.DebugBL("blend bloom bias: %.2f", s.bloom.BloomBias)
	game.DebugBL("blend scene bias: %.2f", s.bloom.SceneBias)

	// HUD + CHAT

	s.hudBatch.Clear()
	s.chatters.OnTick(s, s.hudBatch, cameraMatrix, renderBounds, timeDelta)
	s.hud.OnTick(s, s.hudBatch, cameraMatrix, renderBounds, timeDelta)
	s.overlays.OnTick(s, target, s.hudBatch, timeDelta)
	s.dialogues.OnTick(s, s.hudBatch, renderBounds, timeDelta)
	s.hudBatch.Draw(target)

	game.DebugTR("location: %d, %d", s.player.CurrentLocation.X, s.player.CurrentLocation.Y)

	if game.Controls[*State]().ButtonStart().JustPressed() {
		game.SetActiveStateIntent(game.MenuIntent{
			Background: s,
		})
	}
	if game.Controls[*State]().ButtonSelect().JustPressed() {
		game.SetActiveStateIntent(game.XenologIntent{
			Background: s,
		})
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

func (s *State) AddMob(mob *ShadowMob) {
	s.mobs = append(s.mobs, mob)
}
