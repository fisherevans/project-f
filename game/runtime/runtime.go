package runtime

import (
	"fisherevans.com/project/f/game"
	"fisherevans.com/project/f/game/states/state_selector"
	"github.com/gopxl/pixel/v2"
	"github.com/gopxl/pixel/v2/backends/opengl"
	"image/color"
	"math"
	"os"
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

	ctx := game.NewContext(state_selector.New())

	// Create the fixed-size canvas
	canvas := opengl.NewCanvas(pixel.R(0, 0, game.GameWidth, game.GameHeight))
	canvas.SetSmooth(false)

	targetFps := 60
	frameDuration := time.Second / time.Duration(targetFps)
	lastTime := time.Now()

	for !window.Closed() {
		if window.JustPressed(pixel.KeyF4) {
			os.Exit(0)
		}

		// Calculate the time delta
		loopStartTime := time.Now()
		elapsed := loopStartTime.Sub(lastTime)
		if elapsed < frameDuration {
			time.Sleep(frameDuration - elapsed)
		}
		deltaTime := loopStartTime.Sub(lastTime).Seconds()
		lastTime = loopStartTime

		window.Clear(color.RGBA{R: 40, G: 40, B: 40, A: 255})
		canvas.Clear(color.RGBA{R: 10, G: 10, B: 10, A: 255})

		// Calculate scale and offset for whole-number scaling
		windowWidth, windowHeight := window.Bounds().Size().XY()
		scaleX := math.Floor(windowWidth / game.GameWidth)
		scaleY := math.Floor(windowHeight / game.GameHeight)
		ctx.CanvasScale = math.Min(scaleX, scaleY) // Use the smaller scale
		canvasMatrix := pixel.IM.Scaled(pixel.ZV, ctx.CanvasScale).Moved(window.Bounds().Center())
		windowMousePosition := window.MousePosition()
		ctx.CanvasMousePosition = canvasMatrix.Unproject(windowMousePosition).Add(canvas.Bounds().Center())

		ctx.GetActiveState().OnTick(ctx, window, canvas, deltaTime)

		canvas.Draw(window, canvasMatrix)

		game.RenderDebugLines(window, ctx.PopDebugLines())

		window.Update()
	}
}
