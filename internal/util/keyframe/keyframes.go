package keyframe

import (
	"cmp"
	"slices"

	"fisherevans.com/project/f/internal/util/colors"
	"fisherevans.com/project/f/internal/util/interp"
	"github.com/gopxl/pixel/v2"
	"github.com/rs/zerolog/log"
)

type Group struct {
	timeElapsed     float64
	timeScale       float64
	reverse         bool
	maxTo           float64
	defaultFunction interp.Function
}

func NewGroup(defaultFunction interp.Function) *Group {
	if defaultFunction == nil {
		defaultFunction = interp.Linear
	}
	return &Group{
		defaultFunction: defaultFunction,
		timeScale:       1,
	}
}

func (g *Group) SetTimeScale(scale float64) {
	g.timeScale = scale
}

func (g *Group) Update(timeDelta float64) {
	if g.reverse {
		timeDelta *= -1
	}
	timeDelta *= g.timeScale
	g.timeElapsed = max(min(g.timeElapsed+timeDelta, g.maxTo), 0)
}

func (g *Group) Reset() {
	if g.reverse {
		g.timeElapsed = g.maxTo
	} else {
		g.timeElapsed = 0
	}
}

func (g *Group) Skip() {
	if g.reverse {
		g.timeElapsed = 0
	} else {
		g.timeElapsed = g.maxTo
	}
}

func (g *Group) Progress() float64 {
	return min(g.timeElapsed/g.maxTo, 1)
}

func (g *Group) SetReverse(reverse bool) {
	g.reverse = reverse
}

func (g *Group) AddKeyFrame(from, to float64) *Member {
	return g.AddKeyFrameFn(from, to, nil)
}

func (g *Group) AddKeyFrameFn(from, to float64, function interp.Function) *Member {
	kf := &Member{
		group:      g,
		interpFunc: function,
	}
	kf.AddTransition(from, to)
	return kf
}

type memberTransition struct {
	from, to float64
}

type Member struct {
	group       *Group
	transitions []memberTransition
	interpFunc  interp.Function
}

func (kf *Member) AddTransition(from, to float64) *Member {
	if from > to {
		from, to = to, from
	}
	kf.transitions = append(kf.transitions, memberTransition{from, to})
	slices.SortFunc(kf.transitions, func(a, b memberTransition) int {
		return cmp.Compare(a.from, b.from)
	})
	// ensure none overlap
	for i := 0; i < len(kf.transitions)-1; i++ {
		if kf.transitions[i].to > kf.transitions[i+1].from {
			log.Fatal().Msgf("keyframes overlap, %d:%f passes %d:%f", i, kf.transitions[i].to, i+1, kf.transitions[i+1].from)
		}
	}
	kf.group.maxTo = max(kf.group.maxTo, to)
	return kf
}

func (kf *Member) Progress() float64 {
	progress := 0.0
	fadingIn := true
	for _, t := range kf.transitions {
		if kf.group.timeElapsed < t.from {
			break
		}
		if kf.group.timeElapsed > t.to {
			fadingIn = !fadingIn
			continue
		}
		progress = (kf.group.timeElapsed - t.from) / (t.to - t.from)
		break
	}
	fn := kf.interpFunc
	if fn == nil {
		fn = kf.group.defaultFunction
	}
	progress = fn(progress)
	if !fadingIn {
		progress = 1.0 - progress
	}
	return progress
}

func (kf *Member) InverseProgress() float64 {
	return 1.0 - kf.Progress()
}

func (kf *Member) Alpha(c pixel.RGBA) pixel.RGBA {
	return colors.LayerAlpha(c, kf.Progress())
}
