package resources

import (
	"github.com/gopxl/pixel/v2"
	"github.com/rs/zerolog/log"
	"image"
	"image/draw"
	"sort"
)

type atlasImage struct {
	source   *image.Image
	register func(pixel.Picture, pixel.Rect)
}

var (
	imagesToAtlas []atlasImage
)

func addAtlasImage(source *image.Image, register func(pixel.Picture, pixel.Rect)) {
	imagesToAtlas = append(imagesToAtlas, atlasImage{
		source:   source,
		register: register,
	})
}

func loadAtlas() {
	var images []image.Image
	for _, img := range imagesToAtlas {
		images = append(images, *img.source)
	}
	atlasSource, placements := createAtlasGuillotine(images, SpriteAtlasSize, SpriteAtlasSize)
	atlas := pixel.PictureDataFromImage(atlasSource)
	for i, img := range imagesToAtlas {
		img.register(atlas, placements[i])
	}
	SpriteAtlas = atlas
}

// createAtlasGuillotine tries to pack images into an atlas using a guillotine-like approach.
// It returns the atlas as an *image.RGBA plus the pixel.Rect for each image’s location.
// Larger-first packing is used to help efficiency.
func createAtlasGuillotine(imgs []image.Image, atlasWidth, atlasHeight Pixels) (*image.RGBA, []pixel.Rect) {
	// Sort images largest-to-smallest by area (optional but often helps)
	type indexedImage struct {
		img  image.Image
		idx  int
		area int
	}
	indexed := make([]indexedImage, len(imgs))
	for i, img := range imgs {
		b := img.Bounds()
		w, h := b.Dx(), b.Dy()
		indexed[i] = indexedImage{img, i, w * h}
	}
	sort.Slice(indexed, func(i, j int) bool {
		return indexed[i].area > indexed[j].area
	})

	type freeRect struct {
		x, y, w, h int
	}

	// Create the atlas and init free rectangles
	atlas := image.NewRGBA(image.Rect(0, 0, atlasWidth.Int(), atlasHeight.Int()))
	freeRects := []freeRect{
		{0, 0, atlasWidth.Int(), atlasHeight.Int()}, // The entire space is free initially
	}

	// This will store the final placement for each image, indexed by original order
	placements := make([]pixel.Rect, len(imgs))

	// Helper function: find a freeRect that fits w x h, or return -1 if none
	findRect := func(w, h int) int {
		for i, fr := range freeRects {
			if w <= fr.w && h <= fr.h {
				return i
			}
		}
		return -1
	}

	for id, ii := range indexed {
		img := ii.img
		b := img.Bounds()
		w, h := b.Dx(), b.Dy()

		// Find a free rectangle that can fit this image
		fi := findRect(w, h)
		if fi == -1 {
			log.Fatal().Msgf("atlast is too small (%dx%d) for %d images. Was able to fit %d images before failing", atlasWidth, atlasHeight, len(imgs), id)
			return nil, nil
		}

		// Place the image in that free rectangle
		fr := freeRects[fi]
		// Draw onto atlas (top-left origin for the image package)
		draw.Draw(atlas, image.Rect(fr.x, fr.y, fr.x+w, fr.y+h), img, image.Point{}, draw.Over)

		// Convert to Pixel's bottom-left origin
		// Bottom-left in pixel space: (fr.x, atlasHeight - (fr.y + h))
		// Top-right in pixel space:   (fr.x + w, atlasHeight - fr.y)
		minX := float64(fr.x)
		minY := float64(atlasHeight.Int() - (fr.y + h))
		maxX := float64(fr.x + w)
		maxY := float64(atlasHeight.Int() - fr.y)
		placements[ii.idx] = pixel.R(minX, minY, maxX, maxY)

		// Remove the used freeRect
		freeRects = append(freeRects[:fi], freeRects[fi+1:]...)

		// “Guillotine” split: create new freeRects to the right & below (if there is space)
		// Right split (if there's leftover width)
		if w < fr.w {
			freeRects = append(freeRects, freeRect{
				x: fr.x + w,
				y: fr.y,
				w: fr.w - w,
				h: fr.h,
			})
		}
		// Bottom split (if there's leftover height)
		if h < fr.h {
			freeRects = append(freeRects, freeRect{
				x: fr.x,
				y: fr.y + h,
				w: w,
				h: fr.h - h,
			})
		}
	}

	return atlas, placements
}
