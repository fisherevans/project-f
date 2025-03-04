package resources

import (
	"fisherevans.com/project/f/assets"
	"fmt"
	"github.com/lafriks/go-tiled"
	"github.com/rs/zerolog/log"
	"os"
	"slices"
)

func loadTiledMap(path string, resourceName string, _ []byte) error {
	var tiledMap *tiled.Map
	var err error
	tiledMap, err = tiled.LoadFile(path, tiled.WithFileSystem(assets.FS))
	if err != nil {
		fmt.Printf("error parsing tiledMap: %s", err.Error())
		os.Exit(2)
	}

	log := log.With().Str("path", path).Logger()

	gameMap := &Map{
		Layers:   map[MapLayerName]*Layer{},
		Entities: map[string]*Entity{},
	}

	for _, tiledLayer := range tiledMap.Layers {
		if !slices.Contains(MapLayerOrder, MapLayerName(tiledLayer.Name)) {
			log.Warn().Msgf("Skipping unknown layer %s", tiledLayer.Name)
			continue
		}
		gameLayer := &Layer{}

		for tileId, tiledTile := range tiledLayer.Tiles {
			if tiledTile == nil || tiledTile.IsNil() || tiledTile.Tileset == nil {
				continue
			}
			tileSetX := int(tiledTile.ID) % tiledTile.Tileset.Columns
			tileSetY := int(tiledTile.ID) / tiledTile.Tileset.Columns
			gameTile := &Tile{
				X: tileId % tiledMap.Width,
				Y: tiledMap.Height - (tileId / tiledMap.Width),
				SpriteId: TilesheetSpriteId{
					Tilesheet: tiledTile.Tileset.Name,
					Column:    tileSetX + 1,
					Row:       tileSetY + 1,
				},
			}
			gameLayer.Tiles = append(gameLayer.Tiles, gameTile)
		}
		gameMap.Layers[MapLayerName(tiledLayer.Name)] = gameLayer
	}

	for _, objectGroup := range tiledMap.ObjectGroups {
		if objectGroup.Name != "entities" {
			log.Warn().Msgf("Skipping unknown object group %s", objectGroup.Name)
			continue
		}
		for _, object := range objectGroup.Objects {
			metadata := map[string]any{}
			var entityType string
			for _, property := range object.Properties {
				if property.Name == "type" {
					entityType = property.Value
				} else {
					metadata[property.Name] = property.Value // TODO non string types?
				}
			}

			if entityType == "" {
				log.Warn().Msgf("Skipping entity with no type: %d", object.ID)
				continue
			}
			gameMap.Entities[fmt.Sprintf("tiled-%d", object.ID)] = &Entity{
				X:        int((object.X + float64(tiledMap.TileWidth)/2) / float64(tiledMap.TileWidth)),
				Y:        tiledMap.Height - int((object.Y-float64(tiledMap.TileHeight)/2)/float64(tiledMap.TileHeight)),
				Type:     entityType,
				Metadata: metadata,
			}
		}
	}

	if _, exists := maps[resourceName]; exists {
		log.Fatal().Msgf("Map already exists with name %s", resourceName)
	}
	maps[resourceName] = gameMap
	return nil
}
