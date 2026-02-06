package keyframe

import (
	"fisherevans.com/project/f/internal/util/colors"
	"fisherevans.com/project/f/internal/util/interp"
	"github.com/gopxl/pixel/v2"
)

type Group struct {
	timeElapsed float64
	reverse     bool
	maxFrom     float64
}

func NewGroup() *Group {
	return &Group{}
}

func (g *Group) Update(timeDelta float64) {
	if g.reverse {
		timeDelta *= -1
	}
	g.timeElapsed = max(min(g.timeElapsed+timeDelta, g.maxFrom), 0)
}

func (g *Group) Reset() {
	if g.reverse {
		g.timeElapsed = g.maxFrom
	} else {
		g.timeElapsed = 0
	}
}

func (g *Group) Skip() {
	if g.reverse {
		g.timeElapsed = 0
	} else {
		g.timeElapsed = g.maxFrom
	}
}

func (g *Group) SetReverse(reverse bool) {
	g.reverse = reverse
}

func (g *Group) AddKeyFrame(from, to float64) *Member {
	return g.AddKeyFrameFn(from, to, interp.Linear)
}

func (g *Group) AddKeyFrameFn(from, to float64, function interp.Function) *Member {
	if function == nil {
		function = interp.Linear
	}
	kf := &Member{
		group:      g,
		from:       from,
		to:         to,
		interpFunc: function,
	}
	g.maxFrom = max(g.maxFrom, to)
	return kf
}

type Member struct {
	group      *Group
	from, to   float64
	interpFunc interp.Function
}

func (kf *Member) Progress() float64 {
	if kf.group.timeElapsed < kf.from {
		return 0
	}
	if kf.group.timeElapsed > kf.to {
		return 1
	}
	return kf.interpFunc((kf.group.timeElapsed - kf.from) / (kf.to - kf.from))
}

func (kf *Member) InverseProgress() float64 {
	return 1.0 - kf.Progress()
}

func (kf *Member) Alpha(c pixel.RGBA) pixel.RGBA {
	return colors.LayerAlpha(c, kf.Progress())
}
