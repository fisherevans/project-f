package resources

import (
	"bytes"
	"fmt"
	"github.com/gopxl/pixel/v2"
	"github.com/rs/zerolog/log"
	"image"
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
	Tilesheets       = map[string]*TilesheetMetadata{}
	TilesheetSprites = map[TilesheetSpriteId]*SpriteReference{}

	resourceTilesheets = LocalResource{
		FileRoot:      "tilesheets",
		FileExtension: "png",
		FileLoader:    loadTilesheet,
	}

	SpriteAtlas pixel.Picture
)

type TilesheetMetadata struct {
	Name       string
	TileHeight Pixels
	TileWidth  Pixels
	Columns    int
	Rows       int
	Sprites    map[TilesheetSpriteId]*SpriteReference
	rawImage   image.Image
}

type TilesheetSpriteId struct {
	Tilesheet string `json:"tilesheet"`
	Column    int    `json:"col"`
	Row       int    `json:"row"`
}

func (s TilesheetSpriteId) String() string {
	return fmt.Sprintf("ts:%s,c:%d,r:%d", s.Tilesheet, s.Column, s.Row)
}

func GetTilesheetSprite(tilesheet string, col, row int) *SpriteReference {
	return TilesheetSprites[TilesheetSpriteId{
		Tilesheet: tilesheet,
		Column:    col,
		Row:       row,
	}]
}
func GetTilesheetSpriteId(tilesheet string, col, row int) TilesheetSpriteId {
	return TilesheetSpriteId{
		Tilesheet: tilesheet,
		Column:    col,
		Row:       row,
	}
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

func loadTilesheet(path string, resourceName string, data []byte) error {
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		log.Error().Msgf("Failed to decode image %s: %v", path, err)
		return nil // Continue with the next file
	}

	resourceName, tileWidth, tileHeight := parseTilesheetName(resourceName)

	addAtlasImage(&img, func(atlas pixel.Picture, bounds pixel.Rect) {
		pictureData := pixel.PictureDataFromImage(img)

		sheetWidth := pictureData.Bounds().W()
		sheetHeight := pictureData.Bounds().H()

		tilesheet := &TilesheetMetadata{
			Name:       resourceName,
			TileWidth:  tileWidth,
			TileHeight: tileHeight,
			Rows:       int(sheetHeight / float64(tileHeight)),
			Columns:    int(sheetWidth / float64(tileWidth)),
			rawImage:   img,
			Sprites:    map[TilesheetSpriteId]*SpriteReference{},
		}

		Tilesheets[resourceName] = tilesheet
		log.Info().Msgf("loaded tilesheet %s (%dx%d) with %d columns and %d rows", resourceName, tilesheet.TileWidth, tilesheet.TileHeight, tilesheet.Columns, tilesheet.Rows)

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
				TilesheetSprites[spriteId] = spriteRef
			}
		}
	})
	return nil
}
