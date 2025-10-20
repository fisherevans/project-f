package events

type ReadableObject interface {
	Get(key string) (any, bool)
}

type MutableObject interface {
	ReadableObject
	Set(key string, value any)
}

func (d *Dispatcher) NewObject() MutableObject {
	return &mapState{
		data: make(map[string]any),
	}
}

func NewObject() MutableObject {
	return &mapState{
		data: make(map[string]any),
	}
}

type mapState struct {
	data map[string]any
}

func (s *mapState) Get(key string) (any, bool) {
	if s.data == nil {
		return nil, false
	}
	val, exists := s.data[key]
	return val, exists
}

func (s *mapState) Set(key string, value any) {
	if s.data == nil {
		s.data = make(map[string]any)
	}
	s.data[key] = value
}

func (s *mapState) ToMap() map[string]any {
	return s.data
}
