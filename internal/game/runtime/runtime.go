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
	"fisherevans.com/project/f/internal/game/states/adventure"
	"fisherevans.com/project/f/internal/game/states/combat"
	"fisherevans.com/project/f/internal/game/states/state_selector"
	"fisherevans.com/project/f/internal/util"
	"fisherevans.com/project/f/internal/util/badges/btest"
	"fisherevans.com/project/f/internal/util/textbox/tbtest"
)

func initialState(window *opengl.Window) game.State {
	return state_selector.New(
		state_selector.Destination{
			Name: "Adventure",
			State: func(ctx *game.Context) game.State {
				return adventure.New("map1", ctx.GameSave)
			},
		},
		state_selector.Destination{
			Name: "Combat",
			State: func(ctx *game.Context) game.State {
				return combat.New(ctx.GameSave.NewDeployment(), func(ctx *game.Context, s *combat.State) {
					ctx.Notify("Combat complete!")
					ctx.SwapActiveState(initialState(window))
				})
			},
		},
		state_selector.Destination{
			Name: "Textbox Test",
			State: func(ctx *game.Context) game.State {
				return tbtest.New()
			},
		},
		state_selector.Destination{
			Name: "Badges Test",
			State: func(ctx *game.Context) game.State {
				return btest.New()
			},
		},
	)
}

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

	ctx := game.NewContext(window, initialState(window), "1")

	// Create the fixed-size canvas
	sceneCanvas := shaders.NewCanvas(game.GameWidth, game.GameHeight)
	sceneCanvas.SetSmooth(false)

	lastSceneCanvasScale := ctx.CanvasScale

	createPixelGridCanvas := func() *shaders.Canvas {
		c := shaders.NewCanvas(int(game.GameWidth*ctx.CanvasScale), int(game.GameHeight*ctx.CanvasScale))
		c.SetPixelGridOverlayShader(
			float32(ctx.CanvasScale),
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

		window.Clear(color.RGBA{R: 40, G: 40, B: 40, A: 255})
		sceneCanvas.Clear(ctx.GetActiveState().ClearColor())

		if s := ctx.GetCustomShader(); s != nil {
			s.Apply(sceneCanvas, deltaTime)
		} else {
			sceneCanvas.Reset()
		}

		// Calculate scale and offset for whole-number scaling
		windowWidth, windowHeight := window.Bounds().Size().XY()
		scaleX := math.Floor(windowWidth / game.GameWidth)
		scaleY := math.Floor(windowHeight / game.GameHeight)
		ctx.CanvasScale = math.Min(scaleX, scaleY) // Use the smaller scale
		canvasMatrix := pixel.IM.Scaled(pixel.ZV, ctx.CanvasScale).Moved(window.Bounds().Center())
		windowMousePosition := window.MousePosition()
		ctx.CanvasMousePosition = canvasMatrix.Unproject(windowMousePosition).Add(sceneCanvas.Bounds().Center())

		ctx.Update(window)
		ctx.GetActiveState().OnTick(ctx, sceneCanvas, sceneCanvas.Bounds(), deltaTime)

		if ctx.DebugToggles.F6().JustPressed() || lastSceneCanvasScale != ctx.CanvasScale {
			lastSceneCanvasScale = ctx.CanvasScale
			pixelGridCanvas = createPixelGridCanvas()
			if ctx.DebugToggles.F6().ToggleState() {
				pixelGridCanvas.Reset()
			}
		}
		pixelGridCanvas.Clear(pixel.RGBA{A: 1})

		sceneCanvas.Draw(pixelGridCanvas, pixel.IM.Scaled(pixel.ZV, ctx.CanvasScale).Moved(pixelGridCanvas.Bounds().Center()))
		pixelGridCanvas.Draw(window, pixel.IM.Moved(window.Bounds().Center()))

		runtime.ReadMemStats(&m)
		ctx.DebugTL("Memory: %vMB (Heap %vMB), GCs: %d", m.Alloc/1024/1024, m.HeapAlloc/1024/1024, m.NumGC)
		ctx.DebugTL("%s", frameStats.SummaryFPS())
		ctx.DebugTL("Game Logic %s", gameLogicStats.SummaryMS())
		game.RenderDebugLines(window, ctx.PopDebugLines())
		game.RenderNotifications(window, ctx.PopNotifications(deltaTime))

		gameLogicDur := time.Now().Sub(now).Seconds()
		gameLogicStats.AddFrameTime(gameLogicDur)

		//pixel.NewSprite(resources.SpriteAtlas, resources.SpriteAtlas.Bounds()).Draw(window, pixel.IM.Moved(pixel.V(resources.SpriteAtlasSize/2, 0)))

		window.Update()
	}
}
