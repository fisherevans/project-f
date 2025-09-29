package combat

import (
	"fisherevans.com/project/f/internal/game/rpg"
	"fisherevans.com/project/f/internal/util/colors"
)

type DamageOptions struct {
	Status rpg.StatusType
}

func (s *State) ApplyDamage(damage int, target Combatant, opt *DamageOptions) {
	if damage == 0 {
		return
	}
	target.ApplyDamage(damage)
	color := colors.White.RGBA
	if opt != nil {
		if statusColor, exists := colors.StatusColors[opt.Status]; exists {
			color = statusColor
		}
	}
	s.fx = append(s.fx, NewDamageFX(damage, color, target))
}
