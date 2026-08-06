package util

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestGenerateUUID(t *testing.T) {
	tests := []struct {
		name      string
		assertion func(t *testing.T)
	}{
		{
			name: "should generate valid UUID",
			assertion: func(t *testing.T) {
				result := GenerateUUID()

				assert.NotEqual(t, uuid.Nil, result, "UUID should not be nil")
				assert.NotEmpty(t, result.String(), "UUID string should not be empty")
			},
		},
		{
			name: "should generate UUID v7 when supported",
			assertion: func(t *testing.T) {
				result := GenerateUUID()

				// UUIDv7 has version bits set to 7 (0111)
				// The version is in the 13th byte (bits 48-51 of the UUID)
				version := result.Version()

				// The function should return v7 if NewV7() succeeds, otherwise v4
				assert.Contains(t, []uuid.Version{uuid.Version(4), uuid.Version(7)}, version,
					"UUID should be either v4 or v7")
			},
		},
		{
			name: "should generate unique UUIDs",
			assertion: func(t *testing.T) {
				uuid1 := GenerateUUID()
				uuid2 := GenerateUUID()

				assert.NotEqual(t, uuid1, uuid2, "consecutive UUIDs should be unique")
			},
		},
		{
			name: "should generate parseable UUID string",
			assertion: func(t *testing.T) {
				result := GenerateUUID()
				uuidString := result.String()

				parsed, err := uuid.Parse(uuidString)

				assert.NoError(t, err, "UUID string should be parseable")
				assert.Equal(t, result, parsed, "parsed UUID should match original")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.assertion(t)
		})
	}
}
