package subMerchant

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	orchestratorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	mockServices "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetSubMerchantBalance(t *testing.T) {
	subMerchantID := uuid.NewString()
	parentMerchantID := uuid.NewString()

	tests := []struct {
		name           string
		subMerchantID  string
		query          string
		userClaims     *userModel.UserTokenClaims
		setupMock      func(merchantSvc *mockServices.IMerchantService, orchestratorSvc *mockServices.IOrchestratorService)
		wantStatusCode int
		wantRespBody   string
	}{
		{
			name:           "ERROR: user not found in context",
			subMerchantID:  subMerchantID,
			userClaims:     nil,
			wantStatusCode: http.StatusUnauthorized,
			wantRespBody:   constant.WrapErrApiRespForTest(41, response.ErrTypeAPI, "user not found"),
		},
		{
			name:          "ERROR: missing sub merchant id",
			subMerchantID: "",
			userClaims: &userModel.UserTokenClaims{
				MerchantId: parentMerchantID,
			},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   constant.WrapErrApiRespForTest(40, response.ErrTypeAPI, "submerchant id is required"),
		},
		{
			name:          "ERROR: validate sub merchant parent failed",
			subMerchantID: subMerchantID,
			userClaims: &userModel.UserTokenClaims{
				MerchantId: parentMerchantID,
			},
			setupMock: func(merchantSvc *mockServices.IMerchantService, _ *mockServices.IOrchestratorService) {
				merchantSvc.On("ValidateSubMerchantParent", mock.Anything, parentMerchantID, subMerchantID).
					Once().Return(constant.ErrSomeErrorForUnitTest)
			},
			wantStatusCode: http.StatusUnauthorized,
			wantRespBody:   constant.WrapErrApiRespForTest(41, response.ErrTypeAPI, "incorrect submerchant"),
		},
		{
			name:          "ERROR: unsupported currency",
			subMerchantID: subMerchantID,
			query:         "currency=USD",
			userClaims: &userModel.UserTokenClaims{
				MerchantId: parentMerchantID,
			},
			setupMock: func(merchantSvc *mockServices.IMerchantService, _ *mockServices.IOrchestratorService) {
				merchantSvc.On("ValidateSubMerchantParent", mock.Anything, parentMerchantID, subMerchantID).
					Once().Return(nil)
			},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   constant.WrapErrApiRespForTest(40, response.ErrTypeAPI, "account not found"),
		},
		{
			name:          "ERROR: GetMerchantBalance returns error",
			subMerchantID: subMerchantID,
			userClaims: &userModel.UserTokenClaims{
				MerchantId: parentMerchantID,
			},
			setupMock: func(merchantSvc *mockServices.IMerchantService, orchestratorSvc *mockServices.IOrchestratorService) {
				merchantSvc.On("ValidateSubMerchantParent", mock.Anything, parentMerchantID, subMerchantID).
					Once().Return(nil)
				orchestratorSvc.On("GetMerchantBalance", mock.Anything, mock.Anything).
					Once().Return(nil, constant.ErrSomeErrorForUnitTest)
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   constant.WrapErrApiRespForTest(99, response.ErrTypeUnknown, "some error"),
		},
		{
			name:          "SUCCESS: default usecase is disbursement",
			subMerchantID: subMerchantID,
			userClaims: &userModel.UserTokenClaims{
				MerchantId: parentMerchantID,
			},
			setupMock: func(merchantSvc *mockServices.IMerchantService, orchestratorSvc *mockServices.IOrchestratorService) {
				merchantSvc.On("ValidateSubMerchantParent", mock.Anything, parentMerchantID, subMerchantID).
					Once().Return(nil)
				orchestratorSvc.On("GetMerchantBalance", mock.Anything, mock.MatchedBy(func(req orchestratorModel.GetMerchantBalanceRequest) bool {
					return req.MerchantID == subMerchantID && req.BalanceName == constant.TypeDisbursement
				})).Once().Return(&orchestratorModel.GetMerchantBalanceResponse{
					AvailableBalance: commonModel.Amount{Currency: constant.CurrencyIDR, Value: "10000.00"},
					PendingBalance:   commonModel.Amount{Currency: constant.CurrencyIDR, Value: "1000.00"},
					TotalBalance:     commonModel.Amount{Currency: constant.CurrencyIDR, Value: "11000.00"},
				}, nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody: `{
				"code": "00",
				"message": "Success",
				"data": {
					"availableBalance": {"currency": "IDR", "value": "10000.00"},
					"pendingBalance": {"currency": "IDR", "value": "1000.00"},
					"totalBalance": {"currency": "IDR", "value": "11000.00"}
				}
			}`,
		},
		{
			name:          "SUCCESS: custom usecase with IDR currency",
			subMerchantID: subMerchantID,
			query:         "currency=IDR&usecase=PAYMENT",
			userClaims: &userModel.UserTokenClaims{
				MerchantId: parentMerchantID,
			},
			setupMock: func(merchantSvc *mockServices.IMerchantService, orchestratorSvc *mockServices.IOrchestratorService) {
				merchantSvc.On("ValidateSubMerchantParent", mock.Anything, parentMerchantID, subMerchantID).
					Once().Return(nil)
				orchestratorSvc.On("GetMerchantBalance", mock.Anything, mock.MatchedBy(func(req orchestratorModel.GetMerchantBalanceRequest) bool {
					return req.MerchantID == subMerchantID && req.BalanceName == constant.TypePayment
				})).Once().Return(&orchestratorModel.GetMerchantBalanceResponse{
					AvailableBalance: commonModel.Amount{Currency: constant.CurrencyIDR, Value: "50000.00"},
					PendingBalance:   commonModel.Amount{Currency: constant.CurrencyIDR, Value: "0.00"},
					TotalBalance:     commonModel.Amount{Currency: constant.CurrencyIDR, Value: "50000.00"},
				}, nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody: `{
				"code": "00",
				"message": "Success",
				"data": {
					"availableBalance": {"currency": "IDR", "value": "50000.00"},
					"pendingBalance": {"currency": "IDR", "value": "0.00"},
					"totalBalance": {"currency": "IDR", "value": "50000.00"}
				}
			}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			merchantSvc := mockServices.NewIMerchantService(t)
			orchestratorSvc := mockServices.NewIOrchestratorService(t)
			controller := &SubMerchantController{
				merchantSvc:     merchantSvc,
				orchestratorSvc: orchestratorSvc,
			}

			if test.setupMock != nil {
				test.setupMock(merchantSvc, orchestratorSvc)
			}

			target := "/sub-merchants/" + test.subMerchantID + "/balance"
			if test.query != "" {
				target += "?" + test.query
			}

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, target, nil)

			// Inject chi URL param directly so we can simulate an empty id.
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", test.subMerchantID)
			ctx := context.WithValue(req.Context(), chi.RouteCtxKey, rctx)
			if test.userClaims != nil {
				ctx = context.WithValue(ctx, constant.CtxUserInfoKey, test.userClaims)
			}
			req = req.WithContext(ctx)

			controller.GetSubMerchantBalance(rec, req)

			assert.Equal(t, test.wantStatusCode, rec.Result().StatusCode)
			if test.wantRespBody != "" {
				assert.JSONEq(t, test.wantRespBody, rec.Body.String())
			}
		})
	}
}
