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

func TestCRMPayoutManualProcessingAccountController_Update(t *testing.T) {
	accountUUID := uuid.New().String()

	tests := []struct {
		name           string
		uuid           string
		requestBody    []byte
		modifierMock   func(svc *serviceMocks.IPayoutManualProcessingAccountService)
		wantStatusCode int
		wantRespBody   string
	}{
		{
			name:        "ERROR: Bad Request - Invalid uuid",
			uuid:        "invalid-uuid",
			requestBody: []byte(`{"updatedBy":"admin"}`),
			modifierMock: func(svc *serviceMocks.IPayoutManualProcessingAccountService) {
				// empty modifier
			},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40", "message":"invalid id", "error":{"type":"API_ERROR","message":"invalid id","recommendation":""}, "data":null}`,
		},
		{
			name:        "ERROR: Bad Request - Invalid JSON",
			uuid:        accountUUID,
			requestBody: []byte("{invalid JSON"),
			modifierMock: func(svc *serviceMocks.IPayoutManualProcessingAccountService) {
				// empty modifier
			},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40", "message":"invalid character 'i' looking for beginning of object key string", "error":{"type":"API_ERROR","message":"invalid character 'i' looking for beginning of object key string","recommendation":""}, "data":null}`,
		},
		{
			name:        "ERROR: Bad Request - Failed Validation - Missing updatedBy",
			uuid:        accountUUID,
			requestBody: []byte(`{"status":"INACTIVE"}`),
			modifierMock: func(svc *serviceMocks.IPayoutManualProcessingAccountService) {
				// empty modifier
			},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40", "message":"Key: 'UpdatePayoutManualProcessingAccountRequest.UpdatedBy' Error:Field validation for 'UpdatedBy' failed on the 'required' tag", "error":{"type":"API_ERROR","message":"Key: 'UpdatePayoutManualProcessingAccountRequest.UpdatedBy' Error:Field validation for 'UpdatedBy' failed on the 'required' tag","recommendation":""}, "data":null}`,
		},
		{
			name:        "ERROR: Service error",
			uuid:        accountUUID,
			requestBody: []byte(`{"status":"INACTIVE","updatedBy":"admin"}`),
			modifierMock: func(svc *serviceMocks.IPayoutManualProcessingAccountService) {
				status := constant.StatusInactive
				svc.On("Update", constant.ValueCtxMockType(), &payoutManualProcessingAccountModel.UpdatePayoutManualProcessingAccountRequest{
					UUID:      accountUUID,
					Status:    &status,
					UpdatedBy: "admin",
				}).Once().Return(nil, constant.ErrSomeErrorForUnitTest)
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   `{"code":"99", "message":"some error", "error":{"type":"UNKNOWN","message":"some error","recommendation":""}, "data":null}`,
		},
		{
			name:        "SUCCESS - Update payout manual processing account",
			uuid:        accountUUID,
			requestBody: []byte(`{"status":"INACTIVE","updatedBy":"admin"}`),
			modifierMock: func(svc *serviceMocks.IPayoutManualProcessingAccountService) {
				status := constant.StatusInactive
				svc.On("Update", constant.ValueCtxMockType(), &payoutManualProcessingAccountModel.UpdatePayoutManualProcessingAccountRequest{
					UUID:      accountUUID,
					Status:    &status,
					UpdatedBy: "admin",
				}).Return(&payoutManualProcessingAccountModel.PayoutManualProcessingAccountResponse{
					UUID:      accountUUID,
					Status:    constant.StatusInactive,
					UpdatedBy: "admin",
					UpdatedAt: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
				}, nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code":"00", "message":"Success", "data":{"uuid":"` + accountUUID + `","merchantId":"","merchantName":"","bankCode":"","accountNumber":"","status":"` + constant.StatusInactive + `","updatedBy":"admin","updatedAt":"2023-01-01T00:00:00Z"}}`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			svc := serviceMocks.NewIPayoutManualProcessingAccountService(t)
			test.modifierMock(svc)

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPut, "/crm/v1/payout-manual-processing-accounts/"+test.uuid, bytes.NewBuffer(test.requestBody))

			router := chi.NewRouter()
			router.Put("/crm/v1/payout-manual-processing-accounts/{uuid}", payoutManualProcessingAccountController.New(svc, validator.New()).Update)

			router.ServeHTTP(rec, req)

			require.Equal(t, test.wantStatusCode, rec.Result().StatusCode)
			assert.JSONEq(t, test.wantRespBody, rec.Body.String())
		})
	}
}
