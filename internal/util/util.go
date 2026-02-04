package util

import (
	"github.com/gopxl/pixel/v2"
)

func Ptr[T any](t T) *T {
	return &t
}

func OrDefaultString(value, orDefault string) string {
	if value == "" {
		return orDefault
	}
	return value
}

func R(x1 float64, y1 float64, w int, h int) pixel.Rect {
	return pixel.R(x1, y1, x1+float64(w), y1+float64(h))
}
