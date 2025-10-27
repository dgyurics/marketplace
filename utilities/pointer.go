package utilities

// Ptr returns a pointer to the given value
func Ptr[T any](v T) *T {
	return &v
}

// String returns a pointer to the given string
func String(s string) *string { return &s }

// Int returns a pointer to the given int
func Int(i int) *int { return &i }

// Bool returns a pointer to the given bool
func Bool(b bool) *bool { return &b }

// Value safely dereferences a pointer with a default fallback
func Value[T any](ptr *T, defaultVal T) T {
	if ptr != nil {
		return *ptr
	}
	return defaultVal
}

// StringValue safely dereferences a string pointer
func StringValue(ptr *string, defaultVal string) string {
	if ptr != nil {
		return *ptr
	}
	return defaultVal
}

// IntValue safely dereferences an int pointer
func IntValue(ptr *int, defaultVal int) int {
	if ptr != nil {
		return *ptr
	}
	return defaultVal
}

// BoolValue safely dereferences a bool pointer
func BoolValue(ptr *bool, defaultVal bool) bool {
	if ptr != nil {
		return *ptr
	}
	return defaultVal
}
