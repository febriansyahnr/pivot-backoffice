package commonModel_test

import (
	"testing"
	"time"

	. "github.com/paper-indonesia/pivot-backoffice/internal/model/common"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExportResponse(t *testing.T) {
	input := ExportResponse{
		DownloadURL: "https://",
		ExpiresAt:   time.Date(2025, 3, 1, 17, 0, 0, 0, time.UTC),
	}
	raw, err := input.MarshalBinary()
	require.NoError(t, err)
	require.NotNil(t, raw)

	result := ExportResponse{}
	require.NoError(t, result.UnmarshalBinary(raw))

	assert.Equal(t, input, result)
}
