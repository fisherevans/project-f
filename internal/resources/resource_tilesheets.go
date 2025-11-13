package resources

import (
	"fmt"
	"image"
	_ "image/png"

	"fisherevans.com/project/f/internal/util/pixelutil"
	"github.com/rs/zerolog/log"
)

var (
	tilesheets       = map[string]*SpriteTilesheet{}
	tilesheetSprites = map[string]map[string]*TilesheetCoordinates{}
)

func GetTilesheet(name string) *SpriteTilesheet {
	ts := tilesheets[name]
	if ts == nil {
		log.Error().Msgf("missing tilesheet: %s", name)
	}
	return ts
}

func GetTilesheetNamedSprites(tilesheetName string) []string {
	nameSprites, ok := tilesheetSprites[tilesheetName]
	if !ok {
		log.Fatal().Msgf("tilesheet %s not found", tilesheetName)
	}
	var names []string
	for name := range nameSprites {
		names = append(names, name)
	}
	return names
}

func GetNamedTilesheetSpriteId(tilesheetName, spriteName string) TilesheetSpriteId {
	nameSprites, ok := tilesheetSprites[tilesheetName]
	if !ok {
		log.Fatal().Msgf("tilesheet %s not found", tilesheetName)
	}
	sprite, ok := nameSprites[spriteName]
	if !ok {
		log.Fatal().Msgf("named sprite %s not found in tilesheet %s", spriteName, tilesheetName)
	}
	return TilesheetSpriteId{
		Tilesheet: tilesheetName,
		Column:    sprite.Column,
		Row:       sprite.Row,
	}
}

type SpriteTilesheet struct {
	TileWidth  Pixels `yaml:"tileWidth"`
	TileHeight Pixels `yaml:"tileHeight"`

	// set when loaded
	Columns int `yaml:"-"`
	Rows    int `yaml:"-"`
}

func (t *SpriteTilesheet) init(img image.Image) {
	t.Columns = img.Bounds().Dx() / t.TileWidth.Int()
	t.Rows = img.Bounds().Dy() / t.TileHeight.Int()
}

type TilesheetSpriteId struct {
	Tilesheet string `json:"tilesheet"`
	Column    int    `json:"col"`
	Row       int    `json:"row"`
}

func (s TilesheetSpriteId) String() string {
	return fmt.Sprintf("ts:%s,c:%d,r:%d", s.Tilesheet, s.Column, s.Row)
}

func (s TilesheetSpriteId) From(a *Atlas) pixelutil.BoundedDrawable {
	return a.GetTilesheetSpriteById(s)
}

type TilesheetCoordinates struct {
	Row    int `yaml:"row"`
	Column int `yaml:"column"`
}
