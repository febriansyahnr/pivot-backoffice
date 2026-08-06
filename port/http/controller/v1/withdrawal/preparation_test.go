package withdrawalController_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	c "github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/bankAccount"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/withdrawal"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	s "github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	. "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/withdrawal"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestPreparation(t *testing.T) {
	withdrawalSvc := serviceMocks.NewIWithdrawalService(t)

	handler := New(validator.New(), nil, withdrawalSvc, nil)

	router := chi.NewRouter()
	router.Get("/withdrawals/preparation", handler.Preparation)

	request := &withdrawal.PreparationRequest{
		MerchantId:  "6dae92b4-b6dc-4475-954e-4f8b171fb7ed",
		AccountName: c.TypeWallet,
	}

	userTokenClaims := &user.UserTokenClaims{
		UUID: uuid.NewString(), MerchantId: request.MerchantId,
	}

	tests := []struct {
		name           string
		accountName    string
		userClaim      *user.UserTokenClaims
		setupMock      func()
		wantStatusCode int
		wantRespBody   string
	}{
		{
			name:           "ERROR:User not found",
			wantStatusCode: http.StatusUnauthorized,
			wantRespBody:   c.WrapErrApiRespForTest(41, s.ErrTypeAPI, "user not found"),
		},
		{
			name:           "ERROR:Invalid account name",
			userClaim:      userTokenClaims,
			accountName:    "XXX",
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40","message":"invalid validation","error":{"type":"API_ERROR","details":[{"field":"AccountName","message":"Key: 'PreparationRequest.AccountName' Error:Field validation for 'AccountName' failed on the 'oneof' tag"}],"traceId":""},"data":null}`,
		},
		{
			name:      "ERROR:Some error", // NOSONAR
			userClaim: userTokenClaims,
			setupMock: func() {
				withdrawalSvc.On("Preparation", c.ValueCtxMockType(), request).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   c.WrapErrApiRespForTest(99, s.ErrTypeUnknown, "some error"),
		},
		{
			name:      "SUCCESS", // NOSONAR
			userClaim: userTokenClaims,
			setupMock: func() {
				withdrawalSvc.On("Preparation", c.ValueCtxMockType(), request).Return(&withdrawal.PreparationResponse{
					MerchantId:       request.MerchantId,
					AccountName:      request.AccountName,
					AvailableBalance: 1_250_000.00,
					BankAccounts: []bankAccount.BankAccountResponse{{
						BeneficiaryBankCode:    "002",
						BeneficiaryBankName:    "BANK RAKYAT INDONESIA",
						BeneficiaryAccountNo:   "999966660001",
						BeneficiaryAccountName: "Dummy Simulation",
					}},
				}, nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code":"00","message":"OK","data":{"merchantId":"6dae92b4-b6dc-4475-954e-4f8b171fb7ed","accountName":"WALLET","availableBalance":1250000,"bankAccounts":[{"beneficiaryBankCode":"002","beneficiaryBankName":"BANK RAKYAT INDONESIA","beneficiaryAccountNo":"999966660001","beneficiaryAccountName":"Dummy Simulation"}]}}`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/withdrawals/preparation", nil)

			if test.setupMock != nil {
				test.setupMock()
			}
			if test.accountName == "" {
				test.accountName = c.TypeWallet
			}
			if test.userClaim != nil {
				req = req.WithContext(context.WithValue(req.Context(), c.CtxUserInfoKey, test.userClaim))
			}
			req.URL.RawQuery = fmt.Sprintf("accountName=%s", test.accountName)
			req.Header.Set(c.HeaderXSubMerchantID, "6dae92b4-b6dc-4475-954e-4f8b171fb7ed")

			router.ServeHTTP(rec, req)
			assert.Equal(t, test.wantStatusCode, rec.Result().StatusCode)
			if !assert.JSONEq(t, test.wantRespBody, rec.Body.String()) {
				t.Log("Result:", rec.Body.String())
			}
		})
	}
}
