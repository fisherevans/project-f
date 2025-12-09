package interp

import "math"

type Function func(float64) float64

// clamp clamps t into [0,1].
func clamp(t float64) float64 {
	if t < 0 {
		return 0
	}
	if t > 1 {
		return 1
	}
	return t
}

// Linear returns t in [0,1] unchanged.
func Linear(t float64) float64 {
	return clamp(t)
}

// Smoothstep is C1 continuous; slow start/end, faster middle.
// f(t) = t^2(3-2t)
func Smoothstep(t float64) float64 {
	t = clamp(t)
	return t * t * (3 - 2*t)
}

// SmoothstepParam provides a tunable smoothstep from linear to classic smoothstep.
// s in [0,1]:
//
//	s = 0 → linear (no easing at ends)
//	s = 1 → classic smoothstep (zero slope at ends)
//
// Values in-between interpolate the end slopes.
// Internally this uses a cubic Hermite from (0,0) to (1,1) with endpoint
// derivatives m0=m1=(1-s).
func SmoothstepParam(s float64) Function {
	return func(t float64) float64 {
		t = clamp(t)
		if s < 0 {
			s = 0
		} else if s > 1 {
			s = 1
		}
		m := 1 - s // endpoint slopes; 1=linear, 0=smooth
		t2 := t * t
		t3 := t2 * t
		// Hermite basis on [0,1]
		h01 := -2*t3 + 3*t2
		h10 := t3 - 2*t2 + t
		h11 := t3 - t2
		return h01 + m*h10 + m*h11
	}
}

// SmoothstepAsymmetrical lets you control the start and end smoothing independently.
// sIn, sOut in [0,1]: 0=linear slope at that end, 1=fully smoothed (zero slope).
// Example: SmoothstepAsym(t, 1, 0) eases in strongly but exits linearly.
func SmoothstepAsymmetrical(sIn, sOut float64) Function {
	return func(t float64) float64 {
		t = clamp(t)
		if sIn < 0 {
			sIn = 0
		} else if sIn > 1 {
			sIn = 1
		}
		if sOut < 0 {
			sOut = 0
		} else if sOut > 1 {
			sOut = 1
		}
		m0 := 1 - sIn  // left slope at t=0
		m1 := 1 - sOut // right slope at t=1
		t2 := t * t
		t3 := t2 * t
		// Hermite basis on [0,1]
		h01 := -2*t3 + 3*t2
		h10 := t3 - 2*t2 + t
		h11 := t3 - t2
		return h01 + m0*h10 + m1*h11
	}
}

func EaseOutQuad(t float64) float64 {
	return 1 - (1-t)*(1-t)
}

func EaseOutCubic(t float64) float64 {
	return 1 - math.Pow(1-t, 3)
}

// Smootherstep is C2 continuous; even gentler acceleration/deceleration.
// f(t) = t^3 (t(6t-15)+10)
func Smootherstep(t float64) float64 {
	t = clamp(t)
	return t * t * t * (t*(6*t-15) + 10)
}

// EaseInToLinear starts with zero slope and ends with unit slope.
// Cubic Hermite with m0=0, m1=1, plus a tunable center "bump" that
// preserves endpoints and endpoint slopes.
// Maps 0->0, 1->1, f'(0)=0, f'(1)=1.
// k > 0 raises the middle (stronger ease-in); k < 0 lowers it (stronger ease-out).
func EaseInToLinear(t, k float64) float64 {
	t = clamp(t)
	t2 := t * t
	t3 := t2 * t
	// Base cubic: 2 t^2 - t^3. Bump term: k * t^2 * (1 - t)^2.
	b := (1 - t)
	return 2*t2 - t3 + k*t2*b*b
}

