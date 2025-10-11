package xenolog

import (
	"github.com/gopxl/pixel/v2"
	"github.com/rs/zerolog/log"

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

func NewScreen() *Screen {
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
	s.menuStack = append(s.menuStack, newHomeMenu(s))
	s.CurrentMenu().Enter()
	return s
}

type Menu interface {
	OnTick(target pixel.Target, timeDelta float64)
	Enter()
}

func (s *Screen) CurrentMenu() Menu {
	if len(s.menuStack) == 0 {
		return nil
	}
	return s.menuStack[len(s.menuStack)-1]
}

func (s *Screen) PushMenu(menu Menu) {
	s.menuStack = append(s.menuStack, menu)
	s.CurrentMenu().Enter()
}

func (s *Screen) PopMenu() {
	if s.menuStack == nil || len(s.menuStack) == 1 {
		log.Warn().Msgf("Tried to go back to parent menu, but there was none")
		return
	}
	s.menuStack = s.menuStack[:len(s.menuStack)-1]
	s.CurrentMenu().Enter()
}

func (s *Screen) SwapActiveMenu(menu Menu) {
	s.menuStack[len(s.menuStack)-1] = menu
	s.CurrentMenu().Enter()
}

func (s *Screen) Bounds() pixel.Rect {
	return pixel.R(0, 0, float64(screenWidth), float64(screenHeight))
}

func (s *Screen) OnTick(state *State, targetMatrix pixel.Matrix, target pixel.Target, timeDelta float64) {
	s.batch.Clear()

	s.menuStack[len(s.menuStack)-1].OnTick(s.batch, timeDelta)

	s.canvas.Clear(screenClear)
	s.batch.Draw(s.canvas)
	s.canvas.Draw(target, targetMatrix)

	bloomed := s.bloom.ApplyBloom(s.canvas)
	bloomed.Draw(target, targetMatrix)
}
