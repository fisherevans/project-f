//go:build !js

package speech

import (
	"math"
	"math/rand"
	"strings"
	"time"
	"unicode"

	"github.com/gopxl/beep/v2"
)

// TickGenerator renders the hollow tick sounds (formerly TickHollow preset).
type TickGenerator struct {
	sampleRate beep.SampleRate

	charDuration float64 // seconds
	gapDuration  float64 // seconds
}

// NewTickGenerator creates a generator with default hollow tick timings.
func NewTickGenerator(sr beep.SampleRate) *TickGenerator {
	return &TickGenerator{
		sampleRate:   sr,
		charDuration: 0.04,
		gapDuration:  0.00, // 1,
	}
}

// Speak emits the hollow tick sequence for the provided text.
func (g *TickGenerator) Speak(text string) beep.Streamer {
	if text == "" {
		return beep.Silence(1)
	}

	text = strings.TrimSpace(text)

	var streamers []beep.Streamer

	for _, r := range text {
		if unicode.IsSpace(r) {
			gap := time.Duration(g.gapDuration * 2 * float64(time.Second))
			streamers = append(streamers, beep.Silence(g.sampleRate.N(gap)))
			continue
		}

		tick := g.generateTick()
		gap := beep.Silence(g.sampleRate.N(time.Duration(g.gapDuration * float64(time.Second))))
		streamers = append(streamers, tick, gap)
	}

	if len(streamers) == 0 {
		return beep.Silence(1)
	}
	return beep.Seq(streamers...)
}

func (g *TickGenerator) generateTick() beep.Streamer {
	duration := time.Duration(g.charDuration * float64(time.Second))
	samples := g.sampleRate.N(duration)

	tone := &hollowTickTone{
		sampleRate:   g.sampleRate,
		totalSamples: samples,
		bodyAmount:   0.45,
		bodyFreq:     320,
		toneAmount:   0.6,
		toneFreq:     730,
		ringFreq:     304.7,
	}

	return beep.Take(samples, tone)
}

type hollowTickTone struct {
	sampleRate   beep.SampleRate
	totalSamples int

	sampleCount int

	bodyAmount float64
	bodyFreq   float64
	toneAmount float64
	toneFreq   float64
	ringFreq   float64

	phase     float64
	bodyPhase float64
	tonePhase float64
}

func (t *hollowTickTone) Stream(samples [][2]float64) (int, bool) {
	attack := int(float64(t.totalSamples) * 0.08)
	releaseStart := int(float64(t.totalSamples) * 0.35)
	for i := range samples {
		if t.sampleCount >= t.totalSamples {
			return i, false
		}

		progress := float64(t.sampleCount) / float64(t.totalSamples)

		decay := math.Exp(-6.8 * progress)

		noise := rand.NormFloat64() * 0.2
		noise += math.Sin(t.tonePhase*2*math.Pi) * 0.05

		ring := math.Sin(t.phase*2*math.Pi*t.ringFreq) * 0.12

		body := 0.0
		if t.bodyAmount > 0 {
			bodyDecay := math.Exp(-4.0 * progress)
			body = math.Sin(t.bodyPhase*2*math.Pi) * t.bodyAmount * bodyDecay
			_, t.bodyPhase = math.Modf(t.bodyPhase + t.bodyFreq/float64(t.sampleRate))
		}

		tone := 0.0
		if t.toneAmount > 0 {
			toneDecay := math.Exp(-2.8 * progress)
			tone = math.Sin(t.tonePhase*2*math.Pi) * t.toneAmount * toneDecay
			_, t.tonePhase = math.Modf(t.tonePhase + t.toneFreq/float64(t.sampleRate))
		}

		val := (noise + ring + body + tone) * decay

		if t.sampleCount < attack && attack > 0 {
			val *= float64(t.sampleCount) / float64(attack)
		} else if t.sampleCount > releaseStart {
			releaseSamples := t.totalSamples - releaseStart
			if releaseSamples > 0 {
				remain := t.totalSamples - t.sampleCount
				val *= float64(remain) / float64(releaseSamples)
			}
		}

		val *= 0.4
		samples[i][0] = val
		samples[i][1] = val

		phaseFreq := 260.0
		phaseFreq *= 1.0
		_, t.phase = math.Modf(t.phase + phaseFreq/float64(t.sampleRate))

		t.sampleCount++
	}

	return len(samples), true
}

func (t *hollowTickTone) Err() error { return nil }
