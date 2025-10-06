package typecast

// ToPtr converts an input value of any type to a pointer.
func ToPtr[T any](v T) *T {
	return &v
}
