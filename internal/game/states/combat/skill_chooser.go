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
	initialOrderedSkills []rpg.SkillId
	options              []randomSkillChoice
	totalWeight          int
}

func NewSkillChooser(pool rpg.CombatSkillPool) SkillChooser {
	if pool.Random == nil {
		panic("no random skill pool")
	}
	var options []randomSkillChoice
	for skill, weight := range pool.Random.WeightedSkills {
		options = append(options, randomSkillChoice{skill: &skill, weight: weight})
	}
	rsc := NewRandomSkillChooser(options)
	rsc.SetInitialOrderedSkills(pool.Random.InitialOrderedSkills)
	return rsc
}

func NewRandomSkillChooser(options []randomSkillChoice) *RandomSkillChooser {
	totalWeight := 0
	for _, option := range options {
		totalWeight += option.weight
	}
	return &RandomSkillChooser{
		options:     options,
		totalWeight: totalWeight,
	}
}

func (r *RandomSkillChooser) NextSkill() *rpg.SkillId {
	if len(r.initialOrderedSkills) > 0 {
		next := r.initialOrderedSkills[0]
		r.initialOrderedSkills = r.initialOrderedSkills[1:]
		return &next
	}
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

func (r *RandomSkillChooser) SetInitialOrderedSkills(skills []rpg.SkillId) {
	r.initialOrderedSkills = skills
}
