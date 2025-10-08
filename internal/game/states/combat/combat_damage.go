package combat

import (
	"fisherevans.com/project/f/internal/game/rpg"
	"fisherevans.com/project/f/internal/util/colors"
)

type DamageOptions struct {
	Status rpg.StatusType
}

func (s *State) AdjustHealth(amount int, target Combatant, opt *DamageOptions) {
	if amount == 0 {
		return
	}
	target.AdjustHealth(amount)
	color := colors.Hex(0xed0027)
	if amount > 0 {
		color = colors.Hex(0x1ced00)
	}
	if opt != nil {
		if statusColor, exists := colors.StatusColors[opt.Status]; exists {
			color = statusColor
		}
	}
	s.fx = append(s.fx, NewHealthAdjustFX(amount, color, target))
}
