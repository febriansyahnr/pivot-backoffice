package advanceairepository

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

type TestResponse struct {
	Data string `json:"data"`
}

func TestValidateHttpResponse(t *testing.T) {
	tests := []struct {
		name    string
		input   any
		wantErr bool
		errMsg  string
	}{
		{
			name: "SUCCESS: valid response",
			input: TestResponse{
				Data: "test data",
			},
			wantErr: false,
		},
		{
			name:    "FAIL: invalid json response",
			input:   "{invalid json",
			wantErr: true,
		},
		{
			name:    "FAIL: empty response body",
			input:   "",
			wantErr: true,
		},
		{
			name: "SUCCESS: complex nested response",
			input: TestResponse{
				Data: "complex test data with special chars: !@#$%^&*()",
			},
			wantErr: false,
		},
		{
			name:    "FAIL: malformed JSON",
			input:   `{"name": "John Doe" "age": 30}`,
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			if str, ok := tc.input.(string); ok {
				buf.WriteString(str)
			} else {
				err := json.NewEncoder(&buf).Encode(tc.input)
				assert.NoError(t, err)
			}

			result, err := ValidateHttpResponse[TestResponse](bytes.NewReader(buf.Bytes()))

			if tc.wantErr {
				assert.Error(t, err)
				if tc.errMsg != "" {
					assert.Equal(t, tc.errMsg, err.Error())
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}