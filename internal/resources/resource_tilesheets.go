package resources

import (
	"fisherevans.com/project/f/internal/util"
	"fmt"
	"github.com/gopxl/pixel/v2"
	"github.com/rs/zerolog/log"
	_ "image/png"
	"regexp"
	"strconv"
)

type Pixels int

func (p Pixels) Float() float64 {
	return float64(p)
}

func (p Pixels) Int() int {
	return int(p)
}

const (
	DefaultTileSize Pixels = 16
	MapTileSize     Pixels = DefaultTileSize

	SpriteAtlasSize Pixels = 2048
)

var (
	tilesheets       = map[string]*TilesheetMetadata{}
	tilesheetSprites = map[TilesheetSpriteId]*SpriteReference{}

	SpriteAtlas pixel.Picture
)

func GetTilesheetNames() []string {
	return util.SortedKeys(tilesheets)
}

func GetTilesheet(name string) *TilesheetMetadata {
	ts := tilesheets[name]
	if ts == nil {
		log.Error().Msgf("missing tilesheet: %s", name)
	}
	return ts
}

func GetTilesheetSprite(tilesheet string, col, row int) *SpriteReference {
	return GetTilesheetSpriteById(TilesheetSpriteId{
		Tilesheet: tilesheet,
		Column:    col,
		Row:       row,
	})
}

func GetTilesheetSpriteById(id TilesheetSpriteId) *SpriteReference {
	result := tilesheetSprites[id]
	if result == nil {
		log.Error().Str("tilesheet", id.Tilesheet).Int("row", id.Row).Int("col", id.Column).Msg("tilesheet sprite not found")
	}
	return result
}

type TilesheetMetadata struct {
	Name       string
	TileHeight Pixels
	TileWidth  Pixels
	Columns    int
	Rows       int
	Sprites    map[TilesheetSpriteId]*SpriteReference
}

type TilesheetSpriteId struct {
	Tilesheet string `json:"tilesheet"`
	Column    int    `json:"col"`
	Row       int    `json:"row"`
}

func (s TilesheetSpriteId) String() string {
	return fmt.Sprintf("ts:%s,c:%d,r:%d", s.Tilesheet, s.Column, s.Row)
}

var tilesheetSizeRegex = regexp.MustCompile(`^(.*)-(\d+)x(\d+)$`)

func parseTilesheetName(input string) (string, Pixels, Pixels) {
	m := tilesheetSizeRegex.FindStringSubmatch(input)
	if len(m) == 4 {
		base := m[1]
		x, _ := strconv.Atoi(m[2])
		y, _ := strconv.Atoi(m[3])
		return base, Pixels(x), Pixels(y)
	}
	return input, DefaultTileSize, DefaultTileSize
}

func sliceTilesheet(atlas pixel.Picture, bounds pixel.Rect, resourceName string, tileWidth Pixels, tileHeight Pixels) {
	sheetWidth := bounds.W()
	sheetHeight := bounds.H()

	tilesheet := &TilesheetMetadata{
		Name:       resourceName,
		TileWidth:  tileWidth,
		TileHeight: tileHeight,
		Rows:       int(sheetHeight / float64(tileHeight)),
		Columns:    int(sheetWidth / float64(tileWidth)),
		Sprites:    map[TilesheetSpriteId]*SpriteReference{},
	}
	tilesheets[resourceName] = tilesheet

	for y := 0; y < tilesheet.Rows; y++ {
		for x := 0; x < tilesheet.Columns; x++ {
			spriteId := TilesheetSpriteId{
				Column:    x + 1,
				Row:       tilesheet.Rows - y,
				Tilesheet: tilesheet.Name,
			}
			posX := bounds.Min.X + (float64(x) * tilesheet.TileWidth.Float())
			posY := bounds.Min.Y + (float64(y) * tilesheet.TileHeight.Float())
			r := pixel.R(posX, posY, posX+tilesheet.TileWidth.Float(), posY+tilesheet.TileHeight.Float())
			spriteRef := &SpriteReference{
				Source: atlas,
				Bounds: r,
				Sprite: pixel.NewSprite(atlas, r),
			}
			tilesheet.Sprites[spriteId] = spriteRef
			tilesheetSprites[spriteId] = spriteRef
		}
	}

	log.Info().Msgf("sliced tilesheet %s (%dx%d) with %d columns and %d rows", resourceName, tilesheet.TileWidth, tilesheet.TileHeight, tilesheet.Columns, tilesheet.Rows)
}
