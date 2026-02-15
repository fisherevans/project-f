package screen

import (
	"math"

	"fisherevans.com/project/f/internal/resources"
	"fisherevans.com/project/f/internal/util/colors"
	"github.com/rs/zerolog/log"

	"github.com/gopxl/pixel/v2"
	"github.com/gopxl/pixel/v2/backends/opengl"

	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/game/shaders"
	"fisherevans.com/project/f/internal/game/shaders/bloom"
)

type ScreenHolder interface {
	ToggleOpenState()
}

type RenderContext struct {
	Atlas  *resources.Atlas
	Width  int
	Height int
	Colors Colors
}

type Colors struct {
	Clear     pixel.RGBA
	Dark      pixel.RGBA
	Text      pixel.RGBA
	Highlight pixel.RGBA
}

type Instance[H ScreenHolder] struct {
	holder        H
	renderContext RenderContext

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

func NewScreen[H ScreenHolder](holder H, renderContext RenderContext, initialMenu func(instance *Instance[H]) Menu) *Instance[H] {
	s := &Instance[H]{
		holder:             holder,
		renderContext:      renderContext,
		menuStack:          []Menu{},
		canvas:             shaders.NewCanvas(renderContext.Width, renderContext.Height),
		batch:              renderContext.Atlas.NewBatch(),
		spriteFilterBuffer: newSpriteShader(renderContext),
		bloom: bloom.NewHelper(
			renderContext.Width, renderContext.Height,
			bloom.DefaultBrightnessConfig(),
			bloom.DefaultBlurConfig(),
			bloom.DefaultBlendConfig().WithIntensity(0.5)),
	}
	s.menuStack = append(s.menuStack, initialMenu(s))
	s.CurrentMenu().Enter()
	// setup screen transition
	s.canvas.Clear(colors.Black.RGBA)
	s.lastMenuImage = opengl.NewCanvas(pixel.R(0, 0, float64(renderContext.Width), float64(renderContext.Height)))
	s.lastMenuImage.Clear(colors.Black.RGBA)
	s.screenTransitionMiddleOut = true
	s.screenEffectShader = shaders.NewCanvas(renderContext.Width, renderContext.Height)
	s.screenEffectShader.SetupScreenTransition(s.lastMenuImage.Texture(), .175)
	return s
}

func (s *Instance[H]) Colors() Colors {
	return s.renderContext.Colors
}

func (s *Instance[H]) Width() int {
	return s.renderContext.Width
}

func (s *Instance[H]) Height() int {
	return s.renderContext.Height
}

func (s *Instance[H]) Center() pixel.Vec {
	return pixel.V(s.WidthFloat()/2, s.HeightFloat()/2)
}

func (s *Instance[H]) WidthFloat() float64 {
	return float64(s.renderContext.Width)
}

func (s *Instance[H]) HeightFloat() float64 {
	return float64(s.renderContext.Height)
}

type Menu interface {
	OnTick(target pixel.Target, timeDelta float64)
	Enter()
}

func (s *Instance[H]) onMenuChange(middleOut bool, doAnimate bool) {
	if doAnimate {
		s.screenTransitionElapsed = 0
	}
	s.screenTransitionMiddleOut = middleOut
	s.lastMenuImage.Clear(s.renderContext.Colors.Clear)
	s.canvas.Draw(s.lastMenuImage, pixel.IM.Moved(s.canvas.Bounds().Center()))
}

func (s *Instance[H]) CurrentMenu() Menu {
	if len(s.menuStack) == 0 {
		return nil
	}
	return s.menuStack[len(s.menuStack)-1]
}

func (s *Instance[H]) PushMenuAnimated(menu Menu) {
	s.PushMenu(menu, true)
}

func (s *Instance[H]) PopMenuAnimated() {
	s.PopMenu(true)
}

func (s *Instance[H]) SwapActiveMenuAnimated(menu Menu) {
	s.SwapActiveMenu(menu, true)
}

func (s *Instance[H]) PushMenu(menu Menu, doAnimate bool) {
	s.onMenuChange(true, doAnimate)
	s.menuStack = append(s.menuStack, menu)
	s.CurrentMenu().Enter()
}

func (s *Instance[H]) PopMenu(doAnimate bool) {
	if s.menuStack == nil || len(s.menuStack) == 1 {
		log.Warn().Msgf("Tried to go back to parent menu, but there was none")
		return
	}
	s.onMenuChange(false, doAnimate)
	s.menuStack = s.menuStack[:len(s.menuStack)-1]
	s.CurrentMenu().Enter()
}

func (s *Instance[H]) SwapActiveMenu(menu Menu, doAnimate bool) {
	s.onMenuChange(true, doAnimate)
	s.menuStack[len(s.menuStack)-1] = menu
	s.CurrentMenu().Enter()
}

func (s *Instance[H]) Bounds() pixel.Rect {
	return pixel.R(0, 0, float64(s.renderContext.Width), float64(s.renderContext.Height))
}

func (s *Instance[H]) OnTick(targetMatrix pixel.Matrix, target pixel.Target, timeDelta float64) {
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
	s.canvas.Clear(s.renderContext.Colors.Clear)

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

func (s *Instance[H]) SpriteFilterBuffer() *spriteShader {
	return s.spriteFilterBuffer
}

func (s *Instance[H]) DrawLastScreen() {
	s.doDrawLastScreen = true
}

func (s *Instance[H]) Holder() H {
	return s.holder
}
