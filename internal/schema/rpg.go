package schema

type RPGSkill struct {
    Id          string         `yaml:"id" json:"id"`
    Name        string         `yaml:"name" json:"name"`
    Description string         `yaml:"description" json:"description"`
    Ticks       []RPGSkillTick `yaml:"ticks" json:"ticks"`
}

type RPGSkillTick struct {
    Stance    string            `yaml:"stance,omitempty" json:"stance,omitempty"`
    Effects   []RPGTickEffect   `yaml:"effects,omitempty" json:"effects,omitempty"`
    Animation *RPGTickAnimation `yaml:"animation,omitempty" json:"animation,omitempty"`
}

type RPGTickEffect struct {
    Damage *RPGTickDamage `yaml:"damage,omitempty" json:"damage,omitempty"`
    Status *RPGTickStatus `yaml:"status,omitempty" json:"status,omitempty"`
}

type RPGTickDamage struct {
    Amount   int               `yaml:"amount" json:"amount"`
    Variance int               `yaml:"variance,omitempty" json:"variance,omitempty"`
    MissRate float64           `yaml:"miss_rate,omitempty" json:"miss_rate,omitempty"`
    ScaledBy *RPGDamageScaling `yaml:"scaled_by,omitempty" json:"scaled_by,omitempty"`
}

type RPGDamageScaling struct {
    TargetStatus map[string]map[string]float64 `yaml:"target_status,omitempty" json:"target_status,omitempty"`
    SourceStatus map[string]map[string]float64 `yaml:"source_status,omitempty" json:"source_status,omitempty"`
}

type RPGTickStatus struct {
    Type    string  `yaml:"type" json:"type"`
    Stacks  float64 `yaml:"stacks" json:"stacks"`
    Target  string  `yaml:"target" json:"target"`
    Require string  `yaml:"require,omitempty" json:"require,omitempty"`
}

type RPGTickAnimation struct {
    Source []RPGTransformation `yaml:"source,omitempty" json:"source,omitempty"`
    Target []RPGTransformation `yaml:"target,omitempty" json:"target,omitempty"`
}

type RPGTransformation struct {
    Type        string  `yaml:"type" json:"type"`
    Speed       float64 `yaml:"speed,omitempty" json:"speed,omitempty"`
    Repetitions int     `yaml:"repetitions,omitempty" json:"repetitions,omitempty"`
}

type RPGPrimortal struct {
    Type             string                            `yaml:"type" json:"type"`
    Name             string                            `yaml:"name" json:"name"`
    Description      string                            `yaml:"description" json:"description"`
    XenoLogIndex     int                               `yaml:"xeno_log_index,omitempty" json:"xeno_log_index,omitempty"`
    BaseSync         int                               `yaml:"base_sync" json:"base_sync"`
    UnlockableSkills map[string]RPGUnlockableSkill     `yaml:"unlockable_skills,omitempty" json:"unlockable_skills,omitempty"`
    CombatArchetypes map[string]RPGCombatArchetype     `yaml:"combat_archetypes,omitempty" json:"combat_archetypes,omitempty"`
}

type RPGUnlockableSkill struct {
    Cost          int      `yaml:"cost" json:"cost"`
    Prerequisites []string `yaml:"prerequisites,omitempty" json:"prerequisites,omitempty"`
}

type RPGCombatArchetype struct {
    AdditionalSync         int           `yaml:"additional_sync,omitempty" json:"additional_sync,omitempty"`
    AdditionalSyncVariance int           `yaml:"additional_sync_variance,omitempty" json:"additional_sync_variance,omitempty"`
    SkillPool              *RPGSkillPool `yaml:"skill_pool,omitempty" json:"skill_pool,omitempty"`
}

type RPGSkillPool struct {
    Random *RPGRandomPool `yaml:"random,omitempty" json:"random,omitempty"`
}

type RPGRandomPool struct {
    InitialOrderedSkills []string       `yaml:"initial_ordered_skills,omitempty" json:"initial_ordered_skills,omitempty"`
    WeightedSkills       map[string]int `yaml:"weighted_skills,omitempty" json:"weighted_skills,omitempty"`
}
