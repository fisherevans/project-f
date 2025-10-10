package xenolog

import (
	"github.com/gopxl/pixel/v2"
	"github.com/rs/zerolog/log"

	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/game/shaders"
	"fisherevans.com/project/f/internal/game/shaders/bloom"
	"fisherevans.com/project/f/internal/util/colors"
)

var (
	screenWidth    = 206
	screenHeight   = 128
	screenClear    = colors.HexString("#0476d0")
	uiMask         = colors.HexString("#a6d8ff")
	uiMaskSelected = colors.HexString("#ffffff")
)

type Screen struct {
	canvas *shaders.Canvas
	batch  *pixel.Batch
	bloom  *bloom.Helper

	menuStack []Menu
}

func NewScreen(ctx *game.Context) *Screen {
	s := &Screen{
		menuStack: []Menu{},
		canvas:    shaders.NewCanvas(screenWidth, screenHeight),
		batch:     atlas.NewBatch(),
		bloom: bloom.NewHelper(
			screenWidth, screenHeight,
			bloom.DefaultBrightnessConfig(),
			bloom.DefaultBlurConfig(),
			bloom.DefaultBlendConfig().WithIntensity(0.5)),
	}
	screenContext := newContext(ctx, s)
	s.menuStack = append(s.menuStack, newHomeMenu(screenContext))
	s.CurrentMenu().Enter(screenContext)
	return s
}

type Menu interface {
	OnTick(ctx Context, target pixel.Target, timeDelta float64)
	Enter(ctx Context)
}

type Context struct {
	*game.Context
	PushMenu       func(Menu)
	PopMenu        func()
	SwapActiveMenu func(Menu)
}

func (s *Screen) CurrentMenu() Menu {
	if len(s.menuStack) == 0 {
		return nil
	}
	return s.menuStack[len(s.menuStack)-1]
}

func newContext(ctx *game.Context, s *Screen) Context {
	screenContext := Context{
		Context: ctx,
	}
	screenContext.PushMenu = func(menu Menu) {
		s.menuStack = append(s.menuStack, menu)
		s.CurrentMenu().Enter(screenContext)
	}
	screenContext.PopMenu = func() {
		if s.menuStack == nil || len(s.menuStack) == 1 {
			log.Warn().Msgf("Tried to go back to parent menu, but there was none")
			return
		}
		s.menuStack = s.menuStack[:len(s.menuStack)-1]
		s.CurrentMenu().Enter(screenContext)
	}
	screenContext.SwapActiveMenu = func(menu Menu) {
		s.menuStack[len(s.menuStack)-1] = menu
		s.CurrentMenu().Enter(screenContext)
	}
	return screenContext
}

func (s *Screen) Bounds() pixel.Rect {
	return pixel.R(0, 0, float64(screenWidth), float64(screenHeight))
}

func (s *Screen) OnTick(state *State, ctx *game.Context, targetMatrix pixel.Matrix, target pixel.Target, timeDelta float64) {
	s.batch.Clear()

	s.menuStack[len(s.menuStack)-1].OnTick(newContext(ctx, s), s.batch, timeDelta)

	s.canvas.Clear(screenClear)
	s.batch.Draw(s.canvas)
	s.canvas.Draw(target, targetMatrix)

	bloomed := s.bloom.ApplyBloom(s.canvas)
	bloomed.Draw(target, targetMatrix)
}
