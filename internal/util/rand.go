package util

import (
	"math"
	"math/rand"
	"time"

	"github.com/gopxl/pixel/v2"
)

func RandBetween(low, high float64) float64 {
	if low > high {
		swap := low
		low = high
		high = swap
	}
	diff := high - low
	return low + rand.Float64()*diff
}

// GenerateSpaced generated random, but evenly distributed points within a rectangle
func GenerateSpaced(x0, y0, width, height float64, count int, rng *rand.Rand) []pixel.Vec {
	if count <= 0 || width <= 0 || height <= 0 {
		return nil
	}

	if rng == nil {
		rng = rand.New(rand.NewSource(time.Now().UnixNano()))
	}

	rows, cols := chooseGrid(count, width, height)

	cellW := width / float64(cols)
	cellH := height / float64(rows)

	nCells := rows * cols
	if count > nCells {
		count = nCells
	}

	cells := make([]int, nCells)
	for i := 0; i < nCells; i++ {
		cells[i] = i
	}

	rng.Shuffle(nCells, func(i, j int) {
		cells[i], cells[j] = cells[j], cells[i]
	})

	points := make([]pixel.Vec, 0, count)
	for i := 0; i < count; i++ {
		idx := cells[i]
		r := idx / cols
		c := idx % cols

		cx0 := float64(c) * cellW
		cy0 := float64(r) * cellH

		x := x0 + cx0 + rng.Float64()*cellW
		y := y0 + cy0 + rng.Float64()*cellH

		points = append(points, pixel.Vec{X: x, Y: y})
	}

	return points
}

// chooseGrid picks a rows x cols grid so that:
//   - rows * cols >= count
//   - cells are not smaller than minCellSize if possible
//   - rows * cols is as small as possible (less wasted cells)
func chooseGrid(count int, width, height float64) (rows, cols int) {
	if count <= 0 {
		return 0, 0
	}

	const minCellSize = 2.0 // in pixels; tune to taste

	maxCols := int(math.Floor(width / minCellSize))
	maxRows := int(math.Floor(height / minCellSize))

	if maxCols < 1 {
		maxCols = 1
	}
	if maxRows < 1 {
		maxRows = 1
	}

	// If we cannot satisfy minCellSize with this count, fall back to a near-square grid.
	if maxCols*maxRows < count {
		side := int(math.Ceil(math.Sqrt(float64(count))))
		rows = side
		cols = int(math.Ceil(float64(count) / float64(rows)))
		return
	}

	bestArea := 0
	bestRows, bestCols := 1, count

	for r := 1; r <= maxRows; r++ {
		c := int(math.Ceil(float64(count) / float64(r)))
		if c < 1 || c > maxCols {
			continue
		}
		area := r * c
		if bestArea == 0 || area < bestArea {
			bestArea = area
			bestRows, bestCols = r, c
		}
	}

	// Safety fallback, though the loop above should always find something.
	if bestArea == 0 {
		side := int(math.Ceil(math.Sqrt(float64(count))))
		bestRows = side
		bestCols = int(math.Ceil(float64(count) / float64(bestRows)))
	}

	return bestRows, bestCols
}
