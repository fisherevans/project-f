package util

import (
	"fmt"
	"sort"
)

type FloatStats struct {
	frameTimes []float64
	maxSize    int
}

// NewFrameStats initializes the tracker
func NewFrameStats(maxSize int) *FloatStats {
	return &FloatStats{
		frameTimes: make([]float64, 0, maxSize),
		maxSize:    maxSize,
	}
}

// AddFrameTime adds a new frame time to the tracker
func (fs *FloatStats) AddFrameTime(dt float64) {
	if len(fs.frameTimes) == fs.maxSize {
		// Remove oldest frame time if at capacity
		fs.frameTimes = fs.frameTimes[1:]
	}
	fs.frameTimes = append(fs.frameTimes, dt)
}

// CalculateStats computes average, min, max, and percentiles
func (fs *FloatStats) CalculateStats() (avg, min, max, p50, p95 float64) {
	if len(fs.frameTimes) == 0 {
		return 0, 0, 0, 0, 0
	}

	// Sort a copy for percentile calculations
	sorted := make([]float64, len(fs.frameTimes))
	copy(sorted, fs.frameTimes)
	sort.Float64s(sorted)

	sum := 0.0
	min = sorted[0]
	max = sorted[len(sorted)-1]
	for _, t := range fs.frameTimes {
		sum += t
	}
	avg = sum / float64(len(fs.frameTimes))

	p50 = sorted[len(sorted)/2]
	if len(sorted) > 1 {
		p95 = sorted[int(0.95*float64(len(sorted)))]
	}

	return avg, min, max, p50, p95
}

func (fs *FloatStats) String() string {
	avg, min, max, p50, p95 := fs.CalculateStats()
	return fmt.Sprintf("avg: %.4f, min: %.4f, max: %.4f, p50: %.4f, p95: %.4f", avg, min, max, p50, p95)
}
