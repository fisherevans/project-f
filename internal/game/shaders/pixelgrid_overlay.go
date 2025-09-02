package shaders

import (
	_ "embed"
)

// Uniform names for pixelgrid_overlay.glsl
const ()

//go:embed pixelgrid_overlay.glsl
var pixelgridOverlayFrag string

// SetPixelGridOverlayShader sets the LCD/CRT pixel-grid overlay shader and uniforms.
// Note: bool is passed as int32 (0/1) because the GL uniform helper doesn't handle bools.
func (c *Canvas) SetPixelGridOverlayShader(
	uPixelSize, uScanlineDarken, uGridDarkenX, uGridDarkenY, uSubpixelTint float32) {
	c.SetUniform("uPixelSize", uPixelSize)
	c.SetUniform("uScanlineDarken", uScanlineDarken)
	c.SetUniform("uGridDarkenX", uGridDarkenX)
	c.SetUniform("uGridDarkenY", uGridDarkenY)
	c.SetUniform("uSubpixelTint", uSubpixelTint)
	c.SetFragmentShader(pixelgridOverlayFrag)
}
