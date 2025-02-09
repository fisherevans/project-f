package textbox

import (
	"fisherevans.com/project/f/internal/util"
	"image/color"
	"strings"
)

func (tb *Instance) NewTestAllFeaturesContent(msg string, opts ...ContentOpt) *Content {
	var cgroups []*cgroup
	parts := strings.Split(msg, " ")
	for id, part := range parts {
		if id > 0 {
			cgroups = append(cgroups, newSpaceCgroup(tb.text))
		}
		colorOverride := &cColor{util.StringToColor(part, 0.95, 0.35)}
		cgroups = append(cgroups, newTestFeaturesCGroup(part, colorOverride, tb.text))
	}
	return tb.newContent(cgroups, opts...)
}

func (tb *Instance) NewSimpleContent(msg string, color color.Color, opts ...ContentOpt) *Content {
	var cgroups []*cgroup
	parts := strings.Split(msg, " ")
	for id, part := range parts {
		if id > 0 {
			cgroups = append(cgroups, newSpaceCgroup(tb.text))
		}
		cgroups = append(cgroups, newTestFeaturesCGroup(part, nil, tb.text))
	}
	return tb.newContent(cgroups, opts...)
}
