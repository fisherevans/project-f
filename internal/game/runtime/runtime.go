package runtime

import (
	"image/color"
	"math"
	"os"
	"runtime"
	"time"

	"github.com/gopxl/pixel/v2"
	"github.com/gopxl/pixel/v2/backends/opengl"
	"github.com/rs/zerolog/log"

	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/game/shaders"
	"fisherevans.com/project/f/internal/util"
)

func Run() {
	window := initWindow()
	game.Initialize(window, "1")

	// Setup rendering canvases
	sceneCanvas := shaders.NewCanvas(game.GameWidth, game.GameHeight)
	sceneCanvas.SetSmooth(false)

	var (
		lastSceneCanvasScale = 1.0
		canvasScale          = 1.0
		pixelGridCanvas      = createPixelGridCanvas(canvasScale)
	)

	// Setup recorders (60 FPS for compatibility with QuickTime Player)
	sceneRecorder := NewRecorder("original", sceneCanvas.Canvas, 60)
	pixelGridRecorder := NewRecorder("scaled", pixelGridCanvas.Canvas, 60)

	// Setup frame timing
	last := time.Now()
	frameStats := util.NewFrameStats(600)
	gameLogicStats := util.NewFrameStats(600)
	var m runtime.MemStats

	for !window.Closed() {
		// Frame timing
		now := time.Now()
		deltaTime := now.Sub(last).Seconds()
		frameStats.AddFrameTime(deltaTime)
		last = now

		// Handle exit
		if window.JustPressed(pixel.KeyF4) {
			os.Exit(0)
		}
		if window.JustPressed(pixel.KeySlash) {
			log.Info().Msg("----------------------------------------------------------------------------------------------")
		}

		// Update game state
		game.ApplyIntent()
		game.UpdateControls(window)
		game.Update(window, deltaTime)

		// Render scene to canvas
		renderScene(window, sceneCanvas, deltaTime)
		game.GetActiveState().OnTick(sceneCanvas, sceneCanvas.Bounds(), deltaTime)

		// Update canvas scale and pixel grid canvas
		canvasScale = calculateCanvasScale(window)
		if lastSceneCanvasScale != canvasScale || game.Flags().JustChanged("retro_frame_reset") {
			lastSceneCanvasScale = canvasScale
			pixelGridCanvas = createPixelGridCanvas(canvasScale)
			pixelGridRecorder.UpdateCanvas(pixelGridCanvas.Canvas)
		}

		// Composite and draw to window
		compositeToWindow(window, sceneCanvas, pixelGridCanvas, canvasScale)

		// Debug info
		renderDebugInfo(window, &m, frameStats, gameLogicStats, deltaTime)

		// Handle capture/recording hotkeys
		handleCaptureHotkeys(window, sceneCanvas, sceneRecorder, pixelGridRecorder)

		// Record frames if active
		captureRecordingFrames(sceneRecorder, pixelGridRecorder, deltaTime)

		// Track game logic timing
		gameLogicDur := time.Now().Sub(now).Seconds()
		gameLogicStats.AddFrameTime(gameLogicDur)

		window.Update()
	}
}

func initWindow() *opengl.Window {
	cfg := opengl.WindowConfig{
		Title:     "Project F",
		Bounds:    pixel.R(0, 0, 1280, 800), // steam deck resolution
		Resizable: true,
		VSync:     true,
	}
	window, err := opengl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}
	return window
}

func createPixelGridCanvas(scale float64) *shaders.Canvas {
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

func calculateCanvasScale(window *opengl.Window) float64 {
	windowWidth, windowHeight := window.Bounds().Size().XY()
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

func renderScene(window *opengl.Window, sceneCanvas *shaders.Canvas, deltaTime float64) {
	window.Clear(color.RGBA{R: 40, G: 40, B: 40, A: 255})
	sceneCanvas.Clear(game.GetActiveState().ClearColor())

	if s := game.GetCustomShader(); s != nil {
		s.Apply(sceneCanvas, deltaTime)
	} else {
		sceneCanvas.Reset()
	}
}

func compositeToWindow(window *opengl.Window, sceneCanvas, pixelGridCanvas *shaders.Canvas, scale float64) {
	pixelGridCanvas.Clear(pixel.RGBA{A: 1})
	sceneCanvas.Draw(pixelGridCanvas, pixel.IM.Scaled(pixel.ZV, scale).Moved(pixelGridCanvas.Bounds().Center()))
	pixelGridCanvas.Draw(window, pixel.IM.Moved(window.Bounds().Center()))
}

func renderDebugInfo(window *opengl.Window, m *runtime.MemStats, frameStats, gameLogicStats *util.FloatStats, deltaTime float64) {
	runtime.ReadMemStats(m)
	game.DebugTLf("Memory: %vMB (Heap %vMB), GCs: %d", m.Alloc/1024/1024, m.HeapAlloc/1024/1024, m.NumGC)
	game.DebugTLf("%s", frameStats.SummaryFPS())
	game.DebugTLf("Game Logic %s", gameLogicStats.SummaryMS())
	game.DebugTLf("Routine count: %v", runtime.NumGoroutine())
	game.RenderDebugLines(window, game.PopDebugLines())
	game.RenderNotifications(window, game.PopNotifications(deltaTime))
}

func handleCaptureHotkeys(window *opengl.Window, sceneCanvas *shaders.Canvas, sceneRecorder, pixelGridRecorder *Recorder) {
	if window.JustPressed(pixel.KeyP) {
		if window.Pressed(pixel.KeyLeftControl) {
			sceneRecorder.Toggle()
		} else {
			CopyCanvasToClipboard(sceneCanvas.Canvas)
		}
	}

	if window.JustPressed(pixel.KeyO) {
		if window.Pressed(pixel.KeyLeftControl) {
			pixelGridRecorder.Toggle()
		} else {
			CopyCanvasToClipboard(sceneCanvas.Canvas)
		}
	}
}

func captureRecordingFrames(sceneRecorder, pixelGridRecorder *Recorder, deltaTime float64) {
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
