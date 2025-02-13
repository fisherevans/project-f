package resources

import (
	"bytes"
	"fmt"
	"github.com/gopxl/pixel/v2"
	"github.com/rs/zerolog/log"
	"image"
	_ "image/png"
)

const (
	TileSize        = 16
	TileSizeF64     = float64(TileSize)
	SpriteAtlasSize = 1024
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
	Name     string
	Columns  int
	Rows     int
	Sprites  map[TilesheetSpriteId]*SpriteReference
	rawImage image.Image
}

type TilesheetSpriteId struct {
	Tilesheet string `json:"tilesheet"`
	Column    int    `json:"col"`
	Row       int    `json:"row"`
}

func (s TilesheetSpriteId) String() string {
	return fmt.Sprintf("ts:%s,c:%d,r:%d", s.Tilesheet, s.Column, s.Row)
}

func GetSprite(tilesheet string, col, row int) *SpriteReference {
	return TilesheetSprites[TilesheetSpriteId{
		Tilesheet: tilesheet,
		Column:    col,
		Row:       row,
	}]
}

func loadTilesheet(path string, resourceName string, data []byte) error {
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		log.Error().Msgf("Failed to decode image %s: %v", path, err)
		return nil // Continue with the next file
	}

	addAtlasImage(&img, func(atlas pixel.Picture, bounds pixel.Rect) {
		pictureData := pixel.PictureDataFromImage(img)

		sheetWidth := pictureData.Bounds().W()
		sheetHeight := pictureData.Bounds().H()

		tilesheet := &TilesheetMetadata{
			Name:     resourceName,
			Sprites:  map[TilesheetSpriteId]*SpriteReference{},
			Rows:     int(sheetHeight / TileSize),
			Columns:  int(sheetWidth / TileSize),
			rawImage: img,
		}

		Tilesheets[resourceName] = tilesheet
		log.Info().Msgf("loaded tilesheet %s with %d columns and %d rows", resourceName, tilesheet.Columns, tilesheet.Rows)

		for y := 0; y < tilesheet.Rows; y++ {
			for x := 0; x < tilesheet.Columns; x++ {
				spriteId := TilesheetSpriteId{
					Column:    x + 1,
					Row:       tilesheet.Rows - y,
					Tilesheet: tilesheet.Name,
				}
				posX := bounds.Min.X + (float64(x) * TileSizeF64)
				posY := bounds.Min.Y + (float64(y) * TileSizeF64)
				r := pixel.R(posX, posY, posX+TileSizeF64, posY+TileSizeF64)
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
