package sprite_selector

import (
	"fisherevans.com/project/f/internal/game"
	resources2 "fisherevans.com/project/f/internal/resources"
	"fisherevans.com/project/f/internal/util"
	"github.com/gopxl/pixel/v2"
	"github.com/gopxl/pixel/v2/backends/opengl"
	"slices"
)

type SelectionConsumer func(resources2.SwatchSample)

type SpriteSelector struct {
	win         *opengl.Window
	selected    resources2.SwatchSample
	parentState game.State
	consumer    SelectionConsumer
}

func New(win *opengl.Window, initialSelection resources2.SwatchSample, parent game.State, onSelect SelectionConsumer) game.State {
	return &SpriteSelector{
		win:         win,
		selected:    initialSelection,
		parentState: parent,
		consumer:    onSelect,
	}
}

var selectedOverlaySprite = resources2.GetSprite("ui", 3, 1)

func (s *SpriteSelector) OnTick(ctx *game.Context, target pixel.Target, targetBounds pixel.Rect, timeDelta float64) {
	s.listenToInputs(ctx)
	startY := targetBounds.H()
	for spriteId, spriteRef := range resources2.Tilesheets[s.selected.SpriteId.Tilesheet].Sprites {
		x := float64(spriteId.Column * resources2.TileSize)
		y := startY - float64(spriteId.Row*resources2.TileSize)
		mat := pixel.IM.Moved(pixel.V(x, y))
		spriteRef.Sprite.Draw(target, mat)
		if spriteId == s.selected.SpriteId {
			selectedOverlaySprite.Sprite.Draw(target, mat)
		}
	}
	ctx.DebugTL("selected sprite: (%s)", s.selected.SpriteId)

	ctx.DebugBR("wasd/arrows: change sprite")
	ctx.DebugBR("page up/down or []: change tilesheet")
	ctx.DebugBR("esc/backspace: cancel")
	ctx.DebugBR("enter/tab: select")
}

func (s *SpriteSelector) listenToInputs(ctx *game.Context) {
	win := s.win
	ts := resources2.Tilesheets[s.selected.SpriteId.Tilesheet]
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
	sheets := util.SortedKeys(resources2.Tilesheets)
	next := slices.Index(sheets, s.selected.SpriteId.Tilesheet) + amount
	if next < 0 {
		next += len(sheets)
	}
	if next >= len(sheets) {
		next = 0
	}
	s.selected = resources2.SwatchSample{
		SpriteId: resources2.SpriteId{
			Tilesheet: sheets[next],
			Row:       1,
			Column:    1,
		},
	}
}
