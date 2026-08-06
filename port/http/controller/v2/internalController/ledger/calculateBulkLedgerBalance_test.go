package ledgerController

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	account_model "github.com/paper-indonesia/pivot-backoffice/internal/model/account"
	ledger_model "github.com/paper-indonesia/pivot-backoffice/internal/model/ledger"
	mockSvc "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCalculateBulkLedgerBalance(t *testing.T) {
	testCases := []struct {
		name             string
		requestBody      interface{}
		merchantID       string
		expectedCode     int
		expectedResponse string
		setup            func(mockService *mockSvc.ILedgerService)
	}{
		{
			name: "SUCCESS: Calculate bulk balance",
			requestBody: account_model.CalculateBulkLedgerBalanceRequest{
				AccountIDs: []string{uuid.New().String(), uuid.New().String()},
			},
			merchantID:       uuid.New().String(),
			expectedCode:     http.StatusOK,
			expectedResponse: `{"code":"00","message":"OK","data":{"Balance":320,"Currency":"IDR"}}`,
			setup: func(mockService *mockSvc.ILedgerService) {
				mockService.On("CalculateBulkLedgerBalance", mock.Anything, mock.Anything).Return(&ledger_model.LedgerBalance{
					Balance:  320,
					Currency: constant.CurrencyIDR,
				}, nil)
			},
		},
		{
			name: "SUCCESS: Empty balance",
			requestBody: account_model.CalculateBulkLedgerBalanceRequest{
				AccountIDs: []string{},
			},
			merchantID:       uuid.New().String(),
			expectedCode:     http.StatusOK,
			expectedResponse: `{"code":"00","message":"OK","data":{"Balance":0,"Currency":""}}`,
			setup: func(mockService *mockSvc.ILedgerService) {
				mockService.On("CalculateBulkLedgerBalance", mock.Anything, mock.Anything).Return(&ledger_model.LedgerBalance{
					Balance:  0,
					Currency: "",
				}, nil)
			},
		},
		{
			name:             "ERROR: Invalid JSON body",
			requestBody:      "invalid json",
			merchantID:       uuid.New().String(),
			expectedCode:     http.StatusBadRequest,
			expectedResponse: `{"code":"40","message":"json: cannot unmarshal string into Go value of type account_model.CalculateBulkLedgerBalanceRequest","error":{"type":"API_ERROR","message":"json: cannot unmarshal string into Go value of type account_model.CalculateBulkLedgerBalanceRequest","recommendation":""},"data":null}`,
			setup: func(mockService *mockSvc.ILedgerService) {
			},
		},
		{
			name: "ERROR: Service error",
			requestBody: account_model.CalculateBulkLedgerBalanceRequest{
				AccountIDs: []string{uuid.New().String()},
			},
			merchantID:       uuid.New().String(),
			expectedCode:     http.StatusInternalServerError,
			expectedResponse: `{"code":"99","message":"service error","error":{"type":"UNKNOWN","message":"service error","recommendation":""},"data":null}`,
			setup: func(mockService *mockSvc.ILedgerService) {
				mockService.On("CalculateBulkLedgerBalance", mock.Anything, mock.Anything).Return(nil, errors.New("service error"))
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockLedgerService := mockSvc.NewILedgerService(t)
			tc.setup(mockLedgerService)

			controller := &LedgerController{
				ledgerSvc: mockLedgerService,
			}

			body, _ := json.Marshal(tc.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/ledger/bulk-balance", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			if tc.merchantID != "" {
				req.Header.Set(constant.HeaderXMerchantId, tc.merchantID)
			}
			req = req.WithContext(context.Background())

			rr := httptest.NewRecorder()
			controller.CalculateBulkLedgerBalance(rr, req)

			assert.Equal(t, tc.expectedCode, rr.Code)
			assert.JSONEq(t, tc.expectedResponse, rr.Body.String())
		})
	}
}
