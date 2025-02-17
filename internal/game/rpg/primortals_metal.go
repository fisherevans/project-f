package rpg

var Primortal_Gearmaw = Primortal{
	Type:                "gearmaw",
	Name:                "Gearmaw",
	BaseSync:            18,
	BasePhysicalAttack:  20,
	BasePhysicalDefense: 24,
	BaseAetherAttack:    4,
	BaseAetherDefense:   3,
	BodyType:            BodyTypeMetal,
	Affinity:            SkillTypeKinetic,
}.register()

var Primortal_Stormcore = Primortal{
	Type:                "stormcore",
	Name:                "Stormcore",
	BaseSync:            15,
	BasePhysicalAttack:  6,
	BasePhysicalDefense: 12,
	BaseAetherAttack:    18,
	BaseAetherDefense:   21,
	BodyType:            BodyTypeMetal,
	Affinity:            SkillTypeVoltaic,
}.register()
