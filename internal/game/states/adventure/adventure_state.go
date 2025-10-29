package adventure

import (
	"image/color"
	"math"
	"sort"

	"fisherevans.com/project/f/internal/game/events"
	"fisherevans.com/project/f/internal/game/rpg"
	"github.com/gopxl/pixel/v2"
	"github.com/gopxl/pixel/v2/backends/opengl"
	"github.com/gopxl/pixel/v2/ext/imdraw"
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
	atlas                 = resources.DefaultAtlas()
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

	sceneClear pixel.RGBA
	lightClear pixel.RGBA

	mapWidth, mapHeight int
	underRenderLayers   []renderLayer
	overRenderLayers    []renderLayer
	ambientLightAreas   []resources.AmbientLightArea

	camera Camera
	player string

	entities *EntitySystem

	teleports map[TeleportReference]Teleport
	chatters  *ChatterSystem
	dialogues *DialogueSystem
	overlays  *OverlaySystem
	timers    *timers
	zones     *zones

	hud *Hud

	enteringCombat bool

	sceneBatch  *pixel.Batch
	sceneCanvas *opengl.Canvas

	bloom *bloom.Helper

	lightMapBatch  *pixel.Batch
	lightMapCanvas *opengl.Canvas

	litSceneCanvas *opengl.Canvas

	hudBatch *pixel.Batch
	mobs     []*ShadowMob

	eventDispatcher *events.Dispatcher
	worldState      events.WorldState
	planExecutor    *events.PlanExecutor
	run             *rpg.Run
}

