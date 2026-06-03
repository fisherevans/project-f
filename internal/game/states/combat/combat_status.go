package combat

import (
	"fmt"
	"math"
	"strings"

	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/util"
	"fisherevans.com/project/f/internal/util/interp"
	"fisherevans.com/project/f/internal/util/rng"
	"github.com/gopxl/pixel/v2"
	"github.com/rs/zerolog/log"

	"fisherevans.com/project/f/internal/game/anim"
	"fisherevans.com/project/f/internal/game/rpg"
	"fisherevans.com/project/f/internal/util/colors"
	"fisherevans.com/project/f/internal/util/frames"
	"fisherevans.com/project/f/internal/util/gfx"
	"fisherevans.com/project/f/internal/util/pixelutil"
	"fisherevans.com/project/f/internal/util/sprites"
)

type CombatStatus struct {
	Status               rpg.StatusType
	Stacks               float64
	AccumulativeExposure int
	Apply                func(s *State, appliedTo Combatant, cs *CombatStatus)
	CustomTickDecay      func(s *State, cs *CombatStatus)
	Animation            *anim.AnimatedSprite

	timerAnimation       *anim.AnimatedSprite
	ticksSinceLastUpdate int

	renderLastLevel    rpg.StatusLevel
	renderLastProgress float64
	timeSinceLastApply float64
	timeInCurrentLevel float64
	renderEffects      []*changeAnimation
}

var changeAnimationAlphaKeys = interp.NewKeys().
	WithKey(0, 0).
	WithKey(0.25, 1).
	WithKey(0.75, 1).
	WithKey(1, 0)

type changeAnimation struct {
	animation     *anim.AnimatedSprite
	positionDelta pixel.Vec
	delay         float64
	currentAge    float64
	maxAge        float64
	color         pixel.RGBA
}

func (ca *changeAnimation) Render(target pixel.Target, boxBottomLeft pixel.Matrix, timeDelta float64) {
	ca.delay -= timeDelta
	if ca.delay > 0 || ca.IsComplete() {
		return
	}
	ca.currentAge += timeDelta
	ca.animation.Update(timeDelta)
	progress := ca.currentAge / ca.maxAge
	alpha := changeAnimationAlphaKeys.Interpolate(progress)
	mask := colors.WithAlpha(ca.color, alpha)
	ca.animation.Sprite().DrawColorMask(target, boxBottomLeft.Moved(ca.positionDelta), mask)
}

func (ca *changeAnimation) IsComplete() bool {
	return ca.currentAge >= ca.maxAge
}

func (s *CombatStatus) TotalProgressUntilNextUpdate(currentTickProgress float64) float64 {
	return max(0, min(1, float64(s.Stacks+currentTickProgress)/float64(ticksPerStatusApply)))
}

var statusLevelThresholds = []float64{0, 3, 7}
var statusLevels = []rpg.StatusLevel{rpg.StatusLevel1, rpg.StatusLevel2, rpg.StatusLevel3}

const (
	ticksPerStatusApply        = 4
	maxStacks                  = 15
	statusStackLevel3Threshold = 7.0
	statusStackLevel2Threshold = 3.0
)

func (s *CombatStatus) PreciseLevel(partialTickProgress float64) (rpg.StatusLevel, float64) {
	if partialTickProgress >= 1 || partialTickProgress < 0 {
		log.Error().Float64("partial", partialTickProgress).Msgf("invalid partial tick progress")
	}
	if s.Stacks == 0 {
		return rpg.StatusLevel0, partialTickProgress
	}
	if s.Stacks >= statusStackLevel3Threshold {
		return rpg.StatusLevel3, (s.Stacks - statusStackLevel3Threshold + partialTickProgress) / (maxStacks - statusStackLevel3Threshold)
	}
	if s.Stacks >= statusStackLevel2Threshold {
		return rpg.StatusLevel2, (s.Stacks - statusStackLevel2Threshold + partialTickProgress) / (statusStackLevel3Threshold - statusStackLevel2Threshold)
	}
	return rpg.StatusLevel1, (s.Stacks + partialTickProgress) / statusStackLevel2Threshold
}

func (s *CombatStatus) Level() rpg.StatusLevel {
	level, _ := s.PreciseLevel(0)
	return level
}

