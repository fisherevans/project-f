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
	Close()
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

type Instance struct {
	holder        ScreenHolder
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

func NewScreen(holder ScreenHolder, renderContext RenderContext, initialMenu func(instance *Instance) Menu) *Instance {
	s := &Instance{
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

func (s *Instance) Colors() Colors {
	return s.renderContext.Colors
}

func (s *Instance) Width() int {
	return s.renderContext.Width
}

func (s *Instance) Height() int {
	return s.renderContext.Height
}

func (s *Instance) Center() pixel.Vec {
	return pixel.V(s.WidthFloat()/2, s.HeightFloat()/2)
}

func (s *Instance) WidthFloat() float64 {
	return float64(s.renderContext.Width)
}

func (s *Instance) HeightFloat() float64 {
	return float64(s.renderContext.Height)
}

type Menu interface {
	OnTick(target pixel.Target, timeDelta float64)
	Enter()
}

func (s *Instance) onMenuChange(middleOut bool, doAnimate bool) {
	if doAnimate {
		s.screenTransitionElapsed = 0
	}
	s.screenTransitionMiddleOut = middleOut
	s.lastMenuImage.Clear(s.renderContext.Colors.Clear)
	s.canvas.Draw(s.lastMenuImage, pixel.IM.Moved(s.canvas.Bounds().Center()))
}

func (s *Instance) CurrentMenu() Menu {
	if len(s.menuStack) == 0 {
		return nil
	}
	return s.menuStack[len(s.menuStack)-1]
}

func (s *Instance) PushMenuAnimated(menu Menu) {
	s.PushMenu(menu, true)
}

func (s *Instance) PopMenuAnimated() {
	s.PopMenu(true)
}

func (s *Instance) SwapActiveMenuAnimated(menu Menu) {
	s.SwapActiveMenu(menu, true)
}

func (s *Instance) PushMenu(menu Menu, doAnimate bool) {
	s.onMenuChange(true, doAnimate)
	s.menuStack = append(s.menuStack, menu)
	s.CurrentMenu().Enter()
}

func (s *Instance) PopMenu(doAnimate bool) {
	if s.menuStack == nil || len(s.menuStack) == 1 {
		log.Warn().Msgf("Tried to go back to parent menu, but there was none")
		return
	}
	s.onMenuChange(false, doAnimate)
	s.menuStack = s.menuStack[:len(s.menuStack)-1]
	s.CurrentMenu().Enter()
}

func (s *Instance) SwapActiveMenu(menu Menu, doAnimate bool) {
	s.onMenuChange(true, doAnimate)
	s.menuStack[len(s.menuStack)-1] = menu
	s.CurrentMenu().Enter()
}

func (s *Instance) Bounds() pixel.Rect {
	return pixel.R(0, 0, float64(s.renderContext.Width), float64(s.renderContext.Height))
}

func (s *Instance) OnTick(targetMatrix pixel.Matrix, target pixel.Target, timeDelta float64) {
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

func (s *Instance) SpriteFilterBuffer() *spriteShader {
	return s.spriteFilterBuffer
}

func (s *Instance) DrawLastScreen() {
	s.doDrawLastScreen = true
}

func (s *Instance) Holder() ScreenHolder {
	return s.holder
}
