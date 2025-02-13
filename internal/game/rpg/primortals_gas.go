package rpg

var Primortal_Nebulith = Primortal{
	Type:                "nebulith",
	Name:                "Nebulith",
	BaseSync:            12,
	BasePhysicalAttack:  4,
	BasePhysicalDefence: 19,
	BaseAetherAttack:    20,
	BaseAetherDefence:   14,
	BodyType:            BodyTypeGas,
	Affinity:            SkillTypeSonic,
}.register()

var Primortal_Fumegast = Primortal{
	Type:                "fumegast",
	Name:                "Fumegast",
	BaseSync:            14,
	BasePhysicalAttack:  7,
	BasePhysicalDefence: 2,
	BaseAetherAttack:    15,
	BaseAetherDefence:   8,
	BodyType:            BodyTypeGas,
	Affinity:            SkillTypeAcidic,
}.register()
