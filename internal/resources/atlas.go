package resources

import (
	"fisherevans.com/project/f/internal/util/pixelutil"
	"fmt"
	"github.com/gopxl/pixel/v2"
	"github.com/rs/zerolog/log"
	"image"
	"image/draw"
	"image/png"
	"os"
	"path"
	"sort"
	"strings"
)

var (
	maxSpriteAtlasSize Pixels = 2048
)

type Atlas struct {
	source           *pixel.PictureData
	sprites          map[string]pixelutil.BoundedDrawable
	tilesheetSprites map[TilesheetSpriteId]pixelutil.BoundedDrawable
	frameSprites     map[FrameSpriteId]pixelutil.BoundedDrawable
	fonts            map[string]FontInstance
}

func (a *Atlas) GetSprite(name string) pixelutil.BoundedDrawable {
	sprite, exists := a.sprites[name]
	if !exists {
		log.Fatal().Str("sprite", name).Msg("sprite not found")
	}
	return sprite
}

func (a *Atlas) GetTilesheetSprite(tilesheet string, col, row int) pixelutil.BoundedDrawable {
	return a.GetTilesheetSpriteById(TilesheetSpriteId{
		Tilesheet: tilesheet,
		Column:    col,
		Row:       row,
	})
}

func (a *Atlas) GetTilesheetSpriteById(id TilesheetSpriteId) pixelutil.BoundedDrawable {
	result, exists := a.tilesheetSprites[id]
	if !exists {
		log.Error().Str("tilesheet", id.Tilesheet).Int("row", id.Row).Int("col", id.Column).Msg("tilesheet sprite not found")
	}
	return result
}

func (a *Atlas) GetFrameSprite(frame string, side FrameSide) pixelutil.BoundedDrawable {
	return a.GetFrameSpriteById(FrameSpriteId{
		Frame: frame,
		Side:  side,
	})
}

func (a *Atlas) GetFrameSpriteById(id FrameSpriteId) pixelutil.BoundedDrawable {
	result, exists := a.frameSprites[id]
	if !exists {
		log.Fatal().Str("frame", id.Frame).Str("side", string(id.Side)).Msg("frame sprite not found")
	}
	return result
}

func (a *Atlas) GetFont(name string) FontInstance {
	font, exists := a.fonts[name]
	if !exists {
		log.Error().Str("font", name).Msg("font not found")
	}
	return font
}

type AtlasFilter struct {
	SpritePrefixes []string
	FontNames      []string
}

func CreateAtlas(filter AtlasFilter) *Atlas {
	var atlasedListeners []func(*Atlas, pixel.Rect)
	var images []image.Image

	for spriteName, sprite := range spriteResources {
		include := len(filter.SpritePrefixes) == 0
		for _, prefix := range filter.SpritePrefixes {
			if strings.HasPrefix(spriteName, prefix) {
				include = true
				break
			}
		}
		if !include {
			continue
		}
		atlasedListeners = append(atlasedListeners, func(atlas *Atlas, placement pixel.Rect) {
			if sprite.metadata.Tilesheet != nil {
				tilesheet := sprite.metadata.Tilesheet
				for y := 0; y < tilesheet.Rows; y++ {
					for x := 0; x < tilesheet.Columns; x++ {
						spriteId := TilesheetSpriteId{
							Tilesheet: spriteName,
							Column:    x + 1,
							Row:       tilesheet.Rows - y,
						}
						posX := placement.Min.X + (float64(x) * tilesheet.TileWidth.Float())
						posY := placement.Min.Y + (float64(y) * tilesheet.TileHeight.Float())
						r := pixel.R(posX, posY, posX+tilesheet.TileWidth.Float(), posY+tilesheet.TileHeight.Float())
						atlas.tilesheetSprites[spriteId] = pixelutil.DrawableSprite(pixel.NewSprite(atlas.source, r))
					}
				}
				return
			}
			if sprite.metadata.Frame != nil {
				sf := sprite.metadata.Frame
				registerSide := func(side FrameSide, rect pixel.Rect) {
					frameId := FrameSpriteId{
						Frame: spriteName,
						Side:  side,
					}
					atlas.frameSprites[frameId] = pixelutil.DrawableSprite(pixel.NewSprite(atlas.source, rect))
				}
				top := float64(sf.CutMargin[FrameTop])
				left := float64(sf.CutMargin[FrameLeft])
				bottom := float64(sf.CutMargin[FrameBottom])
				right := float64(sf.CutMargin[FrameRight])
				registerSide(FrameTopLeft, pixel.R(placement.Min.X, placement.Max.Y-top, placement.Min.X+left, placement.Max.Y))
				registerSide(FrameTop, pixel.R(placement.Min.X+left, placement.Max.Y-top, placement.Max.X-right, placement.Max.Y))
				registerSide(FrameTopRight, pixel.R(placement.Max.X-right, placement.Max.Y-top, placement.Max.X, placement.Max.Y))
				registerSide(FrameLeft, pixel.R(placement.Min.X, placement.Min.Y+bottom, placement.Min.X+left, placement.Max.Y-top))
				registerSide(FrameMiddle, pixel.R(placement.Min.X+left, placement.Min.Y+bottom, placement.Max.X-right, placement.Max.Y-top))
				registerSide(FrameRight, pixel.R(placement.Max.X-right, placement.Min.Y+bottom, placement.Max.X, placement.Max.Y-top))
				registerSide(FrameBottomLeft, pixel.R(placement.Min.X, placement.Min.Y, placement.Min.X+left, placement.Min.Y+bottom))
				registerSide(FrameBottom, pixel.R(placement.Min.X+left, placement.Min.Y, placement.Max.X-right, placement.Min.Y+bottom))
				registerSide(FrameBottomRight, pixel.R(placement.Max.X-right, placement.Min.Y, placement.Max.X, placement.Min.Y+bottom))
				return
			}
			atlas.sprites[spriteName] = pixelutil.DrawableSprite(pixel.NewSprite(atlas.source, placement))

		})
		images = append(images, sprite.data)
	}

	for _, fontName := range filter.FontNames {
		instance := CreateFont(fontName)
		images = append(images, instance.Atlas.PictureDataCopy().Image())
		atlasedListeners = append(atlasedListeners, func(atlas *Atlas, placement pixel.Rect) {
			instance.Atlas = instance.Atlas.CloneWithPictureData(atlas.source, placement)
			atlas.fonts[fontName] = instance
		})
	}

	atlasImage, placements := createAtlasGuillotine(images, maxSpriteAtlasSize, maxSpriteAtlasSize)

	atlas := &Atlas{
		source:           pixel.PictureDataFromImage(atlasImage),
		sprites:          map[string]pixelutil.BoundedDrawable{},
		tilesheetSprites: map[TilesheetSpriteId]pixelutil.BoundedDrawable{},
		frameSprites:     map[FrameSpriteId]pixelutil.BoundedDrawable{},
		fonts:            map[string]FontInstance{},
	}

	for id, listener := range atlasedListeners {
		listener(atlas, placements[id])
	}

	return atlas
}

