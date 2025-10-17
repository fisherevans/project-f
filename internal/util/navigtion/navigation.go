package navigtion

import (
	"fisherevans.com/project/f/internal/game/input"
)

type System[T any] struct {
	highlighted Item[T]
	lastAdded   Item[T]
	items       map[Item[T]]*systemItem[T]
	neighbors   map[Item[T]]map[input.Direction]*neighbors[T]
}

func NewSystem[T any]() *System[T] {
	return &System[T]{
		items:     map[Item[T]]*systemItem[T]{},
		neighbors: map[Item[T]]map[input.Direction]*neighbors[T]{},
	}
}

type neighbors[T any] struct {
	items                     []Item[T]
	lastTransitionedFromIndex int
}

func (s *System[T]) addNeighbor(from Item[T], direction input.Direction, to Item[T]) {
	if s.neighbors == nil {
		s.neighbors = map[Item[T]]map[input.Direction]*neighbors[T]{}
	}
	if s.neighbors[from] == nil {
		s.neighbors[from] = map[input.Direction]*neighbors[T]{}
	}
	if s.neighbors[from][direction] == nil {
		s.neighbors[from][direction] = &neighbors[T]{}
	}
	s.neighbors[from][direction].items = append(s.neighbors[from][direction].items, to)
}

func (s *System[T]) AddItem(item Item[T], t T) *NeighborRegistry[T] {
	s.items[item] = &systemItem[T]{
		Item: item,
	}
	if s.highlighted == nil {
		s.Highlight(item, t)
	}
	s.lastAdded = item
	return s.NeighborsOf(item)
}

func (s *System[T]) AddNextToLast(direction input.Direction, t T, items ...Item[T]) *NeighborRegistry[T] {
	var r *NeighborRegistry[T]
	for _, item := range items {
		last := s.lastAdded
		r = s.AddItem(item, t)
		if last != nil {
			r.SetNeighbor(direction.Opposite(), last)
		}
	}
	return r
}

func (s *System[T]) SetLastItem(item Item[T]) {
	s.lastAdded = item
}

func (s *System[T]) Highlight(item Item[T], t T) {
	if s.highlighted != nil {
		s.highlighted.OnUnhighlight(t)
	}
	s.highlighted = item
	if s.highlighted != nil {
		s.highlighted.OnHighlight(t)
	}
}

func (s *System[T]) IsHighlighted(item Item[T]) bool {
	return s.highlighted == item
}

func (s *System[T]) HandleInputs(c *input.Controls, t T) {
	if s.highlighted == nil {
		return
	}
	from := s.highlighted
	if c.ButtonA().JustPressedOrRepeated() {
		from.OnButtonAJustPressed(t)
	}
	if c.ButtonB().JustPressedOrRepeated() {
		from.OnButtonBJustPressed(t)
	}
	if c.ButtonSelect().JustPressedOrRepeated() {
		from.OnButtonSelectJustPressed(t)
	}
	dpadPress := c.DPad().JustPressedOrRepeatedDirection()
	if dpadPress == input.NotPressed {
		return
	}
	from.OnDirectionJustPressed(dpadPress, t)
	var to Item[T]
	// find the new highlight item, call input handlers
	{
		directions := s.neighbors[from]
		if directions == nil {
			return
		}
		neighbors := directions[dpadPress]
		if neighbors == nil {
			return
		}
		if len(neighbors.items) == 0 {
			return
		}
		to = neighbors.items[neighbors.lastTransitionedFromIndex]
	}
	// remember last traversal in the case of multiple neighbors in the same direction
	{
		directions := s.neighbors[to]
		if directions != nil {
			neighbors := directions[dpadPress.Opposite()]
			if neighbors != nil {
				for index, neighborItem := range neighbors.items {
					if neighborItem == from {
						neighbors.lastTransitionedFromIndex = index
						break
					}
				}
			}
		}
	}
	// finally, highlight new node
	s.Highlight(to, t)

}

func (s *System[T]) NeighborsOf(item Item[T]) *NeighborRegistry[T] {
	return &NeighborRegistry[T]{
		s:    s,
		from: item,
	}
}

type systemItem[T any] struct {
	Item[T]
}

type NeighborRegistry[T any] struct {
	s    *System[T]
	from Item[T]
}

func (r *NeighborRegistry[T]) SetNeighbor(direction input.Direction, neighbor Item[T]) *NeighborRegistry[T] {
	r.s.addNeighbor(r.from, direction, neighbor)
	r.s.addNeighbor(neighbor, direction.Opposite(), r.from)
	return r
}

func (r *NeighborRegistry[T]) Item() Item[T] {
	return r.from
}

func (r *NeighborRegistry[T]) Below(other Item[T]) *NeighborRegistry[T] {
	return r.SetNeighbor(input.Up, other)
}

func (r *NeighborRegistry[T]) Above(other Item[T]) *NeighborRegistry[T] {
	return r.SetNeighbor(input.Down, other)
}

func (r *NeighborRegistry[T]) LeftOf(other Item[T]) *NeighborRegistry[T] {
	return r.SetNeighbor(input.Left, other)
}

func (r *NeighborRegistry[T]) RightOf(other Item[T]) *NeighborRegistry[T] {
	return r.SetNeighbor(input.Right, other)
}
