package constant_test

import (
	"testing"

	. "github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/stretchr/testify/assert"
)

func TestTReconCode_Message(t *testing.T) {
	tests := []struct {
		name  string
		input TReconCode
		want  string
	}{
		{
			name:  "invalid amount",
			input: ReconCodeInvalidAmount,
			want:  "Invalid amount",
		},
		{
			name:  "invalid reference",
			input: ReconCodeInvalidReference,
			want:  "Invalid reference",
		},
		{
			name:  "invalid status",
			input: ReconCodeInvalidStatus,
			want:  "Invalid status",
		},
		{
			name:  "invalid date",
			input: ReconCodeIvalidDate,
			want:  "Invalid date",
		},
		{
			name:  "ok status",
			input: ReconCodeOk,
			want:  "OK",
		},
		{
			name:  "unknown code",
			input: TReconCode("unknown"),
			want:  "Unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.input.Message()
			assert.Equal(t, tt.want, got)
		})
	}
}
