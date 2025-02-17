package rpg

var Primortal_Fireslug = Primortal{
	Type:                "fireslug",
	Name:                "Fire Slug",
	BaseSync:            10,
	BasePhysicalAttack:  5,
	BasePhysicalDefense: 4,
	BaseAetherAttack:    20,
	BaseAetherDefense:   15,
	BodyType:            BodyTypeLiquid,
	Affinity:            SkillTypeThermal,
}.register()

var Primortal_Hydrax = Primortal{
	Type:                "hydrax",
	Name:                "Hydrax",
	BaseSync:            14,
	BasePhysicalAttack:  6,
	BasePhysicalDefense: 10,
	BaseAetherAttack:    18,
	BaseAetherDefense:   12,
	BodyType:            BodyTypeLiquid,
	Affinity:            SkillTypeVoltaic,
}.register()
