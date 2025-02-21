package map_editor

import (
	resources "fisherevans.com/project/f/internal/resources"
	"fisherevans.com/project/f/internal/util"
	"github.com/gopxl/pixel/v2"
	"github.com/gopxl/pixel/v2/backends/opengl"
	"image/color"
	"slices"
)

var swatchKeySprites []*pixel.Sprite

func init() {
	for index, _ := range swatchKeys {
		swatchKeySprites = append(swatchKeySprites, resources.GetTilesheetSprite("ui", index+1, 2).Sprite)
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
	s.Canvas = opengl.NewCanvas(pixel.R(0, 0, resources.MapTileSize.Float()*(1+swatchKeyPadding)*float64(len(swatchKeys)), resources.MapTileSize.Float()*2))
	return s
}

var swatchSpriteDelete = resources.GetTilesheetSprite("ui", 1, 1).Sprite
var swatchSpriteDraw = resources.GetTilesheetSprite("ui", 2, 1).Sprite
var swatchSpriteSelected = resources.GetTilesheetSprite("ui", 3, 1).Sprite
var swatchSpriteSelectedNumber = resources.GetTilesheetSprite("ui", 1, 3).Sprite

func (s *Swatch) DrawCanvasOverlay(ctx *Context, win *opengl.Window, target pixel.Target, cameraMatrix pixel.Matrix) {
	if !ctx.MouseInCanvas {
		return
	}

	var sprites []*pixel.Sprite

	if win.Pressed(pixel.KeyLeftShift) {
		sprites = append(sprites, swatchSpriteDelete)
	} else {
		spriteId := resources.Swatches[s.SelectedSwatch].Samples[s.SelectedSample].SpriteId
		sprites = append(sprites, resources.TilesheetSprites[spriteId].Sprite)
		sprites = append(sprites, swatchSpriteDraw)
	}

	drawMatrix := cameraMatrix.Moved(pixel.V(float64(ctx.MouseMapLocation.X*resources.MapTileSize.Int()), float64(ctx.MouseMapLocation.Y*resources.MapTileSize.Int())))
	for _, sprite := range sprites {
		sprite.Draw(target, drawMatrix)
	}
}

func (s *Swatch) DrawSwatch(ctx *Context, win *opengl.Window) {
	s.Canvas.Clear(color.Transparent)
	for index, key := range swatchKeys {
		tile := resources.Swatches[s.SelectedSwatch].Samples[key]
		if tile.SpriteId.Tilesheet == "" {
			continue
		}
		tileMatrix := pixel.IM.
			Moved(pixel.V((resources.MapTileSize.Float()*(1+swatchKeyPadding))*float64(index), 0)).
			Moved(pixel.V(resources.MapTileSize.Float()/2, resources.MapTileSize.Float()/2))
		numberMatrix := tileMatrix.Moved(pixel.V(0, resources.MapTileSize.Float()))
		sprite := resources.TilesheetSprites[tile.SpriteId].Sprite
		sprite.Draw(s.Canvas, tileMatrix)
		swatchKeySprites[index].Draw(s.Canvas, numberMatrix)
		if key == s.SelectedSample {
			swatchSpriteSelected.Draw(s.Canvas, tileMatrix)
			swatchSpriteSelectedNumber.Draw(s.Canvas, numberMatrix)
		}
	}
	s.Canvas.Draw(win, pixel.IM.
		Moved(s.Canvas.Bounds().Center()).
		Moved(pixel.V(resources.MapTileSize.Float()*swatchKeyPadding, resources.MapTileSize.Float()*swatchKeyPadding)).
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
