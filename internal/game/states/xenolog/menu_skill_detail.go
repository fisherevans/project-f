package xenolog

import (
	"fmt"
	"math"
	"strings"

	"github.com/gopxl/pixel/v2"

	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/game/input"
	"fisherevans.com/project/f/internal/game/rpg"
	"fisherevans.com/project/f/internal/game/states/combat/tick_bar"
	"fisherevans.com/project/f/internal/resources"
	"fisherevans.com/project/f/internal/util/colors"
	"fisherevans.com/project/f/internal/util/frames"
	"fisherevans.com/project/f/internal/util/gfx"
	"fisherevans.com/project/f/internal/util/pixelutil"
	"fisherevans.com/project/f/internal/util/sprites"
	"fisherevans.com/project/f/internal/util/textbox"
	"fisherevans.com/project/f/internal/util/textbox/tbcfg"
)

type tickStanceState string

const (
	tickStanceStateWithin   = "In"
	tickStanceStateEntering = "Enters"
	tickStanceStateExiting  = "Exits"
)

var stances = []rpg.CombatStance{
	rpg.TickStanceDefending,
	rpg.TickStanceReflecting,
	rpg.TickStanceVulnerable,
	rpg.TickStanceExposed,
}

var statuses = []rpg.StatusType{
	rpg.StatusBurning,
	rpg.StatusIonized,
	rpg.StatusMending,
	rpg.StatusWarded,
	rpg.StatusPoisoned,
}

var statusLevels = []rpg.StatusLevel{
	rpg.StatusLevel1,
	rpg.StatusLevel2,
	rpg.StatusLevel3,
}
var (
	noneTickSprite          = atlas.GetSprite("2x2")
	tickFrameW              = 140
	tickFrameH              = 68
	smallTextboxSkillDetail = textbox.NewInstance(atlas.GetFont(resources.FontNameFF),
		tbcfg.NewConfig(tickFrameW-4, 9,
			tbcfg.WithExpandMode(tbcfg.ExpandFit),
			tbcfg.HAligned(tbcfg.HAlignLeft),
			tbcfg.VAligned(tbcfg.VAlignMiddle),
			tbcfg.Foreground(colors.XenoLogText.RGBA),
			tbcfg.RenderFrom(gfx.LeftCenter)))
	stanceIcons = sprites.StanceIcons(atlas)
	statusIcons = sprites.StatusIcons(atlas)

	tickBarRenderer = tick_bar.NewRenderer(atlas)
)

type skillDetailMenu struct {
	screen    *Screen
	skillId   rpg.SkillId
	selection int
}

func newSkillDetailMenu(screen *Screen, skillId rpg.SkillId) *skillDetailMenu {
	return &skillDetailMenu{
		screen:  screen,
		skillId: skillId,
	}
}

func (*skillDetailMenu) Enter() {}

