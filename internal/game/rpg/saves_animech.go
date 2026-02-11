package rpg

const BaseAnimechShield = 25
const BaseAnimechSync = 25

type Animech struct {
	SkillSet   *SkillSet        `yaml:"skill_set"`
	Experience int              `yaml:"experience_points"`
	Upgrades   *AnimechUpgrades `yaml:"upgrades"`

	// current run details
	PendingExperience int `yaml:"pending_experience"`
	CurrentSync       int `yaml:"current_sync"`
	CurrentShield     int `yaml:"current_shield"`
}

func (a *Animech) GetMaxShield() int {
	return BaseAnimechShield + a.Upgrades.AdditionalShield()
}

func (a *Animech) GetMaxSync() int {
	return BaseAnimechSync + a.Upgrades.AdditionalSync()
}

func (a *Animech) FillDefaults() {
	if a.SkillSet == nil {
		a.SkillSet = &SkillSet{}
	}
	if a.SkillSet.IsEmpty() {
		a.SkillSet = &SkillSet{
			Skill1: Skill_Jab.Id,
		}
	}
	if a.Upgrades == nil {
		a.Upgrades = &AnimechUpgrades{}
	}
}

type AnimechUpgrades struct {
	ShieldLevel int `yaml:"shield_level"`
	SyncLevel   int `yaml:"sync_level"`
}

func (u *AnimechUpgrades) GetLevel() int {
	return 1 + u.ShieldLevel + u.SyncLevel
}

func (u *AnimechUpgrades) AdditionalSync() int {
	return AnimechUpgradeAdditionalSync(u.GetLevel())
}

func (u *AnimechUpgrades) AdditionalShield() int {
	return AnimechUpgradeAdditionalShield(u.GetLevel())
}

func AnimechUpgradeAdditionalSync(level int) int {
	return level * 10
}

func AnimechUpgradeAdditionalShield(level int) int {
	return level * 5
}

func AnimechUpgradeExperienceRequiredToUpgrade(level int) int {
	return 25 * (level + 1)
}
