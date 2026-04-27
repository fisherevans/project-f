package runtime

import (
	"fmt"
	"image/color"
	"math"
	"os"
	"runtime"
	"time"

	"fisherevans.com/project/f/assets"
	"fisherevans.com/project/f/internal/game/audio"
	"fisherevans.com/project/f/internal/game/overlay"
	"fisherevans.com/project/f/internal/game/rpg"
	"fisherevans.com/project/f/internal/resources"
	"github.com/gopxl/pixel/v2"
	"github.com/gopxl/pixel/v2/backends/opengl"
	"github.com/gopxl/pixel/v2/ext/text"
	"github.com/rs/zerolog/log"

	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/game/shaders"
	"fisherevans.com/project/f/internal/overlays"
	"fisherevans.com/project/f/internal/util"
)

// DevScene is a named entry point exposed to the overlay's scene-select panel.
// Only populated in development builds; empty in release.
type DevScene = overlay.DevScene

type Instance struct {
	saveId             string
	resetIntentFactory func() any
	devScenes          []DevScene
	window             *opengl.Window
	debugDrain         func()
	DebugHighlight     *DebugHighlight
}

func NewInstance(saveId string, resetIntentFactory func() any) *Instance {
	return &Instance{
		saveId:             saveId,
		resetIntentFactory: resetIntentFactory,
		DebugHighlight:     newDebugHighlight(),
	}
}

// WithDevScenes attaches a list of named dev entry points to the instance.
// The overlay's scene-select panel lists these; clicking one calls Factory().
func (i *Instance) WithDevScenes(scenes []DevScene) *Instance {
	i.devScenes = scenes
	return i
}

func (i *Instance) WithDebugDrain(fn func()) *Instance {
	i.debugDrain = fn
	return i
}

func (i *Instance) initialize() {
	if i.window != nil {
		panic("already running")
	}

	cfg := opengl.WindowConfig{
		Title:     "Primortal - Development Build",
		Bounds:    pixel.R(0, 0, 1280, 800), // steam deck resolution
		Resizable: true,
		VSync:     true,
	}
	var err error
	i.window, err = opengl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}

	title := text.New(pixel.V(20, 20), text.Atlas7x13)
	title.Color = pixel.RGB(0.67, 0.84, 0.13) // #aad620
	status := text.New(pixel.V(20, 20), text.Atlas7x13)
	status.Color = pixel.RGB(0.53, 0.53, 0.53)

	initDone := make(chan error, 1)
	go func() {
		// Load all resources (fonts, sprites, audio, etc.)
		// This triggers all deferred initializers registered via resources.RunOnceInitialized()
		resources.Initialize()

		if err := rpg.LoadFromFS(assets.FS); err != nil {
			log.Fatal().Err(err).Msg("failed to load RPG data")
		}
		if err := overlays.LoadFromFS(assets.FS); err != nil {
			log.Fatal().Err(err).Msg("failed to load overlay flows")
		}

		i.DebugHighlight.init()

		registerIntents()

		resources.SetProgress(resources.Progress{Stage: "Starting game"})
		game.Initialize(i.saveId, i.resetIntentFactory())

		// Wait for audio speaker to start streaming before game loop begins
		// This prevents the first ~1-2 seconds of audio from being clipped
		resources.SetProgress(resources.Progress{Stage: "Finalizing audio"})
		audio.WaitUntilReady()

		initDone <- nil
	}()

	lastDotTick := time.Now()
	dotCount := 0
	lastEmitted := resources.Progress{Current: -1}
	for {
		select {
		case err := <-initDone:
			if err != nil {
				panic(err)
			}
			return
		default:
		}

		if i.window.Closed() {
			os.Exit(0)
		}

		if time.Since(lastDotTick) > 250*time.Millisecond {
			dotCount = (dotCount + 1) % 4
			lastDotTick = time.Now()
		}

		p := resources.CurrentProgress()
		if p != lastEmitted {
			emitProgress(p.Stage, p.Current, p.Total)
			lastEmitted = p
		}

		title.Clear()
		title.WriteString("Primortal")

		status.Clear()
		stage := p.Stage
		if stage == "" {
			stage = "Loading"
		}
		if p.Total > 0 {
			pct := int(float64(p.Current) / float64(p.Total) * 100)
			status.WriteString(fmt.Sprintf("%s %d%% (%d/%d)", stage, pct, p.Current, p.Total))
		} else {
			status.WriteString(stage)
			for j := 0; j < dotCount; j++ {
				status.WriteString(".")
			}
		}

		i.window.Clear(pixel.RGB(0, 0, 0))
		wb := i.window.Bounds()
		titleScale := 3.0
		statusScale := 2.0
		titleW := title.Bounds().W() * titleScale
		statusW := status.Bounds().W() * statusScale
		cx, cy := wb.Center().X, wb.Center().Y
		title.Draw(i.window, pixel.IM.Scaled(pixel.ZV, titleScale).Moved(pixel.V(cx-titleW/2, cy+16)))
		status.Draw(i.window, pixel.IM.Scaled(pixel.ZV, statusScale).Moved(pixel.V(cx-statusW/2, cy-24)))
		i.window.Update()
		time.Sleep(16 * time.Millisecond)
	}

}


