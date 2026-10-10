package utilities

// Ptr returns a pointer to the given value
func Ptr[T any](v T) *T {
	return &v
}

// Value safely dereferences a pointer with a default fallback
func Value[T any](ptr *T, defaultVal T) T {
	if ptr != nil {
		return *ptr
	}
	return defaultVal
}
