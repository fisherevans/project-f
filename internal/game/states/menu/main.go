package menu

import (
	"os"
	"strings"

	"github.com/gopxl/pixel/v2"
	"github.com/rs/zerolog/log"

	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/resources"
	"fisherevans.com/project/f/internal/util"
	"fisherevans.com/project/f/internal/util/colors"
	"fisherevans.com/project/f/internal/util/frames"
	"fisherevans.com/project/f/internal/util/gfx"
	"fisherevans.com/project/f/internal/util/textbox"
	"fisherevans.com/project/f/internal/util/textbox/tbcfg"
)

var (
	mainFrame = frames.New("menu/background", atlas,
		frames.WithRenderOrigin(gfx.TopLeft),
	)
	mainText = textbox.NewInstance(atlas.GetFont(resources.FontNameM5x7), tbcfg.NewConfig(
		0,
		0,
		tbcfg.Foreground(colors.Black.RGBA),
		tbcfg.HAligned(tbcfg.HAlignLeft),
		tbcfg.VAligned(tbcfg.VAlignTop),
		tbcfg.ExtraLineSpacing(2),
	))
)

type MainItem struct {
	Label  string
	Action func(s *State, ctx *game.Context)
}

type Main struct {
	items *util.Selectable[MainItem]
}

func (m Main) Render(s *State, ctx *game.Context, matrix pixel.Matrix, target pixel.Target, targetBounds pixel.Rect) {
	m.items.UpdateSelection(ctx.Controls)
	if ctx.Controls.ButtonA().JustPressed() {
		if m.items.SelectedEntry != nil && m.items.SelectedEntry.Value.Action != nil {
			m.items.SelectedEntry.Value.Action(s, ctx)
		}
	}

	var labels []string
	for _, item := range m.items.Entries {
		label := item.Value.Label
		if m.items.SelectedEntry == item {
			label = "{+c:#111,+u}" + label
		} else {
			label = "{+c:#555}" + label
		}
		label += "{-*}"
		labels = append(labels, label)
	}
	content := mainText.NewComplexContent(strings.Join(labels, "\n"))

	frameTopLeft := matrix.Moved(pixel.V(margin, targetBounds.H()-margin))
	frameBounds := pixel.R(0, 0, content.Bounds().W()+padding*2, content.Bounds().H()+padding*2)
	mainFrame.Draw(target, frameBounds, frameTopLeft)
	mainText.Render(ctx, target, frameTopLeft.Moved(pixel.V(padding, -padding)), content, tbcfg.Foreground(colors.White.RGBA))
}

func createMain() Main {
	return Main{
		items: util.NewSimpleSelectable[MainItem]([]MainItem{
			{
				Label:  "Animech",
				Action: nil,
			},
			{
				Label:  "Primortals",
				Action: nil,
			},
			{
				Label:  "Inventory",
				Action: nil,
			},
			{
				Label: "Return to game",
				Action: func(s *State, ctx *game.Context) {
					ctx.SetActiveStateIntent(game.SwapStateIntent{
						State: s.background,
					})
				},
			},
			{
				Label: "Quit",
				Action: func(s *State, ctx *game.Context) {
					log.Info().Msg("Good bye!")
					os.Exit(0)
				},
			},
		}),
	}
}
