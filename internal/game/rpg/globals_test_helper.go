package rpg

func NewTestGlobalValue(key string, value any, exists bool) *GlobalValue {
	return newGlobalValue(key, value, exists)
}

func NewTestGlobals(values map[string]any) Globals {
	return &defaultGlobals{values: values}
}
