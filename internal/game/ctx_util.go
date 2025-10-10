package game

import (
	"math"
)

type ContextUtils struct {
	ctx *Context
}

func (u ContextUtils) TimeCycleSin(speed float64) float64 {
	return (math.Sin(u.ctx.Elapsed*speed*2.0*math.Pi) + 1.0) / 2.0
}
