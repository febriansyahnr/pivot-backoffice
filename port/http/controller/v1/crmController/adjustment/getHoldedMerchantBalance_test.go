package adjustment_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	adjustmentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/adjustment"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	. "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/crmController/adjustment"
)

func TestGetHoldedMerchantBalance(t *testing.T) {
	adjustSvcMock := serviceMocks.NewIAdjustmentService(t)

	tests := []struct {
		name           string
		setupRequest   func() *http.Request
		modifierMock   func()
		wantStatusCode int
		wantRespBody   string
	}{
		{
			name: "ERROR: Missing required query params",
			setupRequest: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "/crm/v1/balances/hold", nil)
			},
			modifierMock:   func() {},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40","errors":{"MerchantId":"Key: 'GetHoldedMerchantBalanceRequest.MerchantId' Error:Field validation for 'MerchantId' failed on the 'required' tag","AccountType":"Key: 'GetHoldedMerchantBalanceRequest.AccountType' Error:Field validation for 'AccountType' failed on the 'required' tag"}}`,
		},
		{
			name: "ERROR: Invalid AccountType value",
			setupRequest: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "/crm/v1/balances/hold?merchantId="+uuid.NewString()+"&accountType=INVALID", nil)
			},
			modifierMock:   func() {},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40","errors":{"AccountType":"Key: 'GetHoldedMerchantBalanceRequest.AccountType' Error:Field validation for 'AccountType' failed on the 'oneof' tag"}}`,
		},
		{
			name: "ERROR: Service returns error",
			setupRequest: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "/crm/v1/balances/hold?merchantId="+uuid.NewString()+"&accountType=PAYMENT", nil)
			},
			modifierMock: func() {
				adjustSvcMock.On(
					"GetHoldedMerchantBalance", mock.AnythingOfType(constant.MockTypeValueContextReference), mock.AnythingOfType("*adjustment.GetHoldedMerchantBalanceRequest"),
				).Once().Return(nil, constant.ErrSomeErrorForUnitTest)
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   `{"code":"99","errors":"some error"}`,
		},
		{
			name: "SUCCESS: Get holded merchant balance",
			setupRequest: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "/crm/v1/balances/hold?merchantId="+uuid.NewString()+"&accountType=PAYMENT", nil)
			},
			modifierMock: func() {
				adjustSvcMock.On(
					"GetHoldedMerchantBalance", mock.AnythingOfType(constant.MockTypeValueContextReference), mock.AnythingOfType("*adjustment.GetHoldedMerchantBalanceRequest"),
				).Once().Return(&adjustmentModel.GetHoldedMerchantBalanceResponse{
					Amount:      75000,
					MerchantID:  "merchant-123",
					AccountType: "PAYMENT",
				}, nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody: `{
				"code": "00",
				"data": {
					"amount": 75000,
					"merchantId": "merchant-123",
					"accountType": "PAYMENT"
				}
			}`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.modifierMock()

			rec := httptest.NewRecorder()
			req := test.setupRequest()

			router := chi.NewRouter()
			router.Get("/crm/v1/balances/hold", New(adjustSvcMock).GetHoldedMerchantBalance)

			router.ServeHTTP(rec, req)

			require.Equal(t, test.wantStatusCode, rec.Result().StatusCode)
			assert.JSONEq(t, test.wantRespBody, rec.Body.String())
		})
	}
}
