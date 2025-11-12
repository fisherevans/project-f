package adventure

import (
	"image/color"
	"math"
	"sort"

	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/game/rpg"
	"fisherevans.com/project/f/internal/game/shaders"
	"fisherevans.com/project/f/internal/game/shaders/bloom"
	"fisherevans.com/project/f/internal/resources"
	"fisherevans.com/project/f/internal/util/colors"
	"github.com/gopxl/pixel/v2"
	"github.com/gopxl/pixel/v2/backends/opengl"
	"github.com/gopxl/pixel/v2/ext/imdraw"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"golang.org/x/image/colornames"
)

const (
	characterSpeed             = 3.0
	globalVariableNamePlayerId = "player_id"
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

	teleports  map[TeleportReference]Teleport
	chatters   *ChatterSystem
	dialogues  *DialogueSystem
	overlays   *OverlaySystem
	timers     *timers
	conditions *Conditions
	zones      *zones

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

	eventDispatcher *Dispatcher
	planExecutor    *PlanExecutor

	globals *observedGlobals
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
		globals:   observeGlobals(game.CurrentSave().Globals),

		sceneBatch:  atlas.NewBatch(),
		sceneCanvas: opengl.NewCanvas(pixel.R(0, 0, game.GameWidth, game.GameHeight)),

		bloom: bloom.NewHelper(
			game.GameWidth,
			game.GameHeight,
			bloom.DefaultBrightnessConfig().
				WithThreshold(1.0).
				WithHighlightColors(shaders.RGBAtoVec3s(
					colors.HexString("#f9324c"), // red coin
					colors.HexString("#feae34"), // torch yellow
					colors.HexString("#f77622"), // torch orange
					colors.HexString("#0069aa"), // water

					colors.HexString("#ed3579"), // red led
					colors.HexString("#4CC9F0"), // blue led
				)),
			bloom.DefaultBlurConfig(),
			bloom.DefaultBlendConfig()),

		lightMapBatch:  atlas.NewBatch(),
		lightMapCanvas: opengl.NewCanvas(pixel.R(0, 0, game.GameWidth, game.GameHeight)),
		litSceneCanvas: opengl.NewCanvas(pixel.R(0, 0, game.GameWidth, game.GameHeight)),

		hudBatch: atlas.NewBatch(),
	}
	a.conditions = NewConditions(a)
	a.planExecutor = NewPlanExecutor(a.processEffects)
	a.entities = NewEntitySystem(a)
	a.hud = NewHud(a)
	a.eventDispatcher = NewDispatcher(a.globals, a.processEffects)
	a.globals.withDispatcher(a.eventDispatcher)
	a.eventDispatcher.Register(a.entities.RegisterEntity("system", MapLocation{}), newSystemEventHandler(a))

	initializeMap(a, m)

	return a
}

func (s *State) Globals() rpg.GlobalsReader {
	return s.globals
}

func (s *State) ClearColor() color.Color {
	return s.sceneClear
}

func (s *State) OnEnter() {
	s.entities.resumeAllSounds()
}

func (s *State) OnExit() {
	s.entities.pauseAllSounds()
}

func (s *State) OnTick(target *shaders.Canvas, targetBounds pixel.Rect, timeDelta float64) {
	game.DebugTLf("delta: %.3f", timeDelta)

	s.entities.Update(timeDelta)
	s.conditions.Update(timeDelta)

	for _, mob := range s.mobs {
		mob.Update(s, timeDelta)
	}

	s.timers.Update(timeDelta, s.eventDispatcher, s)

	s.planExecutor.Update()

	s.camera.Update(s, timeDelta)
	renderBounds, cameraDelta := s.camera.ComputeRenderDetails(s, targetBounds)

	// SCENE

	s.sceneBatch.Clear()
	s.sceneCanvas.Clear(s.ClearColor())

	s.lightMapBatch.Clear()

	for _, thisRenderLayer := range s.underRenderLayers {
		thisRenderLayer.Render(s.sceneBatch, cameraDelta, renderBounds)
	}

	for _, mob := range s.mobs {
		renderLocation := mob.Location.Scaled(resources.MapTileSize.Float())
		mob.Render(s.sceneBatch, cameraDelta.Add(renderLocation))
	}

	// todo limit rendering out of bounds entities
	s.entities.Render(s.sceneBatch, s.lightMapBatch, cameraDelta)

	for _, thisRenderLayer := range s.overRenderLayers {
		thisRenderLayer.Render(s.sceneBatch, cameraDelta, renderBounds)
	}

	s.sceneBatch.Draw(s.sceneCanvas)

	playerEntity, _ := s.entities.GetEntity(s.player)
	var locations []string
	s.entities.occupations.ForEachOccupiedLocation(s.player, func(location MapLocation) {
		locations = append(locations, location.String())
	})
	sort.Strings(locations)
	game.DebugBLf("player locations: %v", locations)
	game.DebugBLf("player behavior enabled: %t (%v)", playerEntity.IsBehaviorEnabled(), s.entities.disabledBehaviors[s.player])

	if game.Window().JustPressed(pixel.KeyB) {
		s.planExecutor.mu.Lock()
		for _, batch := range s.planExecutor.activeBatches {
			log.Info().Int("next effect", batch.nextEffectIndex).
				Any("waiting", batch.waitingFor).
				Any("batch", batch.batch).
				Msg("batch in progress")
		}
		s.planExecutor.mu.Unlock()
	}

	// LIGHTING

	// entities draw lights above
	if game.CurrentSave().SystemSettings.Lighting.LightingComposition == rpg.LightingCompositionAmbient {
		// a little gross
		s.lightMapBatch.Clear()
	}
	s.DrawAmbient(s.lightMapCanvas, cameraDelta, renderBounds, imdraw.New(nil))
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
	s.chatters.OnTick(s, s.hudBatch, cameraDelta, renderBounds, timeDelta)
	s.hud.OnTick(s, s.hudBatch, cameraDelta, renderBounds, timeDelta)
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

func (s *State) ExecuteSystemEffects(effects ...Effect) {
	for _, e := range effects {
		if err := e.FillDefaultsAndValidate(); err != nil {
			log.Error().Err(err).Msg("failed to validate effect")
			continue
		}
		systemEntity, _ := s.entities.GetEntity("system")
		s.processEffects(DispatchedEffect{
			Source: systemEntity,
			Effect: e,
		})
	}
}

func (s *State) processEffects(effects ...DispatchedEffect) {
	for _, dispatched := range effects {
		wasSuccessful := dispatched.Effect.Process(dispatched.Source, s)
		level := zerolog.InfoLevel
		switch dispatched.Effect.(type) {
		case *EffectTriggerMovement, *EffectEntityFaceDirection:
			level = zerolog.DebugLevel
		}
		logEffect(level, dispatched.Source, dispatched.Effect, "event processed %t: %T", wasSuccessful, dispatched.Effect)
	}
}

func (s *State) ExecuteSystemEffectsInOrder(effects ...Effect) {
	s.ExecuteSystemEffects(NewSerialPlan(effects...))
}
