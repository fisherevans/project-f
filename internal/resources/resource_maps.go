package resources

import (
	"encoding/json"
	"fmt"
	"github.com/rs/zerolog/log"
)

var (
	maps = map[string]*Map{}

	resourceMaps = LocalResource{
		FileRoot:        "maps",
		FileExtension:   "json",
		FileLoader:      unmarshaler(&maps, json.Unmarshal),
		ResourceEncoder: jsonEncoder,
	}
)

func GetMap(name string) *Map {
	m := maps[name]
	if m == nil {
		log.Error().Msgf("missing map: %s", name)
	}
	return m
}

type MapLayerName string

const (
	LayerBase      MapLayerName = "base"
	LayerDecor                  = "decor"
	LayerOverlay                = "overlay"
	LayerCollision              = "collision"
)

var MapLayerOrder = []MapLayerName{
	LayerBase,
	LayerDecor,
	LayerOverlay,
	LayerCollision,
}

type Map struct {
	Layers   map[MapLayerName]*Layer `json:"layers"`
	Entities map[string]*Entity      `json:"entities"`
}

func (m *Map) AddEntity(id string, e *Entity) {
	if m.Entities == nil {
		m.Entities = map[string]*Entity{}
	}
	m.Entities[id] = e
}

func (m *Map) RemoveEntity(id string) {
	delete(m.Entities, id)
}

type Layer struct {
	Tiles []*Tile `json:"tiles"`
}

type Tile struct {
	X        int               `json:"x"`
	Y        int               `json:"y"`
	SpriteId TilesheetSpriteId `json:"sprite_id"`
}

type Entity struct {
	X        int            `json:"x"`
	Y        int            `json:"y"`
	Type     string         `json:"type"`
	Metadata map[string]any `json:"metadata"`
}

func (e *Entity) Copy() *Entity {
	next := &Entity{}
	encoded, err := json.Marshal(e)
	if err != nil {
		log.Error().Msgf("Error marshalling Entity: %v", err)
		return next
	}
	if err := json.Unmarshal(encoded, next); err != nil {
		log.Error().Msgf("Error unmarshalling Entity: %v", err)
	}
	return next
}

func (e *Entity) GetStringMetadata(key, defaultValue string) string {
	if e.Metadata == nil {
		return defaultValue
	}
	value, ok := e.Metadata[key]
	if !ok {
		return defaultValue
	}
	str, ok := value.(string)
	if !ok {
		return defaultValue
	}
	return str
}

func (t Tile) String() string {
	return fmt.Sprintf("Tile{x:%d,y:%d,sprite:[%s]}", t.X, t.Y, t.SpriteId.String())
}

func SaveAllMaps() {
	for name := range maps {
		SaveMap(name)
	}
}

func SaveMap(resourceName string) {
	err := save(&maps, resourceMaps, resourceName)
	if err != nil {
		panic(fmt.Sprintf("failed save resource %s: %v", resourceName, err))
	}
}
