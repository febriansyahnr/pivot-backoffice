package util

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestGenerateShortCode(t *testing.T) {

	testCases := []struct {
		name     string
		uniqueID string
	}{
		{
			name:     "SUCCESS: generate short code with unique ID",
			uniqueID: "unique-123",
		},
		{
			name:     "SUCCESS: generate short code with uuid",
			uniqueID: uuid.NewString(),
		},
		{
			name:     "SUCCESS: generate short code without unique ID",
			uniqueID: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := GenerateShortCode(tc.uniqueID)
			assert.NotEmpty(t, result)
		})
	}
}
