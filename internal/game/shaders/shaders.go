package shaders

import (
	_ "embed"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/gopxl/pixel/v2"
	"github.com/gopxl/pixel/v2/backends/opengl"
)

type Options interface {
	SetSwirlShader(
		center mgl32.Vec2,
		radius, swirl, falloff, progress float32,
	)
	Reset()
}

type Canvas struct {
	*opengl.Canvas
	currentShader string
}

func NewCanvas(width, height int) *Canvas {
	c := &Canvas{
		Canvas: opengl.NewCanvas(pixel.R(0, 0, float64(width), float64(height))),
	}
	return c
}

func (c *Canvas) Reset() {
	c.setFragmentShaderIfNeeded("", opengl.BaseCanvasFragmentShader)
}

func (c *Canvas) setFragmentShaderIfNeeded(name, src string) {
	if c.currentShader != name {
		c.currentShader = name
		c.SetFragmentShader(src)
	}
}

func RGBAtoVec3s(rgbs ...pixel.RGBA) []mgl32.Vec3 {
	var vecs []mgl32.Vec3
	for _, rgb := range rgbs {
		vecs = append(vecs, RGBAtoVec3(rgb))
	}
	return vecs
}

func RGBAtoVec3(rgb pixel.RGBA) mgl32.Vec3 {
	return mgl32.Vec3{float32(rgb.R), float32(rgb.G), float32(rgb.B)}
}
