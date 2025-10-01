package combat

import (
	"math/rand"

	"fisherevans.com/project/f/internal/game/rpg"
)

type SkillChooser interface {
	NextSkill() *rpg.SkillId
}

type randomSkillChoice struct {
	skill  *rpg.SkillId
	weight int
}

type RandomSkillChooser struct {
	options     []randomSkillChoice
	totalWeight int
}

func NewRandomSkillChooserEven(skills []rpg.SkillId) SkillChooser {
	var options []randomSkillChoice
	for _, skill := range skills {
		options = append(options, randomSkillChoice{skill: &skill, weight: 1})
	}
	return NewRandomSkillChooser(options)
}

func NewRandomSkillChooser(options []randomSkillChoice) SkillChooser {
	totalWeight := 0
	for _, option := range options {
		totalWeight += option.weight
	}
	return &RandomSkillChooser{
		options:     options,
		totalWeight: totalWeight,
	}
}

func (r RandomSkillChooser) NextSkill() *rpg.SkillId {
	if len(r.options) == 0 {
		panic("no options")
	}
	n := rand.Intn(r.totalWeight)
	for _, option := range r.options {
		n -= option.weight
		if n <= 0 {
			return option.skill
		}
	}
	return r.options[0].skill
}
