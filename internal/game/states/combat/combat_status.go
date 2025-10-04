package combat

import (
	"fmt"
	"math"
	"strings"

	"github.com/gopxl/pixel/v2"

	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/game/anim"
	"fisherevans.com/project/f/internal/game/rpg"
	"fisherevans.com/project/f/internal/util/colors"
	"fisherevans.com/project/f/internal/util/frames"
	"fisherevans.com/project/f/internal/util/gfx"
	"fisherevans.com/project/f/internal/util/pixelutil"
)

type CombatStatus struct {
	Status               rpg.StatusType
	Stacks               float64
	AccumulativeExposure int
	Apply                func(s *State, appliedTo Combatant, cs *CombatStatus)
	CustomTickDecay      func(s *State, cs *CombatStatus)
	Animation            *anim.AnimatedSprite
}

const maxStacks = 14

func (s *CombatStatus) Level() rpg.StatusLevel {
	if s.Stacks >= 7 {
		return rpg.StatusLevel3
	}
	if s.Stacks >= 3 {
		return rpg.StatusLevel2
	}
	if s.Stacks >= 1 {
		return rpg.StatusLevel1
	}
	return rpg.StatusLevel0
}

type AppliedStatuses struct {
	Statuses []*CombatStatus
}

func NewAppliedStatuses() *AppliedStatuses {
	return &AppliedStatuses{
		Statuses: []*CombatStatus{
			{
				Status:    rpg.StatusFortified,
				Animation: anim.NewStaticAnimation(atlas.GetTilesheetSprite("combat/status_icons", 5, 1)),
				Apply: func(s *State, appliedTo Combatant, cs *CombatStatus) {
					// nothing, effects happen when receiving damage
				},
			},
			{
				Status:    rpg.StatusPoisoned,
				Animation: anim.NewStaticAnimation(atlas.GetTilesheetSprite("combat/status_icons", 1, 1)),
				Apply: func(s *State, appliedTo Combatant, cs *CombatStatus) {
					pct := 0.01
					switch cs.Level() {
					case rpg.StatusLevel3:
						pct = 0.03
					case rpg.StatusLevel2:
						pct = 0.02
					}
					dmg := int(math.Max(float64(appliedTo.GetTotalMaxHealth())*pct, 1.0))
					s.AdjustHealth(-dmg, appliedTo, &DamageOptions{
						Status: rpg.StatusPoisoned,
					})
				},
				CustomTickDecay: func(s *State, cs *CombatStatus) {
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
					cs.Stacks -= decayPerTick(cs.Level())
				},
			},
			{
				Status:    rpg.StatusBurning,
				Animation: anim.NewStaticAnimation(atlas.GetTilesheetSprite("combat/status_icons", 2, 1)),
				Apply: func(s *State, appliedTo Combatant, cs *CombatStatus) {
					dmg := cs.AccumulativeExposure / 10
					if dmg < 1 {
						dmg = 1
					}
					s.AdjustHealth(-dmg, appliedTo, &DamageOptions{
						Status: rpg.StatusBurning,
					})
				},
			},
			{
				Status:    rpg.StatusIonized,
				Animation: anim.NewStaticAnimation(atlas.GetTilesheetSprite("combat/status_icons", 3, 1)),
				Apply: func(s *State, appliedTo Combatant, cs *CombatStatus) {
					// nothing, effects happen when applying damage
				},
				CustomTickDecay: func(s *State, cs *CombatStatus) {
					decayPerTick := func(l rpg.StatusLevel) float64 {
						switch l {
						case rpg.StatusLevel3:
							return 0.25
						case rpg.StatusLevel2:
							return 0.5
						default:
							return 0.75
						}
					}
					cs.Stacks -= decayPerTick(cs.Level())
				},
			},
			{
				Status:    rpg.StatusMending,
				Animation: anim.NewStaticAnimation(atlas.GetTilesheetSprite("combat/status_icons", 4, 1)),
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
}

func (a *AppliedStatuses) DecreaseStacks() {
	for _, status := range a.Statuses {
		if status.Stacks <= 0 {
			status.AccumulativeExposure = 0
			continue
		}
		if status.CustomTickDecay != nil {
			status.CustomTickDecay(nil, status)
		} else {
			if status.Stacks > 0 {
				status.Stacks -= 1
			}
		}
		if status.Stacks < 0 {
			status.Stacks = 0
		}
	}
}

func (a *AppliedStatuses) Add(status rpg.StatusType, stacks float64) {
	for _, s := range a.Statuses {
		if s.Status == status {
			s.Stacks += stacks
			if s.Stacks > maxStacks {
				s.Stacks = maxStacks
			}
			return
		}
	}
}

func (a *AppliedStatuses) ApplyEffects(s *State, appliedTo Combatant) {
	for _, status := range a.Statuses {
		if status.Apply != nil && status.Stacks > 0 {
			status.AccumulativeExposure += int(status.Level())
			status.Apply(s, appliedTo, status)
		}
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
		parts = append(parts, fmt.Sprintf("%s:%d(%.1f)", status.Status, status.Level(), status.Stacks))
	}
	return strings.Join(parts, "|")
}

var (
	statusFrame        = frames.New("combat/status_frame", atlas)
	statusLevelSprites = map[rpg.StatusLevel]pixelutil.BoundedDrawable{
		rpg.StatusLevel1: atlas.GetTilesheetSprite("combat/status_level", 1, 1),
		rpg.StatusLevel2: atlas.GetTilesheetSprite("combat/status_level", 2, 1),
		rpg.StatusLevel3: atlas.GetTilesheetSprite("combat/status_level", 3, 1),
	}
)

func (a *AppliedStatuses) Render(ctx *game.Context, originM pixel.Matrix, target pixel.Target, timeDelta float64, origin gfx.OriginLocation) {
	dx := 0.0
	xPadding := 2.0
	dxDirection := -1.0
	if origin == gfx.TopLeft {
		dxDirection = 1
	}
	for _, status := range a.Statuses {
		if status.Stacks == 0 {
			continue
		}
		level := status.Level()

		frameMaskLerp := 0.1
		switch level {
		case rpg.StatusLevel2:
			frameMaskLerp = 0.4
		case rpg.StatusLevel3:
			frameMaskLerp = 0.7
		}
		color := colors.StatusColors[status.Status]
		frameMask := colors.Lerp(colors.White.RGBA, color, frameMaskLerp)
		levelMask := colors.Lerp(colors.White.RGBA, color, 0.75)

		status.Animation.Update(timeDelta)
		icon := status.Animation.Sprite()
		levelSprite := statusLevelSprites[level]
		margin := 2.0
		spacing := 1.0
		w := margin*2 + math.Max(icon.Bounds().W(), levelSprite.Bounds().W())
		h := margin*2 + icon.Bounds().H() + spacing + levelSprite.Bounds().H()

		m := originM.Moved(pixel.V(dxDirection*dx, 0)).Moved(origin.AlignFrom(gfx.BottomLeft, w, h))

		statusFrame.Draw(target, pixel.R(0, 0, w, h), m, frames.WithRenderOrigin(gfx.BottomLeft), frames.WithColor(frameMask))

		levelSprite.DrawColorMask(target, m.Moved(gfx.BottomLeft.Align(levelSprite)).Moved(pixel.V(margin+1, margin)), levelMask) // +1 because rounding bug

		icon.Draw(target, m.Moved(gfx.BottomLeft.Align(icon)).Moved(pixel.V(margin, margin+levelSprite.Bounds().H()+spacing)))

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
