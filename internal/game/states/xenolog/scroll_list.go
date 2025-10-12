package xenolog

import (
	"math"

	"github.com/gopxl/pixel/v2"
	"github.com/rs/zerolog/log"

	"fisherevans.com/project/f/internal/game/input"
	"fisherevans.com/project/f/internal/util/interp"
)

var (
	scrollCursor = atlas.GetSprite("xenolog/scroll_cursor")
)

type scrollListOptions struct {
	targetHeight        int
	verticalItemMargin  int
	verticalListPadding int
	transitionTime      float64
}

type scrollList[T any] struct {
	scrollListOptions
	items  []listItem[T]
	cursor listCursor[T]

	initialized      bool
	highlightedIndex int

	timeTransitioning           float64
	fromDy, currentDy, targetDy float64
	fromCy, currentCy, targetCy float64
}

func newScrollList[T any](opt scrollListOptions, items ...listItem[T]) *scrollList[T] {
	return &scrollList[T]{
		scrollListOptions: opt,
		items:             items,
	}
}

func (s *scrollList[T]) highlightedItem() listItem[T] {
	return s.items[s.highlightedIndex]
}

func (s *scrollList[T]) Add(item listItem[T]) {
	s.items = append(s.items, item)
}

func (s *scrollList[T]) Render(t T, controls *input.Controls, target pixel.Target, timeDelta float64) {
	if !s.initialized {
		s.ensureVisibleNow()
		s.initialized = true
	}
	s.handleInput(t, controls)
	s.updateScroll(timeDelta)
	s.drawList(t, target)
}

func (s *scrollList[T]) drawList(t T, target pixel.Target) {
	baseY := float64(s.targetHeight-s.verticalListPadding) + math.Round(s.currentDy)

	topLeftY := int(baseY)
	nextTopLeftY := topLeftY

	visibleTop := s.targetHeight + s.verticalListPadding
	visibleBottom := -s.verticalItemMargin

	for index := 0; index < len(s.items); index++ {
		item := s.items[index]
		topLeftY = nextTopLeftY
		nextTopLeftY = topLeftY - item.Height() - s.verticalItemMargin

		// Off-screen handling
		if topLeftY > visibleTop {
			item.RenderWasSkipped(t, true)
			continue
		} else if topLeftY < visibleBottom {
			item.RenderWasSkipped(t, false)
			continue
		}

		progress := 0.0
		if index == s.highlightedIndex {
			progress = math.Min(s.timeTransitioning/s.transitionTime, 1.0)
		}
		item.Render(t, topLeftY, target, progress)
	}
	if s.cursor != nil {
		highlightProgress := math.Min(s.timeTransitioning/s.transitionTime, 1.0)
		movementProgress := 1.0
		if s.currentCy != s.targetCy {
			movementProgress = highlightProgress
		}
		s.cursor.Render(t, int(s.currentCy), target, s.highlightedIndex, movementProgress, highlightProgress)
	}
}

type baseListItem[T any] struct{}

func (b baseListItem[T]) RenderWasSkipped(t T, wasAbove bool)           {}
func (b baseListItem[T]) DirectionJustPressed(t T, dir input.Direction) {}
func (b baseListItem[T]) ButtonAJustPressed(t T)                        {}
func (b baseListItem[T]) ButtonStartJustPressed(t T)                    {}
func (b baseListItem[T]) SkipHighlight(t T) bool {
	return false
}

type listItem[T any] interface {
	Render(t T, topLeftY int, target pixel.Target, highlightedProgress float64)
	RenderWasSkipped(t T, wasAbove bool)
	Height() int
	DirectionJustPressed(t T, dir input.Direction)
	ButtonAJustPressed(t T)
	ButtonStartJustPressed(t T)
	SkipHighlight(t T) bool
}

type listCursor[T any] interface {
	Render(t T, centerLeftY int, target pixel.Target, index int, movementProgress, highlightProgress float64)
}

