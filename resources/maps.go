package resources

import (
	"fmt"
)

var (
	Maps = map[string]Map{}

	resourceMaps = LocalResource{
		FileRoot:        "maps",
		FileExtension:   "json",
		FileLoader:      jsonLoader(&Maps),
		ResourceEncoder: jsonEncoder,
	}
)

type MapLayerName string

const (
	LayerBase      MapLayerName = "base"
	LayerDecor                  = "decor"
	LayerOverlay                = "overlay"
	LayerCollision              = "collision"
	LayerEntities               = "entities"
)

var MapLayerOrder = []MapLayerName{
	LayerBase,
	LayerDecor,
	LayerEntities,
	LayerOverlay,
	LayerCollision,
}

type Map struct {
	Layers map[MapLayerName]*Layer `json:"layers"`
}

type Layer struct {
	Tiles []*Tile `json:"tiles"`
}

type Tile struct {
	X        int      `json:"x"`
	Y        int      `json:"y"`
	SpriteId SpriteId `json:"sprite_id"`
}

func (t Tile) String() string {
	return fmt.Sprintf("Tile{x:%d,y:%d,sprite:[%s]}", t.X, t.Y, t.SpriteId.String())
}

func SaveAllMaps() {
	for name := range Maps {
		SaveMap(name)
	}
}

func SaveMap(resourceName string) {
	err := save(&Maps, resourceMaps, resourceName)
	if err != nil {
		panic(fmt.Sprintf("failed save resource %s: %v", resourceName, err))
	}
}
