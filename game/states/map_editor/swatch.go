package map_editor

import (
	"fisherevans.com/project/f/resources"
	"fisherevans.com/project/f/util"
	"github.com/gopxl/pixel/v2"
	"github.com/gopxl/pixel/v2/backends/opengl"
	"image/color"
	"slices"
)

var swatchKeySprites []*pixel.Sprite

func init() {
	for index, _ := range swatchKeys {
		swatchKeySprites = append(swatchKeySprites, resources.GetSprite("ui", index+1, 2).Sprite)
	}
}

const swatchKeyPadding = 0.25

type Swatch struct {
	SelectedSwatch string
	SelectedSample pixel.Button
	Canvas         *opengl.Canvas
}

func newSwatch() *Swatch {
	s := &Swatch{
		SelectedSample: swatchKeys[0],
		SelectedSwatch: "default",
	}
	s.Canvas = opengl.NewCanvas(pixel.R(0, 0, float64(resources.TileSize*(1+swatchKeyPadding)*len(swatchKeys)), resources.TileSizeF64*2))
	return s
}

var swatchSpriteDelete = resources.GetSprite("ui", 1, 1).Sprite
var swatchSpriteDraw = resources.GetSprite("ui", 2, 1).Sprite
var swatchSpriteSelected = resources.GetSprite("ui", 3, 1).Sprite
var swatchSpriteSelectedNumber = resources.GetSprite("ui", 1, 3).Sprite

func (s *Swatch) DrawCanvasOverlay(ctx *Context, win *opengl.Window, canvas *opengl.Canvas, cameraMatrix pixel.Matrix) {
	// draw brush
	if ctx.MouseInCanvas {
		drawMatrix := cameraMatrix.Moved(pixel.V(float64(ctx.MouseMapLocation.X*resources.TileSize), float64(ctx.MouseMapLocation.Y*resources.TileSize)))
		if win.Pressed(pixel.KeyLeftShift) {
			swatchSpriteDelete.Draw(canvas, drawMatrix)
		} else {
			spriteId := resources.Swatches[s.SelectedSwatch].Samples[s.SelectedSample].SpriteId
			sprite := resources.Sprites[spriteId].Sprite
			sprite.Draw(canvas, drawMatrix)
			swatchSpriteDraw.Draw(canvas, drawMatrix)
		}
	}
}

func (s *Swatch) DrawSwatch(ctx *Context, win *opengl.Window, canvas *opengl.Canvas) {
	s.Canvas.Clear(color.Transparent)
	for index, key := range swatchKeys {
		tile := resources.Swatches[s.SelectedSwatch].Samples[key]
		if tile.SpriteId.Tilesheet == "" {
			continue
		}
		tileMatrix := pixel.IM.
			Moved(pixel.V((resources.TileSizeF64*(1+swatchKeyPadding))*float64(index), 0)).
			Moved(pixel.V(resources.TileSizeF64/2, resources.TileSizeF64/2))
		numberMatrix := tileMatrix.Moved(pixel.V(0, resources.TileSizeF64))
		sprite := resources.Sprites[tile.SpriteId].Sprite
		sprite.Draw(s.Canvas, tileMatrix)
		swatchKeySprites[index].Draw(s.Canvas, numberMatrix)
		if key == s.SelectedSample {
			swatchSpriteSelected.Draw(s.Canvas, tileMatrix)
			swatchSpriteSelectedNumber.Draw(s.Canvas, numberMatrix)
		}
	}
	s.Canvas.Draw(win, pixel.IM.
		Moved(s.Canvas.Bounds().Center()).
		Moved(pixel.V(resources.TileSizeF64*swatchKeyPadding, resources.TileSizeF64*swatchKeyPadding)).
		Scaled(pixel.ZV, ctx.CanvasScale))
}

func (s *Swatch) changeSwatch(amount int) {
	swatches := util.SortedKeys(resources.Swatches)
	next := slices.Index(swatches, s.SelectedSwatch) + amount
	if next < 0 {
		next += len(swatches)
	}
	if next >= len(swatches) {
		next = 0
	}
	s.SelectedSwatch = swatches[next]
}
