package util

import (
	"fisherevans.com/project/f/internal/game/input"
)

type SelectableEntry[T any] struct {
	Value           T
	AdjacentEntries map[input.Direction]*SelectableEntry[T]
}

type Selectable[T any] struct {
	Entries       []*SelectableEntry[T]
	SelectedEntry *SelectableEntry[T]
}

func (s *Selectable[T]) UpdateSelection(c *input.Controls) {
	if !c.DPad().JustPressed() {
		return
	}
	next, found := s.SelectedEntry.AdjacentEntries[c.DPad().GetDirection()]
	if !found {
		return
	}
	s.SelectedEntry = next
}

func NewSimpleSelectable[T any](values []T) *Selectable[T] {
	var entries []*SelectableEntry[T]
	for id, v := range values {
		entry := &SelectableEntry[T]{
			Value:           v,
			AdjacentEntries: map[input.Direction]*SelectableEntry[T]{},
		}
		if id > 0 {
			entry.AdjacentEntries[input.Up] = entries[id-1]
			entries[id-1].AdjacentEntries[input.Down] = entry
		}
		entries = append(entries, entry)
	}
	return &Selectable[T]{
		Entries:       entries,
		SelectedEntry: entries[0],
	}
}
