package startup

import (
	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/resources"
	"fisherevans.com/project/f/internal/util/frames"
	"fisherevans.com/project/f/internal/util/gfx"
	"fisherevans.com/project/f/internal/util/textbox"
	"fisherevans.com/project/f/internal/util/textbox/tbcfg"
)

var atlas *resources.Atlas

func init() {
	resources.RunOnceInitialized(func() {
		atlas = resources.CreateAtlas(resources.AtlasFilter{
			DoIncludeSprite: resources.RequireSpritePrefix("startup/", "1x1", "2x2", "common/"),
			FontNames: []string{
				resources.FontNameFF,
				resources.FontNameAddStandard,
			},
		})
		progressFrame = frames.New("common/rounded_frame_2px", atlas, frames.WithRenderOrigin(gfx.Centered))
		smallText = textbox.NewInstance(
			atlas.GetFont(resources.FontNameFF),
			tbcfg.NewConfig(game.GameWidth, progressFrameHeight,
				tbcfg.HAligned(tbcfg.HAlignCenter),
				tbcfg.VAligned(tbcfg.VAlignMiddle),
				tbcfg.WithExpandMode(tbcfg.ExpandFit),
				tbcfg.RenderFrom(gfx.Centered)))
		titleText = textbox.NewInstance(
			atlas.GetFont(resources.FontNameAddStandard),
			tbcfg.NewConfig(game.GameWidth, progressFrameHeight,
				tbcfg.HAligned(tbcfg.HAlignCenter),
				tbcfg.VAligned(tbcfg.VAlignMiddle),
				tbcfg.WithExpandMode(tbcfg.ExpandFit),
				tbcfg.RenderFrom(gfx.Centered)))
	})
}

var centerMatrix = gfx.Moved(game.GameWidth/2, game.GameHeight/2)
