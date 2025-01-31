package util

func MinInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func MaxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func Clamp(min, value, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}
