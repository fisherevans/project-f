package resources

import (
	"bytes"
	"fmt"
	"github.com/gopxl/pixel/v2"
	"image"
	"image/draw"
	_ "image/png"
	"log"
	"slices"
)

const (
	TileSize        = 16
	TileSizeF64     = float64(TileSize)
	SpriteAtlasSize = 1024
)

var (
	Tilesheets = map[string]*TilesheetMetadata{}
	Sprites    = map[SpriteId]*SpriteReference{}

	resourceTilesheets = LocalResource{
		FileRoot:       "tilesheets",
		FileExtension:  "png",
		FileLoader:     loadTilesheet,
		PostProcessing: processTilesheets,
	}

	SpriteAtlas pixel.Picture
)

type TilesheetMetadata struct {
	Name     string
	Columns  int
	Rows     int
	Sprites  map[SpriteId]*SpriteReference
	rawImage image.Image
}

type SpriteReference struct {
	Source    pixel.Picture
	Bounds    pixel.Rect
	Sprite    *pixel.Sprite
	Tilesheet *TilesheetMetadata
}

type SpriteId struct {
	Tilesheet string `json:"tilesheet"`
	Column    int    `json:"col"`
	Row       int    `json:"row"`
}

func (s SpriteId) String() string {
	return fmt.Sprintf("ts:%s,c:%d,r:%d", s.Tilesheet, s.Column, s.Row)
}

func GetSprite(tilesheet string, col, row int) *SpriteReference {
	return Sprites[SpriteId{
		Tilesheet: tilesheet,
		Column:    col,
		Row:       row,
	}]
}

func loadTilesheet(path string, resourceName string, data []byte) error {
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		log.Printf("Failed to decode image %s: %v\n", path, err)
		return nil // Continue with the next file
	}
	pictureData := pixel.PictureDataFromImage(img)

	sheetWidth := pictureData.Bounds().W()
	sheetHeight := pictureData.Bounds().H()

	tilesheet := &TilesheetMetadata{
		Name:     resourceName,
		Sprites:  map[SpriteId]*SpriteReference{},
		Rows:     int(sheetHeight / TileSize),
		Columns:  int(sheetWidth / TileSize),
		rawImage: img,
	}

	Tilesheets[resourceName] = tilesheet
	fmt.Printf("loaded tilesheet %s with %d columns and %d rows\n", resourceName, tilesheet.Columns, tilesheet.Rows)

	return nil
}

func processTilesheets() error {
	tilesheetBounds := map[string]*pixel.Rect{}

	atlasImage := image.NewRGBA(image.Rect(0, 0, int(SpriteAtlasSize), int(SpriteAtlasSize)))
	xOffset := 0
	yOffset := 0
	maxY := 0

	var tilesheetNames []string
	for tilesheetName := range Tilesheets {
		tilesheetNames = append(tilesheetNames, tilesheetName)
	}
	slices.Sort(tilesheetNames)

	for _, tilesheetName := range tilesheetNames {
		tilesheetImage := Tilesheets[tilesheetName].rawImage
		bounds := tilesheetImage.Bounds()
		if xOffset+int(bounds.Dx()) > SpriteAtlasSize {
			xOffset = 0
			yOffset = maxY
		}
		if yOffset+int(bounds.Dy()) > SpriteAtlasSize {
			panic("SpriteAtlasSize is too small")
		}
		xStart := xOffset
		xEnd := xStart + int(bounds.Dx())
		yStart := yOffset
		yEnd := yStart + int(bounds.Dy())
		if yEnd > maxY {
			maxY = yEnd
		}
		rect := image.Rect(xStart, yStart, xEnd, yEnd)

		draw.Draw(atlasImage, rect, tilesheetImage, image.Pt(int(bounds.Min.X), int(bounds.Min.Y)), draw.Over)

		// image.* treats top left as origin, pixel.* treats bottom left as origin
		// swap to pixel rect here because sprite generation uses pixel.Picture
		thisTilesheetBounds := pixel.R(float64(xStart), float64(SpriteAtlasSize)-float64(yEnd), float64(xEnd), float64(SpriteAtlasSize)-float64(yStart))
		tilesheetBounds[tilesheetName] = &thisTilesheetBounds

		xOffset += int(bounds.Dx())
	}

	SpriteAtlas = pixel.PictureDataFromImage(atlasImage)
	for _, tilesheetMetadata := range Tilesheets {
		tsBounds := tilesheetBounds[tilesheetMetadata.Name]

		for y := 0; y < tilesheetMetadata.Rows; y++ {
			for x := 0; x < tilesheetMetadata.Columns; x++ {
				spriteId := SpriteId{
					Column:    x + 1,
					Row:       tilesheetMetadata.Rows - y,
					Tilesheet: tilesheetMetadata.Name,
				}
				posX := tsBounds.Min.X + (float64(x) * TileSizeF64)
				posY := tsBounds.Min.Y + (float64(y) * TileSizeF64)
				r := pixel.R(posX, posY, posX+TileSizeF64, posY+TileSizeF64)
				spriteRef := &SpriteReference{
					Source:    SpriteAtlas,
					Bounds:    r,
					Sprite:    pixel.NewSprite(SpriteAtlas, r),
					Tilesheet: tilesheetMetadata,
				}
				tilesheetMetadata.Sprites[spriteId] = spriteRef
				Sprites[spriteId] = spriteRef
			}
		}
	}

	return nil
}
