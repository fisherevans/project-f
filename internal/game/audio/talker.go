package audio

import (
	"math/rand"
	"strings"
	"sync"
	"time"
)

type Talker struct {
	system          *System
	speedAdjustment float64
	gainAdjustment  float64
	throttleScale   float64

	mu             sync.Mutex
	lastLetterTime time.Time // last time a letter sound was played
	suppressUntil  time.Time // suppress letters until this time (for spaces)
}

// NewTalker creates a new talker with voice characteristics.
// speed: 1.0 = normal, <1.0 = deeper, >1.0 = higher
// gain: 0.0 = normal volume, negative = quieter, positive = louder (in dB)
// throttle: 1.0 = normal, <1.0 = slower, >1.0 = faster (base multiplier)
func (a *System) NewTalker(speed, gain, throttle float64) *Talker {
	return &Talker{
		system:          a,
		speedAdjustment: speed,
		gainAdjustment:  gain,
		throttleScale:   throttle,
	}
}

const baseSpeedShift = .15
const speedVariance = 0.1
const defaultBaseGain = -1.0
const gainVariance = 0.5
const letterThrottle = 60 * time.Millisecond
const minLetterThrottle = 20 * time.Millisecond
const spacePauseDuration = 50 * time.Millisecond

func (t *Talker) Speak(text string, rate float64) {
	text = strings.ToLower(text)
	throttleScale := t.throttleScale / rate
	baseSpeed := 1.0 + baseSpeedShift + t.speedAdjustment + (1.0-rate)*0.1
	baseGain := defaultBaseGain + t.gainAdjustment + (1.0-rate)*0.5
	for _, c := range text {
		speed := baseSpeed + rand.Float64()*speedVariance - speedVariance/2.0
		gain := baseGain + rand.Float64()*gainVariance - gainVariance/2.0

		if c >= 'a' && c <= 'z' {
			// Letters are throttled but play polyphonically
			speed, gain = t.getLetterModulation(c, speed, gain)
			t.playLetter(speed, gain, throttleScale)
		} else if c == ',' || c == '.' || c == ';' {
			if c == ',' || c == ';' {
				gain += -1.0
			}
			// Punctuation always plays
			t.playPunctuation(speed, gain, "end")
		} else if c == '!' {
			// Punctuation always plays
			t.playPunctuation(speed, gain, "exclaim")
		} else if c == '?' {
			// Punctuation always plays
			t.playPunctuation(speed, gain, "question")
		} else if c == ' ' {
			// Spaces suppress letters for a brief period
			t.handleSpace(throttleScale)
		}
	}
}

func (t *Talker) getLetterModulation(c rune, baseSpeed, baseGain float64) (speed, gain float64) {
	// Vowels: slightly higher speed, fuller sound
	vowels := "aeiou"
	if strings.ContainsRune(vowels, c) {
		speed = baseSpeed - 0.05
		gain = baseGain - 0.2
	} else {
		// Consonants: lower speed, quieter
		speed = baseSpeed + 0.05
		gain = baseGain + 0.1
	}
	return
}

func (t *Talker) playLetter(speed, gain, throttleScale float64) {
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
		gain,
		&PlaybackOptions{Speed: speed},
	)

	t.lastLetterTime = now
}

func (t *Talker) playPunctuation(speed, gain float64, soundName string) {
	t.system.PlaySoundOnBus(
		"speech/"+soundName,
		t.system.Buses.SFX,
		gain,
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
