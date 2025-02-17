package rpg

var Primortals = map[PrimortalType]Primortal{}

type PrimortalType string

func (pt PrimortalType) Primortal() Primortal {
	if p, exists := Primortals[pt]; exists {
		return p
	}
	panic("unknown primortal type: " + pt)
}

type BodyType string

const (
	BodyTypeOrganic   BodyType = "organic"
	BodyTypeMetal     BodyType = "metal"
	BodyTypeRock      BodyType = "rock"
	BodyTypeCrystal   BodyType = "crystal"
	BodyTypeSynthetic BodyType = "synthetic"
	BodyTypeLiquid    BodyType = "liquid"
	BodyTypeGas       BodyType = "gas"
	BodyTypeAbyssal   BodyType = "abyssal"
)

type Primortal struct {
	Type PrimortalType
	Name string

	BaseSync            int
	BasePhysicalAttack  int
	BasePhysicalDefense int
	BaseAetherAttack    int
	BaseAetherDefense   int

	BodyType BodyType
	Affinity SkillType
}

func (p Primortal) register() Primortal {
	if _, exists := Primortals[p.Type]; exists {
		panic("duplicate primortal type: " + p.Type)
	}
	Primortals[p.Type] = p
	return p
}
