package bankAccount

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	c "github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/bankAccount"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	s "github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCreate(t *testing.T) {
	bankAccountSvc := serviceMocks.NewIBankAccountService(t)

	handler := New(validator.New(), bankAccountSvc)

	router := chi.NewRouter()
	router.Post("/bank-accounts", handler.Create)

	tests := []struct {
		name           string
		requestBody    string
		setupMock      func()
		wantStatusCode int
		wantRespBody   string
	}{
		{
			name:           "ERROR:Decode request body",
			requestBody:    `A`,
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   c.WrapErrApiRespForTest(40, s.ErrTypeAPI, "invalid character 'A' looking for beginning of value"),
		},
		{
			name:           "ERROR: some error",
			requestBody:    `{"merchantId": "0ee9b6b4-9054-4790-a831-3c480f3b9f0b","beneficiaryBankCode": "001", "beneficiaryBankName": "BCA","beneficiaryAccountNo": "123", "beneficiaryAccountName": "kai", "createdBy": "roger"}`,
			wantStatusCode: http.StatusInternalServerError,
			setupMock: func() {
				bankAccountSvc.On(
					"Create", c.ValueCtxMockType(), mock.AnythingOfType("*bankAccount.CreateBankAccountRequest"),
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
			wantRespBody: c.WrapErrApiRespForTest(99, s.ErrTypeUnknown, "some error"),
		},
		{
			name:           "SUCCESS",
			requestBody:    `{"merchantId": "0ee9b6b4-9054-4790-a831-3c480f3b9f0b","beneficiaryBankCode": "001", "beneficiaryBankName": "BCA","beneficiaryAccountNo": "123", "beneficiaryAccountName": "kai", "createdBy": "roger"}`,
			wantStatusCode: http.StatusOK,
			setupMock: func() {
				bankAccountSvc.On(
					"Create", c.ValueCtxMockType(), mock.AnythingOfType("*bankAccount.CreateBankAccountRequest"),
				).Return(&bankAccount.BankAccountResponse{
					BeneficiaryBankCode:    "001",
					BeneficiaryBankName:    "BCA",
					BeneficiaryAccountNo:   "123",
					BeneficiaryAccountName: "kai",
				}, nil)
			},
			wantRespBody: `{"code": "00","message": "OK","data": {"beneficiaryBankCode":"001","beneficiaryBankName":"BCA","beneficiaryAccountNo":"123","beneficiaryAccountName":"kai"}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/bank-accounts", strings.NewReader(tt.requestBody))

			if tt.setupMock != nil {
				tt.setupMock()
			}

			router.ServeHTTP(rec, req)
			assert.Equal(t, tt.wantStatusCode, rec.Result().StatusCode)
			if !assert.JSONEq(t, tt.wantRespBody, rec.Body.String()) {
				t.Log("Result:", rec.Body.String())
			}
		})
	}
}