func (i *Instance) Run() {
	i.initialize()
	signalReady()

	// Setup overlay
	ov := overlay.New(overlay.Hooks{
		Reset: func() {
			game.SetActiveStateIntent(i.resetIntentFactory())
		},
		SetAudio: func(muted bool, linearVolume float64) {
			s := game.CurrentSave().SystemSettings.Audio
			s.Muted = muted
			s.MasterVolume = linearVolume
			var dB float64
			if muted || linearVolume <= 0 {
				dB = -60
			} else {
				dB = 20 * math.Log10(linearVolume)
			}
			audio.GetSystem().SetMaster(dB)
		},
		SetScale: func(mode string, fixedN int) {
			d := game.CurrentSave().SystemSettings.Display
			d.ScaleMode = mode
			d.FixedScale = fixedN
		},
		ToggleFullscreen: func() {
			d := game.CurrentSave().SystemSettings.Display
			d.Fullscreen = !d.Fullscreen
			overlay.ApplyFullscreen(i.window, d.Fullscreen)
		},
		SaveSettings: func() {
			if err := game.CurrentSave().Save(); err != nil {
				log.Warn().Err(err).Msg("overlay: failed to persist settings")
			}
		},
		DevScenes: i.devScenes,
	})

	// Seed overlay and apply persisted settings
	audioS := game.CurrentSave().SystemSettings.Audio
	ov.InitAudio(audioS.Muted, audioS.MasterVolume)
	if audioS.Muted || audioS.MasterVolume <= 0 {
		audio.GetSystem().SetMaster(-60)
	} else {
		audio.GetSystem().SetMaster(20 * math.Log10(audioS.MasterVolume))
	}

	displayS := game.CurrentSave().SystemSettings.Display
	fixedScale := 0
	if displayS.ScaleMode == rpg.ScaleModeFixed {
		fixedScale = displayS.FixedScale
	}
	ov.InitDisplay(fixedScale, displayS.Fullscreen)
	if displayS.Fullscreen {
		overlay.ApplyFullscreen(i.window, true)
	}

	// Setup rendering canvases
	sceneCanvas := shaders.NewCanvas(game.GameWidth, game.GameHeight)
	sceneCanvas.SetSmooth(false)

	var (
		lastSceneCanvasScale = 1.0
		canvasScale          = 1.0
		pixelGridCanvas      = i.createPixelGridCanvas(canvasScale)
	)

	// Setup recorders (60 FPS for compatibility with QuickTime Player)
	sceneRecorder := NewRecorder("original", sceneCanvas.Canvas, 60)
	pixelGridRecorder := NewRecorder("scaled", pixelGridCanvas.Canvas, 60)

	// Setup frame timing
	last := time.Now()
	frameStats := util.NewFrameStats(600)
	gameLogicStats := util.NewFrameStats(600)
	var m runtime.MemStats

	for !i.window.Closed() {
		// Frame timing
		now := time.Now()
		deltaTime := now.Sub(last).Seconds()
		frameStats.AddFrameTime(deltaTime)
		last = now

		if i.window.JustPressed(pixel.KeyF8) {
			game.SetActiveStateIntent(i.resetIntentFactory())
		}
		if i.window.JustPressed(pixel.KeySlash) {
			log.Info().Msg("----------------------------------------------------------------------------------------------")
		}

		// Update overlay input first — it gets first pick on pointer events
		ov.UpdateInput(i.window, deltaTime)

		// Inject virtual gamepad state before keyboard is merged in UpdateControls
		game.SetVirtualControls(ov.VirtualControls())

		// Scale game time by the speed setting
		gameDelta := deltaTime * game.CurrentSave().SystemSettings.Debugging.GameTimeSpeed

		// Update game state
		game.ApplyIntent()
		if i.debugDrain != nil {
			i.debugDrain()
		}
		game.UpdateControls(i.window)
		game.Update(i.window, gameDelta)

		// Render scene to canvas
		i.renderScene(sceneCanvas, deltaTime)
		game.GetActiveState().OnTick(sceneCanvas, sceneCanvas.Bounds(), gameDelta)

		// Debug highlight overlay (rendered at game resolution, on top of state)
		if i.DebugHighlight.IsActive() {
			i.DebugHighlight.Render(sceneCanvas, gameDelta)
		}

		// Update canvas scale and pixel grid canvas
		canvasScale = overlay.GameCanvasScale(i.window.Bounds())
		if lastSceneCanvasScale != canvasScale {
			lastSceneCanvasScale = canvasScale
			pixelGridCanvas = i.createPixelGridCanvas(canvasScale)
			pixelGridRecorder.UpdateCanvas(pixelGridCanvas.Canvas)
		} else if game.Flags().JustChanged("retro_frame_reset") {
			i.applyPixelGridShader(pixelGridCanvas, canvasScale)
		}

		// Composite and draw to i.window
		i.compositeToWindow(sceneCanvas, pixelGridCanvas, canvasScale, i.window.Bounds())

		// Overlay UI (rendered at window resolution, above game canvas)
		ov.Render(i.window, deltaTime)

		// Debug info
		i.renderDebugInfo(&m, frameStats, gameLogicStats, deltaTime)

		// Handle capture/recording hotkeys
		i.handleCaptureHotkeys(sceneCanvas, sceneRecorder, pixelGridRecorder)

		// Record frames if active
		i.captureRecordingFrames(sceneRecorder, pixelGridRecorder, deltaTime)

		// Track game logic timing
		gameLogicDur := time.Now().Sub(now).Seconds()
		gameLogicStats.AddFrameTime(gameLogicDur)

		i.window.Update()
	}
}

