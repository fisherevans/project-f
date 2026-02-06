package combat

import (
	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/game/anim"
	"fisherevans.com/project/f/internal/game/rpg"
	"fisherevans.com/project/f/internal/util"
	"fisherevans.com/project/f/internal/util/interp"
	"fisherevans.com/project/f/internal/util/pixelutil"
	"github.com/gopxl/pixel/v2"

	"fisherevans.com/project/f/internal/util/colors"
	"fisherevans.com/project/f/internal/util/gfx"
	"fisherevans.com/project/f/internal/util/textbox"
	"fisherevans.com/project/f/internal/util/textbox/tbcfg"
)

const MaxTempo = 10.0
const tempoLevels = 3.0

type Tempo struct {
	current float64
}

func (t *Tempo) GetCurrent() float64 {
	return t.current
}

func (t *Tempo) GetLevel() rpg.TempoLevel {
	if t == nil {
		return rpg.TempoLevel0
	}
	return rpg.TempoLevel(t.current / (MaxTempo / tempoLevels))
}

func (t *Tempo) Increment(n float64) {
	t.current = min(MaxTempo, t.current+n)
}

func (t *Tempo) Reset() {
	t.current = 0
}

func (t *Tempo) Decrease(n float64) {
	t.current = max(0, t.current-n)
}

func newComboText(fontName string) *textbox.Instance {
	return textbox.NewInstance(
		atlas.GetFont(fontName),
		tbcfg.NewConfig(0, 0,
			tbcfg.RenderFrom(gfx.TopLeft),
			tbcfg.VAligned(tbcfg.VAlignTop),
		))
}

// todo make rates tunable
// todo - account for battle speed + time delta
func (t *Tempo) Update(upNext bool, combatant Combatant, timeDelta float64) {
	if t == nil {
		return
	}
	noSkillSelected := upNext && combatant.GetCurrentSkill() == nil && !combatant.IsNextSkillCommitted()
	isInterrupted := combatant.GetCurrentSkill() != nil && combatant.GetCurrentSkill().IsInterrupted()
	if noSkillSelected || isInterrupted {
		t.Decrease(timeDelta * 1.5)
		return
	}
	dur := 1.0
	if combatant.GetCurrentSkill() != nil { // can not have a skill selected, but not need it yet
		dur = float64(combatant.GetCurrentSkill().Duration())
	}
	tempoIncreaseRate := 1.0 / ((dur * 0.2) + 1.0)
	t.Increment(tempoIncreaseRate * timeDelta)
}

var (
	tempoLevel3Border *anim.AnimatedSprite
	tempoLevel2Border pixelutil.BoundedDrawable
	tempoLevel1Border pixelutil.BoundedDrawable
	tempoBase         pixelutil.BoundedDrawable
	tempoName         pixelutil.BoundedDrawable

	tempoBarGradient pixelutil.BoundedDrawable
	tempoBarTick     pixelutil.BoundedDrawable
)

func (t *Tempo) Render(target pixel.Target, center pixel.Matrix, visibility *util.Visibility, timeDelta float64) {
	if t == nil {
		return
	}
	game.DebugBLf("tempo: %.2f (%d)", t.current, t.GetLevel())

	// No alpha fading
	mask := colors.Alpha(1.0)

	topLeft := center //.Moved(tempoTopLeftDelta)

	tempoBase.Draw(target, topLeft)
	tempoName.DrawColorMask(target, topLeft, colors.FromString("#ff00aa").Mul(mask))
	level := t.GetLevel()
	switch level {
	case rpg.TempoLevel1:
		mask := colors.Alpha(interp.Lerp(0.1, 0.25, game.Utils().TimeCycleSin(0.33))).Mul(mask)
		tempoLevel1Border.DrawColorMask(target, topLeft, mask)
	case rpg.TempoLevel2:
		mask := colors.Alpha(interp.Lerp(0.25, 0.5, game.Utils().TimeCycleSin(0.66))).Mul(mask)
		tempoLevel2Border.DrawColorMask(target, topLeft, mask)
	case rpg.TempoLevel3:
		mask := colors.Alpha(0.5).Mul(mask)
		tempoLevel2Border.DrawColorMask(target, topLeft, mask)
		tempoLevel3Border.Update(timeDelta)
		tempoLevel3Border.Sprite().Draw(target, topLeft)
	}

	barMaxWidth := 33.0
	barWidth := float64(t.current) / float64(MaxTempo) * barMaxWidth
	barScale := (1.0 / tempoBarGradient.Bounds().W()) * barWidth
	barMatrix := pixel.IM.Moved(gfx.TopLeft.Align(tempoBarGradient)).ScaledXY(pixel.ZV, pixel.V(barScale, 1.0))
	barTopLeft := topLeft.Moved(gfx.IVec(-6, 2))
	tempoBarGradient.DrawColorMask(target, barMatrix.Chained(barTopLeft), colors.FromString("#ff6ace").Mul(mask))
	for tickNumber := 1; tickNumber < int(tempoLevels); tickNumber++ {
		tickTopLeft := barTopLeft.Moved(gfx.IVec(int(barMaxWidth/float64(tempoLevels)*float64(tickNumber))-1, 0))
		tickMatrix := pixel.IM.Moved(gfx.TopLeft.Align(tempoBarTick)).ScaledXY(pixel.ZV, pixel.V(1.0/tempoBarTick.Bounds().W(), 1.0))
		tempoBarTick.DrawColorMask(target, tickMatrix.Chained(tickTopLeft), colors.Alpha(0.5).Mul(mask))
	}
}
