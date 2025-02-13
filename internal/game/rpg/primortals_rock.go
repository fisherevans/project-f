package rpg

var Primortal_Quarrok = Primortal{
	Type:                "quarrok",
	Name:                "Quarrok",
	BaseSync:            22,
	BasePhysicalAttack:  17,
	BasePhysicalDefence: 20,
	BaseAetherAttack:    5,
	BaseAetherDefence:   7,
	BodyType:            BodyTypeRock,
	Affinity:            SkillTypeKinetic,
}.register()

var Primortal_Lavastone = Primortal{
	Type:                "lavastone",
	Name:                "Lavastone",
	BaseSync:            20,
	BasePhysicalAttack:  10,
	BasePhysicalDefence: 19,
	BaseAetherAttack:    12,
	BaseAetherDefence:   15,
	BodyType:            BodyTypeRock,
	Affinity:            SkillTypeThermal,
}.register()
