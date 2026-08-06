package util

// ToPointer return the pointer of the value
func ValueToPtr[T comparable](v T) *T {
	return &v
}

// PointerValue return value of the pointer
// when the pointer was nil, it will return default value of the type
func ValueOfPtr[T any](p *T) T {
	if p != nil {
		return *p
	}

	var v T
	return v
}

// BoolPtr returns a pointer to the given bool value.
//
// Deprecated: Use ValueToPtr instead.
func BoolPtr(b bool) *bool {
	return &b
}

// ClonePtr creates a deep copy of the pointer's value and returns a new pointer.
// If the source pointer is nil, it returns nil.
func ClonePtr[T any](src *T) *T {
	if src == nil {
		return nil
	}

	dst := *src
	return &dst
}
