package game

type FlagRegister struct {
	flags       map[string]any
	justChanged map[string]bool
}

func newFlags() *FlagRegister {
	return &FlagRegister{
		flags:       make(map[string]any),
		justChanged: make(map[string]bool),
	}
}

func (f *FlagRegister) Set(key string) {
	f.SetValue(key, true)
}

func (f *FlagRegister) SetValue(key string, value any) {
	f.flags[key] = value
	f.justChanged[key] = true
}

func (f *FlagRegister) Get(key string) bool {
	_, exists := f.GetValue(key)
	return exists
}

func (f *FlagRegister) GetValue(key string) (any, bool) {
	v, ok := f.flags[key]
	return v, ok
}

func (f *FlagRegister) JustChanged(key string) bool {
	changed := f.justChanged[key]
	if changed {
		f.justChanged[key] = false
	}
	return changed
}