func (i *Instance) createPixelGridCanvas(scale float64) *shaders.Canvas {
	c := shaders.NewCanvas(int(game.GameWidth*scale), int(game.GameHeight*scale))
	i.applyPixelGridShader(c, scale)
	return c
}

func (i *Instance) applyPixelGridShader(c *shaders.Canvas, scale float64) {
	rf := game.CurrentSave().SystemSettings.RetroFrame
	if rf.DisablePixelGrid {
		c.Reset()
		return
	}
	darkenX, darkenY, subpixel := rf.EffectiveGridValues(int(scale))
	c.SetPixelGridOverlayShader(
		float32(scale),
		float32(rf.ScanlineDarken),
		float32(darkenX),
		float32(darkenY),
		float32(subpixel),
	)
}


func (i *Instance) renderScene(sceneCanvas *shaders.Canvas, deltaTime float64) {
	i.window.Clear(color.RGBA{A: 255})
	sceneCanvas.Clear(game.GetActiveState().ClearColor())
}

func (i *Instance) compositeToWindow(sceneCanvas, pixelGridCanvas *shaders.Canvas, scale float64, wb pixel.Rect) {
	pixelGridCanvas.Clear(pixel.RGBA{A: 1})
	sceneCanvas.Draw(pixelGridCanvas, pixel.IM.Scaled(pixel.ZV, scale).Moved(pixelGridCanvas.Bounds().Center()))
	pixelGridCanvas.Draw(i.window, pixel.IM.Moved(overlay.CanvasCenter(wb, scale)))
}

func (i *Instance) renderDebugInfo(m *runtime.MemStats, frameStats, gameLogicStats *util.FloatStats, deltaTime float64) {
	runtime.ReadMemStats(m)
	wb := i.window.Bounds()
	scale := overlay.GameCanvasScale(wb)
	game.DebugTLf("Win: %.0fx%.0f  Scale: x%.0f  DPR: %.0f  UI: %.1fx",
		wb.W(), wb.H(), scale, overlay.PhysicalDPR(), overlay.OverlayUIScale(wb))
	game.DebugTLf("Memory: %vMB (Heap %vMB), GCs: %d", m.Alloc/1024/1024, m.HeapAlloc/1024/1024, m.NumGC)
	game.DebugTLf("%s", frameStats.SummaryFPS())
	game.DebugTLf("Game Logic %s", gameLogicStats.SummaryMS())
	game.DebugTLf("Routine count: %v", runtime.NumGoroutine())
	game.RenderDebugLines(i.window, game.PopDebugLines())
	game.RenderNotifications(i.window, game.PopNotifications(deltaTime))
	game.Console().OnTick(i.window)
}

func (i *Instance) handleCaptureHotkeys(sceneCanvas *shaders.Canvas, sceneRecorder, pixelGridRecorder *Recorder) {
	if i.window.JustPressed(pixel.KeyP) {
		if i.window.Pressed(pixel.KeyLeftControl) {
			sceneRecorder.Toggle()
		} else {
			CopyCanvasToClipboard(sceneCanvas.Canvas)
		}
	}

	if i.window.JustPressed(pixel.KeyO) {
		if i.window.Pressed(pixel.KeyLeftControl) {
			pixelGridRecorder.Toggle()
		} else {
			CopyCanvasToClipboard(sceneCanvas.Canvas)
		}
	}
}

func (i *Instance) captureRecordingFrames(sceneRecorder, pixelGridRecorder *Recorder, deltaTime float64) {
	if sceneRecorder.IsRecording() {
		if err := sceneRecorder.CaptureFrame(deltaTime); err != nil {
			game.DebugNotificationf("Frame capture error: %v", err)
			sceneRecorder.Stop()
		}
	}

	if pixelGridRecorder.IsRecording() {
		if err := pixelGridRecorder.CaptureFrame(deltaTime); err != nil {
			game.DebugNotificationf("Frame capture error: %v", err)
			pixelGridRecorder.Stop()
		}
	}
}
