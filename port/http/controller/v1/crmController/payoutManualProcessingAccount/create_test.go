package payoutManualProcessingAccount_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	payoutManualProcessingAccountModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payoutManualProcessingAccount"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	payoutManualProcessingAccountController "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/crmController/payoutManualProcessingAccount"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCRMPayoutManualProcessingAccountController_Create(t *testing.T) {
	merchantID := uuid.New().String()

	tests := []struct {
		name           string
		requestBody    []byte
		modifierMock   func(svc *serviceMocks.IPayoutManualProcessingAccountService)
		wantStatusCode int
		wantRespBody   string
	}{
		{
			name:        "ERROR: Bad Request - Invalid JSON",
			requestBody: []byte("{invalid JSON"),
			modifierMock: func(svc *serviceMocks.IPayoutManualProcessingAccountService) {
				// empty modifier
			},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40", "message":"invalid character 'i' looking for beginning of object key string", "error":{"type":"API_ERROR","message":"invalid character 'i' looking for beginning of object key string","recommendation":""}, "data":null}`,
		},
		{
			name:        "ERROR: Bad Request - Invalid merchant id",
			requestBody: []byte(`{"merchantId":"invalid-uuid","bankCode":"BCA","accountNumber":"123456","updatedBy":"admin"}`),
			modifierMock: func(svc *serviceMocks.IPayoutManualProcessingAccountService) {
				// empty modifier
			},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40", "message":"invalid merchant id", "error":{"type":"API_ERROR","message":"invalid merchant id","recommendation":""}, "data":null}`,
		},
		{
			name:        "ERROR: Bad Request - Failed Validation - Missing bankCode",
			requestBody: []byte(`{"merchantId":"` + merchantID + `","accountNumber":"123456","updatedBy":"admin"}`),
			modifierMock: func(svc *serviceMocks.IPayoutManualProcessingAccountService) {
				// empty modifier
			},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40", "message":"Key: 'CreatePayoutManualProcessingAccountRequest.BankCode' Error:Field validation for 'BankCode' failed on the 'required' tag", "error":{"type":"API_ERROR","message":"Key: 'CreatePayoutManualProcessingAccountRequest.BankCode' Error:Field validation for 'BankCode' failed on the 'required' tag","recommendation":""}, "data":null}`,
		},
		{
			name:        "ERROR: Bad Request - Failed Validation - Missing accountNumber",
			requestBody: []byte(`{"merchantId":"` + merchantID + `","bankCode":"BCA","updatedBy":"admin"}`),
			modifierMock: func(svc *serviceMocks.IPayoutManualProcessingAccountService) {
				// empty modifier
			},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40", "message":"Key: 'CreatePayoutManualProcessingAccountRequest.AccountNumber' Error:Field validation for 'AccountNumber' failed on the 'required' tag", "error":{"type":"API_ERROR","message":"Key: 'CreatePayoutManualProcessingAccountRequest.AccountNumber' Error:Field validation for 'AccountNumber' failed on the 'required' tag","recommendation":""}, "data":null}`,
		},
		{
			name:        "ERROR: Bad Request - Failed Validation - Missing updatedBy",
			requestBody: []byte(`{"merchantId":"` + merchantID + `","bankCode":"BCA","accountNumber":"123456"}`),
			modifierMock: func(svc *serviceMocks.IPayoutManualProcessingAccountService) {
				// empty modifier
			},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40", "message":"Key: 'CreatePayoutManualProcessingAccountRequest.UpdatedBy' Error:Field validation for 'UpdatedBy' failed on the 'required' tag", "error":{"type":"API_ERROR","message":"Key: 'CreatePayoutManualProcessingAccountRequest.UpdatedBy' Error:Field validation for 'UpdatedBy' failed on the 'required' tag","recommendation":""}, "data":null}`,
		},
		{
			name:        "ERROR: Service error",
			requestBody: []byte(`{"merchantId":"` + merchantID + `","bankCode":"BCA","accountNumber":"123456","updatedBy":"admin"}`),
			modifierMock: func(svc *serviceMocks.IPayoutManualProcessingAccountService) {
				svc.On("Create", constant.ValueCtxMockType(), &payoutManualProcessingAccountModel.CreatePayoutManualProcessingAccountRequest{
					MerchantID:    merchantID,
					BankCode:      "BCA",
					AccountNumber: "123456",
					UpdatedBy:     "admin",
				}).Once().Return(nil, constant.ErrSomeErrorForUnitTest)
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   `{"code":"99", "message":"some error", "error":{"type":"UNKNOWN","message":"some error","recommendation":""}, "data":null}`,
		},
		{
			name:        "SUCCESS - Create payout manual processing account",
			requestBody: []byte(`{"merchantId":"` + merchantID + `","bankCode":"BCA","accountNumber":"123456","updatedBy":"admin"}`),
			modifierMock: func(svc *serviceMocks.IPayoutManualProcessingAccountService) {
				svc.On("Create", constant.ValueCtxMockType(), &payoutManualProcessingAccountModel.CreatePayoutManualProcessingAccountRequest{
					MerchantID:    merchantID,
					BankCode:      "BCA",
					AccountNumber: "123456",
					UpdatedBy:     "admin",
				}).Return(&payoutManualProcessingAccountModel.PayoutManualProcessingAccountResponse{
					UUID:          "acc-uuid",
					MerchantID:    merchantID,
					BankCode:      "BCA",
					AccountNumber: "123456",
					Status:        constant.StatusActive,
					UpdatedBy:     "admin",
					UpdatedAt:     time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
				}, nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code":"00", "message":"Success", "data":{"uuid":"acc-uuid","merchantId":"` + merchantID + `","merchantName":"","bankCode":"BCA","accountNumber":"123456","status":"` + constant.StatusActive + `","updatedBy":"admin","updatedAt":"2023-01-01T00:00:00Z"}}`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			svc := serviceMocks.NewIPayoutManualProcessingAccountService(t)
			test.modifierMock(svc)

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/crm/v1/payout-manual-processing-accounts", bytes.NewBuffer(test.requestBody))

			router := chi.NewRouter()
			router.Post("/crm/v1/payout-manual-processing-accounts", payoutManualProcessingAccountController.New(svc, validator.New()).Create)

			router.ServeHTTP(rec, req)

			require.Equal(t, test.wantStatusCode, rec.Result().StatusCode)
			assert.JSONEq(t, test.wantRespBody, rec.Body.String())
		})
	}
}
