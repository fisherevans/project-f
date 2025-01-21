package sprite_selector

import (
	"fisherevans.com/project/f/game"
	"fisherevans.com/project/f/resources"
	"fisherevans.com/project/f/util"
	"github.com/gopxl/pixel/v2"
	"github.com/gopxl/pixel/v2/backends/opengl"
	"slices"
)

type SelectionConsumer func(resources.SwatchSample)

type SpriteSelector struct {
	selected    resources.SwatchSample
	parentState game.State
	consumer    SelectionConsumer
}

func New(initialSelection resources.SwatchSample, parent game.State, onSelect SelectionConsumer) game.State {
	return &SpriteSelector{
		selected:    initialSelection,
		parentState: parent,
		consumer:    onSelect,
	}
}

var selectedOverlaySprite = resources.GetSprite("ui", 3, 1)

func (s *SpriteSelector) OnTick(ctx *game.Context, win *opengl.Window, canvas *opengl.Canvas, timeDelta float64) {
	s.listenToInputs(ctx, win)
	startY := canvas.Bounds().H()
	for spriteId, spriteRef := range resources.Tilesheets[s.selected.SpriteId.Tilesheet].Sprites {
		x := float64(spriteId.Column * resources.TileSize)
		y := startY - float64(spriteId.Row*resources.TileSize)
		mat := pixel.IM.Moved(pixel.V(x, y))
		spriteRef.Sprite.Draw(canvas, mat)
		if spriteId == s.selected.SpriteId {
			selectedOverlaySprite.Sprite.Draw(canvas, mat)
		}
	}
	ctx.DebugTL("selected sprite: (%s)", s.selected.SpriteId)

	ctx.DebugBR("wasd/arrows: change sprite")
	ctx.DebugBR("page up/down or []: change tilesheet")
	ctx.DebugBR("esc/backspace: cancel")
	ctx.DebugBR("enter/tab: select")
}

func (s *SpriteSelector) listenToInputs(ctx *game.Context, win *opengl.Window) {
	ts := resources.Tilesheets[s.selected.SpriteId.Tilesheet]
	if win.JustPressed(pixel.KeyUp) || win.Repeated(pixel.KeyUp) {
		s.selected.SpriteId.Row--
		if s.selected.SpriteId.Row <= 0 {
			s.selected.SpriteId.Row += ts.Rows
		}
	}
	if win.JustPressed(pixel.KeyDown) || win.Repeated(pixel.KeyDown) {
		s.selected.SpriteId.Row++
		if s.selected.SpriteId.Row > ts.Rows {
			s.selected.SpriteId.Row = 1
		}
	}
	if win.JustPressed(pixel.KeyLeft) || win.Repeated(pixel.KeyLeft) {
		s.selected.SpriteId.Column--
		if s.selected.SpriteId.Column <= 0 {
			s.selected.SpriteId.Column += ts.Columns
		}
	}
	if win.JustPressed(pixel.KeyRight) || win.Repeated(pixel.KeyRight) {
		s.selected.SpriteId.Column++
		if s.selected.SpriteId.Column > ts.Columns {
			s.selected.SpriteId.Column = 1
		}
	}
	if win.JustPressed(pixel.KeyPageDown) || win.JustPressed(pixel.KeyLeftBracket) {
		s.changeSheet(1)
	}
	if win.JustPressed(pixel.KeyPageUp) || win.JustPressed(pixel.KeyRightBracket) {
		s.changeSheet(-1)
	}
	if win.JustPressed(pixel.KeyEnter) || win.JustPressed(pixel.KeyTab) {
		s.consumer(s.selected)
		ctx.SwapActiveState(s.parentState)
	}
	if win.JustPressed(pixel.KeyBackspace) || win.JustPressed(pixel.KeyEscape) {
		ctx.SwapActiveState(s.parentState)
	}
}

func (s *SpriteSelector) changeSheet(amount int) {
	sheets := util.SortedKeys(resources.Tilesheets)
	next := slices.Index(sheets, s.selected.SpriteId.Tilesheet) + amount
	if next < 0 {
		next += len(sheets)
	}
	if next >= len(sheets) {
		next = 0
	}
	s.selected = resources.SwatchSample{
		SpriteId: resources.SpriteId{
			Tilesheet: sheets[next],
			Row:       1,
			Column:    1,
		},
	}
}
