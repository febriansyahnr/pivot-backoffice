package internalBankAccountController

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/bankAccount"
	mocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetMerchantBankAccount(t *testing.T) {
	// Create a new instance of the controller
	resp := &bankAccount.BankAccount{}

	// Define test cases
	tests := []struct {
		name              string
		merchantID        string
		requesterMerchant string
		setup             func(svc *mocks.IBankAccountService)
		expectedResponse  string
		expectedStatus    int
	}{
		{
			name:              "SUCCESS: Get merchant bank account",
			merchantID:        "merchant123",
			requesterMerchant: "requester123",
			setup: func(svc *mocks.IBankAccountService) {
				svc.On("GetByMerchantID", mock.Anything, mock.Anything).Return(
					resp,
					nil,
				)
			},
			expectedResponse: `{"code":"00","message":"OK","data":{"beneficiaryBankCode":"","beneficiaryBankName":"","beneficiaryAccountNo":"","beneficiaryAccountName":""}}`,
			expectedStatus:   http.StatusOK,
		},
		{
			name:              "Error",
			merchantID:        "merchant123",
			requesterMerchant: "requester123",
			setup: func(svc *mocks.IBankAccountService) {
				svc.On("GetByMerchantID", mock.Anything, mock.Anything).Return(
					nil,
					errors.New("error"),
				)
			},
			expectedResponse: `{"code":"99","message":"error","error":{"type":"UNKNOWN","details":[],"traceId":""},"data":null}`,
			expectedStatus:   http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bankAccountSvc := mocks.NewIBankAccountService(t)
			tt.setup(bankAccountSvc)

			// Create a new HTTP request
			req := httptest.NewRequest("GET", "/merchant/"+tt.merchantID, nil)
			req.Header.Set(constant.HeaderXMerchantId, tt.requesterMerchant)
			w := httptest.NewRecorder()

			// Call the handler
			controller := New(bankAccountSvc)
			controller.GetMerchantBankAccount(w, req)

			// Check the response
			res := w.Result()
			assert.Equal(t, tt.expectedStatus, res.StatusCode)
			assert.JSONEqf(t, tt.expectedResponse, w.Body.String(), fmt.Sprintf("expected: %v ,actual: %v", tt.expectedResponse, w.Body.String()))

			// Additional assertions can be added here for response body, etc.
		})
	}
}
