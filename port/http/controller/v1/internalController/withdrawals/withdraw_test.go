package internalWithdrawalsController

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/withdrawal"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pivot-backoffice/pkg/validatorExt"
)

func TestWithdraw(t *testing.T) {
	validMerchantClaim := &merchant.MerchantAuthTokenClaims{
		MerchantId: uuid.Max.String(),
	}

	validPayload := withdrawal.OpenAPIWithdrawalRequest{
		ReferenceId:  "REF123",
		WithdrawType: "BANK_TRANSFER",
		IsFullAmount: true,
		Amount: commonModel.Amount{
			Currency: "IDR",
			Value:    "100000.00",
		},
		Description: "Test withdrawal",
	}

	testCases := []struct {
		name           string
		merchantClaim  *merchant.MerchantAuthTokenClaims
		payload        interface{}
		setupMock      func(withdrawalSvc *serviceMocks.IWithdrawalService)
		setHeaders     func(req *http.Request)
		wantStatusCode int
		wantRespBody   string
	}{
		{
			name:    "ERROR: Invalid merchant auth",
			payload: validPayload,
			setupMock: func(withdrawalSvc *serviceMocks.IWithdrawalService) {
				// empty setup mock
			},
			wantStatusCode: http.StatusUnauthorized,
			wantRespBody:   `{"code":"merchant_not_found","error":{"details":[{"field":"","message":"Invalid Merchant request"}],"traceId":"","type":"API_ERROR"},"message":"Merchant not found"}`,
		},
		{
			name:          "ERROR: Invalid JSON payload",
			merchantClaim: validMerchantClaim,
			payload:       `{invalid json}`,
			setupMock: func(withdrawalSvc *serviceMocks.IWithdrawalService) {
				// empty setup mock
			},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"general_error","error":{"details":[{"field":"","message":"Please contact our representative team"}],"traceId":"","type":"API_ERROR"},"message":"General error"}`,
		},
		{
			name:          "ERROR: Validation failed - missing referenceId",
			merchantClaim: validMerchantClaim,
			payload: withdrawal.OpenAPIWithdrawalRequest{
				WithdrawType: "BANK_TRANSFER",
				Amount: commonModel.Amount{
					Currency: "IDR",
					Value:    "100000.00",
				},
			},
			setupMock: func(withdrawalSvc *serviceMocks.IWithdrawalService) {
				// empty setup mock
			},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"api_validation_error","error":{"details":[{"field":"referenceId","message":"Make sure referenceId value is fulfilled"}],"traceId":"","type":"API_ERROR"},"message":"The request was invalid, or an error occurred in downstream provider"}`,
		},
		{
			name:          "ERROR: Bank transfer withdrawal - insufficient balance",
			merchantClaim: validMerchantClaim,
			payload:       validPayload,
			setupMock: func(withdrawalSvc *serviceMocks.IWithdrawalService) {
				withdrawalSvc.On("Create",
					mock.Anything,
					mock.AnythingOfType("*withdrawal.WithdrawalRequest"),
				).Return(nil, pkgErrors.New(response.HttpErrInternal, errors.New("insufficient balance")))
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   `{"code":"general_error","error":{"details":[{"field":"","message":"Please contact our representative team"}],"traceId":"","type":"API_ERROR"},"message":"General error"}`,
		},
		{
			name:          "SUCCESS: Balance transfer withdrawal",
			merchantClaim: validMerchantClaim,
			payload: withdrawal.OpenAPIWithdrawalRequest{
				ReferenceId:  "REF456",
				WithdrawType: "BALANCE_TRANSFER",
				BalanceType:  "PAYOUT_BALANCE",
				IsFullAmount: false,
				Amount: commonModel.Amount{
					Currency: "IDR",
					Value:    "50000",
				},
				Description: "Balance transfer",
			},
			setupMock: func(withdrawalSvc *serviceMocks.IWithdrawalService) {
				withdrawalSvc.On("Create",
					mock.Anything,
					mock.AnythingOfType("*withdrawal.WithdrawalRequest"),
				).Return(&withdrawal.WithdrawalProcessResponse{
					Id:          "wd-789",
					Type:        "AUTOMATED",
					AccountName: "PAYMENT",
					Amount: commonModel.Amount{
						Currency: "IDR",
						Value:    "50000",
					},
					Status: "FAILED",
				}, nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code":"00","message":"OK","data":{"id":"wd-789","merchantId":"ffffffff-ffff-ffff-ffff-ffffffffffff","withdrawal":{"referenceId":"REF456","withdrawType":"BALANCE_TRANSFER","balanceType":"PAYOUT_BALANCE","isFullAmount":false,"amount":{"currency":"IDR","value":"50000"},"description":"Balance transfer"},"status":"FAILED","createdAt":"0001-01-01T00:00:00Z","updatedAt":"0001-01-01T00:00:00Z"}}`,
		},
		{
			name:          "SUCCESS: Bank transfer withdrawal",
			merchantClaim: validMerchantClaim,
			payload: withdrawal.OpenAPIWithdrawalRequest{
				ReferenceId:  "REF456",
				WithdrawType: "BANK_TRANSFER",
				IsFullAmount: true,
				Description:  "Bank transfer",
			},
			setupMock: func(withdrawalSvc *serviceMocks.IWithdrawalService) {
				withdrawalSvc.On("Create",
					mock.Anything, mock.Anything,
				).Return(&withdrawal.WithdrawalProcessResponse{
					Id:          "wd-789",
					Type:        "AUTOMATED",
					AccountName: "PAYMENT",
					Amount: commonModel.Amount{
						Currency: constant.CurrencyIDR, Value: "10000",
					},
					Status:    "SUCCESS",
					CreatedAt: time.Date(2025, time.August, 1, 0, 0, 0, 10, time.UTC),
					UpdatedAt: time.Date(2025, time.August, 1, 0, 0, 0, 20, time.UTC),
				}, nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code":"00","message":"OK","data":{"id":"wd-789","merchantId":"ffffffff-ffff-ffff-ffff-ffffffffffff","withdrawal":{"referenceId":"REF456","withdrawType":"BANK_TRANSFER","balanceType":"","isFullAmount":true,"amount":{"currency":"IDR","value":"10000"},"description":"Bank transfer"},"status":"SUCCESS","createdAt":"2025-08-01T00:00:00Z","updatedAt":"2025-08-01T00:00:00Z"}}`,
		},
		{
			name:          "ERROR: Withdrawal service error",
			merchantClaim: validMerchantClaim,
			payload:       validPayload,
			setupMock: func(withdrawalSvc *serviceMocks.IWithdrawalService) {
				withdrawalSvc.On("Create",
					mock.Anything,
					mock.AnythingOfType("*withdrawal.WithdrawalRequest"),
				).Return(nil, pkgErrors.New(response.HttpErrInternal, errors.New("service error")))
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   `{"code":"general_error","error":{"details":[{"field":"","message":"Please contact our representative team"}],"traceId":"","type":"API_ERROR"},"message":"General error"}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			withdrawalSvc := &serviceMocks.IWithdrawalService{}
			tc.setupMock(withdrawalSvc)

			controller := &InternalWithdrawalController{
				withdrawalSvc: withdrawalSvc,
				validate:      validatorExt.New(),
			}

			var reqBody *bytes.Buffer
			if payloadStr, ok := tc.payload.(string); ok {
				reqBody = bytes.NewBufferString(payloadStr)
			} else {
				payloadBytes, _ := json.Marshal(tc.payload)
				reqBody = bytes.NewBuffer(payloadBytes)
			}

			req := httptest.NewRequest(http.MethodPost, "/", reqBody)
			req.Header.Set("Content-Type", "application/json")
			if tc.setHeaders != nil {
				tc.setHeaders(req)
			}

			if tc.merchantClaim != nil {
				ctx := context.WithValue(req.Context(), constant.CtxMerchantInfo, tc.merchantClaim)
				req = req.WithContext(ctx)
			}

			w := httptest.NewRecorder()
			controller.Withdraw(w, req)

			assert.Equal(t, tc.wantStatusCode, w.Code)
			if tc.wantRespBody != "" {
				require.JSONEqf(t, tc.wantRespBody, w.Body.String(), "wantRespBody: %s, got: %s", tc.wantRespBody, w.Body.String())
			}

			withdrawalSvc.AssertExpectations(t)
		})
	}
}