// ScrollPosition returns two values:
// 1. The scroll position as a value from 0 (top) to 1 (bottom)
// 2. The visible ratio of the content (0-1, where 1 means all content is visible)
func (s *scrollList[T]) ScrollPosition() (scrollPct, visibleRatio float64) {
	totalHeight := s.totalHeight()
	if totalHeight <= 0 {
		return 0, 1
	}

	// Calculate visible ratio (clamped to 1.0)
	visibleRatio = math.Min(1.0, float64(s.targetHeight)/totalHeight)

	// Calculate scroll percentage (0-1)
	scrollRange := totalHeight - float64(s.targetHeight)
	if scrollRange <= 0 {
		scrollPct = 0
	} else {
		scrollPct = s.currentDy / scrollRange
		scrollPct = math.Max(0, math.Min(1, scrollPct)) // Clamp to 0-1
	}

	return scrollPct, visibleRatio
}

// totalHeight calculates the total height of all items including vertical margins between them
func (s *scrollList[T]) totalHeight() float64 {
	if len(s.items) == 0 {
		return 0
	}

	total := 0
	for i, item := range s.items {
		total += item.Height()
		if i > 0 {
			total += s.verticalItemMargin
		}
	}

	return float64(total)
}

func (s *scrollList[T]) calcYs(index int) (float64, float64) {
	// Keep selected row visible without unnecessary movement
	rowHeightAbove := 0
	for i := 0; i < index; i++ {
		rowHeightAbove += s.items[i].Height() + s.verticalItemMargin
	}
	topY := int(math.Round(s.currentDy)) + screenHeight - s.verticalListPadding - rowHeightAbove
	bottomY := topY - s.items[index].Height()
	halfRowHeight := float64(topY-bottomY) / 2.0

	targetDy := s.currentDy
	var targetCy float64
	topBound := screenHeight - s.verticalListPadding
	lowerBound := s.verticalListPadding
	if tooHigh := topY - topBound; tooHigh > 0 {
		targetDy -= float64(tooHigh)
		targetCy = float64(topBound) - halfRowHeight
	} else if tooLow := bottomY - lowerBound; tooLow < 0 {
		targetDy -= float64(tooLow)
		targetCy = float64(lowerBound) + halfRowHeight
	} else {
		targetCy = float64(topY+bottomY) / 2.0
	}

	return targetDy, targetCy
}

func (s *scrollList[T]) ensureVisibleNow() {
	s.targetDy, s.targetCy = s.calcYs(s.highlightedIndex)
	s.currentDy = s.targetDy
	s.currentCy = s.targetCy
	s.timeTransitioning = s.transitionTime
}

func (s *scrollList[T]) updateScroll(timeDelta float64) {
	s.timeTransitioning += timeDelta
	progress := interp.Smootherstep(math.Min(1, s.timeTransitioning/s.transitionTime))
	s.currentDy = interp.Lerp(s.fromDy, s.targetDy, progress)
	s.currentCy = interp.Lerp(s.fromCy, s.targetCy, progress)
}

func (s *scrollList[T]) handleInput(t T, controls *input.Controls) {
	if dir := controls.DPad().JustPressedOrRepeatedDirection(); dir != input.NotPressed {
		s.items[s.highlightedIndex].DirectionJustPressed(t, dir)
		if dir == input.Up {
			s.moveHighlight(t, -1)
		} else if dir == input.Down {
			s.moveHighlight(t, 1)
		}
	}
	if controls.ButtonA().JustPressed() {
		s.items[s.highlightedIndex].ButtonAJustPressed(t)
	}
	if controls.ButtonStart().JustPressed() {
		s.items[s.highlightedIndex].ButtonStartJustPressed(t)
	}
}

func (s *scrollList[T]) moveHighlight(t T, dir int) {
	if s.highlightedIndex+dir < 0 || s.highlightedIndex+dir >= len(s.items) {
		log.Warn().Msgf("tried to move highlight by %d, but it was out of bounds", dir)
		return
	}
	newHighlight := s.highlightedIndex + dir
	for newHighlight >= 0 && newHighlight < len(s.items) && s.items[newHighlight].SkipHighlight(t) {
		newHighlight += dir
	}
	if newHighlight < 0 || newHighlight >= len(s.items) {
		return
	}
	s.highlight(newHighlight)
}

func (s *scrollList[T]) highlight(newHighlight int) {
	s.fromDy = s.currentDy
	s.fromCy = s.currentCy
	s.targetDy, s.targetCy = s.calcYs(newHighlight)

	s.timeTransitioning = 0
	s.highlightedIndex = newHighlight
}

func (s *scrollList[T]) highlightLast() {
	s.highlight(len(s.items) - 1)
}
