package rpg

type ItemType string

type Inventory struct {
	Items []*Item `yaml:"items"`
}

type Item struct {
	Type     ItemType `yaml:"type"`
	Quantity int      `yaml:"quantity"`
}

var (
	Item_Towel ItemType = "towel"
)