func New(i game.AdventureIntent) game.State {
	m := resources.GetMap(i.MapName)
	a := &State{
		teleports: make(map[TeleportReference]Teleport),
		camera:    NewStaticCamera(pixel.Vec{}),
		chatters:  NewChatterSystem(),
		dialogues: NewDialogueSystem(),
		overlays:  NewOverlaySystem(),
		timers:    newTimers(),
		zones:     newZones(),
		run:       &rpg.Run{},

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
	a.planExecutor = events.NewPlanExecutor(a.processEffects)
	a.entities = NewEntitySystem(a)
	a.hud = NewHud(func() int { return a.run.Elythium })
	a.worldState = events.NewWorldState(a.run)
	a.eventDispatcher = events.NewDispatcher(a.worldState, a.processEffects)

	a.bloom.Threshold = 1.0
	a.bloom.HighlightColors = shaders.RGBAtoVec3s(
		colors.HexString("#f9324c"), // red coin
		colors.HexString("#feae34"), // torch yellow
		colors.HexString("#f77622"), // torch orange
		colors.HexString("#0069aa"), // water

		colors.HexString("#ed3579"), // red led
		colors.HexString("#4CC9F0"), // blue led
	)
	a.eventDispatcher.Register(events.NewBasicEntityContext("system"), newSystemEventHandler(a))

	initializeMap(a, m)
	return a
}

func (s *State) ClearColor() color.Color {
	return s.sceneClear
}

func (s *State) OnTick(target *shaders.Canvas, targetBounds pixel.Rect, timeDelta float64) {
	game.DebugTL("delta: %.3f", timeDelta)

	s.entities.Update(timeDelta)

	for _, mob := range s.mobs {
		mob.Update(s, timeDelta)
	}

	s.timers.Update(timeDelta, s.eventDispatcher, s)

	s.planExecutor.Update()

	s.camera.Update(s, timeDelta)
	renderBounds, cameraMatrix := s.camera.ComputeRenderDetails(s, targetBounds)

	// SCENE

	s.sceneBatch.Clear()
	s.sceneCanvas.Clear(s.ClearColor())

	s.lightMapBatch.Clear()

	for _, thisRenderLayer := range s.underRenderLayers {
		thisRenderLayer.Render(s.sceneBatch, cameraMatrix, renderBounds)
	}

	for _, mob := range s.mobs {
		renderLocation := mob.Location.Scaled(resources.MapTileSize.Float())
		mob.Render(s.sceneBatch, cameraMatrix.Moved(renderLocation))
	}

	// todo limit rendering out of bounds entities
	s.entities.Render(s.sceneBatch, s.lightMapBatch, cameraMatrix)

	for _, thisRenderLayer := range s.overRenderLayers {
		thisRenderLayer.Render(s.sceneBatch, cameraMatrix, renderBounds)
	}

	s.sceneBatch.Draw(s.sceneCanvas)

	var locations []string
	s.entities.occupations.ForEachOccupiedLocation(s.player, func(location MapLocation) {
		locations = append(locations, location.String())
	})
	sort.Strings(locations)
	game.DebugBL("player locations: %v", locations)

	// LIGHTING

	// entities draw lights above
	if game.CurrentSave().SystemSettings.Lighting.LightingComposition == rpg.LightingCompositionAmbient {
		// a little gross
		s.lightMapBatch.Clear()
	}
	s.DrawAmbient(s.lightMapCanvas, cameraMatrix, renderBounds, imdraw.New(nil))
	s.lightMapCanvas.SetComposeMethod(pixel.ComposeScreen)
	s.lightMapBatch.Draw(s.lightMapCanvas)

	s.litSceneCanvas.Clear(colornames.Black)
	s.litSceneCanvas.SetComposeMethod(pixel.ComposeOver)
	s.sceneCanvas.Draw(s.litSceneCanvas, pixel.IM.Moved(targetBounds.Center()))
	switch game.CurrentSave().SystemSettings.Lighting.LightingMode {
	case rpg.LightingModeBlended:
		s.litSceneCanvas.SetComposeMethod(pixel.ComposeMultiply)
		s.lightMapCanvas.Draw(s.litSceneCanvas, pixel.IM.Moved(targetBounds.Center()))
	case rpg.LightingModeOver:
		s.litSceneCanvas.SetComposeMethod(pixel.ComposeOver)
		s.lightMapCanvas.Draw(s.litSceneCanvas, pixel.IM.Moved(targetBounds.Center()))
	case rpg.LightingModeOff:
	}
	s.litSceneCanvas.Draw(target, pixel.IM.Moved(targetBounds.Center()))

	// BLOOM

	switch game.CurrentSave().SystemSettings.Lighting.BloomMode {
	case rpg.BloomModeOff:
	case rpg.BloomModeOverThreshold, rpg.BloomModeOverBlurred:
		oldPasses := s.bloom.Passes
		if game.CurrentSave().SystemSettings.Lighting.BloomMode == rpg.BloomModeOverThreshold {
			s.bloom.Passes = 0
		} else {
		}
		bloomed := s.bloom.GenerateBloomCanvas(s.sceneCanvas)
		s.bloom.Passes = oldPasses
		target.Clear(pixel.RGBA{})
		bloomed.Draw(target, pixel.IM.Moved(targetBounds.Center()))
	case rpg.BloomModeBlended:
		bloomed := s.bloom.ApplyBloom(s.sceneCanvas)
		bloomed.Draw(target, pixel.IM.Moved(targetBounds.Center()))
	}

	// HUD + CHAT

	s.hudBatch.Clear()
	s.chatters.OnTick(s, s.hudBatch, cameraMatrix, renderBounds, timeDelta)
	s.hud.OnTick(s, s.hudBatch, cameraMatrix, renderBounds, timeDelta)
	s.overlays.OnTick(s, target, s.hudBatch, timeDelta)
	s.dialogues.OnTick(s, s.hudBatch, renderBounds, timeDelta)
	s.hudBatch.Draw(target)

	if game.Controls[*State]().ButtonSelect().JustPressed() {
		game.SetActiveStateIntent(game.MenuIntent{
			Background: s,
		})
	}
	if game.Controls[*State]().ButtonStart().JustPressed() {
		game.SetActiveStateIntent(game.XenologIntent{
			Background: s,
		})
	}
}

func (s *State) AddMob(mob *ShadowMob) {
	s.mobs = append(s.mobs, mob)
}

func (s *State) setWorldState(key string, value any, id string) {
	oldValue := s.worldState.Get(key)
	s.worldState.Set(key, value)
	s.eventDispatcher.Dispatch(&events.EventWorldStateUpdated{
		Key:      key,
		NewValue: value,
		OldValue: oldValue,
	})
}

func (s *State) ExecuteSystemEffects(effects ...events.Effect) {
	for _, e := range effects {
		s.processEffects(events.DispatchedEffect{
			Source: events.NewBasicEntityContext("system"),
			Effect: e,
		})
	}
}

func (s *State) ExecuteSystemEffectsInOrder(effects ...events.Effect) {
	s.ExecuteSystemEffects(events.NewSerialPlan(effects...))
}
