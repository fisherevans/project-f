package adventure

import (
	"math"

	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/resources"
	"fisherevans.com/project/f/internal/util/colors"
	"fisherevans.com/project/f/internal/util/frames"
	"fisherevans.com/project/f/internal/util/gfx"
	"fisherevans.com/project/f/internal/util/textbox"
	"fisherevans.com/project/f/internal/util/textbox/tbcfg"
	"github.com/gopxl/pixel/v2"
)

var defaultTooltipTime = 4.0

var tooltipFrame *frames.Instance
var tooltipText *textbox.Instance

func init() {
	resources.RunOnceInitialized(func() {
		tooltipFrame = frames.New("adventure/tooltip_frame", atlas, frames.WithRenderOrigin(gfx.TopCenter))
		tooltipText = textbox.NewInstance(atlas.GetFont(resources.FontNameM3x6), tbcfg.NewConfig(game.GameWidth/2, 20,
			tbcfg.HAligned(tbcfg.HAlignCenter),
			tbcfg.RenderFrom(gfx.TopCenter),
			tbcfg.Foreground(colors.FromString("#e0dec7"))))
	})
}

type Tooltips struct {
	active []*tooltip
	dy     float64
}

func NewTooltips() *Tooltips {
	return &Tooltips{}
}

func (t *Tooltips) Add(message string) {
	t.active = append(t.active, &tooltip{
		content:  tooltipText.NewSimpleContent(message),
		duration: defaultTooltipTime + tooltipFadeDuration*2,
	})
}

var tooltipFadeDuration = 1.0

func (t *Tooltips) OnTick(target pixel.Target, timeDelta float64) {
	t.dy = math.Max(t.dy-t.dy*timeDelta*0.5-3*timeDelta, 0)
	for i, tt := range t.active {
		tt.age += timeDelta
		if tt.age > tt.duration {
			t.active = append(t.active[:i], t.active[i+1:]...)
			t.dy += tt.content.Bounds().H() + 6 + 4
		}
	}
	if len(t.active) == 0 {
		t.dy = 0
		return
	}
	topMiddle := gfx.Moved(game.GameWidth/2, game.GameHeight-30).Moved(pixel.V(0, -math.Round(t.dy)))
	for _, tt := range t.active {
		maskAlpha := 1.0
		if tt.age < tooltipFadeDuration {
			maskAlpha = tt.age / tooltipFadeDuration
		} else if tt.duration-tt.age < tooltipFadeDuration {
			maskAlpha = (tt.duration - tt.age) / tooltipFadeDuration
		}
		mask := colors.WithAlpha(colors.White.RGBA, maskAlpha)
		w := tt.content.Bounds().W() + 16
		h := tt.content.Bounds().H() + 6
		tooltipFrame.Draw(target, pixel.R(0, 0, w, h), topMiddle, frames.WithColor(mask))
		tt.content.Render(target, topMiddle.Moved(pixel.V(0, 7)), tbcfg.ColorMask(mask))
		topMiddle = topMiddle.Moved(pixel.V(0, -h-4))
	}
}

type tooltip struct {
	content  *textbox.Content
	duration float64
	age      float64
}
