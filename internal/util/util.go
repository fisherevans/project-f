package util

func Ptr[T any](t T) *T {
	return &t
}

func OrDefaultString(value, orDefault string) string {
	if value == "" {
		return orDefault
	}
	return value
}
