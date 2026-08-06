package internalWithdrawalsController

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/bankAccount"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/withdrawal"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
)

func TestGetBankAccountList(t *testing.T) {
	validMerchantClaim := &merchant.MerchantAuthTokenClaims{
		MerchantId: uuid.NewString(),
	}

	testCases := []struct {
		name           string
		merchantClaim  *merchant.MerchantAuthTokenClaims
		setupMock      func(withdrawalSvc *serviceMocks.IWithdrawalService)
		setHeaders     func(req *http.Request)
		wantStatusCode int
		wantRespBody   string
	}{
		{
			name: "ERROR: Invalid merchant auth",
			setupMock: func(withdrawalSvc *serviceMocks.IWithdrawalService) {
				// empty setup mock
			},
			wantStatusCode: http.StatusUnauthorized,
			wantRespBody:   `{"code":"merchant_not_found","error":{"details":[{"field":"","message":"Invalid Merchant request"}],"traceId":"","type":"API_ERROR"},"message":"Merchant not found"}`,
		},
		{
			name:          "SUCCESS: Get bank account list without sub merchant",
			merchantClaim: validMerchantClaim,
			setupMock: func(withdrawalSvc *serviceMocks.IWithdrawalService) {
				withdrawalSvc.On("Preparation", 
					mock.Anything,
					&withdrawal.PreparationRequest{
						MerchantId:  validMerchantClaim.MerchantId,
						AccountName: constant.AccountNamePayment,
					},
				).Return(&withdrawal.PreparationResponse{
					BankAccounts: []bankAccount.BankAccountResponse{
						{
							BeneficiaryBankCode: "014",
							BeneficiaryBankName: "Bank BCA",
						},
					},
				}, nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code":"00","data":{"bankAccounts":[{"beneficiaryBankCode":"014","beneficiaryBankName":"Bank BCA","beneficiaryAccountNo":"","beneficiaryAccountName":""}]},"message":"OK"}`,
		},
		{
			name:          "SUCCESS: Get bank account list with sub merchant",
			merchantClaim: validMerchantClaim,
			setHeaders: func(req *http.Request) {
				req.Header.Set(constant.HeaderXSubMerchantID, "sub-merchant-id")
			},
			setupMock: func(withdrawalSvc *serviceMocks.IWithdrawalService) {
				withdrawalSvc.On("Preparation", 
					mock.Anything,
					&withdrawal.PreparationRequest{
						MerchantId:  "sub-merchant-id",
						AccountName: constant.AccountNamePayment,
					},
				).Return(&withdrawal.PreparationResponse{
					BankAccounts: []bankAccount.BankAccountResponse{
						{
							BeneficiaryBankCode: "008",
							BeneficiaryBankName: "Bank Mandiri",
						},
					},
				}, nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code":"00","data":{"bankAccounts":[{"beneficiaryBankCode":"008","beneficiaryBankName":"Bank Mandiri","beneficiaryAccountNo":"","beneficiaryAccountName":""}]},"message":"OK"}`,
		},
		{
			name:          "ERROR: Withdrawal service error",
			merchantClaim: validMerchantClaim,
			setupMock: func(withdrawalSvc *serviceMocks.IWithdrawalService) {
				withdrawalSvc.On("Preparation", 
					mock.Anything,
					&withdrawal.PreparationRequest{
						MerchantId:  validMerchantClaim.MerchantId,
						AccountName: constant.AccountNamePayment,
					},
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
			}

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tc.setHeaders != nil {
				tc.setHeaders(req)
			}

			if tc.merchantClaim != nil {
				ctx := context.WithValue(req.Context(), constant.CtxMerchantInfo, tc.merchantClaim)
				req = req.WithContext(ctx)
			}

			w := httptest.NewRecorder()
			controller.GetBankAccountList(w, req)

			assert.Equal(t, tc.wantStatusCode, w.Code)
			if tc.wantRespBody != "" {
				require.JSONEq(t, tc.wantRespBody, w.Body.String())
			}

			withdrawalSvc.AssertExpectations(t)
		})
	}
}
