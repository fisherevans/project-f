package overlays

import (
	"fmt"
	"io/fs"
	"strings"

	"fisherevans.com/project/f/internal/schema"
	"fisherevans.com/project/f/internal/util/highlighter"
	"github.com/gopxl/pixel/v2"
	"gopkg.in/yaml.v3"
)

var NamedRects = map[string]schema.OverlayRect{}
var Flows = map[string]schema.OverlayFlow{}

func LoadFromFS(fsys fs.FS) error {
	NamedRects = map[string]schema.OverlayRect{}
	Flows = map[string]schema.OverlayFlow{}

	rectsData, err := fs.ReadFile(fsys, "overlays/_rects.yaml")
	if err == nil {
		if err := yaml.Unmarshal(rectsData, &NamedRects); err != nil {
			return fmt.Errorf("parsing _rects.yaml: %w", err)
		}
	}

	entries, err := fs.ReadDir(fsys, "overlays")
	if err != nil {
		return nil // directory doesn't exist yet
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".yaml") {
			continue
		}
		if entry.Name() == "_rects.yaml" {
			continue
		}
		data, err := fs.ReadFile(fsys, "overlays/"+entry.Name())
		if err != nil {
			return fmt.Errorf("reading %s: %w", entry.Name(), err)
		}
		var flow schema.OverlayFlow
		if err := yaml.Unmarshal(data, &flow); err != nil {
			return fmt.Errorf("parsing %s: %w", entry.Name(), err)
		}
		name := strings.TrimSuffix(entry.Name(), ".yaml")
		Flows[name] = flow
	}
	return nil
}

func GetFlow(name string) (schema.OverlayFlow, bool) {
	f, ok := Flows[name]
	return f, ok
}

func ResolveTargets(flow schema.OverlayFlow) ([]highlighter.Target, error) {
	return ResolveTargetsWithExtra(flow, nil)
}

func ResolveTargetsWithExtra(flow schema.OverlayFlow, extraRects map[string]schema.OverlayRect) ([]highlighter.Target, error) {
	var targets []highlighter.Target
	for i, ot := range flow.Targets {
		area, err := resolveRect(ot, extraRects)
		if err != nil {
			return nil, fmt.Errorf("target[%d]: %w", i, err)
		}
		t := highlighter.NewTarget(area)
		if ot.Message != nil {
			msg := highlighter.NewMessage(ot.Message.Text, parseMessagePlacement(ot.Message.Placement))
			if ot.Message.Wrap > 0 {
				msg = msg.Wrapped(ot.Message.Wrap)
			}
			t = t.WithMessage(msg)
		}
		if ot.Badge != nil {
			badge := highlighter.NewBadge(parseBadgePlacement(ot.Badge.Placement))
			if ot.Badge.Label != "" {
				badge = badge.WithLabel(ot.Badge.Label)
			}
			t = t.WithBadge(badge)
		}
		if ot.NoPadding {
			t = t.NoPadding()
		}
		targets = append(targets, t)
	}
	return targets, nil
}

func resolveRect(ot schema.OverlayTarget, extraRects map[string]schema.OverlayRect) (pixel.Rect, error) {
	if ot.Region != nil {
		r := ot.Region
		return pixel.R(float64(r.X), float64(r.Y), float64(r.X+r.W), float64(r.Y+r.H)), nil
	}
	if ot.Rect != "" {
		if extraRects != nil {
			if r, ok := extraRects[ot.Rect]; ok {
				return pixel.R(float64(r.X), float64(r.Y), float64(r.X+r.W), float64(r.Y+r.H)), nil
			}
		}
		if r, ok := NamedRects[ot.Rect]; ok {
			return pixel.R(float64(r.X), float64(r.Y), float64(r.X+r.W), float64(r.Y+r.H)), nil
		}
		return pixel.Rect{}, fmt.Errorf("unknown named rect: %s", ot.Rect)
	}
	return pixel.Rect{}, fmt.Errorf("target must have either rect or region")
}

func parseMessagePlacement(s string) highlighter.MessagePlacement {
	switch s {
	case "top":
		return highlighter.MessageOnTop
	case "left":
		return highlighter.MessageOnLeft
	case "right":
		return highlighter.MessageOnRight
	default:
		return highlighter.MessageOnBottom
	}
}

func parseBadgePlacement(s string) highlighter.BadgePlacement {
	switch s {
	case "top_right":
		return highlighter.BadgeInTopRight
	case "bottom_right":
		return highlighter.BadgeInBottomRight
	case "top_middle":
		return highlighter.BadgeOnTopMiddle
	default:
		return highlighter.BadgeOnBottomMiddle
	}
}
