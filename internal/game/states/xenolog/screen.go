package xenolog

import (
	"math"

	"github.com/rs/zerolog/log"

	"github.com/gopxl/pixel/v2"
	"github.com/gopxl/pixel/v2/backends/opengl"

	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/game/input"
	"fisherevans.com/project/f/internal/game/shaders"
	"fisherevans.com/project/f/internal/game/shaders/bloom"
	"fisherevans.com/project/f/internal/util/badges"
	"fisherevans.com/project/f/internal/util/colors"
	"fisherevans.com/project/f/internal/util/pixelutil"
)

var (
	screenWidth  = 206
	screenHeight = 128

	badgeButtonStyle = badges.ButtonColorStyle{
		Action:    colorText,
		Button:    colorDark,
		Highlight: colorHighlight,
	}

	colorDark      = colors.XenoLogDark.RGBA
	colorClear     = colors.XenoLogClear.RGBA
	colorText      = colors.XenoLogText.RGBA
	colorHighlight = colors.XenoLogHighlight.RGBA
)

type Screen struct {
	state  *State
	canvas *shaders.Canvas
	batch  *pixel.Batch
	bloom  *bloom.Helper

	spriteFilterBuffer *spriteShader

	lastMenuImage             *opengl.Canvas
	screenEffectShader        *shaders.Canvas
	screenTransitionElapsed   float64
	screenTransitionMiddleOut bool

	menuStack        []Menu
	doDrawLastScreen bool
}

func NewScreen(state *State, primortalsEnabled bool) *Screen {
	s := &Screen{
		state:              state,
		menuStack:          []Menu{},
		canvas:             shaders.NewCanvas(screenWidth, screenHeight),
		batch:              atlas.NewBatch(),
		spriteFilterBuffer: newSpriteShader(),
		bloom: bloom.NewHelper(
			screenWidth, screenHeight,
			bloom.DefaultBrightnessConfig(),
			bloom.DefaultBlurConfig(),
			bloom.DefaultBlendConfig().WithIntensity(0.5)),
	}
	s.menuStack = append(s.menuStack, newHomeMenu(s, primortalsEnabled))
	s.CurrentMenu().Enter()
	// setup screen transition
	s.canvas.Clear(colors.Black.RGBA)
	s.lastMenuImage = opengl.NewCanvas(pixel.R(0, 0, float64(screenWidth), float64(screenHeight)))
	s.lastMenuImage.Clear(colors.Black.RGBA)
	s.screenTransitionMiddleOut = true
	s.screenEffectShader = shaders.NewCanvas(screenWidth, screenHeight)
	s.screenEffectShader.SetupScreenTransition(s.lastMenuImage.Texture(), .175)
	return s
}

type Menu interface {
	OnTick(target pixel.Target, timeDelta float64)
	Enter()
}

func (s *Screen) onMenuChange(middleOut bool, doAnimate bool) {
	if doAnimate {
		s.screenTransitionElapsed = 0
	}
	s.screenTransitionMiddleOut = middleOut
	s.lastMenuImage.Clear(colorClear)
	s.canvas.Draw(s.lastMenuImage, pixel.IM.Moved(s.canvas.Bounds().Center()))
}

func (s *Screen) CurrentMenu() Menu {
	if len(s.menuStack) == 0 {
		return nil
	}
	return s.menuStack[len(s.menuStack)-1]
}

func (s *Screen) PushMenuAnimated(menu Menu) {
	s.PushMenu(menu, true)
}

func (s *Screen) PopMenuAnimated() {
	s.PopMenu(true)
}

func (s *Screen) SwapActiveMenuAnimated(menu Menu) {
	s.SwapActiveMenu(menu, true)
}

func (s *Screen) PushMenu(menu Menu, doAnimate bool) {
	s.onMenuChange(true, doAnimate)
	s.menuStack = append(s.menuStack, menu)
	s.CurrentMenu().Enter()
}

func (s *Screen) PopMenu(doAnimate bool) {
	if s.menuStack == nil || len(s.menuStack) == 1 {
		log.Warn().Msgf("Tried to go back to parent menu, but there was none")
		return
	}
	s.onMenuChange(false, doAnimate)
	s.menuStack = s.menuStack[:len(s.menuStack)-1]
	s.CurrentMenu().Enter()
}

func (s *Screen) SwapActiveMenu(menu Menu, doAnimate bool) {
	s.onMenuChange(true, doAnimate)
	s.menuStack[len(s.menuStack)-1] = menu
	s.CurrentMenu().Enter()
}

func (s *Screen) Bounds() pixel.Rect {
	return pixel.R(0, 0, float64(screenWidth), float64(screenHeight))
}

func (s *Screen) OnTick(state *State, targetMatrix pixel.Matrix, target pixel.Target, timeDelta float64) {
	s.screenTransitionElapsed += timeDelta
	s.screenEffectShader.UpdateScreenTransition(
		float32(game.TimeElapsed()),
		float32(math.Max(0.0001, s.screenTransitionElapsed)), // max with 0 to allow for delay into transition at start
		s.screenTransitionMiddleOut,
		s.lastMenuImage.Texture())

	s.spriteFilterBuffer.Clear()
	s.batch.Clear()

	s.menuStack[len(s.menuStack)-1].OnTick(s.batch, timeDelta)

	// canvas clear after OnTick so that snapshotLastScreen can grab the last image
	s.canvas.Clear(colorClear)

	if s.doDrawLastScreen {
		s.doDrawLastScreen = false
		s.lastMenuImage.Draw(s.canvas, pixel.IM.Moved(s.canvas.Bounds().Center()))
	}

	s.batch.Draw(s.canvas)
	s.spriteFilterBuffer.Render(s.canvas)
	s.canvas.Draw(s.screenEffectShader, pixel.IM.Moved(s.canvas.Bounds().Center()))
	s.screenEffectShader.Draw(target, targetMatrix)

	bloomed := s.bloom.ApplyBloom(s.screenEffectShader)
	bloomed.Draw(target, targetMatrix)
}

func (s *Screen) drawLastScreen() {
	s.doDrawLastScreen = true
}

func arrowSprite(dir input.Direction) pixelutil.BoundedDrawable {
	switch dir {
	case input.Up:
		return arrowUp
	case input.Right:
		return arrowRight
	case input.Down:
		return arrowDown
	case input.Left:
		return arrowLeft
	default:
		panic("invalid direction")
	}
}
