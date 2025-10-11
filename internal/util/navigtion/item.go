package navigtion

import (
	"fisherevans.com/project/f/internal/game/input"
)

type Item[T any] interface {
	OnHighlight(T)
	OnUnhighlight(T)
	OnButtonAJustPressed(T)
	OnButtonBJustPressed(T)
	OnDirectionJustPressed(input.Direction, T)
}

type BaseItem[T any] struct{}

func (i BaseItem[T]) OnHighlight(T)                             {}
func (i BaseItem[T]) OnUnhighlight(T)                           {}
func (i BaseItem[T]) OnButtonAJustPressed(T)                    {}
func (i BaseItem[T]) OnButtonBJustPressed(T)                    {}
func (i BaseItem[T]) OnDirectionJustPressed(input.Direction, T) {}

type SimpleItem[T any] struct {
	OnHighlightHandler   func(T)
	OnUnhighlightHandler func(T)
	OnButtonAHandler     func(T)
	OnButtonBHandler     func(T)
	OnDirectionHandler   func(input.Direction, T)
}

func NewSimpleItem[T any]() *SimpleItem[T] {
	return &SimpleItem[T]{}
}

func (i *SimpleItem[T]) WithHighlightHandler(f func(T)) *SimpleItem[T] {
	i.OnHighlightHandler = f
	return i
}

func (i *SimpleItem[T]) WithUnhighlightHandler(f func(T)) *SimpleItem[T] {
	i.OnUnhighlightHandler = f
	return i
}

func (i *SimpleItem[T]) WithButtonAHandler(f func(T)) *SimpleItem[T] {
	i.OnButtonAHandler = f
	return i
}

func (i *SimpleItem[T]) WithButtonBHandler(f func(T)) *SimpleItem[T] {
	i.OnButtonBHandler = f
	return i
}

func (i *SimpleItem[T]) WithDirectionHandler(f func(input.Direction, T)) *SimpleItem[T] {
	i.OnDirectionHandler = f
	return i
}

func (i *SimpleItem[T]) OnHighlight(t T) {
	if i.OnHighlightHandler != nil {
		i.OnHighlightHandler(t)
	}
}

func (i *SimpleItem[T]) OnUnhighlight(t T) {
	if i.OnUnhighlightHandler != nil {
		i.OnUnhighlightHandler(t)
	}
}

func (i *SimpleItem[T]) OnButtonAJustPressed(t T) {
	if i.OnButtonAHandler != nil {
		i.OnButtonAHandler(t)
	}
}

func (i *SimpleItem[T]) OnButtonBJustPressed(t T) {
	if i.OnButtonBHandler != nil {
		i.OnButtonBHandler(t)
	}
}

func (i *SimpleItem[T]) OnDirectionJustPressed(dir input.Direction, t T) {
	if i.OnDirectionHandler != nil {
		i.OnDirectionHandler(dir, t)
	}
}
