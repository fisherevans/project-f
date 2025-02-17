package rpg

var Primortal_Quarrok = Primortal{
	Type:                "quarrok",
	Name:                "Quarrok",
	BaseSync:            22,
	BasePhysicalAttack:  17,
	BasePhysicalDefense: 20,
	BaseAetherAttack:    5,
	BaseAetherDefense:   7,
	BodyType:            BodyTypeRock,
	Affinity:            SkillTypeKinetic,
}.register()

var Primortal_Lavastone = Primortal{
	Type:                "lavastone",
	Name:                "Lavastone",
	BaseSync:            20,
	BasePhysicalAttack:  10,
	BasePhysicalDefense: 19,
	BaseAetherAttack:    12,
	BaseAetherDefense:   15,
	BodyType:            BodyTypeRock,
	Affinity:            SkillTypeThermal,
}.register()
