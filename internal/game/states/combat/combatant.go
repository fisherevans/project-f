package combat

import (
	"math"

	"github.com/gopxl/pixel/v2"

	"fisherevans.com/project/f/internal/game/rpg"
	"fisherevans.com/project/f/internal/util/colors"
)

type Combatant interface {
	Name() string
	ApplyDamage(damage int)
	GetStats() CombatantStats
	GetTempo() *Tempo
	Update(timeDelta float64)
	GetColorMask() pixel.RGBA

	GetCurrentSkill() *SkillInstance
	SetCurrentSkill(skill *SkillInstance)

	IsNextSkillCommitted() bool
	PopNextSkill() *rpg.SkillId
	PeekNextSkill() *rpg.SkillId
}

type HealthState struct {
	Max        int
	Target     int
	Current    float64
	AdjustRate float64
}

func NewHealthState(max int) *HealthState {
	return &HealthState{
		Max:        max,
		Current:    float64(max),
		Target:     max,
		AdjustRate: 0.1,
	}
}

func (h *HealthState) GetCurrentInt() int {
	return int(math.Round(h.Current))
}

func (h *HealthState) AdjustTarget(amount int) int {
	h.Target += amount
	if h.Target < 0 {
		remainder := h.Target
		h.Target = 0
		return remainder
	}
	if h.Target > h.Max {
		remainder := h.Target - h.Max
		h.Target = h.Max
		return remainder
	}
	return 0
}

func (h *HealthState) Update(timeDelta float64) {
	if math.Round(h.Current) == float64(h.Target) {
		return
	}
	sign := 1.0
	if h.Current > float64(h.Target) {
		sign = -1
	}
	diff := h.AdjustRate * timeDelta * float64(h.Max)
	maxDiff := math.Abs(float64(h.Target) - h.Current)
	diff = math.Min(diff, maxDiff)
	h.Current += sign * diff
}

type CombatantStats struct {
	Stance rpg.CombatStance
}

type DamageFlashMask struct {
	FullMask         pixel.RGBA
	RecoveryDuration float64

	timeRecovered float64
}

func NewDamageFlashMask() *DamageFlashMask {
	return &DamageFlashMask{
		FullMask:         pixel.RGBA{1, 0.4, 0.4, 1},
		RecoveryDuration: 0.5,
	}
}

func (d *DamageFlashMask) Update(timeDelta float64) {
	d.timeRecovered += timeDelta
}

func (d *DamageFlashMask) getMask() pixel.RGBA {
	recovered := pixel.RGBA{1, 1, 1, 1}
	recoveryProgress := math.Min(1, d.timeRecovered/d.RecoveryDuration)
	if recoveryProgress >= 1 {
		return recovered
	}
	return colors.Lerp(d.FullMask, recovered, recoveryProgress)
}

func (d *DamageFlashMask) damaged() {
	d.timeRecovered = 0
}

type CurrentCombatantSkills struct {
	CurrentSkill *SkillInstance
}

func NewCurrentCombatantSkills() *CurrentCombatantSkills {
	return &CurrentCombatantSkills{}
}

func (c *CurrentCombatantSkills) GetCurrentSkill() *SkillInstance {
	return c.CurrentSkill
}

func (c *CurrentCombatantSkills) SetCurrentSkill(skill *SkillInstance) {
	c.CurrentSkill = skill
}