func (v *skillDetailMenu) OnTick(target pixel.Target, timeDelta float64) {
	v.handleInput()
	skill := v.skillId.Get()
	titleTxt := newTextRenderer(target, titleTextbox)
	smallTxt := newTextRenderer(target, smallTextboxSkillDetail).withOpts(tbcfg.RenderFrom(gfx.LeftCenter), tbcfg.HAligned(tbcfg.HAlignLeft))

	tickSpacing := 16
	tbOpt := tick_bar.NewDrawOptions(8, tickSpacing)
	tickBarHeight := tickBarRenderer.HeightOf(skill, tbOpt)
	tickBarX := 25
	tickBarY := screenHeight - (screenHeight-tickBarHeight)/2
	tickCenters := tickBarRenderer.Draw(
		v.screen.spriteFilterBuffer.Target(),
		pixel.IM.Moved(gfx.IVec(tickBarX, tickBarY)),
		&skill,
		tbOpt).TickDotCenterMatrices

	arrowLeft6px.DrawColorMask(target, tickCenters[v.selection].Moved(gfx.IVec(10, 0)), flashingHighlight())
	for i := 0; i < len(tickCenters)-1; i++ {
		leftCenter := tickCenters[i].Moved(gfx.IVec(8, -tickSpacing/2))
		noneTickSprite.DrawColorMask(target, leftCenter, colors.XenoLogDark.RGBA)
		noneTickSprite.DrawColorMask(target, leftCenter.Moved(gfx.IVec(2, 0)), colors.XenoLogDark.RGBA)
	}

	x := 51
	y := screenHeight - 14

	smallTxt.render("skill details", x, y, colors.XenoLogDark.RGBA, tbcfg.RenderFrom(gfx.LeftCenter))
	y -= 9
	titleTxt.render(skill.Name, x, y, colors.XenoLogHighlight.RGBA, tbcfg.RenderFrom(gfx.LeftCenter))
	y -= 14
	_, h := smallTxt.render(skill.Description, x, y, colors.XenoLogText.RGBA, tbcfg.RenderFrom(gfx.LeftCenter))

	y -= h + 4

	tickFrameR := pixel.R(0, 0, float64(tickFrameW), float64(tickFrameH))
	tickFrameTopLeft := pixel.IM.Moved(gfx.IVec(x-4, y))
	frame2px.Draw(target, tickFrameR, tickFrameTopLeft, frames.WithColor(colors.XenoLogDark.RGBA), frames.WithRenderOrigin(gfx.TopLeft))
	frame2pxBorder.Draw(target, tickFrameR, tickFrameTopLeft, frames.WithColor(colors.XenoLogText.RGBA), frames.WithRenderOrigin(gfx.TopLeft))

	y -= 7
	printTickDetail := func(pl printLine) {
		v.renderPrintLine(pl, target, smallTxt, x, &y)
	}

	printTickDetail(newPrintLine(stringContent(fmt.Sprintf("{+u,+c:xenolog_highlight}Tick %d{-u}:", v.selection+1))))

	effects := newEffectDetails()

	tick := skill.Ticks[v.selection]
	if tick.StanceType != rpg.TickStanceNone {
		stanceState := tickStanceStateWithin
		if v.selection == 0 || skill.Ticks[v.selection-1].StanceType != tick.StanceType {
			stanceState = tickStanceStateEntering
		} else if v.selection == len(skill.Ticks)-1 || skill.Ticks[v.selection+1].StanceType != tick.StanceType {
			stanceState = tickStanceStateExiting
		}
		effects.addStance(tick.StanceType)
		effects.addGeneralLine(newPrintLine(stringContent(stanceState + " {+u}" + tick.StanceType.String() + "{-u} stance")).withMargin())
	}

	for _, e := range tick.Effects {
		damageSummary(e.Damage, e.Self, effects)
		damageStatusSummary(e.Status, e.Self, effects)
	}

	preEffectsY := y

	for _, pl := range effects.generalLines {
		printTickDetail(pl)
	}

	for _, pl := range effects.targetDamageLines {
		printTickDetail(pl)
	}

	for _, pl := range effects.selfDamageLines {
		pl.contents = append(pl.contents, stringContent("to {+c:xenolog_highlight}self"))
		printTickDetail(pl)
	}

	for _, stance := range stances {
		if effects.stances[stance] {
			c := []lineContents{
				stringContent(fmt.Sprintf("{+c:xenolog_highlight}%s", stance.String())),
				spriteContent(stanceIcons[stance]).mask(colors.XenoLogHighlight.RGBA),
				stringContent("{+c:xenolog_clear}(stance)"),
			}
			printTickDetail(newPrintLine(c...))
			printTickDetail(newPrintLine(stringContent(stance.Description())).withMargin())
		}
	}

	for _, status := range statuses {
		if effects.statuses[status] {
			sprite := spriteContent(statusIcons[status]).mask(colors.XenoLogHighlight.RGBA)
			if status == rpg.StatusMending {
				sprite = sprite.moved(0, -1)
			}
			c := []lineContents{
				stringContent(fmt.Sprintf("{+c:xenolog_highlight}%s", status.PastTense())),
				sprite,
				stringContent("{+c:xenolog_clear}(status)"),
			}
			printTickDetail(newPrintLine(c...))
			printTickDetail(newPrintLine(stringContent(status.Description())).withMargin())
		}
	}

	if y == preEffectsY {
		printTickDetail(newPrintLine(stringContent("No effects")))
	}

}

func (v *skillDetailMenu) renderPrintLine(pl printLine, target pixel.Target, smallTxt *textRenderer, x int, yPtr *int) {
	thisX := x
	for _, c := range pl.contents {
		mask := colors.XenoLogText.RGBA
		if c.color != nil {
			mask = *c.color
		}
		if c.text != "" {
			dx, _ := smallTxt.render(c.text, thisX+c.dx, *yPtr+c.dy, mask)
			thisX += dx + 3
		}
		if c.image != nil {
			imgDx := c.image.Bounds().W() / 2
			imgDy := int((c.image.Bounds().H() - 5) / 2)
			c.image.DrawColorMask(v.screen.spriteFilterBuffer.Target(), pixel.IM.Moved(gfx.IVec(thisX+int(imgDx)+c.dx, *yPtr+imgDy+c.dy)), mask)
			thisX += int(c.image.Bounds().W()) + 2
		}
	}

	*yPtr -= 10 + pl.margin
}

func (v *skillDetailMenu) handleInput() {
	controls := game.Controls[*State]()
	if controls.ButtonB().JustPressed() || controls.ButtonSelect().JustPressed() {
		v.screen.PopMenu()
	}
	if controls.DPad().JustPressed() {
		delta := 0
		switch controls.DPad().JustPressedDirection() {
		case input.Up:
			delta = -1
		case input.Down:
			delta = 1
		}
		v.selection += delta
		if v.selection < 0 {
			v.selection = 0
		} else if v.selection >= len(v.skillId.Get().Ticks) {
			v.selection = len(v.skillId.Get().Ticks) - 1
		}
	}
}

type printLine struct {
	contents []lineContents
	margin   int
}

func newPrintLine(contents ...lineContents) printLine {
	return printLine{contents: contents}
}

func (pl printLine) withMargin() printLine {
	pl.margin = 3
	return pl
}

