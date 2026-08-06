package withdrawalController_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	c "github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/withdrawal"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	s "github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	. "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/withdrawal"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreate(t *testing.T) {
	userSvc := serviceMocks.NewIUserService(t)
	withdrawalSvc := serviceMocks.NewIWithdrawalService(t)

	handler := New(validator.New(), nil, withdrawalSvc, userSvc)

	router := chi.NewRouter()
	router.Post("/withdrawals", handler.Create)

	userTokenClaims := &user.UserTokenClaims{
		UUID:       uuid.NewString(),
		MerchantId: "3fc96de8-f65e-4b16-90a1-e2a00d1bae29",
	}
	requestBody := `{"accountName":"PAYMENT","destination":"BANK_TRANSFER","beneficiaryBankCode": "001","beneficiaryAccountNo": "123","amount": 10000}`

	tests := []struct {
		name           string
		userClaim      *user.UserTokenClaims
		requestBody    string
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
			name:           "ERROR:Decode request body",
			userClaim:      userTokenClaims,
			requestBody:    `A`,
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   c.WrapErrApiRespForTest(40, s.ErrTypeAPI, "invalid character 'A' looking for beginning of value"),
		},
		{
			name:           "ERROR:BeneficiaryAccountNo is required field",
			userClaim:      userTokenClaims,
			requestBody:    `{"accountName":"PAYMENT", "destination":"BANK_TRANSFER", "beneficiaryBankCode":"1", "amount":10000}`,
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40","message":"invalid validation","error":{"type":"API_ERROR","details":[{"field":"BeneficiaryAccountNo","message":"Key: 'WithdrawalRequest.BeneficiaryAccountNo' Error:Field validation for 'BeneficiaryAccountNo' failed on the 'required_if' tag"}],"traceId":""},"data":null}`,
		},
		{
			name:        "ERROR:Invalid PIN",
			userClaim:   userTokenClaims,
			requestBody: requestBody,
			setupMock: func() {
				userSvc.On(
					"CheckCurrentPin", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType(),
				).Once().Return(pkgErrs.New(s.HttpErrRequest, c.ErrInvalidPIN))
			},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   c.WrapErrApiRespForTest(40, s.ErrTypeAPI, "invalid pin"),
		},
		{
			name:        "ERROR:Some error",
			userClaim:   userTokenClaims,
			requestBody: requestBody,
			setupMock: func() {
				userSvc.On(
					"CheckCurrentPin", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType(),
				).Return(nil)
				withdrawalSvc.On(
					"Create", c.ValueCtxMockType(), mock.AnythingOfType("*withdrawal.WithdrawalRequest"),
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   c.WrapErrApiRespForTest(99, s.ErrTypeUnknown, "some error"),
		},
		{
			name:        "SUCCESS",
			userClaim:   userTokenClaims,
			requestBody: requestBody,
			setupMock: func() {
				withdrawalSvc.On(
					"Create", c.ValueCtxMockType(), mock.AnythingOfType("*withdrawal.WithdrawalRequest"),
				).Return(&withdrawal.WithdrawalProcessResponse{
					Id:          "497856fe-a03f-4e18-b9db-042085f27e11",
					AccountName: c.TypePayment,
					Type:        c.WithdrawalManual,
					Status:      c.StatusSuccess,
				}, nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code": "00","message": "OK","data": {"id": "497856fe-a03f-4e18-b9db-042085f27e11","accountName":"PAYMENT","type":"MANUAL","status": "SUCCESS"}}`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/withdrawals", strings.NewReader(test.requestBody))

			if test.setupMock != nil {
				test.setupMock()
			}
			if test.userClaim != nil {
				req = req.WithContext(context.WithValue(req.Context(), c.CtxUserInfoKey, test.userClaim))
			}
			req.Header.Set(c.HeaderXRequestPIN, "MTIzNDU2") // Base64 encoded

			router.ServeHTTP(rec, req)
			assert.Equal(t, test.wantStatusCode, rec.Result().StatusCode)
			if !assert.JSONEq(t, test.wantRespBody, rec.Body.String()) {
				t.Log("Result:", rec.Body.String())
			}
		})
	}
}
