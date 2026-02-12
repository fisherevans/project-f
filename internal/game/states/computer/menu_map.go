package computer

import (
	"fisherevans.com/project/f/internal/game/states/xenolog/screen"
	"fisherevans.com/project/f/internal/resources"
	"fisherevans.com/project/f/internal/util/gfx"
	"fisherevans.com/project/f/internal/util/textbox"
	"fisherevans.com/project/f/internal/util/textbox/tbcfg"
	"github.com/gopxl/pixel/v2"

	"fisherevans.com/project/f/internal/game"
)

// Variables declared in init_vars.go
var (
	smallText *textbox.Instance
)

func init() {
	resources.RunOnceInitialized(func() {
		smallText = textbox.NewInstance(atlas.GetFont(resources.FontNameFF57i),
			tbcfg.NewConfig(screenWidth, 10,
				tbcfg.Foreground(screenColors.Text),
				tbcfg.RenderFrom(gfx.TopCenter),
				tbcfg.VAligned(tbcfg.VAlignTop),
				tbcfg.HAligned(tbcfg.HAlignCenter),
			))
	})
}

type mapMenu struct {
	screen *screen.Instance
}

func newMapMenu(screen *screen.Instance) screen.Menu {
	home := &mapMenu{
		screen: screen,
	}
	return home
}

func (s *mapMenu) Enter() {
}

func (s *mapMenu) OnTick(target pixel.Target, timeDelta float64) {
	if game.Controls[*State]().ButtonB().JustPressed() {
		s.screen.Holder().Close()
	}
	content := smallText.NewSimpleContent("Hello, world!")
	content.Render(target, pixel.IM.Moved(s.screen.Center()))
}
