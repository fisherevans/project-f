package runtime

import (
	"sync"

	"fisherevans.com/project/f/internal/resources"
	"fisherevans.com/project/f/internal/util/highlighter"
	"github.com/gopxl/pixel/v2"
)

type DebugHighlight struct {
	mu      sync.Mutex
	drawer  *highlighter.Drawer
	targets []highlighter.Target
	current int
	remain  float64
	perStep float64
}

func newDebugHighlight() *DebugHighlight {
	return &DebugHighlight{}
}

func (dh *DebugHighlight) init() {
	atlas := resources.DefaultAtlas()
	dh.drawer = highlighter.NewDrawer(atlas, resources.FontNameM3x6)
}

func (dh *DebugHighlight) Show(targets []highlighter.Target, durationPerStep float64) {
	dh.mu.Lock()
	defer dh.mu.Unlock()
	dh.targets = targets
	dh.current = 0
	dh.perStep = durationPerStep
	dh.remain = durationPerStep
	if len(targets) > 0 {
		dh.drawer.SetTargetArea(targets[0], false)
	}
}

func (dh *DebugHighlight) Dismiss() {
	dh.mu.Lock()
	defer dh.mu.Unlock()
	dh.targets = nil
	dh.current = 0
	dh.remain = 0
	dh.drawer.Dismiss(true)
}

func (dh *DebugHighlight) Render(target pixel.Target, dt float64) {
	dh.mu.Lock()
	defer dh.mu.Unlock()
	if dh.drawer == nil {
		return
	}
	if len(dh.targets) == 0 {
		return
	}
	dh.remain -= dt
	if dh.remain <= 0 {
		dh.current++
		if dh.current >= len(dh.targets) {
			dh.targets = nil
			dh.drawer.Dismiss(true)
			dh.drawer.Render(target, dt)
			return
		}
		dh.drawer.SetTargetArea(dh.targets[dh.current], true)
		dh.remain = dh.perStep
	}
	dh.drawer.Render(target, dt)
}

func (dh *DebugHighlight) IsActive() bool {
	dh.mu.Lock()
	defer dh.mu.Unlock()
	return len(dh.targets) > 0 || (dh.drawer != nil && !dh.drawer.IsDismissed())
}
