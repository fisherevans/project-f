package rpg

var Primortal_Fangroot = Primortal{
	Type:                "fangroot",
	Name:                "Fangroot",
	BaseSync:            12,
	BasePhysicalAttack:  15,
	BasePhysicalDefence: 13,
	BaseAetherAttack:    5,
	BaseAetherDefence:   7,
	BodyType:            BodyTypeOrganic,
	Affinity:            SkillTypeKinetic,
}.register()

var Primortal_Viraglow = Primortal{
	Type:                "viraglow",
	Name:                "Viraglow",
	BaseSync:            14,
	BasePhysicalAttack:  8,
	BasePhysicalDefence: 10,
	BaseAetherAttack:    12,
	BaseAetherDefence:   15,
	BodyType:            BodyTypeOrganic,
	Affinity:            SkillTypeGamma,
}.register()
