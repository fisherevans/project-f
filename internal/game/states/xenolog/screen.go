package xenolog

import (
	"github.com/gopxl/pixel/v2"
	"github.com/rs/zerolog/log"

	"fisherevans.com/project/f/internal/game/shaders"
	"fisherevans.com/project/f/internal/game/shaders/bloom"
	"fisherevans.com/project/f/internal/util/badges"
	"fisherevans.com/project/f/internal/util/colors"
)

var (
	screenWidth  = 206
	screenHeight = 128

	badgeButtonStyle = badges.ButtonColorStyle{
		Action:    colors.XenoLogText.RGBA,
		Button:    colors.XenoLogDark.RGBA,
		Highlight: colors.XenoLogHighlight.RGBA,
	}
	badgeASelect     = badges.Using(atlas).ButtonAction("a", "select", badgeButtonStyle)
	badgeBBack       = badges.Using(atlas).ButtonAction("b", "back", badgeButtonStyle)
	badgeSelectClose = badges.Using(atlas).ButtonAction("select", "close", badgeButtonStyle)
	badgeF1Reset     = badges.Using(atlas).ButtonAction("f1", "reset", badgeButtonStyle)
	badgeADetails    = badges.Using(atlas).ButtonAction("a", "details", badgeButtonStyle)
	badgeAUnlock     = badges.Using(atlas).ButtonAction("a", "unlock", badgeButtonStyle)

	arrowUp    = atlas.GetTilesheetSprite("common/arrows_5px", 1, 1)
	arrowRight = atlas.GetTilesheetSprite("common/arrows_5px", 2, 1)
	arrowDown  = atlas.GetTilesheetSprite("common/arrows_5px", 3, 1)
	arrowLeft  = atlas.GetTilesheetSprite("common/arrows_5px", 4, 1)

	dot = atlas.GetTilesheetSprite("common/symbols_5px", 1, 1)
)

type Screen struct {
	canvas *shaders.Canvas
	batch  *pixel.Batch
	bloom  *bloom.Helper

	spriteShader *spriteShader

	menuStack []Menu
}

func NewScreen() *Screen {
	s := &Screen{
		menuStack:    []Menu{},
		canvas:       shaders.NewCanvas(screenWidth, screenHeight),
		batch:        atlas.NewBatch(),
		spriteShader: newSpriteShader(),
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
	s.spriteShader.Clear()
	s.batch.Clear()
	s.canvas.Clear(colors.XenoLogClear.RGBA)

	s.menuStack[len(s.menuStack)-1].OnTick(s.batch, timeDelta)

	s.batch.Draw(s.canvas)
	s.spriteShader.Render(s.canvas)
	s.canvas.Draw(target, targetMatrix)

	bloomed := s.bloom.ApplyBloom(s.canvas)
	bloomed.Draw(target, targetMatrix)
}
