package rpg

var Primortal_Fireslug = Primortal{
	Type:                "fireslug",
	Name:                "Fire Slug",
	BaseSync:            10,
	BasePhysicalAttack:  5,
	BasePhysicalDefence: 4,
	BaseAetherAttack:    20,
	BaseAetherDefence:   15,
	BodyType:            BodyTypeLiquid,
	Affinity:            SkillTypeThermal,
}.register()

var Primortal_Hydrax = Primortal{
	Type:                "hydrax",
	Name:                "Hydrax",
	BaseSync:            14,
	BasePhysicalAttack:  6,
	BasePhysicalDefence: 10,
	BaseAetherAttack:    18,
	BaseAetherDefence:   12,
	BodyType:            BodyTypeLiquid,
	Affinity:            SkillTypeVoltaic,
}.register()
