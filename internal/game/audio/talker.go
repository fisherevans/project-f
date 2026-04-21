//go:build !js

package audio

import (
	"math"
	"math/rand"
	"strings"
	"sync"
	"time"

	"fisherevans.com/project/f/internal/util/interp"
)

type Talker struct {
	system          *System
	speedAdjustment float64
	volume          float64
	throttleScale   float64

	mu             sync.Mutex
	lastLetterTime time.Time // last time a letter sound was played
	suppressUntil  time.Time // suppress letters until this time (for spaces)
}

// NewTalker creates a new talker with voice characteristics.
// speed: 1.0 = normal, <1.0 = deeper, >1.0 = higher
// volume: silent 0-1 loud
// throttle: 1.0 = normal, <1.0 = slower, >1.0 = faster (base multiplier)
func (a *System) NewTalker(speed, volume, throttle float64) *Talker {
	return &Talker{
		system:          a,
		speedAdjustment: speed,
		volume:          volume,
		throttleScale:   throttle,
	}
}

const baseSpeedShift = .15
const speedVariance = 0.1
const minVolumeVarianceScale = 0.8
const letterThrottle = 60 * time.Millisecond
const minLetterThrottle = 20 * time.Millisecond
const spacePauseDuration = 50 * time.Millisecond

func (t *Talker) Speak(text string, rate float64) {
	text = strings.ToLower(text)
	throttleScale := t.throttleScale / rate
	baseSpeed := 1.0 + baseSpeedShift + t.speedAdjustment + (1.0-rate)*0.1
	baseVolume := t.volume * (math.Log(rate)*0.5 + 1.0)
	for _, c := range text {
		speed := baseSpeed + rand.Float64()*speedVariance - speedVariance/2.0
		minVolume := baseVolume * minVolumeVarianceScale
		volume := interp.Lerp(minVolume, baseVolume, rand.Float64())

		if c >= 'a' && c <= 'z' {
			// Letters are throttled but play polyphonically
			speed, volume = t.getLetterModulation(c, speed, volume)
			t.playLetter(speed, volume, throttleScale)
		} else if c == ',' || c == '.' || c == ';' {
			if c == ',' || c == ';' {
				volume *= 0.9
			}
			// Punctuation always plays
			t.playPunctuation(speed, volume, "end")
		} else if c == '!' {
			// Punctuation always plays
			t.playPunctuation(speed, volume, "exclaim")
		} else if c == '?' {
			// Punctuation always plays
			t.playPunctuation(speed, volume, "question")
		} else if c == ' ' {
			// Spaces suppress letters for a brief period
			t.handleSpace(throttleScale)
		}
	}
}

func (t *Talker) getLetterModulation(c rune, baseSpeed, baseVolume float64) (speed, volume float64) {
	// Vowels: slightly higher speed, fuller sound
	vowels := "aeiou"
	if strings.ContainsRune(vowels, c) {
		speed = baseSpeed - 0.05
		volume = baseVolume * 0.8
	} else {
		// Consonants: lower speed, quieter
		speed = baseSpeed + 0.05
		volume = baseVolume
	}
	return
}

func (t *Talker) playLetter(speed, volume, throttleScale float64) {
	t.mu.Lock()
	defer t.mu.Unlock()

	now := time.Now()

	// Check if we're in a suppression period (from a space)
	if now.Before(t.suppressUntil) {
		return
	}

	// Adjust throttle based on throttle (higher throttle = shorter throttle)
	throttleDur := max(time.Duration(float64(letterThrottle)*throttleScale), minLetterThrottle)

	// Check throttle - only play if enough time has passed since last letter
	if now.Sub(t.lastLetterTime) < throttleDur {
		return
	}

	// Play the letter sound (polyphonic - no callback needed)
	t.system.PlaySoundOnBus(
		"speech/letter",
		t.system.Buses.SFX,
		volume,
		&PlaybackOptions{Speed: speed},
	)

	t.lastLetterTime = now
}

func (t *Talker) playPunctuation(speed, volume float64, soundName string) {
	t.system.PlaySoundOnBus(
		"speech/"+soundName,
		t.system.Buses.SFX,
		volume,
		&PlaybackOptions{Speed: speed},
	)
	t.lastLetterTime = time.Now()
}

func (t *Talker) handleSpace(throttleScale float64) {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Adjust pause duration based on throttle (higher throttle = shorter pause)
	pauseDuration := time.Duration(float64(spacePauseDuration) * throttleScale)

	// Set suppression time to create a pause
	t.suppressUntil = time.Now().Add(pauseDuration)
}
