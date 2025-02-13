package rpg

var Primortal_Polyvex = Primortal{
	Type:                "polyvex",
	Name:                "Polyvex",
	BaseSync:            18,
	BasePhysicalAttack:  9,
	BasePhysicalDefence: 10,
	BaseAetherAttack:    12,
	BaseAetherDefence:   15,
	BodyType:            BodyTypeSynthetic,
	Affinity:            SkillTypeAcidic,
}.register()

var Primortal_Cyronex = Primortal{
	Type:                "cyronex",
	Name:                "Cyronex",
	BaseSync:            16,
	BasePhysicalAttack:  8,
	BasePhysicalDefence: 10,
	BaseAetherAttack:    15,
	BaseAetherDefence:   23,
	BodyType:            BodyTypeSynthetic,
	Affinity:            SkillTypeGamma,
}.register()
