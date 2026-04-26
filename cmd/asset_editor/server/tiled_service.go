package server

import (
	"encoding/xml"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

type TiledService struct {
	assetsDir string
}

func NewTiledService(assetsDir string) *TiledService {
	return &TiledService{assetsDir: assetsDir}
}

type TiledEntityRef struct {
	MapFile    string            `json:"mapFile"`
	ObjectID   int               `json:"objectId"`
	EntityID   string            `json:"entityId,omitempty"`
	ObjectType string            `json:"objectType,omitempty"`
	X          float64           `json:"x"`
	Y          float64           `json:"y"`
	Properties map[string]string `json:"properties"`
}

type TiledHandlerUsage struct {
	HandlerName string           `json:"handlerName"`
	Entities    []TiledEntityRef `json:"entities"`
}

type tmxMap struct {
	XMLName    xml.Name       `xml:"map"`
	ObjectGroups []tmxObjectGroup `xml:"objectgroup"`
}

type tmxObjectGroup struct {
	Objects []tmxObject `xml:"object"`
}

type tmxObject struct {
	ID         int            `xml:"id,attr"`
	Type       string         `xml:"type,attr"`
	X          float64        `xml:"x,attr"`
	Y          float64        `xml:"y,attr"`
	Properties tmxProperties  `xml:"properties"`
}

type tmxProperties struct {
	Props []tmxProperty `xml:"property"`
}

type tmxProperty struct {
	Name  string `xml:"name,attr"`
	Value string `xml:"value,attr"`
	Inner string `xml:",chardata"`
}

func (t *TiledService) ListHandlerUsages() ([]TiledHandlerUsage, error) {
	mapsDir := filepath.Join(t.assetsDir, "tiled_maps")
	usageMap := map[string][]TiledEntityRef{}

	err := filepath.WalkDir(mapsDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() || !strings.HasSuffix(d.Name(), ".tmx") {
			return nil
		}
		if strings.Contains(path, "/rules/") {
			return nil
		}

		data, readErr := os.ReadFile(path)
		if readErr != nil {
			return nil
		}

		rel, _ := filepath.Rel(t.assetsDir, path)

		var m tmxMap
		if xmlErr := xml.Unmarshal(data, &m); xmlErr != nil {
			return nil
		}

		for _, group := range m.ObjectGroups {
			for _, obj := range group.Objects {
				props := map[string]string{}
				var scriptRef string
				var entityID string
				for _, p := range obj.Properties.Props {
					val := p.Value
					if val == "" {
						val = strings.TrimSpace(p.Inner)
					}
					props[p.Name] = val
					if p.Name == "script_ref" {
						scriptRef = val
					}
					if p.Name == "entity_id" {
						entityID = val
					}
				}
				if scriptRef == "" {
					continue
				}
				delete(props, "script_ref")
				ref := TiledEntityRef{
					MapFile:    rel,
					ObjectID:   obj.ID,
					EntityID:   entityID,
					ObjectType: obj.Type,
					X:          obj.X,
					Y:          obj.Y,
					Properties: props,
				}
				usageMap[scriptRef] = append(usageMap[scriptRef], ref)
			}
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walking tiled maps: %w", err)
	}

	var result []TiledHandlerUsage
	for name, entities := range usageMap {
		result = append(result, TiledHandlerUsage{
			HandlerName: name,
			Entities:    entities,
		})
	}
	return result, nil
}
