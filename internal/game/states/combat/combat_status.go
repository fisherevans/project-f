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

type StatusLevel int

const (
	StatusLevel1 StatusLevel = 1
	StatusLevel2 StatusLevel = 2
	StatusLevel3 StatusLevel = 3
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

func (s *CombatStatus) Level() StatusLevel {
	if s.Stacks >= 7 {
		return StatusLevel3
	}
	if s.Stacks >= 3 {
		return StatusLevel2
	}
	return StatusLevel1
}

type AppliedStatuses struct {
	Statuses []*CombatStatus
}

func NewAppliedStatuses() *AppliedStatuses {
	return &AppliedStatuses{
		Statuses: []*CombatStatus{
			{
				Status:    rpg.StatusPoisoned,
				Animation: anim.NewStaticAnimation(atlas.GetTilesheetSprite("combat/status_icons", 1, 1)),
				Apply: func(s *State, appliedTo Combatant, cs *CombatStatus) {
					// todo change to % of max health
					s.ApplyDamage(1, appliedTo, &DamageOptions{
						Status: rpg.StatusPoisoned,
					})
				},
				CustomTickDecay: func(s *State, cs *CombatStatus) {
					decayPerTick := func(l StatusLevel) float64 {
						switch l {
						case StatusLevel3:
							return 0.5
						case StatusLevel2:
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
					s.ApplyDamage(dmg, appliedTo, &DamageOptions{
						Status: rpg.StatusBurning,
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
		parts = append(parts, fmt.Sprintf("%c:%d", status.Status[0], status.Stacks))
	}
	return strings.Join(parts, "|")
}

var (
	statusFrame        = frames.New("combat/status_frame", atlas)
	statusLevelSprites = map[StatusLevel]pixelutil.BoundedDrawable{
		StatusLevel1: atlas.GetTilesheetSprite("combat/status_level", 1, 1),
		StatusLevel2: atlas.GetTilesheetSprite("combat/status_level", 2, 1),
		StatusLevel3: atlas.GetTilesheetSprite("combat/status_level", 3, 1),
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
		case StatusLevel2:
			frameMaskLerp = 0.4
		case StatusLevel3:
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
