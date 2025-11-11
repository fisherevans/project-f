package adventure

import (
	"math"
	"math/rand/v2"

	"fisherevans.com/project/f/internal/game/audio"
)

var (
	FootstepFalloff  = NewFalloffFromNearFar(3, 12)
	ExplosionFalloff = NewFalloffFromNearFar(8, 48)
	ChatterFalloff   = NewFalloffFromNearFar(2, 4)
	StandardFalloff  = NewFalloffFromNearFar(3, 12)
)

type Falloff struct {
	NearDistance float64
	FarDistance  float64

	rolloffDBPerOctave float64
	floorDb            float64
}

// NewFalloffFromNearFar lets you define the curve as:
// - 0 dB attenuation from 0..near
// - floorDB attenuation at and beyond far
// The rolloff in dB/octave is computed so the curve hits floorDB at far.
func NewFalloffFromNearFar(near, far float64) Falloff {
	// Guard against invalid ranges
	if far <= near {
		far = near * 1.0001
	}
	floorDb := audio.SilenceGainDbThreshold
	// Solve: atten(far) = -R * log2(far/near) = floorDB  =>  R = -floorDB / log2(far/near)
	rolloff := -floorDb / math.Log2(far/near)
	return Falloff{
		NearDistance:       near,
		FarDistance:        far,
		floorDb:            floorDb,
		rolloffDBPerOctave: rolloff,
	}
}

// AttenuationDB returns a negative dB value to add to gain based on distance.
func (f Falloff) AttenuationDB(dist float64) float64 {
	if dist <= f.NearDistance {
		return 0
	}
	if f.FarDistance > 0 && dist >= f.FarDistance {
		return f.floorDb
	}
	curve := 0.3 // >1 = steeper near, flatter far. <1 = flatter near, steeper far.
	octaves := math.Log2(math.Pow(dist/f.NearDistance, curve))
	atten := -f.rolloffDBPerOctave * octaves
	return atten
}

func createStepSoundsHard() SoundSupplier {
	return NewSoundSupplierPool(false,
		NewSoundPool(true, "steps/hard/heavy_1", "steps/hard/heavy_2", "steps/hard/heavy_3"),
		NewSoundPool(true, "steps/hard/soft_1", "steps/hard/soft_2", "steps/hard/soft_3"),
	)
}

func createStepSoundsSoft() SoundSupplier {
	return NewSoundPool(true, "steps/soft/1", "steps/soft/2", "steps/soft/3", "steps/soft/4", "steps/soft/5", "steps/soft/6")
}

type SoundSupplier interface {
	Next() string
}

type SoundPool struct {
	names    []string
	isRandom bool
	next     int
}

func NewSoundPool(isRandom bool, names ...string) *SoundPool {
	return &SoundPool{
		names:    names,
		isRandom: isRandom,
	}
}

func (s *SoundPool) Next() string {
	if s.isRandom {
		s.next = rand.IntN(len(s.names))
	} else {
		s.next++
		if s.next >= len(s.names) {
			s.next = 0
		}
	}
	return s.names[s.next]
}

type SoundSupplierPool struct {
	suppliers []SoundSupplier
	isRandom  bool
	next      int
}

func NewSoundSupplierPool(isRandom bool, suppliers ...SoundSupplier) *SoundSupplierPool {
	return &SoundSupplierPool{
		suppliers: suppliers,
		isRandom:  isRandom,
	}
}

func (s *SoundSupplierPool) Next() string {
	if s.isRandom {
		s.next = rand.IntN(len(s.suppliers))
	} else {
		s.next++
		if s.next >= len(s.suppliers) {
			s.next = 0
		}
	}
	return s.suppliers[s.next].Next()
}