type lineContents struct {
	text   string
	image  pixelutil.BoundedDrawable
	dx, dy int
	color  *pixel.RGBA
}

func stringContent(text string) lineContents {
	return lineContents{text: text}
}

func spriteContent(sprite pixelutil.BoundedDrawable) lineContents {
	return lineContents{image: sprite}
}

func (v lineContents) mask(color pixel.RGBA) lineContents {
	v.color = &color
	return v
}

func (v lineContents) moved(dx, dy int) lineContents {
	v.dx, v.dy = dx, dy
	return v
}

func line(contents ...lineContents) []lineContents {
	return contents
}

type effectDetails struct {
	generalLines      []printLine
	targetDamageLines []printLine
	selfDamageLines   []printLine
	statuses          map[rpg.StatusType]bool
	stances           map[rpg.CombatStance]bool
}

func newEffectDetails() *effectDetails {
	return &effectDetails{
		statuses: map[rpg.StatusType]bool{},
		stances:  map[rpg.CombatStance]bool{},
	}
}

func (e *effectDetails) addGeneralLine(pl printLine) {
	e.generalLines = append(e.generalLines, pl)
}

func (e *effectDetails) addDamageLine(pl printLine, toSelf bool) {
	if toSelf {
		e.selfDamageLines = append(e.selfDamageLines, pl)
	} else {
		e.targetDamageLines = append(e.targetDamageLines, pl)
	}
}

func (e *effectDetails) addStatus(status rpg.StatusType) {
	e.statuses[status] = true
}

func (e *effectDetails) addStance(stance rpg.CombatStance) {
	e.stances[stance] = true
}

func damageSummary(damage *rpg.SkillTickDamage, toSelf bool, e *effectDetails) {
	if damage == nil {
		return
	}
	var amountString string
	if damage.RandomVariance == 0 {
		amountString = fmt.Sprintf("%d", damage.Amount)
	} else {
		low := damage.Amount - damage.RandomVariance/2
		high := damage.Amount + damage.RandomVariance/2
		amountString = fmt.Sprintf("%d-%d", low, high)
	}
	e.addDamageLine(newPrintLine(stringContent("Deals "+amountString+" damage")), toSelf)

	if damage.MissRate > 0 {
		e.addDamageLine(newPrintLine(stringContent(fmt.Sprintf("Has a %.f%% chance to miss", damage.MissRate))), toSelf)
	}

	damageScaleByEffects(damage.ScaledBy.TargetStatus, "the target is", toSelf, e)
	damageScaleByEffects(damage.ScaledBy.SourceStatus, "you are", toSelf, e)
}

func damageScaleByEffects(scalers map[rpg.StatusType]map[rpg.StatusLevel]float64, theyIs string, toSelf bool, e *effectDetails) {
	for _, status := range statuses {
		var levelStrings []string
		levelScales, ok := scalers[status]
		if !ok {
			continue
		}
		for _, level := range statusLevels {
			scale, ok := levelScales[level]
			if !ok {
				continue
			}
			e.addStatus(status)
			levelStrings = append(levelStrings, "{+c:xenolog_highlight}"+FormatFloat(scale)+"x{-c}")
		}
		if len(levelStrings) > 0 {
			e.addDamageLine(newPrintLine(stringContent(fmt.Sprintf("If %s {+u}%s{-u}:",
				theyIs,
				status.CurrentTense(),
			))), false) // always false, since it's talking about modifying the last line
			e.addDamageLine(newPrintLine(
				spriteContent(arrowRight).mask(colors.XenoLogClear.RGBA).moved(0, 1),
				stringContent(fmt.Sprintf("Deal %s more damage",
					strings.Join(levelStrings, "/"),
				)),
			), false) // always false, since it's talking about modifying the last line
		}
	}
}

func damageStatusSummary(status *rpg.SkillTickStatus, toSelf bool, e *effectDetails) {
	if status == nil {
		return
	}
	var suffix string
	if status.RequireExistingStacks == rpg.SkillTickStatusRequireExistingStacks {
		suffix = " if stacks exist"
	} else if status.RequireExistingStacks == rpg.SkillTickStatusRequireNoStacks {
		suffix = " if not stacks are present"
	}
	e.addStatus(status.Status)
	e.addDamageLine(newPrintLine(stringContent(fmt.Sprintf("Applies %s stacks of {+u}%s{-u}%s",
		FormatFloat(status.Stacks), status.Status.PastTense(), suffix))).withMargin(), toSelf)
}

func FormatFloat(f float64) string {
	// Round to 2 decimals first to avoid tiny float noise
	f = math.Round(f*100) / 100

	// If it's basically an integer (within 0.05), drop decimals
	if math.Abs(f-math.Round(f)) < 0.05 {
		return fmt.Sprintf("%.0f", math.Round(f))
	}

	// If it has only one meaningful decimal, show one
	if math.Abs(f*10-math.Round(f*10)) < 0.05 {
		return fmt.Sprintf("%.1f", math.Round(f*10)/10)
	}

	// Otherwise show two
	return fmt.Sprintf("%.2f", f)
}
