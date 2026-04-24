package schema

// Pixels is an integer pixel dimension used by sprite metadata.
type Pixels int

func (p Pixels) Float() float64 {
    return float64(p)
}

func (p Pixels) Int() int {
    return int(p)
}
