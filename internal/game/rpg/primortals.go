package rpg

var Primortals = map[PrimortalType]Primortal{}
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
	Type                   PrimortalType
	Name                   string
	BaseSync               int
	UnlockableSkills       []UnlockableSkill
	AdditionalCombatSkills []SkillId
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

// kinetic - little roll up rock guy - mix between sandshrew and geodude
var Primortal_Pumbl = Primortal{
	Type:     "pumbl",
	Name:     "Pumbl",
	BaseSync: 70,
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
	AdditionalCombatSkills: []SkillId{
		Skill_Guard.Id,
	},
}.register()

// thermal - little lizard guy - heat fins that glow
var Primortal_Scintail = Primortal{
	Type:     "scintail",
	Name:     "Scintail",
	BaseSync: 50,
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
	AdditionalCombatSkills: []SkillId{
		Skill_Brace.Id,
		Skill_Strike.Id,
	},
}.register()

// voltaic - eel, with stubby little feet - can stand up right
var Primortal_Volteel = Primortal{
	Type:     "volteel",
	Name:     "Volteel",
	BaseSync: 40,
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
	AdditionalCombatSkills: []SkillId{
		Skill_Strike.Id,
	},
}.register()

// voltaic - arcmander

// corrosive - flying insect with pincers
var Primortal_Toxmidge = Primortal{
	Type:     "toxmidge",
	Name:     "Toxmidge",
	BaseSync: 50,
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
	AdditionalCombatSkills: []SkillId{
		Skill_Guard.Id,
		Skill_Jab.Id,
	},
}.register()

// mutagenic - little mushroom guy, base turns into legs
var Primortal_Myceli = Primortal{
	Type:     "myceli",
	Name:     "Myceli",
	BaseSync: 80,
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
	AdditionalCombatSkills: []SkillId{
		Skill_Brace.Id,
		Skill_Jab.Id,
	},
}.register()
