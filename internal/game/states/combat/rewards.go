package combat

import (
	"fmt"
	"strings"

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
	"github.com/rs/zerolog/log"
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

	exiting bool

	enterKeyGroup   *keyframe.Group
	keyBottomAppear *keyframe.Member
	keyTopAppear    *keyframe.Member
	keyPlus         *keyframe.Member
	keyTotal        *keyframe.Member
	keyButton       *keyframe.Member

	exitKeyGroup  *keyframe.Group
	keyTopExit    *keyframe.Member
	keyBottomExit *keyframe.Member

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

func newRewardModals(reward game.CombatReward) []*rewardModal {
	var modals []*rewardModal
	if reward.ExperiencePoints > 0 {
		nextLevel := 1
		if u := game.CurrentSave().Animech.Upgrades; u != nil {
			nextLevel = u.GetLevel() + 1
		}
		required := rpg.AnimechUpgradeExperienceRequiredToUpgrade(nextLevel)
		modals = append(modals, newRewardExperience(game.CurrentSave().Animech.AnimechExperience, reward.ExperiencePoints, required))
	}
	if reward.ExperiencePoints > 0 {
		current := 0
		if prog, ok := game.CurrentSave().Primortals[reward.ResearchType]; ok {
			current = prog.ResearchPoints
		}
		required := -1
		primortal, ok := rpg.Primortals[reward.ResearchType]
		if !ok {
			log.Fatal().Str("type", string(reward.ResearchType)).Msg("primortal not found")
		}
		for skillId, skillReqs := range primortal.UnlockableSkills {
			if game.CurrentSave().IsSkillUnlocked(skillId) {
				continue
			}
			available := true
			for _, prereq := range skillReqs.Prerequisites {
				if !game.CurrentSave().IsSkillUnlocked(prereq) {
					available = false
					break
				}
			}
			if available && required < skillReqs.Cost {
				required = skillReqs.Cost
			}
		}
		modals = append(modals, newRewardResearch(reward.ResearchType, current, reward.ResearchPoints, required))
	}
	return modals
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
	enterKg := keyframe.NewGroup(interp.Smoothstep)
	exitKg := keyframe.NewGroup(interp.Smoothstep)
	if nextAward < 0 {
		message = "No " + strings.ToLower(message) + " available"
	} else if from+awarded >= nextAward {
		message += " available!"
	} else {
		message += fmt.Sprintf(" at: %d", nextAward)
	}
	return &rewardModal{
		from:      from,
		awarded:   awarded,
		nextAward: nextAward,

		enterKeyGroup:   enterKg,
		keyBottomAppear: enterKg.AddKeyFrame(0.0, 0.5),
		keyTopAppear:    enterKg.AddKeyFrame(0.25, 0.75),
		keyPlus:         enterKg.AddKeyFrame(0.25, 2.25),
		keyTotal:        enterKg.AddKeyFrame(1.5, 2.5),
		keyButton:       enterKg.AddKeyFrame(2.5, 3.0),

		exitKeyGroup:  exitKg,
		keyTopExit:    exitKg.AddKeyFrame(0, 0.25),
		keyBottomExit: exitKg.AddKeyFrame(0.5, 0.75),

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

func (r *rewardModal) Exit() {
	r.exiting = true
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
	if r.exiting {
		r.enterKeyGroup.Skip()
		r.exitKeyGroup.Update(timeDelta)
	} else {
		if renderButton {
			r.enterKeyGroup.Update(timeDelta)
		} else {
			r.enterKeyGroup.Skip()
		}
	}

	topDelta := pixel.V(0, -10*(r.keyTopAppear.InverseProgress()+r.keyTopExit.Progress()))
	topMask := colors.Alpha(r.keyTopAppear.Progress() * r.keyTopExit.InverseProgress())
	bottomDelta := pixel.V(0, 10*(r.keyBottomAppear.InverseProgress()+r.keyBottomExit.Progress()))
	bottomMask := colors.Alpha(r.keyBottomAppear.Progress() * r.keyBottomExit.InverseProgress())

	// TOP

	// logo frame
	combatStatFrame.Draw(target, gfx.R(52, 65), topMiddle.Moved(topDelta), frames.WithRenderOrigin(gfx.TopCenter), frames.WithColor(topMask))

	// sprite, actually center (shhhh)
	topMiddle = topMiddle.Moved(gfx.IVec(0, -24-2))
	r.sprite.DrawColorMask(target, topMiddle.Moved(topDelta), topMask)

	// name
	topMiddle = topMiddle.Moved(gfx.IVec(0, -24))
	r.textName.Render(target, topMiddle.Moved(topDelta), tbcfg.RenderFrom(gfx.TopCenter), tbcfg.RenderFrom(gfx.TopCenter), tbcfg.ColorMask(topMask))

	// BOTTOM

	topMiddle = topMiddle.Moved(bottomDelta)

	// stat frame
	topMiddle = topMiddle.Moved(gfx.IVec(0, -12))
	rewardsFrame.Draw(target, gfx.R(rewardWidth, rewardHeight), topMiddle, frames.WithRenderOrigin(gfx.TopCenter), frames.WithColor(bottomMask))

	// title
	topMiddle = topMiddle.Moved(gfx.IVec(0, -3))
	r.textTitle.Render(target, topMiddle, tbcfg.RenderFrom(gfx.TopCenter), tbcfg.ColorMask(greyLight.Mul(bottomMask)))

	// math equation
	topMiddle = topMiddle.Moved(gfx.IVec(0, -16))
	padding := 4
	totalWidth := r.textFrom.Width() + padding + r.textPlus.Width() + padding + int(rewardArrow.Bounds().W()) + padding + r.textTotal.Width()
	topLeft := topMiddle.Moved(gfx.IVec(-totalWidth/2, 0))
	r.textFrom.Render(target, topLeft, tbcfg.RenderFrom(gfx.LeftCenter), tbcfg.ColorMask(r.colorDark.Mul(bottomMask)), tbcfg.HAligned(tbcfg.HAlignLeft), tbcfg.VAligned(tbcfg.VAlignMiddle))
	topLeft = topLeft.Moved(gfx.IVec(r.textFrom.Width()+padding, 0))
	r.textPlus.Render(target, topLeft, tbcfg.RenderFrom(gfx.LeftCenter), tbcfg.ColorMask(r.keyPlus.Alpha(colors.White.RGBA).Mul(bottomMask)), tbcfg.HAligned(tbcfg.HAlignLeft), tbcfg.VAligned(tbcfg.VAlignMiddle))
	topLeft = topLeft.Moved(gfx.IVec(r.textPlus.Width()+padding, 0))
	rewardArrow.DrawColorMask(target, topLeft.Moved(gfx.LeftCenter.Align(rewardArrow).Add(gfx.IVec(0, 1))), r.keyTotal.Alpha(greyLight).Mul(bottomMask))
	topLeft = topLeft.Moved(gfx.IVec(int(rewardArrow.Bounds().W())+padding, 0))
	r.textTotal.Render(target, topLeft, tbcfg.RenderFrom(gfx.LeftCenter), tbcfg.ColorMask(r.keyTotal.Alpha(r.color).Mul(bottomMask)), tbcfg.HAligned(tbcfg.HAlignLeft), tbcfg.VAligned(tbcfg.VAlignMiddle))

	// bar
	topMiddle = topMiddle.Moved(gfx.IVec(0, -7))
	topLeft = topMiddle.Moved(gfx.IVec(-rewardBarWidth/2, 0))
	gfx.DrawRect(atlas, target, topLeft, gfx.TopLeft, rewardBarWidth, 4, greyDark.Mul(bottomMask))
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
	gfx.DrawRect(atlas, target, topLeft, gfx.TopLeft, fromBarWidth, 4, r.colorDark.Mul(bottomMask))
	gfx.DrawRect(atlas, target, topLeft.Moved(gfx.IVec(fromBarWidth, 0)), gfx.TopLeft, awardedBarWidth, 4, r.color.Mul(bottomMask))
	if markerLocation > 0 {
		gfx.DrawRect(atlas, target, topLeft.Moved(gfx.IVec(markerLocation, 0)), gfx.TopLeft, 1, 4, r.keyPlus.Alpha(colors.White.RGBA).Mul(bottomMask))
	}

	// message
	topMiddle = topMiddle.Moved(gfx.IVec(0, -8))
	messageMask := greyDark
	if r.nextAward >= 0 && total > r.nextAward {
		messageMask = colors.Lerp(r.colorLight, colors.White.RGBA, game.Utils().TimeCycleSin(0.5))
	}
	r.textMessage.Render(target, topMiddle, tbcfg.RenderFrom(gfx.TopCenter), tbcfg.ColorMask(r.keyTotal.Alpha(messageMask).Mul(bottomMask)))

	if !renderButton {
		return
	}
	topMiddle = topMiddle.Moved(gfx.IVec(0, -13))
	r.rewardNextBadge.RenderWithColorMask(target, topMiddle, gfx.TopCenter, colors.Alpha(r.keyButton.Progress()).Mul(bottomMask))
}
