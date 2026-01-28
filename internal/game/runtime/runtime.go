package runtime

import (
	"image/color"
	"math"
	"os"
	"runtime"
	"time"

	"fisherevans.com/project/f/internal/game/states/adventure"
	"fisherevans.com/project/f/internal/game/states/combat"
	"fisherevans.com/project/f/internal/game/states/menu"
	"fisherevans.com/project/f/internal/game/states/startup"
	"fisherevans.com/project/f/internal/game/states/state_selector"
	"fisherevans.com/project/f/internal/game/states/title"
	"fisherevans.com/project/f/internal/game/states/xenolog"
	"fisherevans.com/project/f/internal/resources"
	"github.com/gopxl/pixel/v2"
	"github.com/gopxl/pixel/v2/backends/opengl"
	"github.com/gopxl/pixel/v2/ext/text"
	"github.com/rs/zerolog/log"

	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/game/audio"
	"fisherevans.com/project/f/internal/game/shaders"
	"fisherevans.com/project/f/internal/util"
)

type Instance struct {
	window *opengl.Window
}

func NewInstance() *Instance {
	return &Instance{}
}

func (i *Instance) Run() {
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

	loading := text.New(pixel.V(20, 20), text.Atlas7x13)
	loading.Color = pixel.RGB(1, 1, 1)
	loading.WriteString("Loading...")
	loading.Draw(i.window, pixel.IM)
	i.window.Update()

	// Load all resources (fonts, sprites, audio, etc.)
	// This triggers all deferred initializers registered via resources.RunOnceInitialized()
	resources.Initialize()

	game.RegisterStateFactory(adventure.New)
	game.RegisterStateFactory(combat.New)
	game.RegisterStateFactory(menu.New)
	game.RegisterStateFactory(xenolog.New)
	game.RegisterStateFactory(state_selector.New)
	game.RegisterStateFactory(title.New)
	game.RegisterStateFactory(startup.NewDevice)
	game.RegisterStateFactory(startup.NewCopyright)
	game.RegisterStateFactory(startup.NewDeveloper)
	game.RegisterStateFactory(startup.NewControls)
	game.RegisterStateFactory(game.DoSwapStateIntent)

	game.Initialize("default", game.StartupDeviceIntent{})

	// Wait for audio speaker to start streaming before game loop begins
	// This prevents the first ~1-2 seconds of audio from being clipped
	audio.WaitUntilReady()

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

		// Handle exit
		if i.window.JustPressed(pixel.KeyF4) {
			os.Exit(0)
		}
		if i.window.JustPressed(pixel.KeySlash) {
			log.Info().Msg("----------------------------------------------------------------------------------------------")
		}

		// Update game state
		game.ApplyIntent()
		game.UpdateControls(i.window)
		game.Update(i.window, deltaTime)

		// Render scene to canvas
		i.renderScene(sceneCanvas, deltaTime)
		game.GetActiveState().OnTick(sceneCanvas, sceneCanvas.Bounds(), deltaTime)

		// Update canvas scale and pixel grid canvas
		canvasScale = i.calculateCanvasScale()
		if lastSceneCanvasScale != canvasScale || game.Flags().JustChanged("retro_frame_reset") {
			lastSceneCanvasScale = canvasScale
			pixelGridCanvas = i.createPixelGridCanvas(canvasScale)
			pixelGridRecorder.UpdateCanvas(pixelGridCanvas.Canvas)
		}

		// Composite and draw to i.window
		i.compositeToWindow(sceneCanvas, pixelGridCanvas, canvasScale)

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
	if !game.CurrentSave().SystemSettings.RetroFrame.DisablePixelGrid {
		c.SetPixelGridOverlayShader(
			float32(scale),
			float32(game.CurrentSave().SystemSettings.RetroFrame.ScanlineDarken),
			float32(game.CurrentSave().SystemSettings.RetroFrame.GridDarkenX),
			float32(game.CurrentSave().SystemSettings.RetroFrame.GridDarkenY),
			float32(game.CurrentSave().SystemSettings.RetroFrame.SubpixelTint),
		)
	}
	return c
}

func (i *Instance) calculateCanvasScale() float64 {
	windowWidth, windowHeight := i.window.Bounds().Size().XY()
	scaleX := math.Floor(windowWidth / game.GameWidth)
	scaleY := math.Floor(windowHeight / game.GameHeight)
	scale := math.Min(scaleX, scaleY)

	if scale < 0.5 {
		return 0.5
	} else if scale < 1.0 {
		return 1.0
	}
	return scale
}

func (i *Instance) renderScene(sceneCanvas *shaders.Canvas, deltaTime float64) {
	i.window.Clear(color.RGBA{R: 40, G: 40, B: 40, A: 255})
	sceneCanvas.Clear(game.GetActiveState().ClearColor())

	if s := game.GetCustomShader(); s != nil {
		s.Apply(sceneCanvas, deltaTime)
	} else {
		sceneCanvas.Reset()
	}
}

func (i *Instance) compositeToWindow(sceneCanvas, pixelGridCanvas *shaders.Canvas, scale float64) {
	pixelGridCanvas.Clear(pixel.RGBA{A: 1})
	sceneCanvas.Draw(pixelGridCanvas, pixel.IM.Scaled(pixel.ZV, scale).Moved(pixelGridCanvas.Bounds().Center()))
	pixelGridCanvas.Draw(i.window, pixel.IM.Moved(i.window.Bounds().Center()))
}

func (i *Instance) renderDebugInfo(m *runtime.MemStats, frameStats, gameLogicStats *util.FloatStats, deltaTime float64) {
	runtime.ReadMemStats(m)
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
