package rpg

var Primortals = map[PrimortalType]Primortal{}
var XenoLogEntries = map[int]PrimortalType{}
var MaxXenoLogEntryIndex = 0
var reservedUnlockableSkill = map[SkillId]bool{}

const SkillCostUnavailable = 0

type PrimortalType string

func (pt PrimortalType) Primortal() Primortal {
	if p, exists := Primortals[pt]; exists {
		return p
	}
	panic("unknown primortal type: " + pt)
}

type Primortal struct {
	XenoLogIndex     int
	Type             PrimortalType
	Name             string
	Description      string
	BaseSync         int
	UnlockableSkills []UnlockableSkill
	CombatArchetypes map[string]PrimortalCombatArchetype
}

type PrimortalCombatArchetype struct {
	AdditionalSync         int             `yaml:"additional_sync"`
	AdditionalSyncVariance int             `yaml:"additional_sync_variance"`
	SkillPool              CombatSkillPool `yaml:"skill_pool"`
}

type CombatSkillPool struct {
	Random *CombatSkillPoolRandom `yaml:"random"`
}

type CombatSkillPoolRandom struct {
	WeightedSkills map[SkillId]int `yaml:"weighted_skills"`
}

type UnlockableSkill struct {
	SkillId SkillId
	Cost    int
}

func (p Primortal) register() Primortal {
	if _, exists := Primortals[p.Type]; exists {
		panic("duplicate primortal type: " + p.Type)
	}
	for _, us := range p.UnlockableSkills {
		if taken, _ := reservedUnlockableSkill[us.SkillId]; taken {
			panic("duplicate unlockable skill: " + us.SkillId)
		}
	}
	if p.XenoLogIndex < 0 || p.XenoLogIndex > 1000 {
		panic("invalid XenoLogIndex: " + string(p.XenoLogIndex))
	}
	if p.XenoLogIndex != 0 {
		if _, exists := XenoLogEntries[p.XenoLogIndex]; exists {
			panic("duplicate XenoLogIndex: " + string(p.XenoLogIndex))
		}
		XenoLogEntries[p.XenoLogIndex] = p.Type
		if p.XenoLogIndex > MaxXenoLogEntryIndex {
			MaxXenoLogEntryIndex = p.XenoLogIndex
		}
	}
	Primortals[p.Type] = p
	return p
}

var DefaultUnlockableSkills = []UnlockableSkill{
	{
		SkillId: Skill_Strike.Id,
		Cost:    5,
	},
	{
		SkillId: Skill_Brace.Id,
		Cost:    5,
	},
}

var Primortal_Dummy = Primortal{
	Type:        "dummy",
	Name:        "Dummy",
	Description: "A test robot to hit for fun.",
	BaseSync:    50,
	CombatArchetypes: map[string]PrimortalCombatArchetype{
		"default": {
			SkillPool: CombatSkillPool{
				Random: &CombatSkillPoolRandom{
					WeightedSkills: map[SkillId]int{
						Skill_Brace.Id: 40,
						Skill_Guard.Id: 20,
						Skill_Jab.Id:   10,
					},
				},
			},
		},
	},
}.register()

// kinetic
var Primortal_Pumbl = Primortal{
	Type:         "pumbl",
	Name:         "Pumbl",
	XenoLogIndex: 3,
	Description:  "Little roll up rock guy. A mix between Sandshrew and Geodude.",
	BaseSync:     30,
	UnlockableSkills: []UnlockableSkill{
		{
			SkillId: Skill_ShoulderRoll.Id,
			Cost:    2,
		},
		{
			SkillId: Skill_CurlUp.Id,
			Cost:    5,
		},
	},
	CombatArchetypes: map[string]PrimortalCombatArchetype{
		"default": {
			AdditionalSync:         10,
			AdditionalSyncVariance: 5,
			SkillPool: CombatSkillPool{
				Random: &CombatSkillPoolRandom{
					WeightedSkills: map[SkillId]int{
						Skill_ShoulderRoll.Id: 40,
						Skill_CurlUp.Id:       40,
						Skill_Guard.Id:        20,
					},
				},
			},
		},
	},
}.register()

