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
