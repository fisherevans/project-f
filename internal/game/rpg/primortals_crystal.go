package rpg

var Primortal_Shardskulk = Primortal{
	Type:                "shardskulk",
	Name:                "Shardskulk",
	BaseSync:            14,
	BasePhysicalAttack:  12,
	BasePhysicalDefence: 21,
	BaseAetherAttack:    9,
	BaseAetherDefence:   3,
	BodyType:            BodyTypeCrystal,
	Affinity:            SkillTypeSonic,
}.register()

var Primortal_Glimmerfang = Primortal{
	Type:                "glimmerfang",
	Name:                "Glimmerfang",
	BaseSync:            16,
	BasePhysicalAttack:  7,
	BasePhysicalDefence: 9,
	BaseAetherAttack:    14,
	BaseAetherDefence:   15,
	BodyType:            BodyTypeCrystal,
	Affinity:            SkillTypeMagnetic,
}.register()
