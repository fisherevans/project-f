package adventure

import (
	"image/color"
	"math"
	"sort"

	"fisherevans.com/project/f/internal/game/events"
	"fisherevans.com/project/f/internal/game/input"
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
	baseRenderLayers    []renderLayer
	overlayRenderLayers []renderLayer
	ambientLightAreas   []resources.AmbientLightArea

	camera Camera
	player *Player

	entities   map[EntityId]Entity
	tileStates map[MapLocation]*TileState
	teleports  map[TeleportReference]Teleport
	chatters   *ChatterSystem
	dialogues  *DialogueSystem
	overlays   *OverlaySystem
	timers     *timers
	zones      *zones

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

	eventDispatcher    *events.Dispatcher
	systemEffects      []events.DispatchedEffect
	worldState         events.WorldState
	planExecutor       *events.PlanExecutor
	run                *rpg.Run
	movementController *MovementController
	behaviors          map[EntityId]EntityBehavior
}

func New(i game.AdventureIntent) game.State {
	m := resources.GetMap(i.MapName)
	a := &State{
		entities:   make(map[EntityId]Entity),
		tileStates: make(map[MapLocation]*TileState),
		teleports:  make(map[TeleportReference]Teleport),
		camera:     NewStaticCamera(pixel.Vec{}),
		chatters:   NewChatterSystem(),
		dialogues:  NewDialogueSystem(),
		overlays:   NewOverlaySystem(),
		timers:     newTimers(),
		zones:      newZones(),
		run:        &rpg.Run{},

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

		eventDispatcher:    events.NewDispatcher(),
		planExecutor:       events.NewPlanExecutor(),
		movementController: NewMovementController(),
		behaviors:          make(map[EntityId]EntityBehavior),
	}
	a.hud = NewHud(func() int { return a.run.Elythium })
	a.worldState = events.NewWorldState(a.run)

	a.bloom.Threshold = 1.0
	a.bloom.HighlightColors = shaders.RGBAtoVec3s(
		colors.HexString("#f9324c"), // red coin
		colors.HexString("#feae34"), // torch yellow
		colors.HexString("#f77622"), // torch orange
		colors.HexString("#0069aa"), // water

		colors.HexString("#ed3579"), // red led
		colors.HexString("#4CC9F0"), // blue led
	)
	a.eventDispatcher.Register(events.NewEphemeralEntityContext("system"), newSystemEventHandler(a))

	initializeMap(a, m)
	return a
}

func (s *State) ClearColor() color.Color {
	return s.sceneClear
}

func (s *State) OnTick(target *shaders.Canvas, targetBounds pixel.Rect, timeDelta float64) {
	game.DebugTL("delta: %.3f", timeDelta)

	// 1. Update behaviors (input handling, AI decisions)
	for _, behavior := range s.behaviors {
		behavior.Update(s, timeDelta)
	}

	// 2. Update movement controller (handles all tile transitions with time delta chaining)
	s.movementController.Update(s, timeDelta)

	// 3. Update animations and lights for moveable entities
	s.updateAnimationsAndLights(timeDelta)

	// 4. Update other systems
	for _, mob := range s.mobs {
		mob.Update(s, timeDelta)
	}

	s.timers.Update(timeDelta, s.eventDispatcher, s)

	// 4. Process effects
	s.processEffects(s.PopSystemEffects())
	s.processEffects(s.eventDispatcher.Flush(s.worldState))
	s.processEffects(s.planExecutor.GetNextEffects())

	// 5. Update camera
	s.camera.Update(s, timeDelta)
	renderBounds, cameraMatrix := s.camera.ComputeRenderDetails(s, targetBounds)

	// SCENE

	playerMovement := s.movementController.Get(s.player.GetEntityId())
	if playerMovement != nil {
		game.DebugBR("player moving: %.1f", playerMovement.ConstantMovement())
	}

	s.sceneBatch.Clear()
	s.sceneCanvas.Clear(s.ClearColor())

	// todo limit rendering out of bounds tiles

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

	// build light map
	s.lightMapBatch.Clear()
	switch game.CurrentSave().SystemSettings.Lighting.LightingComposition {
	case rpg.LightingCompositionFull:
		for _, entity := range s.locationSortedEntities() {
			renderLocation := entity.RenderMapLocation().Scaled(resources.MapTileSize.Float())
			entity.RenderLight(s.lightMapBatch, cameraMatrix.Moved(renderLocation))
		}
	case rpg.LightingCompositionAmbient:
	}
	s.lightMapCanvas.Clear(colors.Black.RGBA)
	s.DrawAmbient(s.lightMapCanvas, cameraMatrix, renderBounds, imdraw.New(nil))
	s.lightMapCanvas.SetComposeMethod(pixel.ComposeScreen)
	s.lightMapBatch.Draw(s.lightMapCanvas)

	// render "lit" scene
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

	playerLoc := s.player.Location()
	game.DebugTR("location: %d, %d", playerLoc.X, playerLoc.Y)

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
		if iL.X != jL.X {
			return iL.X < jL.X
		}
		return i < j
	})
	return sortedEntities
}

