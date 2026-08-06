package withdrawalController_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	c "github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/withdrawal"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	s "github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	. "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/withdrawal"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetById(t *testing.T) {
	service := serviceMocks.NewIWithdrawalService(t)

	handler := New(validator.New(), nil, service, nil)

	router := chi.NewRouter()
	router.Get("/withdrawals/{account}/{id}", handler.GetById)

	userTokenClaims := &user.UserTokenClaims{
		UUID:       uuid.NewString(),
		MerchantId: uuid.NewString(),
	}

	tests := []struct {
		name            string
		account         string
		userTokenClaims *user.UserTokenClaims
		id              string
		setupMock       func()
		wantStatusCode  int
		wantRespBody    string
	}{
		{
			name:           "ERROR:Invalid path URL",
			account:        "invalid",
			id:             "12345",
			wantStatusCode: http.StatusNotFound,
			wantRespBody:   c.WrapErrApiRespForTest(44, s.ErrTypeAPI, "invalid path URL"),
		},
		{
			name:           "ERROR:Invalid transaction id format",
			id:             "ERROR",
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   c.WrapErrApiRespForTest(40, s.ErrTypeAPI, "invalid transaction id format"),
		},
		{
			name:           "ERROR:User not found",
			id:             uuid.NewString(),
			wantStatusCode: http.StatusUnauthorized,
			wantRespBody:   c.WrapErrApiRespForTest(41, s.ErrTypeAPI, "user not found"),
		},
		{
			name:            "ERROR:Some error",
			id:              uuid.NewString(),
			userTokenClaims: userTokenClaims,
			setupMock: func() {
				service.On(
					"GetById", c.ValueCtxMockType(), mock.AnythingOfType("*withdrawal.WithdrawalDetailRequest"),
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   c.WrapErrApiRespForTest(99, s.ErrTypeUnknown, "some error"),
		},
		{
			name:            "SUCCESS",
			id:              uuid.NewString(),
			userTokenClaims: userTokenClaims,
			setupMock: func() {
				service.On(
					"GetById", c.ValueCtxMockType(), mock.AnythingOfType("*withdrawal.WithdrawalDetailRequest"),
				).Return(&withdrawal.WithdrawalDetailResponse{
					Id:                     "1",
					CreatedBy:              "2",
					Amount:                 3,
					Status:                 "4",
					BankReferenceNo:        "5",
					BeneficiaryBankName:    "6",
					BeneficiaryAccountNo:   "7",
					BeneficiaryAccountName: "8",
					Type:                   "9",
				}, nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code":"00","message":"OK","data":{"id":"1","createdAt":"0001-01-01T00:00:00Z","updatedAt":"0001-01-01T00:00:00Z","createdBy":"2","type":"9","amount":3,"status":"4","bankReferenceNo":"5","beneficiaryBankName":"6","beneficiaryAccountNo":"7","beneficiaryAccountName":"8"}}`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.account == "" {
				test.account = "payments"
			}
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/withdrawals/%s/%s", test.account, test.id), nil)

			if test.setupMock != nil {
				test.setupMock()
			}
			if test.userTokenClaims != nil {
				req = req.WithContext(context.WithValue(req.Context(), c.CtxUserInfoKey, test.userTokenClaims))
			}

			router.ServeHTTP(rec, req)
			assert.Equal(t, test.wantStatusCode, rec.Result().StatusCode)
			if !assert.JSONEq(t, test.wantRespBody, rec.Body.String()) {
				t.Log("Result:", rec.Body.String())
			}
		})
	}
}
