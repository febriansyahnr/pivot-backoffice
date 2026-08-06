package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSliceExtractOrDefault(t *testing.T) {
	t.Run("Extract string from valid index", func(t *testing.T) {
		slice := []string{"a", "b", "c"}
		result := SliceExtractOrDefault(slice, 1, "default")
		assert.Equal(t, "b", result)
	})

	t.Run("Extract int from valid index", func(t *testing.T) {
		slice := []int{1, 2, 3}
		result := SliceExtractOrDefault(slice, 0, 999)
		assert.Equal(t, 1, result)
	})

	t.Run("Return default when index out of bounds", func(t *testing.T) {
		slice := []string{"a", "b"}
		result := SliceExtractOrDefault(slice, 5, "default")
		assert.Equal(t, "default", result)
	})

	t.Run("Return default for empty slice", func(t *testing.T) {
		var slice []int
		result := SliceExtractOrDefault(slice, 0, 42)
		assert.Equal(t, 42, result)
	})

	t.Run("Extract float from valid index", func(t *testing.T) {
		slice := []float64{1.1, 2.2, 3.3}
		result := SliceExtractOrDefault(slice, 2, 9.9)
		assert.Equal(t, 3.3, result)
	})
}
