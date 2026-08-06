package snapCoreRepository_test

import (
	"context"
	"errors"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	snapCoreModelQr "github.com/paper-indonesia/pivot-backoffice/internal/model/snapCore/qr"
	. "github.com/paper-indonesia/pivot-backoffice/internal/repository/snapCore"
	httpMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/httpRequestExt"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSnapCoreRepositoryInquiryStatusQris(t *testing.T) {
	successResponse := `{
		"code": "SNP-CR-00",
		"message": "OK",
		"data": {
			"responseCode": "2005500",
			"responseMessage": "Successful",
			"uuid": "test-uuid-123",
			"transactionId": "TXN123456",
			"partnerReferenceNo": "PARTNER-REF-001",
			"acquirerReferenceNo": "ACQ-REF-001",
			"status": "PAID",
			"qrType": "DYNAMIC",
			"acquirer": "BCA",
			"amount": {
				"value": "100000.00",
				"currency": "IDR"
			}
		}
	}`

	notFoundResponse := `{
		"code": "SNP-CR-04",
		"message": "Not Found",
		"data": {
			"responseCode": "4045500",
			"responseMessage": "Transaction Not Found"
		}
	}`

	conflictResponse := `{
		"code": "SNP-CR-09",
		"message": "Conflict",
		"data": {
			"responseCode": "4095500",
			"responseMessage": "Conflict"
		}
	}`

	testCases := []struct {
		name           string
		request        *snapCoreModelQr.InquiryStatusQrMpmRequest
		wantError      bool
		expectedStatus string
		setupMock      func(mockHttp *httpMocks.IHTTPRequest)
	}{
		{
			name:           "SUCCESS: inquiry QRIS returns PAID status",
			request:        &snapCoreModelQr.InquiryStatusQrMpmRequest{QrisUUID: "test-uuid-123"},
			wantError:      false,
			expectedStatus: "PAID",
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On(
					"GET",
					mock.Anything,
					constant.StringMockType(),
					mock.AnythingOfType(constant.MockTypeMapStringStringReference),
				).Return([]byte(successResponse), 200, nil)
			},
		},
		{
			name:      "SUCCESS: inquiry QRIS returns 404 not found",
			request:   &snapCoreModelQr.InquiryStatusQrMpmRequest{QrisUUID: "non-existent-uuid"},
			wantError: false,
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On(
					"GET",
					mock.Anything,
					constant.StringMockType(),
					mock.AnythingOfType(constant.MockTypeMapStringStringReference),
				).Return([]byte(notFoundResponse), 404, nil)
			},
		},
		{
			name:      "SUCCESS: inquiry QRIS returns 409 conflict",
			request:   &snapCoreModelQr.InquiryStatusQrMpmRequest{QrisUUID: "conflict-uuid"},
			wantError: false,
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On(
					"GET",
					mock.Anything,
					constant.StringMockType(),
					mock.AnythingOfType(constant.MockTypeMapStringStringReference),
				).Return([]byte(conflictResponse), 409, nil)
			},
		},
		{
			name:      "ERROR: HTTP request error",
			request:   &snapCoreModelQr.InquiryStatusQrMpmRequest{QrisUUID: "test-uuid"},
			wantError: true,
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On(
					"GET",
					mock.Anything,
					constant.StringMockType(),
					mock.AnythingOfType(constant.MockTypeMapStringStringReference),
				).Return([]byte(""), 0, errors.New("network error"))
			},
		},
		{
			name:      "ERROR: invalid JSON response",
			request:   &snapCoreModelQr.InquiryStatusQrMpmRequest{QrisUUID: "test-uuid"},
			wantError: true,
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On(
					"GET",
					mock.Anything,
					constant.StringMockType(),
					mock.AnythingOfType(constant.MockTypeMapStringStringReference),
				).Return([]byte(`{invalid json}`), 200, nil)
			},
		},
		{
			name:      "ERROR: 400 bad request",
			request:   &snapCoreModelQr.InquiryStatusQrMpmRequest{QrisUUID: "test-uuid"},
			wantError: true,
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On(
					"GET",
					mock.Anything,
					constant.StringMockType(),
					mock.AnythingOfType(constant.MockTypeMapStringStringReference),
				).Return([]byte(`{"code":"4005500","message":"Bad Request"}`), 400, nil)
			},
		},
		{
			name:      "ERROR: 500 internal server error",
			request:   &snapCoreModelQr.InquiryStatusQrMpmRequest{QrisUUID: "test-uuid"},
			wantError: true,
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On(
					"GET",
					mock.Anything,
					constant.StringMockType(),
					mock.AnythingOfType(constant.MockTypeMapStringStringReference),
				).Return([]byte(`{"code":"5005500","message":"Internal Server Error"}`), 500, nil)
			},
		},
		{
			name:      "ERROR: 502 bad gateway",
			request:   &snapCoreModelQr.InquiryStatusQrMpmRequest{QrisUUID: "test-uuid"},
			wantError: true,
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On(
					"GET",
					mock.Anything,
					constant.StringMockType(),
					mock.AnythingOfType(constant.MockTypeMapStringStringReference),
				).Return([]byte(`{"code":"5025500","message":"Bad Gateway"}`), 502, nil)
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

			result, err := repo.InquiryStatusQris(context.Background(), tc.request)

			if tc.wantError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				if tc.expectedStatus != "" && result.Data != nil {
					assert.Equal(t, tc.expectedStatus, result.Data.Status)
				}
			}

			mockHttp.AssertExpectations(t)
		})
	}
}
