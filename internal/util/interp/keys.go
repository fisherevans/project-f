package interp

import (
	"cmp"
	"slices"
)

type key struct {
	progress, value float64
}

type Keys struct {
	keys []key
	fn   Function
}

func NewKeys() *Keys {
	return &Keys{
		fn: Linear,
	}
}

func (k *Keys) WithFunction(fn Function) *Keys {
	k.fn = fn
	return k
}

func (k *Keys) WithKey(progress, value float64) *Keys {
	k.keys = append(k.keys, key{progress, value})
	slices.SortFunc(k.keys, func(a, b key) int {
		return cmp.Compare(a.progress, b.progress)
	})
	return k
}

func (k *Keys) Interpolate(progress float64) float64 {
	if len(k.keys) == 0 {
		return 0
	}
	if progress <= k.keys[0].progress {
		return k.keys[0].value
	}
	if progress >= k.keys[len(k.keys)-1].progress {
		return k.keys[len(k.keys)-1].value
	}
	for i := 0; i < len(k.keys)-1; i++ {
		if k.keys[i].progress <= progress && k.keys[i+1].progress >= progress {
			a := k.keys[i].value
			b := k.keys[i+1].value
			p := k.fn((progress - k.keys[i].progress) / (k.keys[i+1].progress - k.keys[i].progress))
			return Lerp(a, b, p)
		}
	}
	return k.keys[len(k.keys)-1].value
}
