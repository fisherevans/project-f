package combat

import (
	"math"

	"github.com/gopxl/pixel/v2"

	"fisherevans.com/project/f/internal/game/rpg"
	"fisherevans.com/project/f/internal/util/colors"
)

type Combatant interface {
	Name() string
	AdjustHealth(amount int)
	GetStats() rpg.CombatantStats
	Update(timeDelta float64)
	GetColorMask() pixel.RGBA
	GetStatuses() *AppliedStatuses
	IsDead() bool
	IsPlayer() bool
	GetRenderer() *CombatantRenderer

	GetCurrentSkill() *SkillInstance
	SetCurrentSkill(skill *SkillInstance)

	IsNextSkillCommitted() bool
	PopNextSkill() *rpg.SkillId
	PeekNextSkill() *rpg.SkillId
	GetTotalMaxHealth() int
}

type HealthState struct {
	Max        int
	Target     int
	Current    float64
	AdjustRate float64
}

func NewFullHealthState(max int) *HealthState {
	return NewHealthState(max, max)
}

func NewHealthState(current, max int) *HealthState {
	return &HealthState{
		Max:        max,
		Current:    float64(current),
		Target:     current,
		AdjustRate: 1.0, //0.1, // 1 == turn off
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

type HealthFlash struct {
	DamageMask       pixel.RGBA
	HealMask         pixel.RGBA
	RecoveryDuration float64

	wasDamaged    bool
	timeRecovered float64
}

func NewDamageFlashMask() *HealthFlash {
	return &HealthFlash{
		DamageMask:       pixel.RGBA{R: 1, G: 0.4, B: 0.4, A: 1},
		HealMask:         pixel.RGBA{R: 0.4, G: 1, B: 0.4, A: 1},
		RecoveryDuration: 0.5,
	}
}

func (d *HealthFlash) Update(timeDelta float64) {
	d.timeRecovered += timeDelta
}

func (d *HealthFlash) getMask() pixel.RGBA {
	recovered := pixel.RGBA{R: 1, G: 1, B: 1, A: 1}
	recoveryProgress := math.Min(1, d.timeRecovered/d.RecoveryDuration)
	if recoveryProgress >= 1 {
		return recovered
	}
	mask := d.HealMask
	if d.wasDamaged {
		mask = d.DamageMask
	}
	return colors.Lerp(mask, recovered, recoveryProgress)
}

func (d *HealthFlash) damaged() {
	d.wasDamaged = true
	d.timeRecovered = 0
}

func (d *HealthFlash) healed() {
	d.wasDamaged = false
	d.timeRecovered = 0
}

func (d *HealthFlash) basedOnAdjust(amount int) {
	if amount == 0 {
		return
	}
	if amount > 0 {
		d.healed()
	} else {
		d.damaged()
	}
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
