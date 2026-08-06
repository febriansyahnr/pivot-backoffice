package internalFeeController

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	feeModel "github.com/paper-indonesia/pivot-backoffice/internal/model/fee"
	mocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCalculateWalletTransactionFee(t *testing.T) {
	// Create a new instance of the controller
	resp := &feeModel.TrxFeeOnBehalfMetadata{}
	payload := &feeModel.GetTransactionFeeRequest{}
	// Define test cases
	tests := []struct {
		name              string
		merchantID        string
		requesterMerchant string
		setup             func(svc *mocks.IFeeService)
		setupBody         func() []byte
		expectedResponse  string
		expectedStatus    int
	}{
		{
			name:              "SUCCESS: Get merchant bank account",
			merchantID:        "merchant123",
			requesterMerchant: "requester123",
			setup: func(svc *mocks.IFeeService) {
				svc.On("GetTransactionFeeOnBehalf", mock.Anything, mock.Anything).Return(
					resp,
					nil,
				)
			},
			setupBody: func() []byte {
				b, _ := json.Marshal(payload)
				return b
			},
			expectedResponse: `{"code":"00","message":"OK","data":{"type":"","amountType":"","amount":0,"percentage":0,"finalAmount":0}}`,
			expectedStatus:   http.StatusOK,
		},
		{
			name:              "ERROR: Bad request",
			merchantID:        "merchant123",
			requesterMerchant: "requester123",
			setup: func(svc *mocks.IFeeService) {
			},
			setupBody: func() []byte {
				return nil
			},
			expectedResponse: `{"code":"40","message":"invalid request payload","error":{"type":"API_ERROR","details":[],"traceId":""},"data":null}`,
			expectedStatus:   http.StatusBadRequest,
		},
		{
			name:              "Error",
			merchantID:        "merchant123",
			requesterMerchant: "requester123",
			setup: func(svc *mocks.IFeeService) {
				svc.On("GetTransactionFeeOnBehalf", mock.Anything, mock.Anything).Return(
					nil,
					errors.New("error"),
				)
			},
			setupBody: func() []byte {
				b, _ := json.Marshal(payload)
				return b
			},
			expectedResponse: `{"code":"99","message":"error","error":{"type":"UNKNOWN","details":[],"traceId":""},"data":null}`,
			expectedStatus:   http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			feeSvc := mocks.NewIFeeService(t)
			tt.setup(feeSvc)

			// Create a new HTTP request
			req := httptest.NewRequest("POST", "/"+tt.merchantID, bytes.NewReader(tt.setupBody()))
			req.Header.Set(constant.HeaderXMerchantId, tt.requesterMerchant)
			w := httptest.NewRecorder()

			// Call the handler
			controller := New(feeSvc)
			controller.CalculateWalletTransactionFee(w, req)

			// Check the response
			res := w.Result()
			assert.Equal(t, tt.expectedStatus, res.StatusCode)
			assert.JSONEqf(t, tt.expectedResponse, w.Body.String(), fmt.Sprintf("expected: %v ,actual: %v", tt.expectedResponse, w.Body.String()))

			// Additional assertions can be added here for response body, etc.
		})
	}
}