func (s *CombatStatus) generateChangeEffects(animName string, mask pixel.RGBA, areaW, areaH float64) {
	log.Info().Any("status", s.Status).Str("anim", animName).Msg("generating change effects")
	s.renderEffects = nil
	deltas := util.GenerateSpaced(0, 0, areaW, areaH, 6, rng.NewV1())
	for _, delta := range deltas {
		duration := util.RandBetween(statusFadeOut/2.0, statusFadeOut)
		s.renderEffects = append(s.renderEffects, &changeAnimation{
			animation:     anim.Load(atlas, "combat/status_change:"+animName).WithDuration(duration),
			positionDelta: delta,
			delay:         util.RandBetween(0, 0.5),
			maxAge:        duration,
			color:         mask,
		})
	}
}

type AppliedStatuses struct {
	Statuses []*CombatStatus
}

func NewAppliedStatuses() *AppliedStatuses {
	statusIcons := sprites.StatusIcons(atlas)
	out := &AppliedStatuses{
		Statuses: []*CombatStatus{
			{
				Status:    rpg.StatusWarded,
				Animation: anim.NewStaticAnimation(statusIcons[rpg.StatusWarded]),
				Apply: func(s *State, appliedTo Combatant, cs *CombatStatus) {
					// nothing, effects happen when receiving damage
				},
			},
			{
				Status:    rpg.StatusPoisoned,
				Animation: anim.NewStaticAnimation(statusIcons[rpg.StatusPoisoned]),
				Apply: func(s *State, appliedTo Combatant, cs *CombatStatus) {
					var pct float64
					switch cs.Level() {
					case rpg.StatusLevel3:
						pct = 0.07
					case rpg.StatusLevel2:
						pct = 0.05
					default:
						pct = 0.03
					}
					dmg := int(math.Max(float64(appliedTo.GetTotalMaxHealth())*pct, 1.0))
					s.AdjustHealth(-dmg, appliedTo, &DamageOptions{
						Status: rpg.StatusPoisoned,
					})
				},
				CustomTickDecay: func(s *State, cs *CombatStatus) {
					// decay slower the higher the level
					decayPerTick := func(l rpg.StatusLevel) float64 {
						switch l {
						case rpg.StatusLevel3:
							return 0.5
						case rpg.StatusLevel2:
							return 0.75
						default:
							return 1.0
						}
					}
					cs.Stacks -= decayPerTick(cs.Level()) / ticksPerStatusApply
				},
			},
			{
				Status:    rpg.StatusBurning,
				Animation: anim.NewStaticAnimation(statusIcons[rpg.StatusBurning]),
				Apply: func(s *State, appliedTo Combatant, cs *CombatStatus) {
					x := float64(cs.AccumulativeExposure)
					maxDamage := max(1, float64(appliedTo.GetTotalMaxHealth())*0.1) // limit to 10%
					damage := int(min(math.Ceil(x*0.1), maxDamage))
					log.Info().Int("exposure", cs.AccumulativeExposure).Int("damage", damage).Msg("combat/status_burning")
					s.AdjustHealth(-damage, appliedTo, &DamageOptions{
						Status: rpg.StatusBurning,
					})
				},
				CustomTickDecay: func(s *State, cs *CombatStatus) {
					// decay faster the higher the level
					decayPerTick := func(l rpg.StatusLevel) float64 {
						switch l {
						case rpg.StatusLevel3:
							return 2
						case rpg.StatusLevel2:
							return 1
						default:
							return 0.5
						}
					}
					cs.Stacks -= decayPerTick(cs.Level()) / ticksPerStatusApply
				},
			},
			{
				Status:    rpg.StatusIonized,
				Animation: anim.NewStaticAnimation(statusIcons[rpg.StatusIonized]),
				Apply: func(s *State, appliedTo Combatant, cs *CombatStatus) {
					// nothing, effects happen when applying damage
				},
				CustomTickDecay: func(s *State, cs *CombatStatus) {
					// very slow decay, in order to build up and utilize it
					decayPerTick := func(l rpg.StatusLevel) float64 {
						switch l {
						case rpg.StatusLevel3:
							return 0.2
						case rpg.StatusLevel2:
							return 0.4
						default:
							return 0.6
						}
					}
					cs.Stacks -= decayPerTick(cs.Level()) / ticksPerStatusApply
				},
			},
			{
				Status:    rpg.StatusMending,
				Animation: anim.NewStaticAnimation(statusIcons[rpg.StatusMending]),
				Apply: func(s *State, appliedTo Combatant, cs *CombatStatus) {
					healthPerTick := func(l rpg.StatusLevel) int {
						switch l {
						case rpg.StatusLevel3:
							return 3
						case rpg.StatusLevel2:
							return 2
						default:
							return 1
						}
					}
					s.AdjustHealth(healthPerTick(cs.Level()), appliedTo, &DamageOptions{
						Status: rpg.StatusMending,
					})
				},
			},
		},
	}
	for _, status := range out.Statuses {
		status.timeInCurrentLevel = math.MaxFloat64
		status.timerAnimation = anim.Load(atlas, "combat/status_timer:default")
	}
	return out
}