func (s *State) updateAnimationsAndLights(timeDelta float64) {
	// Update player animations
	if s.player != nil {
		s.updatePlayerAnimations(timeDelta)
	}

	// Update NPC animations
	for _, entity := range s.entities {
		if npc, ok := entity.(*NPC); ok {
			s.updateNPCAnimations(npc, timeDelta)
		}
	}

	// Update other entities that have Update() methods
	for _, entity := range s.entities {
		// DynamicEntity, LightEntity, etc. still have Update() methods
		switch e := entity.(type) {
		case *DynamicEntity:
			e.Update(s, timeDelta)
		case *LightEntity:
			e.Update(s, timeDelta)
		}
	}
}

func (s *State) updatePlayerAnimations(timeDelta float64) {
	movement := s.movementController.Get(s.player.GetEntityId())
	if movement == nil {
		return
	}

	animations, ok := s.player.Animations[movement.MoveState()]
	if !ok {
		animations, ok = s.player.Animations[MoveStateIdle]
	}
	if !ok {
		return
	}

	animation, ok := animations[movement.FacingDirection()]
	if !ok {
		animation = animations[input.Down]
	}
	if animation == nil {
		return
	}

	// Reset animation only if direction changed (not on state changes during continuous movement)
	if s.player.lastAnimationDirection != movement.FacingDirection() {
		animation.Reset()
		s.player.lastAnimationDirection = movement.FacingDirection()
	}
	s.player.lastAnimationState = movement.MoveState()

	// Update animation with speed multiplier when moving
	if movement.IsMoving() {
		animation.Update(timeDelta * movement.GetCurrentSpeed())
	} else {
		animation.Update(timeDelta)
	}

	// Update player light
	if s.player.Lights != nil {
		light, ok := s.player.Lights[movement.MoveState()]
		if !ok {
			light, ok = s.player.Lights[MoveStateIdle]
		}
		if light != nil {
			light.Update(timeDelta)
		}
	}
}

func (s *State) updateNPCAnimations(npc *NPC, timeDelta float64) {
	if npc.movementState == nil {
		return
	}

	animations, ok := npc.Animations[npc.movementState.MoveState()]
	if !ok {
		animations, ok = npc.Animations[MoveStateIdle]
	}
	if !ok {
		return
	}

	animation, ok := animations[npc.movementState.FacingDirection()]
	if !ok {
		animation = animations[input.Down]
	}
	if animation == nil {
		return
	}

	// Reset animation only if direction changed (not on state changes during continuous movement)
	if npc.lastAnimationDirection != npc.movementState.FacingDirection() {
		animation.Reset()
		npc.lastAnimationDirection = npc.movementState.FacingDirection()
	}
	npc.lastAnimationState = npc.movementState.MoveState()

	// Update animation with speed multiplier when moving
	if npc.movementState.IsMoving() {
		animation.Update(timeDelta * npc.movementState.GetCurrentSpeed())
	} else {
		animation.Update(timeDelta)
	}
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

func (s *State) setWorldState(key string, value any, id string) {
	oldValue := s.worldState.Get(key)
	s.worldState.Set(key, value)
	s.eventDispatcher.Dispatch(&events.EventWorldStateUpdated{
		Key:      key,
		NewValue: value,
		OldValue: oldValue,
	})
}

func (s *State) AddSystemEffect(e events.Effect) {
	s.systemEffects = append(s.systemEffects, events.DispatchedEffect{
		Source: events.NewEphemeralEntityContext("system"),
		Effect: e,
	})
}

func (s *State) AddSerialSystemEffects(effects ...events.Effect) {
	s.systemEffects = append(s.systemEffects, events.DispatchedEffect{
		Source: events.NewEphemeralEntityContext("system"),
		Effect: events.Effect{
			Plan: events.NewSerialPlan(effects...),
		},
	})
}

func (s *State) PopSystemEffects() []events.DispatchedEffect {
	out := s.systemEffects
	s.systemEffects = nil
	return out
}
