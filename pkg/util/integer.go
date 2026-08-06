package util

import (
	"fmt"
	"math"
)

// SafeInt64ToUint64 safely converts an int64 to uint64,
// handling potential overflow (CWE-190).
// If the value is negative, it returns an error.
func SafeInt64ToUint64(value int64) (uint64, error) {
	if value < 0 {
		return 0, fmt.Errorf("cannot convert negative value %d to uint64", value)
	}
	return uint64(value), nil
}

// MustInt64ToUint64 converts an int64 to uint64 without returning an error.
// If the value is negative, it returns 0.
// This is a utility function for cases where an error cannot be handled
// and a reasonable fallback is needed.
func MustInt64ToUint64(value int64) uint64 {
	if value < 0 {
		// Return default value for negative inputs
		return 0
	}
	return uint64(value)
}

// SafeUint64ToInt64 safely converts a uint64 to int64,
// handling potential overflow (CWE-190).
// If the value exceeds math.MaxInt64, it returns an error.
func SafeUint64ToInt64(value uint64) (int64, error) {
	if value > uint64(math.MaxInt64) {
		return 0, fmt.Errorf("uint64 value %d exceeds max int64 value", value)
	}
	return int64(value), nil
}

// MustUint64ToInt64 converts a uint64 to int64, but returns a string representation
// if the value would overflow int64.
// This is useful when the value needs to be preserved as a string if it's too large.
func MustUint64ToInt64(value uint64) interface{} {
	if value > uint64(math.MaxInt64) {
		return fmt.Sprintf("%d", value)
	}
	return int64(value)
}

// SafeInt32ToUint32 safely converts an int32 to uint32,
// handling potential overflow (CWE-190).
// If the value is negative, it returns an error.
func SafeInt32ToUint32(value int32) (uint32, error) {
	if value < 0 {
		return 0, fmt.Errorf("cannot convert negative value %d to uint32", value)
	}
	return uint32(value), nil
}

// SafeUint32ToInt32 safely converts a uint32 to int32,
// handling potential overflow (CWE-190).
// If the value exceeds math.MaxInt32, it returns an error.
func SafeUint32ToInt32(value uint32) (int32, error) {
	if value > uint32(math.MaxInt32) {
		return 0, fmt.Errorf("uint32 value %d exceeds max int32 value", value)
	}
	return int32(value), nil
}

// SafeIntToUint16 safely converts an int to uint16,
// handling potential overflow (CWE-190).
// If the value is negative or exceeds math.MaxUint16, it returns an error.
func SafeIntToUint16(value int) (uint16, error) {
	if value < 0 {
		return 0, fmt.Errorf("cannot convert negative value %d to uint16", value)
	}
	if value > math.MaxUint16 {
		return 0, fmt.Errorf("value %d exceeds max uint16 value (%d)", value, math.MaxUint16)
	}
	return uint16(value), nil
}

// MustIntToUint16 converts an int to uint16 without returning an error.
// If the value is negative, it returns 0.
// If the value exceeds math.MaxUint16, it returns math.MaxUint16.
// This is a utility function for cases where an error cannot be handled
// and a reasonable fallback is needed.
func MustIntToUint16(value int) uint16 {
	if value < 0 {
		return 0
	}
	if value > math.MaxUint16 {
		return math.MaxUint16
	}
	return uint16(value)
}

// SafeIntToUint32 safely converts an int to uint32,
// handling potential overflow (CWE-190).
// If the value is negative or exceeds math.MaxUint32, it returns an error.
func SafeIntToUint32(value int) (uint32, error) {
	if value < 0 {
		return 0, fmt.Errorf("cannot convert negative value %d to uint32", value)
	}
	if value > math.MaxUint32 {
		return 0, fmt.Errorf("value %d exceeds max uint32 value", value)
	}
	return uint32(value), nil
}

// MustIntToUint32 converts an int to uint32 without returning an error.
// If the value is negative, it returns 0.
// If the value exceeds math.MaxUint32, it returns math.MaxUint32.
// This is a utility function for cases where an error cannot be handled
// and a reasonable fallback is needed.
func MustIntToUint32(value int) uint32 {
	if value < 0 {
		return 0
	}
	if value > math.MaxUint32 {
		return math.MaxUint32
	}
	return uint32(value)
}

// SafeUint16ToUint32 safely converts a uint16 to uint32.
// This is always safe as uint16 can fit within uint32,
// but is provided for consistency with other conversion functions.
func SafeUint16ToUint32(value uint16) uint32 {
	return uint32(value)
}

// MustUint16ToUint32 safely converts a uint16 to uint32.
// This is always safe as uint16 can fit within uint32,
// but is provided for consistency with other "Must" conversion functions.
func MustUint16ToUint32(value uint16) uint32 {
	return uint32(value)
}