// SmootherstepParam provides a tunable path from linear to classic smootherstep (C2).
// s in [0,1]:
//
//	s = 0 → linear (y=t) with end slopes 1
//	s = 1 → classic smootherstep (6t^5 - 15t^4 + 10t^3) with zero end slopes & curvature
//
// We use quintic Hermite with endpoint first-derivatives m0=m1=(1-s) and zero
// second-derivatives (k0=k1=0), ensuring C2 continuity across the range.
func SmootherstepParam(s float64) Function {
	if s < 0 {
		s = 0
	} else if s > 1 {
		s = 1
	}
	m := 1 - s // endpoint slopes: 1 at s=0 (linear), 0 at s=1 (smootherstep)
	return func(t float64) float64 {
		t = clamp(t)
		t2 := t * t
		t3 := t2 * t
		t4 := t3 * t
		t5 := t4 * t
		// Quintic Hermite basis (value, slope, curvature at both ends)
		H00 := 1 - 10*t3 + 15*t4 - 6*t5
		H10 := t - 6*t3 + 8*t4 - 3*t5
		H20 := 0.5*t2 - 1.5*t3 + 1.5*t4 - 0.5*t5
		H01 := 10*t3 - 15*t4 + 6*t5
		H11 := -4*t3 + 7*t4 - 3*t5
		H21 := 0.5*t3 - t4 + 0.5*t5
		// y0=0, y1=1, k0=k1=0, m0=m1=m
		return /*y0*/ 0*H00 + m*H10 + 0*H20 + /*y1*/ 1*H01 + m*H11 + 0*H21
	}
}

// SmootherstepAsymmetrical lets you set independent smoothing at start/end (C2).
// sIn, sOut in [0,1]: 0 = linear slope at that end, 1 = fully smoothed (zero slope).
// Uses quintic Hermite with zero end curvatures (k0=k1=0) and m0=1-sIn, m1=1-sOut.
func SmootherstepAsymmetrical(sIn, sOut float64) Function {
	if sIn < 0 {
		sIn = 0
	} else if sIn > 1 {
		sIn = 1
	}
	if sOut < 0 {
		sOut = 0
	} else if sOut > 1 {
		sOut = 1
	}
	m0 := 1 - sIn
	m1 := 1 - sOut
	return func(t float64) float64 {
		t = clamp(t)
		t2 := t * t
		t3 := t2 * t
		t4 := t3 * t
		t5 := t4 * t
		H00 := 1 - 10*t3 + 15*t4 - 6*t5
		H10 := t - 6*t3 + 8*t4 - 3*t5
		H20 := 0.5*t2 - 1.5*t3 + 1.5*t4 - 0.5*t5
		H01 := 10*t3 - 15*t4 + 6*t5
		H11 := -4*t3 + 7*t4 - 3*t5
		H21 := 0.5*t3 - t4 + 0.5*t5
		return 0*H00 + m0*H10 + 0*H20 + 1*H01 + m1*H11 + 0*H21
	}
}

// EaseInOutSine: sinusoidal ease; very smooth in/out.
// f(t) = 0.5*(1 - cos(pi*t))
func EaseInOutSine(t float64) float64 {
	t = clamp(t)
	return 0.5 * (1 - math.Cos(math.Pi*t))
}

type TimedProgress struct {
	duration      float64
	interpolation Function
	progress      float64
	reverse       bool
}

func NewTimedProgress(duration float64, interpolation Function) *TimedProgress {
	return &TimedProgress{
		duration:      duration,
		interpolation: interpolation,
	}
}

func (t *TimedProgress) Update(timeDelta float64) {
	pDelta := timeDelta / t.duration
	if t.reverse {
		pDelta = -pDelta
	}
	t.progress = clamp(t.progress + pDelta)
}

func (t *TimedProgress) IsComplete() bool {
	if t.reverse {
		return t.progress <= 0
	}
	return t.progress >= 1
}

func (t *TimedProgress) Progress() float64 {
	return t.interpolation(t.progress)
}

func (t *TimedProgress) Reset() {
	if t.reverse {
		t.progress = 1
	} else {
		t.progress = 0
	}
}

func (t *TimedProgress) Reverse() {
	t.reverse = !t.reverse
}

func (t *TimedProgress) SetFunction(f Function) {
	t.interpolation = f
}

func Lerp(from, to, progress float64) float64 {
	progress = clamp(progress)
	delta := to - from
	return from + delta*progress
}

func ParabolaUpLeft(x float64) float64 {
	return -math.Pow(x-1, 2) + 1
}

func ParabolaUpRight(x float64) float64 {
	return math.Pow(x, 2)
}
