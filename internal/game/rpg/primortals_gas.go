package rpg

var Primortal_Nebulith = Primortal{
	Type:                "nebulith",
	Name:                "Nebulith",
	BaseSync:            12,
	BasePhysicalAttack:  4,
	BasePhysicalDefense: 19,
	BaseAetherAttack:    20,
	BaseAetherDefense:   14,
	BodyType:            BodyTypeGas,
	Affinity:            SkillTypeSonic,
}.register()

var Primortal_Fumegast = Primortal{
	Type:                "fumegast",
	Name:                "Fumegast",
	BaseSync:            14,
	BasePhysicalAttack:  7,
	BasePhysicalDefense: 2,
	BaseAetherAttack:    15,
	BaseAetherDefense:   8,
	BodyType:            BodyTypeGas,
	Affinity:            SkillTypeAcidic,
}.register()
