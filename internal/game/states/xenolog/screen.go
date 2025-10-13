package xenolog

import (
	"math"

	"github.com/rs/zerolog/log"

	"github.com/gopxl/pixel/v2"
	"github.com/gopxl/pixel/v2/backends/opengl"

	"fisherevans.com/project/f/internal/game"
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

	spriteFilterBuffer *spriteShader

	lastMenuImage             *opengl.Canvas
	screenEffectShader        *shaders.Canvas
	screenTransitionElapsed   float64
	screenTransitionMiddleOut bool

	menuStack []Menu
}

func NewScreen() *Screen {
	s := &Screen{
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
	s.menuStack = append(s.menuStack, newHomeMenu(s))
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

func (s *Screen) snapshotLastScreen(middleOut bool) {
	s.screenTransitionElapsed = 0
	s.screenTransitionMiddleOut = middleOut
	s.lastMenuImage.Clear(colors.XenoLogClear.RGBA)
	s.canvas.Draw(s.lastMenuImage, pixel.IM.Moved(s.canvas.Bounds().Center()))
}

func (s *Screen) CurrentMenu() Menu {
	if len(s.menuStack) == 0 {
		return nil
	}
	return s.menuStack[len(s.menuStack)-1]
}

func (s *Screen) PushMenu(menu Menu) {
	s.snapshotLastScreen(true)
	s.menuStack = append(s.menuStack, menu)
	s.CurrentMenu().Enter()
}

func (s *Screen) PopMenu() {
	if s.menuStack == nil || len(s.menuStack) == 1 {
		log.Warn().Msgf("Tried to go back to parent menu, but there was none")
		return
	}
	s.snapshotLastScreen(false)
	s.menuStack = s.menuStack[:len(s.menuStack)-1]
	s.CurrentMenu().Enter()
}

func (s *Screen) SwapActiveMenu(menu Menu) {
	s.snapshotLastScreen(true)
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
	s.canvas.Clear(colors.XenoLogClear.RGBA)

	s.batch.Draw(s.canvas)
	s.spriteFilterBuffer.Render(s.canvas)
	s.canvas.Draw(s.screenEffectShader, pixel.IM.Moved(s.canvas.Bounds().Center()))
	s.screenEffectShader.Draw(target, targetMatrix)

	bloomed := s.bloom.ApplyBloom(s.screenEffectShader)
	bloomed.Draw(target, targetMatrix)
}
