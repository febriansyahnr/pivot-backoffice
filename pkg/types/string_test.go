package types_test

import (
	"encoding/json"
	"testing"

	. "github.com/paper-indonesia/pivot-backoffice/pkg/types"

	"github.com/stretchr/testify/assert"
)

func TestString(t *testing.T) {
	type Data struct {
		Number String `json:"number"`
	}

	tests := []struct {
		input         []byte
		wantResult    String
		wantSerialize string
	}{
		{
			input:         []byte(`{"number": "01"}`),
			wantResult:    "01",
			wantSerialize: `{"number": "01"}`,
		},
		{
			input:         []byte(`{"number": null}`),
			wantResult:    "",
			wantSerialize: `{"number": ""}`,
		},
		{
			input:         []byte(`{"number": 2024}`),
			wantResult:    "2024",
			wantSerialize: `{"number": "2024"}`,
		},
		{
			input:         []byte(`{"number": 1}`),
			wantResult:    "1",
			wantSerialize: `{"number": "1"}`,
		},
	}
	for _, tt := range tests {
		data := Data{}
		assert.NoError(t, json.Unmarshal(tt.input, &data))
		assert.Equal(t, tt.wantResult, data.Number)

		raw, err := json.Marshal(data)
		assert.NoError(t, err)
		assert.JSONEq(t, string(tt.wantSerialize), string(raw))
	}
}
