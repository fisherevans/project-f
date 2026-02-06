package combat

import (
	"fmt"

	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/game/anim"
	"fisherevans.com/project/f/internal/game/rpg"
	"fisherevans.com/project/f/internal/resources"
	"fisherevans.com/project/f/internal/util/badges"
	"fisherevans.com/project/f/internal/util/colors"
	"fisherevans.com/project/f/internal/util/frames"
	"fisherevans.com/project/f/internal/util/gfx"
	"fisherevans.com/project/f/internal/util/interp"
	"fisherevans.com/project/f/internal/util/keyframe"
	"fisherevans.com/project/f/internal/util/pixelutil"
	"fisherevans.com/project/f/internal/util/textbox"
	"fisherevans.com/project/f/internal/util/textbox/tbcfg"
	"github.com/gopxl/pixel/v2"
)

var (
	rewardsFrame  *frames.Instance
	rewardTextFF  *textbox.Instance
	rewardText36  *textbox.Instance
	rewardTextAdd *textbox.Instance
	rewardArrow   pixelutil.BoundedDrawable
)

func init() {
	resources.RunOnceInitialized(func() {
		rewardsFrame = frames.New("combat/rewards/frame", atlas)
		rewardTextFF = textbox.NewInstance(atlas.GetFont(resources.FontNameFF), tbcfg.NewConfig(rewardWidth, 9,
			tbcfg.Foreground(colors.White.RGBA),
			tbcfg.HAligned(tbcfg.HAlignCenter),
			tbcfg.VAligned(tbcfg.VAlignTop)))
		rewardText36 = textbox.NewInstance(atlas.GetFont(resources.FontNameM3x6), tbcfg.NewConfig(rewardWidth, 9,
			tbcfg.Foreground(colors.White.RGBA),
			tbcfg.HAligned(tbcfg.HAlignCenter),
			tbcfg.VAligned(tbcfg.VAlignTop)))
		rewardTextAdd = textbox.NewInstance(atlas.GetFont(resources.FontNameAddStandard), tbcfg.NewConfig(rewardWidth, 9,
			tbcfg.Foreground(colors.White.RGBA),
			tbcfg.HAligned(tbcfg.HAlignCenter),
			tbcfg.VAligned(tbcfg.VAlignTop)))
		rewardArrow = atlas.GetTilesheetSprite("common/arrows_5px", 2, 1)
	})
}

type rewardModal struct {
	from      int
	awarded   int
	nextAward int

	keyGroup   *keyframe.Group
	keyPlus    *keyframe.Member
	keyTotal   *keyframe.Member
	keyMessage *keyframe.Member

	textName    *textbox.Content
	textTitle   *textbox.Content
	textFrom    *textbox.Content
	textPlus    *textbox.Content
	textTotal   *textbox.Content
	textMessage *textbox.Content

	rewardNextBadge *badges.ButtonAction

	colorDark, color, colorLight pixel.RGBA
	sprite                       pixelutil.BoundedDrawable
}

func newRewardExperience(from, awarded, nextAward int) *rewardModal {
	return newRewardModal(
		atlas.GetSprite("animech/icon_animech"),
		"Animech", "Experience", "Level up",
		from, awarded, nextAward,
		blueDark, blue, blueLight)
}

func newRewardResearch(primortal rpg.PrimortalType, from, awarded, nextAward int) *rewardModal {
	return newRewardModal(
		anim.LoadTilesheetAnimation(atlas, "primortals/"+string(primortal), "default").Sprite(),
		primortal.Primortal().Name, "Research Points", "New skill",
		from, awarded, nextAward,
		purpleDark, purple, purpleLight)
}

func newRewardModal(sprite pixelutil.BoundedDrawable, name, title, message string, from, awarded, nextAward int, dark, color, light pixel.RGBA) *rewardModal {
	kg := keyframe.NewGroup()
	if from+awarded >= nextAward {
		message += " available!"
	} else {
		message += fmt.Sprintf(" at: %d", nextAward)
	}
	return &rewardModal{
		from:       from,
		awarded:    awarded,
		nextAward:  nextAward,
		keyGroup:   kg,
		keyPlus:    kg.AddKeyFrameFn(0.5, 3.0, interp.Smootherstep),
		keyTotal:   kg.AddKeyFrame(2.0, 3.0),
		keyMessage: kg.AddKeyFrame(3.0, 4.0),

		sprite: sprite,

		textName:    rewardText36.NewSimpleContent(name),
		textTitle:   rewardText36.NewSimpleContent(title),
		textFrom:    rewardTextFF.NewSimpleContent(fmt.Sprintf("%d", from)),
		textPlus:    rewardTextAdd.NewSimpleContent(fmt.Sprintf("+%d", awarded)),
		textTotal:   rewardTextFF.NewSimpleContent(fmt.Sprintf("%d", from+awarded)),
		textMessage: rewardTextFF.NewSimpleContent(message),

		rewardNextBadge: badges.Using(atlas).ButtonAction("a", "Next", badges.ButtonStyleStandard),

		colorDark:  dark,
		color:      color,
		colorLight: light,
	}
}

var (
	rewardWidth    = 84
	rewardBarWidth = 70
	rewardHeight   = 45

	greyDark  = colors.FromString("#636363")
	greyLight = colors.FromString("#c7cfcc")

	blueDark  = colors.FromString("#0f96b9")
	blue      = colors.FromString("#00caff")
	blueLight = colors.FromString("#85e6ff")

	purpleDark  = colors.FromString("#630fb8")
	purple      = colors.FromString("#ae61fa")
	purpleLight = colors.FromString("#cb97ff")
)