func (a *AppliedStatuses) Add(status rpg.StatusType, stacks float64) {
	for idx, s := range a.Statuses {
		if s.Status == status {
			if s.Stacks == 0 {
				// move this status to the end for rendering
				// this only works because we return immediately
				a.Statuses = append(append(a.Statuses[:idx], a.Statuses[idx+1:]...), s)
			}
			s.Stacks += stacks
			if s.Stacks > maxStacks {
				s.Stacks = maxStacks
			}
			return
		}
	}
}

func (a *AppliedStatuses) OnCombatTick(s *State, appliedTo Combatant) {
	for _, status := range a.Statuses {
		if status.Stacks <= 0 {
			status.AccumulativeExposure = 0
			status.ticksSinceLastUpdate = 0
			continue
		}
		status.ticksSinceLastUpdate++
		if status.ticksSinceLastUpdate >= ticksPerStatusApply {
			if status.Apply != nil {
				status.AccumulativeExposure += int(status.Level())
				status.Apply(s, appliedTo, status)
			}
			status.ticksSinceLastUpdate = 0
		}
		// decrease stacks
		if status.CustomTickDecay != nil {
			status.CustomTickDecay(nil, status)
		} else {
			status.Stacks -= 1.0 / ticksPerStatusApply
		}
		status.Stacks = max(0, status.Stacks)
	}
}

func (a *AppliedStatuses) HasStatus(status rpg.StatusType) bool {
	for _, s := range a.Statuses {
		if s.Status != status {
			continue
		}
		if s.Stacks > 0 {
			return true
		}
		return false
	}
	return false
}

func (a *AppliedStatuses) String() string {
	var parts []string
	for _, status := range a.Statuses {
		if status.Stacks == 0 {
			continue
		}
		level, progress := status.PreciseLevel(0)
		parts = append(parts, fmt.Sprintf("%s:%d+%.1f(%.1f)", status.Status, level, progress, status.Stacks))
	}
	return strings.Join(parts, "|")
}

var (
	statusFrame        *frames.Instance
	statusBorder       pixelutil.BoundedDrawable
	statusLevelSprites map[rpg.StatusLevel]pixelutil.BoundedDrawable
)

const statusFadeOut = 2.0
const statusSlideDur = 0.5
const statusApplyFadeDur = 1

