package rpg

var Primortal_Nightbane = Primortal{
	Type:                "nightbane",
	Name:                "Nightbane",
	BaseSync:            18,
	BasePhysicalAttack:  10,
	BasePhysicalDefence: 4,
	BaseAetherAttack:    18,
	BaseAetherDefence:   11,
	BodyType:            BodyTypeAbyssal,
	Affinity:            SkillTypeAbyssal,
}.register()

var Primortal_Darkvein = Primortal{
	Type:                "darkvein",
	Name:                "Darkvein",
	BaseSync:            16,
	BasePhysicalAttack:  9,
	BasePhysicalDefence: 6,
	BaseAetherAttack:    22,
	BaseAetherDefence:   15,
	BodyType:            BodyTypeAbyssal,
	Affinity:            SkillTypeGamma,
}.register()
