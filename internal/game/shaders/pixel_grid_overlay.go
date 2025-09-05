package shaders

import (
	_ "embed"
)

//go:embed pixel_grid_overlay.glsl
var pixelGridOverlayFrag string

func (c *Canvas) SetPixelGridOverlayShader(
	uPixelSize, uScanlineDarken, uGridDarkenX, uGridDarkenY, uSubpixelTint float32) {
	c.SetUniform("uPixelSize", uPixelSize)
	c.SetUniform("uScanlineDarken", uScanlineDarken)
	c.SetUniform("uGridDarkenX", uGridDarkenX)
	c.SetUniform("uGridDarkenY", uGridDarkenY)
	c.SetUniform("uSubpixelTint", uSubpixelTint)
	c.setFragmentShaderIfNeeded("pixel_grid", pixelGridOverlayFrag)
}
