package rpg

var Primortal_Gearmaw = Primortal{
	Type:                "gearmaw",
	Name:                "Gearmaw",
	BaseSync:            18,
	BasePhysicalAttack:  20,
	BasePhysicalDefence: 24,
	BaseAetherAttack:    4,
	BaseAetherDefence:   3,
	BodyType:            BodyTypeMetal,
	Affinity:            SkillTypeKinetic,
}.register()

var Primortal_Stormcore = Primortal{
	Type:                "stormcore",
	Name:                "Stormcore",
	BaseSync:            15,
	BasePhysicalAttack:  6,
	BasePhysicalDefence: 12,
	BaseAetherAttack:    18,
	BaseAetherDefence:   21,
	BodyType:            BodyTypeMetal,
	Affinity:            SkillTypeVoltaic,
}.register()
