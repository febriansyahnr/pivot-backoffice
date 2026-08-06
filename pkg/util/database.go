package util

import (
	"encoding/json"
	"fmt"
)

// ScanJSON is a generic function that scans JSON data from a database value into a destination struct.
// It handles nil values by returning nil (no error) and verifies that the input value is a byte slice.
// If the input is not a byte slice, it returns an error.
func ScanJSON[T any](value interface{}, dest T) error {
	if value == nil {
		return nil
	}

	b, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("expected []byte, got %T", value)
	}

	return json.Unmarshal(b, dest)
}
