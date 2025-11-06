package menu

import (
	"image/color"
	"math"

	"github.com/gopxl/pixel/v2"

	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/game/shaders"
	"fisherevans.com/project/f/internal/resources"
	"fisherevans.com/project/f/internal/util/colors"
	"fisherevans.com/project/f/internal/util/gfx"
)

var atlas = resources.DefaultAtlas()

var (
	menuContentWidth = 100
	menuMargin       = 6
	menuPadding      = 5
	itemPadding      = 3
	transitionTime   = 0.25
)

type State struct {
	background game.State
	transition float64
	exiting    bool

	menuStack []*Menu
	batch     *pixel.Batch
}

func New(i game.MenuIntent) game.State {
	s := &State{
		background: i.Background,
		transition: 0,
		batch:      atlas.NewBatch(),
	}
	s.menuStack = []*Menu{createMainMenu(s)}
	return s
}

func (s *State) ClearColor() color.Color {
	return colors.Black.RGBA
}

func (s *State) OnTick(target *shaders.Canvas, targetBounds pixel.Rect, timeDelta float64) {
	if s.background != nil {
		s.background.OnTick(target, targetBounds, 0)
	}

	s.batch.Clear()

	transitionDt := timeDelta
	if s.exiting {
		transitionDt *= -1
	}
	s.transition = math.Min(1, s.transition+transitionDt*(1.0/transitionTime))
	if s.exiting && s.transition <= 0 {
		if err := game.CurrentSave().Save(); err != nil {
			game.DebugNotificationf("failed to save: " + err.Error())
		}
		game.SetActiveStateIntent(game.SwapStateIntent{
			State: s.background,
		})
		return
	}

	fadeColor := colors.WithAlphaTodoFix(colors.Black.RGBA, 0.7*s.transition)
	gfx.DrawRect(atlas, s.batch, pixel.IM, gfx.BottomLeft, game.GameWidth, game.GameHeight, fadeColor)

	if game.Controls[*State]().ButtonSelect().JustPressed() {
		s.exiting = !s.exiting
	}
	if !s.exiting {
		s.menuStack[len(s.menuStack)-1].HandleInputs()
	}

	topLeft := pixel.IM.Moved(pixel.V(-100.0*(1.0-s.transition), game.GameHeight-10))
	for menuIndex, menu := range s.menuStack {
		if menuIndex > 0 {
			gfx.DrawRect(atlas, s.batch, pixel.IM, gfx.BottomLeft, game.GameWidth, game.GameHeight, colors.WithAlpha(colors.Black.RGBA, 0.2))
		}
		menu.Render(s, topLeft.Moved(gfx.IVec(20*menuIndex+menuMargin, -5*menuIndex-menuMargin)), s.batch, targetBounds)
	}

	s.batch.Draw(target)

	game.DebugBLf("retro frame: %v", game.CurrentSave().SystemSettings.RetroFrame)
}

func (s *State) PopMenu() {
	if len(s.menuStack) <= 1 {
		s.exiting = true
	} else {
		s.menuStack = s.menuStack[:len(s.menuStack)-1]
	}
}

func (s *State) PushMenu(menu *Menu) {
	s.menuStack = append(s.menuStack, menu)
}
