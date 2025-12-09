package interp

import (
	"cmp"
	"slices"
)

type key struct {
	progress, value float64
	// fn is what is applied BEFORE this key (to get to it)
	fn Function
}

type Keys struct {
	keys            []key
	defaultFunction Function
}

func NewKeys() *Keys {
	return &Keys{
		defaultFunction: Linear,
	}
}

func (k *Keys) WithDefaultFunction(fn Function) *Keys {
	k.defaultFunction = fn
	return k
}

func (k *Keys) WithKey(progress, value float64) *Keys {
	return k.WithKeyFn(progress, value, nil)
}

// WithKeyFn lets you supply a fn which is used leading up to the value
func (k *Keys) WithKeyFn(progress, value float64, fn Function) *Keys {
	k.keys = append(k.keys, key{progress: progress, value: value, fn: fn})
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
			fn := k.defaultFunction
			if k.keys[i+1].fn != nil {
				fn = k.keys[i+1].fn
			}
			p := fn((progress - k.keys[i].progress) / (k.keys[i+1].progress - k.keys[i].progress))
			return Lerp(a, b, p)
		}
	}
	return k.keys[len(k.keys)-1].value
}
