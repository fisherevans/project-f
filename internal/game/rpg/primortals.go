package rpg

var Primortals = map[PrimortalType]Primortal{}

type PrimortalType string

func (pt PrimortalType) Primortal() Primortal {
	if p, exists := Primortals[pt]; exists {
		return p
	}
	panic("unknown primortal type: " + pt)
}

type Primortal struct {
	Type PrimortalType
	Name string

	BaseSync int
}

func (p Primortal) register() Primortal {
	if _, exists := Primortals[p.Type]; exists {
		panic("duplicate primortal type: " + p.Type)
	}
	Primortals[p.Type] = p
	return p
}

var Primortal_Blob = Primortal{
	Type:     "blob",
	Name:     "blob",
	BaseSync: 10,
}.register()
