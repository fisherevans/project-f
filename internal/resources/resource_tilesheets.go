package resources

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"image"
	_ "image/png"
)

var (
	tilesheets = map[string]*SpriteTilesheet{}
)

func GetTilesheet(name string) *SpriteTilesheet {
	ts := tilesheets[name]
	if ts == nil {
		log.Error().Msgf("missing tilesheet: %s", name)
	}
	return ts
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
