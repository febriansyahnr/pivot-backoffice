package snapCoreRepository

import (
	"context"
	"errors"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	snapCoreVAModel "github.com/paper-indonesia/pivot-backoffice/internal/model/snapCore/virtualAccount"
	httpMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/httpRequestExt"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSnapCoreRepositoryInquiryStatusVAPayment(t *testing.T) {
	successResponsePaid := `{
		"data": {
			"responseCode": "2002800",
			"responseMessage": "Successful",
			"virtualAccountData": {
				"trxDateTime": "2025-01-10T10:30:00Z",
				"paidAmount": {"currency": "IDR", "value": "100000.00"}
			}
		}
	}`

	successResponsePending := `{
		"data": {
			"responseCode": "4002800",
			"responseMessage": "Pending"
		}
	}`

	testCases := []struct {
		name           string
		wantError      bool
		expectedIsPaid bool
		setupMock      func(mockHttp *httpMocks.IHTTPRequest)
	}{
		{
			name:           "SUCCESS: inquiry VA payment returns PAID status",
			wantError:      false,
			expectedIsPaid: true,
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On(
					"POST",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.Anything,
					mock.AnythingOfType(constant.MockTypeMapStringStringReference),
				).Return([]byte(successResponsePaid), 200, nil)
			},
		},
		{
			name:           "SUCCESS: inquiry VA payment returns PENDING status",
			wantError:      false,
			expectedIsPaid: false,
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On(
					"POST",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.Anything,
					mock.AnythingOfType(constant.MockTypeMapStringStringReference),
				).Return([]byte(successResponsePending), 200, nil)
			},
		},
		{
			name:      "ERROR: HTTP request error",
			wantError: true,
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On(
					"POST",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.Anything,
					mock.AnythingOfType(constant.MockTypeMapStringStringReference),
				).Return([]byte(""), 0, errors.New("network error"))
			},
		},
		{
			name:      "ERROR: invalid JSON response",
			wantError: true,
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On(
					"POST",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.Anything,
					mock.AnythingOfType(constant.MockTypeMapStringStringReference),
				).Return([]byte(`{invalid json}`), 200, nil)
			},
		},
		{
			name:      "ERROR: 400 bad request",
			wantError: true,
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On(
					"POST",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.Anything,
					mock.AnythingOfType(constant.MockTypeMapStringStringReference),
				).Return([]byte(`{"responseCode":"4002400","responseMessage":"Bad Request"}`), 400, nil)
			},
		},
		{
			name:      "ERROR: 500 internal server error",
			wantError: true,
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On(
					"POST",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.Anything,
					mock.AnythingOfType(constant.MockTypeMapStringStringReference),
				).Return([]byte(`{"responseCode":"5002400","responseMessage":"Internal Server Error"}`), 500, nil)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			mockHttp := httpMocks.NewIHTTPRequest(t)
			tc.setupMock(mockHttp)

			conf := &config.Config{SnapCoreConfig: config.SnapCoreConfig{BaseUrl: "https://api.example.com"}}
			secret := &config.Secret{SnapCoreSecret: struct {
				InternalServiceKey string "mapstructure:\"INTERNAL_SERVICE_KEY\""
			}{InternalServiceKey: "test-key"}}

			repo := New(conf, secret, mockLogger, mockHttp)

			request := &snapCoreVAModel.InquiryStatusVARequest{
				VirtualAccount: "7663123400000012",
			}

			result, err := repo.InquiryStatusVirtualAccount(context.Background(), request)

			if tc.wantError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tc.expectedIsPaid, result.IsPaid())
			}

			mockHttp.AssertExpectations(t)
		})
	}
}
