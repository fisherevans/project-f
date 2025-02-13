package runtime

import (
	game "fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/game/states/adventure"
	"fisherevans.com/project/f/internal/game/states/state_selector"
	"fisherevans.com/project/f/internal/game/states/tools/map_editor"
	"fisherevans.com/project/f/internal/util"
	"github.com/gopxl/pixel/v2"
	"github.com/gopxl/pixel/v2/backends/opengl"
	"image/color"
	"math"
	"os"
	"runtime"
	"time"
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

	ctx := game.NewContext(state_selector.New(
		state_selector.Destination{
			Name: "Adventure",
			State: func() game.State {
				return adventure.New("dummy")
			},
		},
		state_selector.Destination{
			Name: "Map Editor",
			State: func() game.State {
				return map_editor.New(window)
			},
		},
	), "1")

	// Create the fixed-size canvas
	canvas := opengl.NewCanvas(pixel.R(0, 0, game.GameWidth, game.GameHeight))
	canvas.SetSmooth(false)

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
		canvas.Clear(ctx.GetActiveState().ClearColor())

		// Calculate scale and offset for whole-number scaling
		windowWidth, windowHeight := window.Bounds().Size().XY()
		scaleX := math.Floor(windowWidth / game.GameWidth)
		scaleY := math.Floor(windowHeight / game.GameHeight)
		ctx.CanvasScale = math.Min(scaleX, scaleY) // Use the smaller scale
		canvasMatrix := pixel.IM.Scaled(pixel.ZV, ctx.CanvasScale).Moved(window.Bounds().Center())
		windowMousePosition := window.MousePosition()
		ctx.CanvasMousePosition = canvasMatrix.Unproject(windowMousePosition).Add(canvas.Bounds().Center())

		ctx.Update(window)
		ctx.GetActiveState().OnTick(ctx, canvas, canvas.Bounds(), deltaTime)

		canvas.Draw(window, canvasMatrix)

		runtime.ReadMemStats(&m)
		ctx.DebugTL("Memory: %vMB (Heap %vMB), GCs: %d", m.Alloc/1024/1024, m.HeapAlloc/1024/1024, m.NumGC)
		ctx.DebugTL("Tile Delta %s", frameStats)
		ctx.DebugTL("Game Logic %s", gameLogicStats)
		game.RenderDebugLines(window, ctx.PopDebugLines())
		game.RenderNotifications(window, ctx.PopNotifications(deltaTime))

		gameLogicDur := time.Now().Sub(now).Seconds()
		gameLogicStats.AddFrameTime(gameLogicDur)

		//pixel.NewSprite(resources.SpriteAtlas, resources.SpriteAtlas.Bounds()).Draw(window, pixel.IM.Moved(pixel.V(resources.SpriteAtlasSize/2, 0)))

		window.Update()
	}
}