func (a *Atlas) NewBatch() *pixel.Batch {
	return pixel.NewBatch(&pixel.TrianglesData{}, a.source)
}

func (a *Atlas) Dump(dir, name string) {
	f, err := os.Create(path.Join(dir, fmt.Sprintf("%s.png", name)))
	if err != nil {
		log.Error().Msgf("Failed to create atlas file %s: %v", name, err)
		return
	}
	defer f.Close()

	if err := png.Encode(f, a.source.Image()); err != nil {
		log.Error().Msgf("Failed to encode atlas file %s: %v", name, err)
	}
}

// createAtlasGuillotine tries to pack images into an atlas using a guillotine-like approach.
// It returns the atlas as an *image.RGBA plus the pixel.Rect for each image’s location.
// Larger-first packing is used to help efficiency.
func createAtlasGuillotine(sourceImages []image.Image, atlasWidth, atlasHeight Pixels) (*image.RGBA, []pixel.Rect) {
	// Sort images largest-to-smallest by area (optional but often helps)
	type indexedImage struct {
		image         image.Image
		originalIndex int
		area          int
	}
	indexedImages := make([]indexedImage, len(sourceImages))
	for id, sourceImage := range sourceImages {
		b := sourceImage.Bounds()
		w, h := b.Dx(), b.Dy()
		indexedImages[id] = indexedImage{sourceImage, id, w * h}
	}
	sort.Slice(indexedImages, func(i, j int) bool {
		return indexedImages[i].area > indexedImages[j].area
	})

	type rect struct {
		x, y, w, h int
	}

	// Create the atlasImage and init free rectangles
	atlasImage := image.NewRGBA(image.Rect(0, 0, atlasWidth.Int(), atlasHeight.Int()))
	availableRects := []rect{
		{0, 0, atlasWidth.Int(), atlasHeight.Int()}, // The entire space is free initially
	}

	// This will store the final placement for each image, indexed by original order
	placements := make([]pixel.Rect, len(sourceImages))

	// Helper function: find a rect that fits w x h, or return -1 if none
	findRect := func(w, h int) int {
		for i, fr := range availableRects {
			if w <= fr.w && h <= fr.h {
				return i
			}
		}
		return -1
	}

	for id, ii := range indexedImages {
		img := ii.image
		b := img.Bounds()
		w, h := b.Dx(), b.Dy()

		// Find a free rectangle that can fit this image
		fi := findRect(w, h)
		if fi == -1 {
			log.Fatal().Msgf("atlast is too small (%dx%d) for %d images. Was able to fit %d images before failing", atlasWidth, atlasHeight, len(sourceImages), id)
			return nil, nil
		}

		// Place the image in that free rectangle
		fr := availableRects[fi]
		// Draw onto atlasImage (top-left origin for the image package)
		draw.Draw(atlasImage, image.Rect(fr.x, fr.y, fr.x+w, fr.y+h), img, b.Min, draw.Over)

		// Convert to Pixel's bottom-left origin
		// Bottom-left in pixel space: (fr.x, atlasHeight - (fr.y + h))
		// Top-right in pixel space:   (fr.x + w, atlasHeight - fr.y)
		minX := float64(fr.x)
		minY := float64(atlasHeight.Int() - (fr.y + h))
		maxX := float64(fr.x + w)
		maxY := float64(atlasHeight.Int() - fr.y)
		placements[ii.originalIndex] = pixel.R(minX, minY, maxX, maxY)

		// Remove the used rect
		availableRects = append(availableRects[:fi], availableRects[fi+1:]...)

		// “Guillotine” split: create new availableRects to the right & below (if there is space)
		// Right split (if there's leftover width)
		if w < fr.w {
			availableRects = append(availableRects, rect{
				x: fr.x + w,
				y: fr.y,
				w: fr.w - w,
				h: fr.h,
			})
		}
		// Bottom split (if there's leftover height)
		if h < fr.h {
			availableRects = append(availableRects, rect{
				x: fr.x,
				y: fr.y + h,
				w: w,
				h: fr.h - h,
			})
		}
	}

	freeSpace := 0
	for _, r := range availableRects {
		freeSpace += r.w * r.h
	}
	totalSpace := atlasWidth.Int() * atlasHeight.Int()
	log.Info().Msgf("Atlas packed %d images, using %d%% of available pixels.", len(sourceImages), (totalSpace-freeSpace)*100/totalSpace)

	return atlasImage, placements
}
