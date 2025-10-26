package adventure

import (
	"math"

	"github.com/gopxl/pixel/v2"
	"github.com/gopxl/pixel/v2/ext/imdraw"
	"github.com/rs/zerolog/log"

	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/game/shaders"
	"fisherevans.com/project/f/internal/util/colors"
)

type Overlay interface {
	Id() string
	OnTick(s *State, target *pixel.Batch, timeDelta float64)
	SetIsActive(bool)
	GetIsActive() bool
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

func (o *OverlaySystem) OnTick(s *State, shader shaders.Options, target *pixel.Batch, timeDelta float64) {
	var remaining []Overlay
	for _, overlay := range o.overlays {
		overlay.OnTick(s, target, timeDelta)
		if overlay.GetIsActive() {
			remaining = append(remaining, overlay)
		} else {
			log.Info().Str("overlay", overlay.Id()).Msg("overlay deactivated")
		}
	}
	o.overlays = remaining
}

func (o *OverlaySystem) Deactivate(id string) {
	for _, overlay := range o.overlays {
		if overlay.Id() == id {
			overlay.SetIsActive(false)
		}
	}
}

type BaseOverlay struct {
	OverlayId       string
	DurationSeconds float64
	AutoDeactivate  bool

	IsComplete bool
	IsActive   bool

	elapsedSeconds float64
}

func NewBaseOverlay(id string, durationSeconds float64, autoDeactivate bool) *BaseOverlay {
	return &BaseOverlay{
		OverlayId:       id,
		DurationSeconds: durationSeconds,
		AutoDeactivate:  autoDeactivate,
		IsActive:        true,
	}
}

func (o *BaseOverlay) Id() string {
	return o.OverlayId
}

func (o *BaseOverlay) GetIsActive() bool {
	return o.IsActive
}

func (o *BaseOverlay) SetIsActive(v bool) {
	o.IsActive = v
}

func (o *BaseOverlay) OnTick(s *State, target *pixel.Batch, timeDelta float64) {
	o.elapsedSeconds += timeDelta
	if !o.IsComplete && o.elapsedSeconds >= o.DurationSeconds {
		o.IsComplete = true
		s.planExecutor.MarkFadeComplete(o.Id())
		if o.AutoDeactivate {
			o.IsActive = false
		}
	}
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

func (f *FadeOverlay) OnTick(s *State, target *pixel.Batch, timeDelta float64) {
	f.BaseOverlay.OnTick(s, target, timeDelta)

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
}
