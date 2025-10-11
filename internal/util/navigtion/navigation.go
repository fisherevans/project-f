package navigtion

import (
	"fisherevans.com/project/f/internal/game/input"
)

type BaseItem[T any] struct{}

func (i BaseItem[T]) OnHighlight(T)                             {}
func (i BaseItem[T]) OnUnhighlight(T)                           {}
func (i BaseItem[T]) OnButtonAJustPressed(T)                    {}
func (i BaseItem[T]) OnButtonBJustPressed(T)                    {}
func (i BaseItem[T]) OnDirectionJustPressed(input.Direction, T) {}

type Item[T any] interface {
	OnHighlight(T)
	OnUnhighlight(T)
	OnButtonAJustPressed(T)
	OnButtonBJustPressed(T)
	OnDirectionJustPressed(input.Direction, T)
}

type System[T any] struct {
	highlighted Item[T]
	lastAdded   Item[T]
	items       map[Item[T]]*systemItem[T]
	neighbors   map[Item[T]]map[input.Direction]Item[T]
}

func NewSystem[T any]() *System[T] {
	return &System[T]{
		items:     map[Item[T]]*systemItem[T]{},
		neighbors: map[Item[T]]map[input.Direction]Item[T]{},
	}
}

func (s *System[T]) addNeighbor(from Item[T], direction input.Direction, to Item[T]) {
	if s.neighbors == nil {
		s.neighbors = map[Item[T]]map[input.Direction]Item[T]{}
	}
	if s.neighbors[from] == nil {
		s.neighbors[from] = map[input.Direction]Item[T]{}
	}
	s.neighbors[from][direction] = to
}

func (s *System[T]) AddItem(item Item[T], t T) *NeighborRegistry[T] {
	s.items[item] = &systemItem[T]{
		Item: item,
	}
	if s.highlighted == nil {
		s.Highlight(item, t)
	}
	return s.NeighborsOf(item)
}

func (s *System[T]) AddNextToLast(direction input.Direction, t T, items ...Item[T]) *NeighborRegistry[T] {
	var r *NeighborRegistry[T]
	for _, item := range items {
		r = s.AddItem(item, t)
		if s.lastAdded != nil {
			r.SetNeighbor(direction.Opposite(), s.lastAdded)
		}
		s.lastAdded = item
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
	if c.ButtonA().JustPressedOrRepeated() {
		s.highlighted.OnButtonAJustPressed(t)
	}
	if c.ButtonB().JustPressedOrRepeated() {
		s.highlighted.OnButtonBJustPressed(t)
	}
	dpadPress := c.DPad().JustPressedOrRepeatedDirection()
	if dpadPress == input.NotPressed {
		return
	}
	s.highlighted.OnDirectionJustPressed(dpadPress, t)
	neighbors := s.neighbors[s.highlighted]
	if neighbors == nil {
		return
	}
	neighbor := neighbors[dpadPress]
	if neighbor == nil {
		return
	}
	s.Highlight(neighbor, t)
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
	return r.SetNeighbor(input.Down, other)
}

func (r *NeighborRegistry[T]) Above(other Item[T]) *NeighborRegistry[T] {
	return r.SetNeighbor(input.Up, other)
}

func (r *NeighborRegistry[T]) LeftOf(other Item[T]) *NeighborRegistry[T] {
	return r.SetNeighbor(input.Left, other)
}

func (r *NeighborRegistry[T]) RightOf(other Item[T]) *NeighborRegistry[T] {
	return r.SetNeighbor(input.Right, other)
}

func (r *NeighborRegistry[T]) GetNeighborInDirection(direction input.Direction) *NeighborRegistry[T] {
	if r.s.neighbors[r.from] == nil {
		return nil
	}
	if r.s.neighbors[r.from][direction] == nil {
		return nil
	}
	return r.s.NeighborsOf(r.s.neighbors[r.from][direction])
}
