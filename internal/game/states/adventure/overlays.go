package adventure

import (
	"math"

	"github.com/gopxl/pixel/v2"
	"github.com/gopxl/pixel/v2/ext/imdraw"

	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/util/colors"
)

type Overlay interface {
	OnTick(ctx *game.Context, s *State, target pixel.Target, timeDelta float64) bool
}

type OverlaySystem struct {
	overlays []Overlay
}

func NewOverlaySystem() *OverlaySystem {
	return &OverlaySystem{}
}

func (o *OverlaySystem) Add(overlay Overlay) {
	o.overlays = append(o.overlays, overlay)
}

func (o *OverlaySystem) OnTick(ctx *game.Context, s *State, target pixel.Target, timeDelta float64) {
	var remaining []Overlay
	for _, overlay := range o.overlays {
		if !overlay.OnTick(ctx, s, target, timeDelta) {
			remaining = append(remaining, overlay)
		}
	}
	o.overlays = remaining
}

type FadeOverlay struct {
	FromColor       pixel.RGBA
	ToColor         pixel.RGBA
	DurationSeconds float64
	AutoComplete    bool
	OnComplete      func(ctx *game.Context, s *State)

	IsComplete bool

	imd            *imdraw.IMDraw
	elapsedSeconds float64
	lastComplete   bool
}

func NewFadeOverlay(fromColor pixel.RGBA, toColor pixel.RGBA, durationSeconds float64, autoComplete bool, onComplete func(ctx *game.Context, s *State)) *FadeOverlay {
	return &FadeOverlay{
		FromColor:       fromColor,
		ToColor:         toColor,
		DurationSeconds: durationSeconds,
		AutoComplete:    autoComplete,
		OnComplete:      onComplete,
		imd:             imdraw.New(nil),
	}
}

func (f *FadeOverlay) OnTick(ctx *game.Context, s *State, target pixel.Target, timeDelta float64) bool {
	// update progress
	f.elapsedSeconds += timeDelta
	progress := math.Min(1, f.elapsedSeconds/f.DurationSeconds)

	// draw color
	fadeColor := colors.Lerp(f.FromColor, f.ToColor, progress)
	f.imd.Clear()
	f.imd.Color = fadeColor
	f.imd.Push(pixel.V(0, 0), pixel.V(game.GameWidth, game.GameHeight))
	f.imd.Rectangle(0) // draw filled rectangle
	f.imd.Draw(target)
	//gfx.DrawRect(atlas, target, pixel.IM, gfx.BottomLeft, game.GameWidth, game.GameHeight, fadeColor)

	// mark complete
	if f.elapsedSeconds >= f.DurationSeconds && f.AutoComplete {
		f.IsComplete = true
	}
	if f.IsComplete && !f.lastComplete {
		f.lastComplete = true
		if f.OnComplete != nil {
			f.OnComplete(ctx, s)
		}
	}
	return f.IsComplete
}
