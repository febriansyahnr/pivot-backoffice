package fraudnetrepository

import (
	"bytes"
	"encoding/json"
	"testing"

	fdscommon "github.com/paper-indonesia/pivot-backoffice/internal/model/fdsProcessor/fdsCommon"
	fraudnetmodel "github.com/paper-indonesia/pivot-backoffice/internal/model/fdsProcessor/fraudNet"
	"github.com/stretchr/testify/assert"
)

func toPtr[T any](v T) *T {
	return &v
}

func TestCreateBasicAuth(t *testing.T) {
	tests := []struct {
		name     string
		username string
		password string
		expected string
	}{
		{
			name:     "SUCCESS: simple credentials",
			username: "Aladdin",
			password: "OpenSesame",
			expected: "Basic QWxhZGRpbjpPcGVuU2VzYW1l",
		},
		{
			name:     "SUCCESS: empty password",
			username: "demo",
			password: "",
			expected: "Basic ZGVtbzo=",
		},
		{
			name:     "SUCCESS: empty username",
			username: "",
			password: "password",
			expected: "Basic OnBhc3N3b3Jk",
		},
		{
			name:     "SUCCESS: special characters",
			username: "user!@#",
			password: "p@$$w0rd",
			expected: "Basic dXNlciFAIzpwQCQkdzByZA==",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := CreateBasicAuth(tc.username, tc.password)
			assert.Equal(t, tc.expected, got)
		})
	}
}

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
			name: "SUCCESS: valid response ",
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

func TestCheckRequestMapToCommonResponse(t *testing.T) {
	tests := []struct {
		name     string
		input    *fraudnetmodel.MarketplaceCheckResponse
		expected *fdscommon.CheckResponse
	}{
		{
			name: "SUCCESS: valid mapping",
			input: &fraudnetmodel.MarketplaceCheckResponse{
				Success: true,
				Code:    toPtr("200"),
				Source:  toPtr("fraudnet"),
				Message: "ok",
				Data: fraudnetmodel.MarketplaceCheckData{
					ID:        "txn123",
					Timer:     123,
					RiskScore: 80,
					RiskGroup: "HIGH",
					Link:      "https://example.com/check",
					Tags: []fraudnetmodel.MarketplaceCheckTags{
						{
							ID:        "tag1",
							Action:    toPtr("block"),
							Name:      "Fraudulent",
							Source:    "fn",
							Type:      "risk",
							State:     toPtr("active"),
							Weight:    toPtr(10),
							RiskScore: toPtr(90),
							RiskGroup: toPtr("HIGH"),
							Link:      toPtr("https://example.com/tag1"),
						},
					},
				},
			},
			expected: &fdscommon.CheckResponse{
				Success: true,
				Code:    toPtr("200"),
				Source:  toPtr("fraudnet"),
				Message: "ok",
				Data: fdscommon.CheckData{
					ID:        "txn123",
					Timer:     123,
					RiskScore: 80,
					RiskGroup: "HIGH",
					Link:      "https://example.com/check",
					Tags: []fdscommon.CheckTags{
						{
							ID:        "tag1",
							Action:    toPtr("block"),
							Name:      "Fraudulent",
							Source:    "fn",
							Type:      "risk",
							State:     toPtr("active"),
							Weight:    toPtr(10),
							RiskScore: toPtr(90),
							RiskGroup: toPtr("HIGH"),
							Link:      toPtr("https://example.com/tag1"),
						},
					},
				},
			},
		},
		{
			name:     "FAIL: nil input returns nil",
			input:    nil,
			expected: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := CheckRequestMapToCommonResponse(tc.input)
			assert.Equal(t, tc.expected, got)
		})
	}
}

func TestUpdateRequestMapToCommonResponse(t *testing.T) {
	tests := []struct {
		name     string
		input    *fraudnetmodel.MarketplaceUpdateResponse
		expected *fdscommon.UpdateResponse
	}{
		{
			name: "SUCCESS: valid mapping",
			input: &fraudnetmodel.MarketplaceUpdateResponse{
				Success: true,
				Code:    toPtr("202"),
				Source:  toPtr("fraudnet"),
				Message: toPtr("update successful"),
				Data: fraudnetmodel.MarketplaceUpdateData{
					ID:    "txn789",
					Link:  "https://example.com/update",
					Timer: 456,
				},
			},
			expected: &fdscommon.UpdateResponse{
				Success: true,
				Code:    toPtr("202"),
				Source:  toPtr("fraudnet"),
				Message: toPtr("update successful"),
				Data: fdscommon.UpdateData{
					ID:    "txn789",
					Link:  "https://example.com/update",
					Timer: 456,
				},
			},
		},
		{
			name:     "FAIL: nil input returns nil",
			input:    nil,
			expected: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := UpdateRequestMapToCommonResponse(tc.input)
			assert.Equal(t, tc.expected, got)
		})
	}
}
