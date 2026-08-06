package merchantRcn

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/vccSettlement"
	mockService "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestMerchantRcnController_InquiryTransactions(t *testing.T) {
	validRequest := vccSettlement.VccTransactionInquiryRequest{
		RcnId:        "rcn-123",
		RecordType:   "ST",
		BillingCycle: "02",
		PostingDate:  "20250201",
	}

	invalidRequest := vccSettlement.VccTransactionInquiryRequest{
		RcnId:        "rcn-123",
		RecordType:   "ST",
		BillingCycle: "991",
		PostingDate:  "202502011",
	}

	expectedResponse := &vccSettlement.VccTransactionInquiryResponse{
		PartnerReferenceNo: "partner-ref-123",
	}

	testCases := []struct {
		name           string
		requestBody    interface{}
		merchantID     string
		mockSetup      func(vccSettlementSvc *mockService.IVccSettlementService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:        "Success",
			requestBody: validRequest,
			merchantID:  "merchant-456",
			mockSetup: func(vccSettlementSvc *mockService.IVccSettlementService) {
				vccSettlementSvc.On("RcnTransactionInquiry",
					mock.Anything,
					mock.AnythingOfType("*vccSettlement.VccTransactionInquiryRequest"),
				).Return(expectedResponse, nil).Once()
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"code":"00","message":"OK","data":{"partnerReferenceNo":"partner-ref-123"}}`,
		},
		{
			name:           "Failure - Invalid JSON body",
			requestBody:    "invalid-json",
			merchantID:     "merchant-456",
			mockSetup:      func(vccSettlementSvc *mockService.IVccSettlementService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"code":"field_format_invalid","message":"Format Field is invalid","error":{"type":"API_ERROR","details":[{"field":"","message":"invalid request payload"}],"traceId":""}}`,
		},
		{
			name:           "Failure - Missing merchant ID header",
			requestBody:    validRequest,
			merchantID:     "",
			mockSetup:      func(vccSettlementSvc *mockService.IVccSettlementService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"code":"field_format_invalid","message":"Format Field is invalid","error":{"type":"API_ERROR","details":[{"field":"","message":"incorrect merchant id"}],"traceId":""}}`,
		},
		{
			name:           "Failure - Validation error",
			requestBody:    invalidRequest,
			merchantID:     "merchant-456",
			mockSetup:      func(vccSettlementSvc *mockService.IVccSettlementService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"code":"field_format_invalid","message":"Format Field is invalid","error":{"type":"API_ERROR","details":[{"field":"","message":"Key: 'VccTransactionInquiryRequest.BillingCycle' Error:Field validation for 'BillingCycle' failed on the 'len' tag\nKey: 'VccTransactionInquiryRequest.PostingDate' Error:Field validation for 'PostingDate' failed on the 'datetime' tag"}],"traceId":""}}`,
		},
		{
			name:        "Failure - Service error",
			requestBody: validRequest,
			merchantID:  "merchant-456",
			mockSetup: func(vccSettlementSvc *mockService.IVccSettlementService) {
				vccSettlementSvc.On("RcnTransactionInquiry",
					mock.Anything,
					mock.AnythingOfType("*vccSettlement.VccTransactionInquiryRequest"),
				).Return(nil, errors.New("service error")).Once()
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"code":"general_error","message":"General error","error":{"type":"API_ERROR","details":[{"field":"","message":"Please contact our representative team"}],"traceId":""}}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup mocks
			mockMerchantSvc := mockService.NewIMerchantRcnService(t)
			mockVccSettlementSvc := mockService.NewIVccSettlementService(t)
			mockValidator := validator.New()

			tc.mockSetup(mockVccSettlementSvc)

			// Create controller
			mc := New(mockMerchantSvc, mockValidator, WithVccSettlementService(mockVccSettlementSvc))

			// Prepare request body
			var requestBody []byte
			if str, ok := tc.requestBody.(string); ok {
				requestBody = []byte(str)
			} else {
				requestBody, _ = json.Marshal(tc.requestBody)
			}

			// Create HTTP request
			req := httptest.NewRequest(http.MethodPost, "/merchant/rcns/transactions/inquiry", bytes.NewBuffer(requestBody))
			req = req.WithContext(context.Background())
			if tc.merchantID != "" {
				req.Header.Set(constant.HeaderXMerchantId, tc.merchantID)
			}

			// Create response recorder
			rr := httptest.NewRecorder()

			// Execute handler
			handler := http.HandlerFunc(mc.InquiryTransactions)
			handler.ServeHTTP(rr, req)

			if rr.Body.String() != "" {
				t.Logf("Handler response body: %s", rr.Body.String())
			}

			// Assertions
			assert.Equal(t, tc.expectedStatus, rr.Code)
			assert.JSONEqf(t, tc.expectedBody, rr.Body.String(), "expected: %s | actual: %s", tc.expectedBody, rr.Body.String())
			mockVccSettlementSvc.AssertExpectations(t)
		})
	}
}