func (r *rewardModal) Render(target pixel.Target, topMiddle pixel.Matrix, renderButton bool, timeDelta float64) {
	r.keyGroup.Update(timeDelta)
	if !renderButton {
		r.keyGroup.Skip()
	}

	combatStatFrame.Draw(target, gfx.R(52, 65), topMiddle, frames.WithRenderOrigin(gfx.TopCenter))

	// sprite, actually center (shhhh)
	topMiddle = topMiddle.Moved(gfx.IVec(0, -24-2))
	r.sprite.Draw(target, topMiddle)

	// name
	topMiddle = topMiddle.Moved(gfx.IVec(0, -24))
	r.textName.Render(target, topMiddle, tbcfg.RenderFrom(gfx.TopCenter), tbcfg.RenderFrom(gfx.TopCenter), tbcfg.ColorMask(colors.White.RGBA))

	// stat frame
	topMiddle = topMiddle.Moved(gfx.IVec(0, -12))
	rewardsFrame.Draw(target, gfx.R(rewardWidth, rewardHeight), topMiddle, frames.WithRenderOrigin(gfx.TopCenter))

	// title
	topMiddle = topMiddle.Moved(gfx.IVec(0, -3))
	r.textTitle.Render(target, topMiddle, tbcfg.RenderFrom(gfx.TopCenter), tbcfg.ColorMask(greyLight))

	// math equation
	topMiddle = topMiddle.Moved(gfx.IVec(0, -16))
	padding := 4
	totalWidth := r.textFrom.Width() + padding + r.textPlus.Width() + padding + int(rewardArrow.Bounds().W()) + padding + r.textTotal.Width()
	topLeft := topMiddle.Moved(gfx.IVec(-totalWidth/2, 0))
	r.textFrom.Render(target, topLeft, tbcfg.RenderFrom(gfx.LeftCenter), tbcfg.ColorMask(r.colorDark), tbcfg.HAligned(tbcfg.HAlignLeft), tbcfg.VAligned(tbcfg.VAlignMiddle))
	topLeft = topLeft.Moved(gfx.IVec(r.textFrom.Width()+padding, 0))
	r.textPlus.Render(target, topLeft, tbcfg.RenderFrom(gfx.LeftCenter), tbcfg.ColorMask(r.keyPlus.Alpha(colors.White.RGBA)), tbcfg.HAligned(tbcfg.HAlignLeft), tbcfg.VAligned(tbcfg.VAlignMiddle))
	topLeft = topLeft.Moved(gfx.IVec(r.textPlus.Width()+padding, 0))
	rewardArrow.DrawColorMask(target, topLeft.Moved(gfx.LeftCenter.Align(rewardArrow).Add(gfx.IVec(0, 1))), r.keyTotal.Alpha(greyLight))
	topLeft = topLeft.Moved(gfx.IVec(int(rewardArrow.Bounds().W())+padding, 0))
	r.textTotal.Render(target, topLeft, tbcfg.RenderFrom(gfx.LeftCenter), tbcfg.ColorMask(r.keyTotal.Alpha(r.color)), tbcfg.HAligned(tbcfg.HAlignLeft), tbcfg.VAligned(tbcfg.VAlignMiddle))

	// bar
	topMiddle = topMiddle.Moved(gfx.IVec(0, -7))
	topLeft = topMiddle.Moved(gfx.IVec(-rewardBarWidth/2, 0))
	gfx.DrawRect(atlas, target, topLeft, gfx.TopLeft, rewardBarWidth, 4, greyDark)
	var fromBarWidth, awardedBarWidth, markerLocation int
	total := r.from + r.awarded
	if total > r.nextAward {
		fromBarWidth = int(float64(rewardBarWidth) * float64(r.from) / float64(total))
		awardedBarWidth = rewardBarWidth - fromBarWidth
		markerLocation = int(float64(rewardBarWidth) * float64(r.nextAward) / float64(total))
	} else {
		fromBarWidth = int(float64(rewardBarWidth) * float64(r.from) / float64(r.nextAward))
		awardedBarWidth = int(float64(rewardBarWidth) * float64(r.awarded) / float64(r.nextAward))
	}
	awardedBarWidth = int(float64(awardedBarWidth) * r.keyPlus.Progress())
	gfx.DrawRect(atlas, target, topLeft, gfx.TopLeft, fromBarWidth, 4, r.colorDark)
	gfx.DrawRect(atlas, target, topLeft.Moved(gfx.IVec(fromBarWidth, 0)), gfx.TopLeft, awardedBarWidth, 4, r.color)
	if markerLocation > 0 {
		gfx.DrawRect(atlas, target, topLeft.Moved(gfx.IVec(markerLocation, 0)), gfx.TopLeft, 1, 4, r.keyPlus.Alpha(colors.White.RGBA))
	}

	// message
	topMiddle = topMiddle.Moved(gfx.IVec(0, -8))
	messageMask := greyDark
	if total > r.nextAward {
		messageMask = colors.Lerp(r.colorLight, colors.White.RGBA, game.Utils().TimeCycleSin(0.5))
	}
	r.textMessage.Render(target, topMiddle, tbcfg.RenderFrom(gfx.TopCenter), tbcfg.ColorMask(r.keyMessage.Alpha(messageMask)))

	if !renderButton {
		return
	}
	topMiddle = topMiddle.Moved(gfx.IVec(0, -13))
	r.rewardNextBadge.RenderWithColorMask(target, topMiddle, gfx.TopCenter, colors.Alpha(r.keyMessage.Progress()))
}
