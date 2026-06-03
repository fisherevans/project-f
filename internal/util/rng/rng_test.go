package rng

import "testing"

func TestSetSeedReproducible(t *testing.T) {
	SetSeed(42)
	a := []float64{Float64(), Float64(), Float64()}
	ai := []int{IntN(1000), IntN(1000), IntN(1000)}

	SetSeed(42)
	b := []float64{Float64(), Float64(), Float64()}
	bi := []int{IntN(1000), IntN(1000), IntN(1000)}

	for i := range a {
		if a[i] != b[i] {
			t.Fatalf("Float64 draw %d differs after reseed: %v vs %v", i, a[i], b[i])
		}
		if ai[i] != bi[i] {
			t.Fatalf("IntN draw %d differs after reseed: %d vs %d", i, ai[i], bi[i])
		}
	}
}

func TestDifferentSeedsDiffer(t *testing.T) {
	SetSeed(1)
	a := Float64()
	SetSeed(2)
	b := Float64()
	if a == b {
		t.Fatalf("different seeds produced the same first draw: %v", a)
	}
}

func TestNewV1Reproducible(t *testing.T) {
	SetSeed(7)
	r1 := NewV1()
	SetSeed(7)
	r2 := NewV1()
	if r1.Int63() != r2.Int63() {
		t.Fatal("NewV1 generators seeded from the same stream position diverged")
	}
}

func TestIntNRange(t *testing.T) {
	SetSeed(99)
	for i := 0; i < 1000; i++ {
		if v := IntN(10); v < 0 || v >= 10 {
			t.Fatalf("IntN(10) out of range: %d", v)
		}
	}
}
