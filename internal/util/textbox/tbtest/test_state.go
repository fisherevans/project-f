package tbtest

import (
	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/game/input"
	"fisherevans.com/project/f/internal/resources"
	"fisherevans.com/project/f/internal/util/colors"
	"fisherevans.com/project/f/internal/util/gfx"
	"fisherevans.com/project/f/internal/util/textbox"
	"fisherevans.com/project/f/internal/util/textbox/tbcfg"
	"github.com/gopxl/pixel/v2"
	"image/color"
)

var (
	alignments = []tbcfg.Alignment{
		tbcfg.AlignLeft,
		tbcfg.AlignCenter,
		tbcfg.AlignRight,
	}

	expands = []tbcfg.ExpandMode{
		tbcfg.ExpandFull,
		tbcfg.ExpandFit,
	}
)

type State struct {
	aligned int
	expand  int
}

func New() *State {
	return &State{}
}

func (s *State) ClearColor() color.Color {
	return colors.Black.RGBA
}

var (
	atlas = resources.CreateAtlas(resources.AtlasFilter{
		FontNames: []string{resources.FontNameFF},
	})
	tb      = textbox.NewInstance(atlas.GetFont(resources.FontNameFF), tbcfg.NewConfig(100, tbcfg.Foreground(colors.White.RGBA)))
	content = tb.NewComplexContent("Hello, world! How are you doing today?")
)

func (s *State) OnTick(ctx *game.Context, target pixel.Target, targetBounds pixel.Rect, timeDelta float64) {

	if ctx.Controls.DPad().DirectionJustPressed(input.Up) {
		s.expand++
		if s.expand >= len(expands) {
			s.expand = 0
		}
	}
	ctx.DebugBR("Up: toggle expand (%s)", expands[s.expand].Name())

	if ctx.Controls.DPad().DirectionJustPressed(input.Down) {
		s.aligned++
		if s.aligned >= len(alignments) {
			s.aligned = 0
		}
	}
	ctx.DebugBR("Down: toggle alignment (%s)", alignments[s.aligned].Name())

	opts := []tbcfg.ConfigOpt{
		tbcfg.WithExpandMode(expands[s.expand]),
		tbcfg.Aligned(alignments[s.aligned]),
	}

	left := 0.0
	right := targetBounds.Max.X

	top := targetBounds.Max.Y
	bottom := 0.0

	atlas.GetSprite("2x2").Draw(target, pixel.IM.Moved(pixel.V(right/2, top/2)))

	ctx.DebugBR("left %.1f, right %.1f, top %.1f, bottom %.1f", left, right, top, bottom)

	tb.Render(ctx, target, pixel.IM.Moved(pixel.V(left, top)), content, append(opts, tbcfg.RenderFrom(gfx.TopLeft))...)
	tb.Render(ctx, target, pixel.IM.Moved(pixel.V(right, top)), content, append(opts, tbcfg.RenderFrom(gfx.TopRight))...)
	tb.Render(ctx, target, pixel.IM.Moved(pixel.V(right, bottom)), content, append(opts, tbcfg.RenderFrom(gfx.BottomRight))...)
	tb.Render(ctx, target, pixel.IM.Moved(pixel.V(left, bottom)), content, append(opts, tbcfg.RenderFrom(gfx.BottomLeft))...)
	tb.Render(ctx, target, pixel.IM.Moved(pixel.V(right/2.0, top/2.0)), content, append(opts, tbcfg.RenderFrom(gfx.Centered))...)

}
