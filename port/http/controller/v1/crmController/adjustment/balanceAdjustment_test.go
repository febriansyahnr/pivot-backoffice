package adjustment_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	adjustModel "github.com/paper-indonesia/pivot-backoffice/internal/model/adjustment"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	. "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/crmController/adjustment"
)

func TestCreateBalanceAdjustment(t *testing.T) {
	adjustSvcMock := serviceMocks.NewIAdjustmentService(t)

	router := chi.NewRouter()
	router.Post("/balance/adjustment", New(adjustSvcMock).CreateBalanceAdjustment)

	validRequest := adjustModel.MerchantBalanceAdjustmentRequest{
		MerchantId:  "123e4567-e89b-12d3-a456-426614174000",
		ReferenceId: "REF-123",
		BalanceType: constant.AdjustmentPayoutBalanceDestination,
		Currency:    "IDR",
		Credit:      100000,
		Debit:       0,
		CreatedBy:   "admin@example.com",
		Remarks:     "Balance adjustment test",
	}

	validResponse := &adjustModel.ManualAdjustmentHistory{
		UUID:        "adjustment-uuid-123",
		MerchantID:  validRequest.MerchantId,
		ReferenceID: validRequest.ReferenceId,
		Currency:    validRequest.Currency,
		Amount:      validRequest.Credit,
		CreatedBy:   validRequest.CreatedBy,
		Notes:       validRequest.Remarks,
	}

	tests := []struct {
		name           string
		requestBody    interface{}
		setupMock      func()
		wantStatusCode int
		wantResponse   string
	}{
		{
			name:        "ERROR: Invalid JSON payload",
			requestBody: `{invalid json}`,
			setupMock: func() {
				// No service call expected
			},
			wantStatusCode: http.StatusBadRequest,
			wantResponse:   `{"code":"40","errors":"invalid character 'i' looking for beginning of object key string"}`,
		},
		{
			name: "ERROR: Validation failed - missing merchantId",
			requestBody: adjustModel.MerchantBalanceAdjustmentRequest{
				ReferenceId: "REF-123",
				BalanceType: constant.AdjustmentPayoutBalanceDestination,
				Currency:    "IDR",
				Credit:      100000,
				CreatedBy:   "admin@example.com",
				Remarks:     "Test remarks",
			},
			setupMock: func() {
				// No service call expected due to validation failure
			},
			wantStatusCode: http.StatusBadRequest,
			wantResponse:   `{"code":"40","errors":{"MerchantId":"Key: 'MerchantBalanceAdjustmentRequest.MerchantId' Error:Field validation for 'MerchantId' failed on the 'required' tag"}}`,
		},
		{
			name: "ERROR: Validation failed - invalid UUID format",
			requestBody: adjustModel.MerchantBalanceAdjustmentRequest{
				MerchantId:  "invalid-uuid",
				ReferenceId: "REF-123",
				BalanceType: constant.AdjustmentPayoutBalanceDestination,
				Currency:    "IDR",
				Credit:      100000,
				CreatedBy:   "admin@example.com",
				Remarks:     "Test remarks",
			},
			setupMock: func() {
				// No service call expected due to validation failure
			},
			wantStatusCode: http.StatusBadRequest,
			wantResponse:   `{"code":"40","errors":{"MerchantId":"Key: 'MerchantBalanceAdjustmentRequest.MerchantId' Error:Field validation for 'MerchantId' failed on the 'uuid' tag"}}`,
		},
		{
			name: "ERROR: Service returns invalid balance type error",
			requestBody: adjustModel.MerchantBalanceAdjustmentRequest{
				MerchantId:  "123e4567-e89b-12d3-a456-426614174000",
				ReferenceId: "REF-123",
				BalanceType: "INVALID_BALANCE",
				Currency:    "IDR",
				Credit:      100000,
				CreatedBy:   "admin@example.com",
				Remarks:     "Test remarks",
			},
			setupMock: func() {
				adjustSvcMock.On(
					"CreateMerchantBalanceAdjustment",
					mock.Anything,
					mock.AnythingOfType("*adjustment.MerchantBalanceAdjustmentRequest"),
				).Once().Return(nil, pkgErrs.New(response.HttpErrRequest, constant.ErrAdjustmentInvalidBalanceType))
			},
			wantStatusCode: http.StatusBadRequest,
			wantResponse:   `{"code":"40","errors":"invalid balance type"}`,
		},
		{
			name: "ERROR: Validation failed - invalid currency",
			requestBody: adjustModel.MerchantBalanceAdjustmentRequest{
				MerchantId:  "123e4567-e89b-12d3-a456-426614174000",
				ReferenceId: "REF-123",
				BalanceType: constant.AdjustmentPayoutBalanceDestination,
				Currency:    "USD",
				Credit:      100000,
				CreatedBy:   "admin@example.com",
				Remarks:     "Test remarks",
			},
			setupMock: func() {
				// No service call expected due to validation failure
			},
			wantStatusCode: http.StatusBadRequest,
			wantResponse:   `{"code":"40","errors":{"Currency":"Key: 'MerchantBalanceAdjustmentRequest.Currency' Error:Field validation for 'Currency' failed on the 'oneof' tag"}}`,
		},
		{
			name: "ERROR: Validation failed - both credit and debit are zero",
			requestBody: adjustModel.MerchantBalanceAdjustmentRequest{
				MerchantId:  "123e4567-e89b-12d3-a456-426614174000",
				ReferenceId: "REF-123",
				BalanceType: constant.AdjustmentPayoutBalanceDestination,
				Currency:    "IDR",
				Credit:      0,
				Debit:       0,
				CreatedBy:   "admin@example.com",
				Remarks:     "Test remarks",
			},
			setupMock: func() {
				// No service call expected due to validation failure
			},
			wantStatusCode: http.StatusBadRequest,
			wantResponse:   `{"code":"40","errors":{"Credit":"Key: 'MerchantBalanceAdjustmentRequest.Credit' Error:Field validation for 'Credit' failed on the 'required_if' tag","Debit":"Key: 'MerchantBalanceAdjustmentRequest.Debit' Error:Field validation for 'Debit' failed on the 'required_if' tag"}}`,
		},
		{
			name:        "ERROR: Service returns merchant not found",
			requestBody: validRequest,
			setupMock: func() {
				adjustSvcMock.On(
					"CreateMerchantBalanceAdjustment",
					mock.Anything,
					mock.AnythingOfType("*adjustment.MerchantBalanceAdjustmentRequest"),
				).Once().Return(nil, pkgErrs.New(response.HttpErrUnprocessableContent, constant.ErrMerchantNotFound))
			},
			wantStatusCode: http.StatusUnprocessableEntity,
			wantResponse:   `{"code":"45","errors":"merchant not found"}`,
		},
		{
			name:        "ERROR: Service returns database error",
			requestBody: validRequest,
			setupMock: func() {
				adjustSvcMock.On(
					"CreateMerchantBalanceAdjustment",
					mock.Anything,
					mock.AnythingOfType("*adjustment.MerchantBalanceAdjustmentRequest"),
				).Once().Return(nil, pkgErrs.New(response.HttpErrDatabase, errors.New("database connection failed")))
			},
			wantStatusCode: http.StatusInternalServerError,
			wantResponse:   `{"code":"98","errors":"database connection failed"}`,
		},
		{
			name:        "ERROR: Service returns invalid amount error",
			requestBody: validRequest,
			setupMock: func() {
				adjustSvcMock.On(
					"CreateMerchantBalanceAdjustment",
					mock.Anything,
					mock.AnythingOfType("*adjustment.MerchantBalanceAdjustmentRequest"),
				).Once().Return(nil, pkgErrs.New(response.HttpErrRequest, constant.ErrInvalidAmount))
			},
			wantStatusCode: http.StatusBadRequest,
			wantResponse:   `{"code":"40","errors":"invalid amount"}`,
		},
		{
			name:        "SUCCESS: Credit adjustment",
			requestBody: validRequest,
			setupMock: func() {
				adjustSvcMock.On(
					"CreateMerchantBalanceAdjustment",
					mock.Anything,
					mock.AnythingOfType("*adjustment.MerchantBalanceAdjustmentRequest"),
				).Once().Return(validResponse, nil)
			},
			wantStatusCode: http.StatusOK,
			wantResponse:   `{"code":"00","data":{"uuid":"adjustment-uuid-123","merchantId":"123e4567-e89b-12d3-a456-426614174000","type":"","action":"","currency":"IDR","amount":100000,"referenceId":"REF-123","status":"SUCCESS","notes":"Balance adjustment test","createdBy":"admin@example.com","createdAt":"0001-01-01T00:00:00Z","updatedAt":"0001-01-01T00:00:00Z"}}`,
		},
		{
			name: "SUCCESS: Debit adjustment",
			requestBody: adjustModel.MerchantBalanceAdjustmentRequest{
				MerchantId:  "123e4567-e89b-12d3-a456-426614174000",
				ReferenceId: "REF-DEBIT-456",
				BalanceType: constant.AdjustmentPaymentBalanceDestination,
				Currency:    "IDR",
				Credit:      0,
				Debit:       50000,
				CreatedBy:   "admin@example.com",
				Remarks:     "Debit adjustment test",
			},
			setupMock: func() {
				debitResponse := &adjustModel.ManualAdjustmentHistory{
					UUID:        "debit-adjustment-uuid-456",
					MerchantID:  "123e4567-e89b-12d3-a456-426614174000",
					ReferenceID: "REF-DEBIT-456",
					Currency:    "IDR",
					Amount:      -50000, // Negative for debit
					CreatedBy:   "admin@example.com",
					Notes:       "Debit adjustment test",
				}
				adjustSvcMock.On(
					"CreateMerchantBalanceAdjustment",
					mock.Anything,
					mock.AnythingOfType("*adjustment.MerchantBalanceAdjustmentRequest"),
				).Once().Return(debitResponse, nil)
			},
			wantStatusCode: http.StatusOK,
			wantResponse:   `{"code":"00","data":{"uuid":"debit-adjustment-uuid-456","merchantId":"123e4567-e89b-12d3-a456-426614174000","type":"","action":"","currency":"IDR","amount":-50000,"referenceId":"REF-DEBIT-456","status":"SUCCESS","notes":"Debit adjustment test","createdBy":"admin@example.com","createdAt":"0001-01-01T00:00:00Z","updatedAt":"0001-01-01T00:00:00Z"}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			var reqBody *bytes.Buffer
			if str, ok := tt.requestBody.(string); ok {
				// Invalid JSON string
				reqBody = bytes.NewBufferString(str)
			} else {
				// Valid struct - marshal to JSON
				jsonData, err := json.Marshal(tt.requestBody)
				require.NoError(t, err)
				reqBody = bytes.NewBuffer(jsonData)
			}

			req := httptest.NewRequest(http.MethodPost, "/balance/adjustment", reqBody)
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatusCode, w.Code)
			
			if tt.wantResponse != "" {
				require.JSONEq(t, tt.wantResponse, w.Body.String())
			}

			// Verify all mock expectations were met
			adjustSvcMock.AssertExpectations(t)
		})
	}
}