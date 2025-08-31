package adventure

import (
	"math"

	"github.com/gopxl/pixel/v2"
	"github.com/gopxl/pixel/v2/ext/imdraw"

	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/game/shaders"
	"fisherevans.com/project/f/internal/util/colors"
)

type Overlay interface {
	OnTick(ctx *game.Context, s *State, target *pixel.Batch, timeDelta float64) bool
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

func (o *OverlaySystem) OnTick(ctx *game.Context, s *State, shader shaders.Options, target *pixel.Batch, timeDelta float64) {
	var remaining []Overlay
	for _, overlay := range o.overlays {
		if !overlay.OnTick(ctx, s, target, timeDelta) {
			remaining = append(remaining, overlay)
		}
	}
	o.overlays = remaining
}

type BaseOverlay struct {
	DurationSeconds float64
	AutoComplete    bool
	OnComplete      func(ctx *game.Context, s *State)

	IsComplete bool

	elapsedSeconds float64
	lastComplete   bool
}

func NewBaseOverlay(durationSeconds float64, autoComplete bool, onComplete func(ctx *game.Context, s *State)) *BaseOverlay {
	return &BaseOverlay{
		DurationSeconds: durationSeconds,
		AutoComplete:    autoComplete,
		OnComplete:      onComplete,
	}
}

func (o *BaseOverlay) OnTick(ctx *game.Context, s *State, target *pixel.Batch, timeDelta float64) bool {
	o.elapsedSeconds += timeDelta
	if o.elapsedSeconds >= o.DurationSeconds && o.AutoComplete {
		o.IsComplete = true
	}
	if o.IsComplete && !o.lastComplete {
		o.lastComplete = true
		if o.OnComplete != nil {
			o.OnComplete(ctx, s)
		}
	}
	return o.IsComplete
}

func (o *BaseOverlay) Progress() float64 {
	return math.Min(1, o.elapsedSeconds/o.DurationSeconds)
}

type FadeOverlay struct {
	*BaseOverlay
	FromColor pixel.RGBA
	ToColor   pixel.RGBA

	// 1 = from→to, 5 = from→to→from→to→from; evenly over DurationSeconds
	TransitionCount int // number of color transitions across the duration (min 1)

	imd *imdraw.IMDraw
}

func NewFadeOverlay(fromColor pixel.RGBA, toColor pixel.RGBA, transitions int, base *BaseOverlay) *FadeOverlay {
	return &FadeOverlay{
		BaseOverlay:     base,
		FromColor:       fromColor,
		ToColor:         toColor,
		TransitionCount: transitions,
		imd:             imdraw.New(nil),
	}
}

func (f *FadeOverlay) OnTick(ctx *game.Context, s *State, target *pixel.Batch, timeDelta float64) bool {
	isComplete := f.BaseOverlay.OnTick(ctx, s, target, timeDelta)

	// draw color with multiple transitions support
	p := f.Progress()
	count := f.TransitionCount
	if count < 1 {
		count = 1
	}
	pos := p * float64(count)
	seg := int(math.Floor(pos))
	if seg >= count {
		seg = count - 1
	}
	t := pos - float64(seg)

	var fadeColor pixel.RGBA
	if seg%2 == 0 {
		fadeColor = colors.Lerp(f.FromColor, f.ToColor, t)
	} else {
		fadeColor = colors.Lerp(f.ToColor, f.FromColor, t)
	}

	f.imd.Clear()
	f.imd.Color = fadeColor
	f.imd.Push(pixel.V(0, 0), pixel.V(game.GameWidth, game.GameHeight))
	f.imd.Rectangle(0) // draw filled rectangle
	f.imd.Draw(target)
	//gfx.DrawRect(atlas, target, pixel.IM, gfx.BottomLeft, game.GameWidth, game.GameHeight, fadeColor)

	return isComplete
}