// thermal
var Primortal_Scintail = Primortal{
	Type:         "scintail",
	Name:         "Scintail",
	XenoLogIndex: 43,
	Description:  "A cute lizard guy with heat fins that glow.",
	BaseSync:     25,
	UnlockableSkills: []UnlockableSkill{
		{
			SkillId: Skill_Cinder.Id,
			Cost:    3,
		},
		{
			SkillId: Skill_Searline.Id,
			Cost:    7,
		},
	},
	CombatArchetypes: map[string]PrimortalCombatArchetype{
		"default": {
			AdditionalSync:         5,
			AdditionalSyncVariance: 10,
			SkillPool: CombatSkillPool{
				Random: &CombatSkillPoolRandom{
					WeightedSkills: map[SkillId]int{
						Skill_Cinder.Id:   20,
						Skill_Searline.Id: 40,
						Skill_Brace.Id:    10,
						Skill_Strike.Id:   20,
					},
				},
			},
		},
	},
}.register()

// voltaic
var Primortal_Volteel = Primortal{
	Type:         "volteel",
	Name:         "Volteel",
	XenoLogIndex: 8,
	Description:  "An eel with stubby little feet that can stand up right.",
	BaseSync:     20,
	UnlockableSkills: []UnlockableSkill{
		{
			SkillId: Skill_ArcDart.Id,
			Cost:    2,
		},
		{
			SkillId: Skill_ZapWrap.Id,
			Cost:    4,
		},
	},
	CombatArchetypes: map[string]PrimortalCombatArchetype{
		"default": {
			AdditionalSync:         10,
			AdditionalSyncVariance: 10,
			SkillPool: CombatSkillPool{
				Random: &CombatSkillPoolRandom{
					WeightedSkills: map[SkillId]int{
						Skill_ArcDart.Id: 20,
						Skill_ZapWrap.Id: 40,
						Skill_Strike.Id:  10,
					},
				},
			},
		},
	},
}.register()

// voltaic - arcmander

// corrosive
var Primortal_Toxmidge = Primortal{
	Type:         "toxmidge",
	Name:         "Toxmidge",
	XenoLogIndex: 9,
	Description:  "A flying gross flying insect that squirts acid.",
	BaseSync:     15,
	UnlockableSkills: []UnlockableSkill{
		{
			SkillId: Skill_AcidSting.Id,
			Cost:    3,
		},
		{
			SkillId: Skill_Molt.Id,
			Cost:    7,
		},
	},
	CombatArchetypes: map[string]PrimortalCombatArchetype{
		"default": {
			AdditionalSync:         5,
			AdditionalSyncVariance: 5,
			SkillPool: CombatSkillPool{
				Random: &CombatSkillPoolRandom{
					WeightedSkills: map[SkillId]int{
						Skill_AcidSting.Id: 20,
						Skill_Molt.Id:      10,
						Skill_Guard.Id:     20,
						Skill_Jab.Id:       20,
					},
				},
			},
		},
	},
}.register()

// mutagenic
var Primortal_Myceli = Primortal{
	Type:         "myceli",
	Name:         "Myceli",
	XenoLogIndex: 19,
	Description:  "A fun guy with tentacle legs.",
	BaseSync:     40,
	UnlockableSkills: []UnlockableSkill{
		{
			SkillId: Skill_PhotoSurge.Id,
			Cost:    4,
		},
		{
			SkillId: Skill_MendSpores.Id,
			Cost:    5,
		},
	},
	CombatArchetypes: map[string]PrimortalCombatArchetype{
		"default": {
			AdditionalSync:         0,
			AdditionalSyncVariance: 10,
			SkillPool: CombatSkillPool{
				Random: &CombatSkillPoolRandom{
					WeightedSkills: map[SkillId]int{
						Skill_PhotoSurge.Id: 40,
						Skill_MendSpores.Id: 10,
						Skill_Jab.Id:        20,
						Skill_Brace.Id:      20,
					},
				},
			},
		},
	},
}.register()
