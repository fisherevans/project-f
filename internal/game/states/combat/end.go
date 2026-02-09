package combat

import (
	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/resources"
	"fisherevans.com/project/f/internal/util/badges"
	"fisherevans.com/project/f/internal/util/colors"
	"fisherevans.com/project/f/internal/util/frames"
	"fisherevans.com/project/f/internal/util/gfx"
	"fisherevans.com/project/f/internal/util/interp"
	"fisherevans.com/project/f/internal/util/keyframe"
	"fisherevans.com/project/f/internal/util/textbox"
	"fisherevans.com/project/f/internal/util/textbox/tbcfg"
	"github.com/gopxl/pixel/v2"
)

var (
	endLineSpacing = 5
	endPadding     = 6
	endFrame       *frames.Instance
	endTextTitle   *textbox.Instance
	endTextRewards *textbox.Instance
)

func init() {
	resources.RunOnceInitialized(func() {
		endFrame = frames.New("combat/rewards/frame", atlas)
		endTextTitle = textbox.NewInstance(atlas.GetFont(resources.FontNameM5x7), tbcfg.NewConfig(game.GameWidth, 9,
			tbcfg.Foreground(colors.White.RGBA),
			tbcfg.HAligned(tbcfg.HAlignCenter),
			tbcfg.VAligned(tbcfg.VAlignTop),
			tbcfg.Foreground(colors.White.RGBA)))
		endTextRewards = textbox.NewInstance(atlas.GetFont(resources.FontNameFF), tbcfg.NewConfig(game.GameWidth, 9,
			tbcfg.Foreground(colors.White.RGBA),
			tbcfg.HAligned(tbcfg.HAlignCenter),
			tbcfg.VAligned(tbcfg.VAlignTop),
			tbcfg.Foreground(greyLight)))
	})
}

type endModal struct {
	contents  []*textbox.Content
	frameR    pixel.Rect
	nextBadge *badges.ButtonAction

	kg          *keyframe.Group
	keyContents []*keyframe.Member
	keyButton   *keyframe.Member
}

func newEndWonModal(opponentName string, reward game.CombatReward) *endModal {
	title := opponentName + " was neutralized!"
	var extraLines []string
	if reward.ExperiencePoints > 0 {
		extraLines = append(extraLines, "Your Animech gained experience")
	}
	if reward.ResearchPoints > 0 {
		extraLines = append(extraLines, "Specimen research was captured")
	}
	return newEndModal(title, extraLines, "next")
}

func newEndLostModal() *endModal {
	title := "You've been defeated!"
	extraLines := []string{
		"Your soul was de-synced with your Animech...",
		"You've lost all progress from this run.",
	}
	return newEndModal(title, extraLines, "return to your ship")
}

func newEndModal(title string, extraLines []string, nextAction string) *endModal {
	kg := keyframe.NewGroup(interp.Smoothstep)
	contents := []*textbox.Content{
		endTextTitle.NewSimpleContent(title),
	}
	delay := 1.0
	keySkip, keyDur := 0.25, 0.5
	keyContents := []*keyframe.Member{
		kg.AddKeyFrame(delay, delay+keyDur),
	}
	nextFrom := delay + keySkip
	for _, line := range extraLines {
		contents = append(contents, endTextRewards.NewSimpleContent(line))
		keyContents = append(keyContents, kg.AddKeyFrame(nextFrom, nextFrom+keyDur))
		nextFrom += keySkip
	}
	keyButton := kg.AddKeyFrame(nextFrom, nextFrom+keyDur)
	width, height := 0, 0
	for idx, c := range contents {
		if idx > 0 {
			height += endLineSpacing
		}
		cw := int(c.Bounds().W())
		ch := int(c.Bounds().H())
		width = max(width, cw)
		height += ch
	}
	return &endModal{
		contents:    contents,
		kg:          kg,
		keyContents: keyContents,
		keyButton:   keyButton,
		frameR:      gfx.R(width+endPadding*4, height+endPadding*2),
		nextBadge:   badges.Using(atlas).ButtonAction("a", nextAction, badges.ButtonStyleStandard),
	}
}

func (m *endModal) exit() {
	m.kg.SetReverse(true)
	m.kg.SetTimeScale(1.5)
}

func (m *endModal) render(target pixel.Target, middle pixel.Matrix, timeDelta float64, renderButton bool) {
	m.kg.Update(timeDelta)

	topMiddle := middle.Moved(pixel.V(0, m.frameR.H()/2))
	topMiddle = topMiddle.Moved(pixel.V(0, 15*m.keyContents[0].InverseProgress()))
	frameMask := colors.WithAlpha(colors.White.RGBA, m.keyContents[0].Progress())

	endFrame.Draw(target, m.frameR, topMiddle, frames.WithRenderOrigin(gfx.TopCenter), frames.WithColor(frameMask))

	topMiddle = topMiddle.Moved(gfx.IVec(0, -4))
	for id, content := range m.contents {
		mask := colors.WithAlpha(colors.White.RGBA, m.keyContents[id].Progress())
		content.Render(target, topMiddle, tbcfg.HAlignedCenter(), tbcfg.VAlignedTop(), tbcfg.ColorMask(mask), tbcfg.RenderFrom(gfx.TopCenter))
		topMiddle = topMiddle.Moved(gfx.IVec(0, -int(content.Bounds().H())))
		topMiddle = topMiddle.Moved(gfx.IVec(0, -endLineSpacing))
	}

	if !renderButton {
		return
	}

	topMiddle = topMiddle.Moved(gfx.IVec(0, -8))
	buttonMask := colors.WithAlpha(colors.White.RGBA, m.keyButton.Progress())
	m.nextBadge.RenderWithColorMask(target, topMiddle, gfx.TopCenter, buttonMask)
}
