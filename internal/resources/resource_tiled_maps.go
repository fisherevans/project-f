package resources

import (
	"fmt"
	"os"
	"slices"

	"github.com/lafriks/go-tiled"
	"github.com/rs/zerolog/log"

	"fisherevans.com/project/f/assets"
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

	loggedTiles := map[string]bool{}
	tileToSpriteId := func(tileId uint32, tileset *tiled.Tileset) TilesheetSpriteId {
		tileSetX := int(tileId) % tileset.Columns
		tileSetY := int(tileId) / tileset.Columns
		id := TilesheetSpriteId{
			Tilesheet: tileset.Name,
			Column:    tileSetX + 1,
			Row:       tileSetY + 1,
		}
		tiledName := fmt.Sprintf("%s/%d", tileset.Name, tileId)
		if !loggedTiles[tiledName] {
			loggedTiles[tiledName] = true
			log.Debug().Msgf("Tiled %#v // GID: %d", id, tileId)
		}
		return id
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
			gameTile := &Tile{
				X:        tileId % tiledMap.Width,
				Y:        tiledMap.Height - (tileId / tiledMap.Width),
				SpriteId: tileToSpriteId(tiledTile.ID, tiledTile.Tileset),
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
			for _, property := range object.Properties {
				metadata[property.Name] = property.Value // TODO non string types?
			}

			e := &Entity{
				ID:         int(object.ID),
				X:          int((object.X + float64(tiledMap.TileWidth)/2) / float64(tiledMap.TileWidth)),
				Y:          tiledMap.Height - int((object.Y-float64(tiledMap.TileHeight)/2)/float64(tiledMap.TileHeight)),
				Properties: metadata,
			}

			if tiledTile, err := tiledMap.TileGIDToTile(object.GID); err == nil {
				e.SpriteId = tileToSpriteId(tiledTile.ID, tiledTile.Tileset)
			}

			gameMap.Entities[fmt.Sprintf("tiled-%d", object.ID)] = e
		}
	}

	if _, exists := maps[resourceName]; exists {
		log.Fatal().Msgf("Map already exists with name %s", resourceName)
	}
	maps[resourceName] = gameMap
	return nil
}