func (a *AppliedStatuses) Render(originM pixel.Matrix, target pixel.Target, timeDelta float64, origin gfx.OriginLocation, tickProgress float64) {
	dx := 0.0
	xPadding := 2.0
	dxDirection := -1.0
	if origin == gfx.TopLeft {
		dxDirection = 1
	}
	for _, status := range a.Statuses {
		level, levelProgress := status.PreciseLevel(0)
		status.timeInCurrentLevel += timeDelta
		generateEffectDirection := 0
		if status.renderLastLevel != level {
			generateEffectDirection = int(level) - int(status.renderLastLevel)
			status.timeInCurrentLevel = 0
			status.renderLastLevel = level
		}
		alpha := 1.0
		color := colors.StatusColors[status.Status]
		status.Animation.Update(timeDelta)
		icon := status.Animation.Sprite()
		levelSprite := statusLevelSprites[level]
		margin := 2.0
		spacing := 1.0
		w := margin*2 + math.Max(icon.Bounds().W(), levelSprite.Bounds().W())
		h := margin*2 + icon.Bounds().H() + spacing + levelSprite.Bounds().H()
		if level == rpg.StatusLevel0 {
			if status.timeInCurrentLevel > statusFadeOut {
				moveDelta := float64(w + xPadding)
				moveProgress := 1.0 - max(0, min(1, (status.timeInCurrentLevel-statusFadeOut)/statusSlideDur))
				dx += moveProgress * moveDelta
				continue
			}
			alpha = 1.0 - (status.timeInCurrentLevel / statusFadeOut)
		}
		frameMaskLerp := 0.1
		switch level {
		case rpg.StatusLevel2:
			frameMaskLerp = 0.4
		case rpg.StatusLevel3:
			frameMaskLerp = 0.7
		}

		if generateEffectDirection > 0 {
			status.generateChangeEffects("up", colors.Lerp(color, colors.White.RGBA, 0.25), w, h)
		} else if generateEffectDirection < 0 {
			status.generateChangeEffects("down", colors.Lerp(color, colors.Black.RGBA, 0.25), w, h)
		}

		frameMask := colors.Lerp(colors.White.RGBA, color, frameMaskLerp)
		levelMask := colors.Lerp(colors.White.RGBA, color, 0.75)

		m := originM.Moved(pixel.V(dxDirection*dx, 0)).Moved(origin.AlignFrom(gfx.BottomLeft, w, h))

		statusFrame.Draw(target, pixel.R(0, 0, w, h), m, frames.WithRenderOrigin(gfx.BottomLeft), frames.WithColor(colors.WithAlpha(frameMask, alpha)))

		progressFrameH := int((h - 2.0) * levelProgress)
		progressFrameW := int(w - 2.0)

		gfx.DrawRect(atlas, target, m.Moved(gfx.IVec(1, 1)), gfx.BottomLeft, progressFrameW, progressFrameH, colors.WithAlpha(colors.White.RGBA, 0.1*alpha))

		for id := len(status.renderEffects) - 1; id >= 0; id-- {
			e := status.renderEffects[id]
			e.Render(target, m, timeDelta)
			if e.IsComplete() {
				status.renderEffects = append(status.renderEffects[:id], status.renderEffects[id+1:]...)
			}
		}

		timerProgress := (float64(status.ticksSinceLastUpdate) + tickProgress) / float64(ticksPerStatusApply)
		game.DebugBRf("timerProgress: %.3f", timerProgress)
		if timerProgress < status.renderLastProgress {
			status.timeSinceLastApply = 0
		}
		status.renderLastProgress = timerProgress
		status.timeSinceLastApply += timeDelta

		// border + flash on effect
		status.timerAnimation.SetProgress(timerProgress)
		timerSprite := status.timerAnimation.Sprite()
		timerMatrix := m.Moved(gfx.BottomLeft.Align(timerSprite))
		applyFlashProgress := min(1.0, status.timeSinceLastApply/statusApplyFadeDur)
		if applyFlashProgress < 1 {
			borderAlpha := 1.0 - applyFlashProgress
			borderAlpha = borderAlpha * alpha
			game.DebugBRf("border alpha: %.3f", borderAlpha)
			statusBorder.DrawColorMask(target, timerMatrix, colors.WithAlpha(color, borderAlpha))
		}
		applyFlashProgressHalf := min(1.0, status.timeSinceLastApply/(statusApplyFadeDur/2.0))
		timerAlpha := alpha
		timerFlashAmount := 0.75
		timerColor := colors.Lerp(colors.White.RGBA, color, applyFlashProgressHalf*timerFlashAmount+1.0-timerFlashAmount)
		timerSprite.DrawColorMask(target, timerMatrix, colors.WithAlpha(timerColor, timerAlpha))

		levelSprite.DrawColorMask(target, m.Moved(gfx.BottomLeft.Align(levelSprite)).Moved(pixel.V(margin+1, margin)), colors.WithAlpha(levelMask, alpha)) // +1 because rounding bug

		icon.DrawColorMask(target, m.Moved(gfx.BottomLeft.Align(icon)).Moved(pixel.V(margin, margin+levelSprite.Bounds().H()+spacing)), colors.Alpha(alpha))

		dx += w + xPadding
	}
}

func (a *AppliedStatuses) GetLevels() map[rpg.StatusType]rpg.StatusLevel {
	levels := map[rpg.StatusType]rpg.StatusLevel{}
	for _, status := range a.Statuses {
		levels[status.Status] = status.Level()
	}
	return levels
}

func (a *AppliedStatuses) ReduceResult(reductions map[rpg.StatusType]float64) {
	for _, status := range a.Statuses {
		reduction, exists := reductions[status.Status]
		if !exists {
			continue
		}
		status.Stacks -= reduction
		if status.Stacks < 0 {
			status.Stacks = 0
		}
	}
}
