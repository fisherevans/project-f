package tbtest

import (
	"image/color"

	"github.com/gopxl/pixel/v2"

	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/game/input"
	"fisherevans.com/project/f/internal/game/shaders"
	"fisherevans.com/project/f/internal/resources"
	"fisherevans.com/project/f/internal/util/colors"
	"fisherevans.com/project/f/internal/util/gfx"
	"fisherevans.com/project/f/internal/util/textbox"
	"fisherevans.com/project/f/internal/util/textbox/tbcfg"
)

var (
	halignments = []tbcfg.HAlignment{
		tbcfg.HAlignLeft,
		tbcfg.HAlignCenter,
		tbcfg.HAlignRight,
	}

	valignments = []tbcfg.VAlignment{
		tbcfg.VAlignTop,
		tbcfg.VAlignMiddle,
		tbcfg.VAlignBottom,
	}

	expands = []tbcfg.ExpandMode{
		tbcfg.ExpandFull,
		tbcfg.ExpandFit,
	}
)

type State struct {
	haligned int
	valigned int
	expand   int
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
	tb      = textbox.NewInstance(atlas.GetFont(resources.FontNameFF), tbcfg.NewConfig(100, 50, tbcfg.Foreground(colors.SkillTypeKinetic.RGBA)))
	content = tb.NewComplexContent("Hello, world! How are you doing today?")
)

func (s *State) OnTick(ctx *game.Context, target *shaders.Canvas, targetBounds pixel.Rect, timeDelta float64) {

	if ctx.Controls.DPad().DirectionJustPressed(input.Up) {
		s.expand++
		if s.expand >= len(expands) {
			s.expand = 0
		}
	}
	ctx.DebugBR("Up: toggle expand (%s)", expands[s.expand].Name())

	if ctx.Controls.DPad().DirectionJustPressed(input.Right) {
		s.haligned++
		if s.haligned >= len(halignments) {
			s.haligned = 0
		}
	}
	ctx.DebugBR("Right: toggle h alignment (%s)", halignments[s.haligned].Name())

	if ctx.Controls.DPad().DirectionJustPressed(input.Left) {
		s.valigned++
		if s.valigned >= len(valignments) {
			s.valigned = 0
		}
	}
	ctx.DebugBR("Left: toggle v alignment (%s)", valignments[s.valigned].Name())

	opts := []tbcfg.ConfigOpt{
		tbcfg.WithExpandMode(expands[s.expand]),
		tbcfg.HAligned(halignments[s.haligned]),
		tbcfg.VAligned(valignments[s.valigned]),
	}

	left := 0.0
	right := targetBounds.Max.X

	top := targetBounds.Max.Y
	bottom := 0.0

	atlas.GetSprite("2x2").Draw(target, pixel.IM.Moved(pixel.V(right/2, top/2)))

	ctx.DebugBR("left %.1f, right %.1f, top %.1f, bottom %.1f", left, right, top, bottom)

	render := func(vec pixel.Vec, origin gfx.OriginLocation) {
		gfx.DrawRect(atlas, target, pixel.IM.Moved(vec), origin, tb.GetConfig().BoxWidth, tb.GetConfig().BoxHeight, colors.HexColor("#111"))
		tb.Render(ctx, target, pixel.IM.Moved(vec), content, append(opts, tbcfg.RenderFrom(origin))...)

	}

	render(pixel.V(left, top), gfx.TopLeft)
	render(pixel.V(right, top), gfx.TopRight)
	render(pixel.V(right, bottom), gfx.BottomRight)
	render(pixel.V(left, bottom), gfx.BottomLeft)
	render(pixel.V(right/2.0, top/2.0), gfx.Centered)

}
