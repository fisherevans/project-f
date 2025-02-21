package sprite_selector

import (
	"fisherevans.com/project/f/internal/game"
	resources "fisherevans.com/project/f/internal/resources"
	"fisherevans.com/project/f/internal/util"
	"github.com/gopxl/pixel/v2"
	"github.com/gopxl/pixel/v2/backends/opengl"
	"slices"
)

type SelectionConsumer func(*resources.SwatchSample)

type SpriteSelector struct {
	game.BaseState
	win                  *opengl.Window
	selected             *resources.SwatchSample
	backState, nextState game.State
	consumer             SelectionConsumer
	batch                *pixel.Batch
}

func New(win *opengl.Window, initialSelection *resources.SwatchSample, backState, nextState game.State, onSelect SelectionConsumer) game.State {
	return &SpriteSelector{
		win:       win,
		selected:  initialSelection,
		backState: backState,
		nextState: nextState,
		consumer:  onSelect,
		batch:     pixel.NewBatch(&pixel.TrianglesData{}, resources.SpriteAtlas),
	}
}

var selectedOverlaySprite = resources.GetTilesheetSprite("ui", 3, 1)

func (s *SpriteSelector) OnTick(ctx *game.Context, target pixel.Target, targetBounds pixel.Rect, timeDelta float64) {
	s.batch.Clear()

	s.listenToInputs(ctx)
	startY := targetBounds.H()
	for spriteId, spriteRef := range resources.Tilesheets[s.selected.SpriteId.Tilesheet].Sprites {
		x := float64(spriteId.Column * resources.MapTileSize.Int())
		y := startY - float64(spriteId.Row*resources.MapTileSize.Int())
		mat := pixel.IM.Moved(pixel.V(x, y))
		spriteRef.Sprite.Draw(s.batch, mat)
		if spriteId == s.selected.SpriteId {
			selectedOverlaySprite.Sprite.Draw(s.batch, mat)
		}
	}

	s.batch.Draw(target)

	ctx.DebugTL("selected sprite: (%s)", s.selected.SpriteId)

	ctx.DebugBR("wasd/arrows: change sprite")
	ctx.DebugBR("page up/down or []: change tilesheet")
	ctx.DebugBR("esc/backspace: cancel")
	ctx.DebugBR("enter/tab: select")
}

func (s *SpriteSelector) listenToInputs(ctx *game.Context) {
	win := s.win
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
		ctx.SwapActiveState(s.nextState)
	}
	if win.JustPressed(pixel.KeyBackspace) || win.JustPressed(pixel.KeyEscape) {
		ctx.SwapActiveState(s.backState)
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
	s.selected = &resources.SwatchSample{
		SpriteId: resources.TilesheetSpriteId{
			Tilesheet: sheets[next],
			Row:       1,
			Column:    1,
		},
	}
}
