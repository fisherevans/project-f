package runtime

import (
	"image/color"
	"math"
	"os"
	"runtime"
	"time"

	"github.com/gopxl/pixel/v2"
	"github.com/gopxl/pixel/v2/backends/opengl"

	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/game/shaders"
	"fisherevans.com/project/f/internal/util"
)

func Run() {
	cfg := opengl.WindowConfig{
		Title:     "Project F",
		Bounds:    pixel.R(0, 0, game.GameWidth*3.5, game.GameHeight*5),
		Resizable: true,
		VSync:     true,
	}
	window, err := opengl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}

	game.Initialize(window, "1")

	// Create the fixed-size canvas
	sceneCanvas := shaders.NewCanvas(game.GameWidth, game.GameHeight)
	sceneCanvas.SetSmooth(false)

	lastSceneCanvasScale := 1.0
	canvasScale := 1.0

	createPixelGridCanvas := func() *shaders.Canvas {
		c := shaders.NewCanvas(int(game.GameWidth*canvasScale), int(game.GameHeight*canvasScale))
		c.SetPixelGridOverlayShader(
			float32(canvasScale),
			0.0125,
			0.05, //0.0375,
			0.05, //0.0375,
			0.05,
		)
		return c
	}
	pixelGridCanvas := createPixelGridCanvas()

	last := time.Now()
	frameStats := util.NewFrameStats(600)
	gameLogicStats := util.NewFrameStats(600)

	var m runtime.MemStats

	for !window.Closed() {
		if window.JustPressed(pixel.KeyF4) {
			os.Exit(0)
		}

		// Calculate the time delta
		now := time.Now()
		deltaTime := now.Sub(last).Seconds()
		frameStats.AddFrameTime(deltaTime)
		last = now

		game.ApplyIntent()

		window.Clear(color.RGBA{R: 40, G: 40, B: 40, A: 255})
		sceneCanvas.Clear(game.GetActiveState().ClearColor())

		if s := game.GetCustomShader(); s != nil {
			s.Apply(sceneCanvas, deltaTime)
		} else {
			sceneCanvas.Reset()
		}

		// Calculate scale and offset for whole-number scaling
		windowWidth, windowHeight := window.Bounds().Size().XY()
		scaleX := math.Floor(windowWidth / game.GameWidth)
		scaleY := math.Floor(windowHeight / game.GameHeight)
		canvasScale = math.Min(scaleX, scaleY) // Use the smaller scale
		if canvasScale < 0.5 {
			canvasScale = 0.5
		} else if canvasScale < 1.0 {
			canvasScale = 1.0
		}

		// Calculate mouse position in canvas space, if needed
		//canvasMatrix := pixel.IM.Scaled(pixel.ZV, canvasScale).Moved(window.Bounds().Center())
		//windowMousePosition := window.MousePosition()
		//canvasMousePosition = canvasMatrix.Unproject(windowMousePosition).Add(sceneCanvas.Bounds().Center())

		game.UpdateControls(window)
		game.Update(window, deltaTime)
		game.GetActiveState().OnTick(sceneCanvas, sceneCanvas.Bounds(), deltaTime)

		if game.DebugToggles().F6().JustPressed() || lastSceneCanvasScale != canvasScale {
			lastSceneCanvasScale = canvasScale
			pixelGridCanvas = createPixelGridCanvas()
			if game.DebugToggles().F6().ToggleState() {
				pixelGridCanvas.Reset()
			}
		}
		pixelGridCanvas.Clear(pixel.RGBA{A: 1})

		sceneCanvas.Draw(pixelGridCanvas, pixel.IM.Scaled(pixel.ZV, canvasScale).Moved(pixelGridCanvas.Bounds().Center()))
		pixelGridCanvas.Draw(window, pixel.IM.Moved(window.Bounds().Center()))

		runtime.ReadMemStats(&m)
		game.DebugTL("Memory: %vMB (Heap %vMB), GCs: %d", m.Alloc/1024/1024, m.HeapAlloc/1024/1024, m.NumGC)
		game.DebugTL("%s", frameStats.SummaryFPS())
		game.DebugTL("Game Logic %s", gameLogicStats.SummaryMS())
		game.RenderDebugLines(window, game.PopDebugLines())
		game.RenderNotifications(window, game.PopNotifications(deltaTime))

		gameLogicDur := time.Now().Sub(now).Seconds()
		gameLogicStats.AddFrameTime(gameLogicDur)

		//pixel.NewSprite(resources.SpriteAtlas, resources.SpriteAtlas.Bounds()).Draw(window, pixel.IM.Moved(pixel.V(resources.SpriteAtlasSize/2, 0)))

		if window.JustPressed(pixel.KeyP) {
			CopyCanvasToClipboard(sceneCanvas.Canvas)
		}
		if window.JustPressed(pixel.KeyO) {
			CopyCanvasToClipboard(pixelGridCanvas.Canvas)
		}

		window.Update()
	}
}
