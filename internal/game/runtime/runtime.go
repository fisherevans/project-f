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

	game.Ctx = game.NewContext(window, "1")

	// Create the fixed-size canvas
	sceneCanvas := shaders.NewCanvas(game.GameWidth, game.GameHeight)
	sceneCanvas.SetSmooth(false)

	lastSceneCanvasScale := game.Ctx.CanvasScale

	createPixelGridCanvas := func() *shaders.Canvas {
		c := shaders.NewCanvas(int(game.GameWidth*game.Ctx.CanvasScale), int(game.GameHeight*game.Ctx.CanvasScale))
		c.SetPixelGridOverlayShader(
			float32(game.Ctx.CanvasScale),
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

		game.Ctx.ApplyIntent()

		window.Clear(color.RGBA{R: 40, G: 40, B: 40, A: 255})
		sceneCanvas.Clear(game.Ctx.GetActiveState().ClearColor())

		if s := game.Ctx.GetCustomShader(); s != nil {
			s.Apply(sceneCanvas, deltaTime)
		} else {
			sceneCanvas.Reset()
		}

		// Calculate scale and offset for whole-number scaling
		windowWidth, windowHeight := window.Bounds().Size().XY()
		scaleX := math.Floor(windowWidth / game.GameWidth)
		scaleY := math.Floor(windowHeight / game.GameHeight)
		game.Ctx.CanvasScale = math.Min(scaleX, scaleY) // Use the smaller scale
		canvasMatrix := pixel.IM.Scaled(pixel.ZV, game.Ctx.CanvasScale).Moved(window.Bounds().Center())
		windowMousePosition := window.MousePosition()
		game.Ctx.CanvasMousePosition = canvasMatrix.Unproject(windowMousePosition).Add(sceneCanvas.Bounds().Center())

		game.Ctx.Update(window, deltaTime)
		game.Ctx.GetActiveState().OnTick(game.Ctx, sceneCanvas, sceneCanvas.Bounds(), deltaTime)

		if game.Ctx.DebugToggles.F6().JustPressed() || lastSceneCanvasScale != game.Ctx.CanvasScale {
			lastSceneCanvasScale = game.Ctx.CanvasScale
			pixelGridCanvas = createPixelGridCanvas()
			if game.Ctx.DebugToggles.F6().ToggleState() {
				pixelGridCanvas.Reset()
			}
		}
		pixelGridCanvas.Clear(pixel.RGBA{A: 1})

		sceneCanvas.Draw(pixelGridCanvas, pixel.IM.Scaled(pixel.ZV, game.Ctx.CanvasScale).Moved(pixelGridCanvas.Bounds().Center()))
		pixelGridCanvas.Draw(window, pixel.IM.Moved(window.Bounds().Center()))

		runtime.ReadMemStats(&m)
		game.Ctx.DebugTL("Memory: %vMB (Heap %vMB), GCs: %d", m.Alloc/1024/1024, m.HeapAlloc/1024/1024, m.NumGC)
		game.Ctx.DebugTL("%s", frameStats.SummaryFPS())
		game.Ctx.DebugTL("Game Logic %s", gameLogicStats.SummaryMS())
		game.RenderDebugLines(window, game.Ctx.PopDebugLines())
		game.RenderNotifications(window, game.Ctx.PopNotifications(deltaTime))

		gameLogicDur := time.Now().Sub(now).Seconds()
		gameLogicStats.AddFrameTime(gameLogicDur)

		//pixel.NewSprite(resources.SpriteAtlas, resources.SpriteAtlas.Bounds()).Draw(window, pixel.IM.Moved(pixel.V(resources.SpriteAtlasSize/2, 0)))

		window.Update()
	}
}
