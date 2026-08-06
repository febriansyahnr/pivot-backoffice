package crmPaymentMethodController

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetupConfig(t *testing.T) {
	svc := serviceMocks.NewIPaymentMethodService(t)

	validMerchantID := "57b31991-838d-4505-b04c-6d5bf14328c2"
	validPaymentMethodID := "3fe9a257-6575-4532-97b6-60ecc263e89e"

	tests := []struct {
		name            string
		merchantID      string
		paymentMethodID string
		requestBody     []byte
		modifierMock    func()
		wantStatusCode  int
		wantRespBody    string
	}{
		{
			name:            "ERROR: Invalid merchantID format",
			merchantID:      "invalid",
			paymentMethodID: "",
			modifierMock: func() {
				// empty modifier
			},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40","errors":"id is required"}`,
		},
		{
			name:            "ERROR: Invalid paymentMethodID format",
			merchantID:      validMerchantID,
			paymentMethodID: "invalid",
			modifierMock: func() {
				// empty modifier
			},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40","errors":"paymentMethodId is required"}`,
		},
		{
			name:            "ERROR: Bad Request - Invalid JSON",
			merchantID:      validMerchantID,
			paymentMethodID: validPaymentMethodID,
			requestBody:     []byte("{invalid JSON"),
			modifierMock: func() {
				// empty modifier
			},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40", "errors":"invalid character 'i' looking for beginning of object key string"}`,
		},
		{
			name:            "ERROR: Bad Request - Failed Validation",
			merchantID:      validMerchantID,
			paymentMethodID: validPaymentMethodID,
			requestBody:     []byte(`{"client_transaction_id": "12345abcde"}`),
			modifierMock: func() {
				// empty modifier
			},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40", "errors":{"ChannelType":"Key: 'SetupPaymentMethodConfigRequest.ChannelType' Error:Field validation for 'ChannelType' failed on the 'oneof' tag"}}`,
		},
		{
			name:            "ERROR: Service error",
			merchantID:      validMerchantID,
			paymentMethodID: validPaymentMethodID,
			requestBody:     []byte(`{"channelType":"DIRECT","channelConfig":{},"partnerConfig":{"virtualAccount":[{"binPrefix":"1402","type":"OPEN_STATIC","integrationMethod":"SERVER"}]}}`),
			modifierMock: func() {
				svc.On("SetupConfig", constant.ValueCtxMockType(), constant.PtrSetupPaymentMethodConfigRequest()).
					Once().Return(constant.ErrSomeErrorForUnitTest)
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   `{"code":"99", "errors":"some error"}`,
		},
		{
			name:            "SUCCESS",
			merchantID:      validMerchantID,
			paymentMethodID: validPaymentMethodID,
			requestBody:     []byte(`{"channelType":"DIRECT","channelConfig":{},"partnerConfig":{"virtualAccount":[{"binPrefix":"1402","type":"OPEN_STATIC","integrationMethod":"SERVER"}]}}`),
			modifierMock: func() {
				svc.On("SetupConfig", constant.ValueCtxMockType(), constant.PtrSetupPaymentMethodConfigRequest()).
					Return(nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code":"00", "data":{"updated":true}}`,
		},
		{
			name:            "SUCCESS:Payment method CREDIT_CARD with partner config",
			merchantID:      validMerchantID,
			paymentMethodID: validPaymentMethodID,
			requestBody:     []byte(`{"channelType":"AGGREGATOR","channelConfig":{},"partnerConfig":{"card":[{"acquirer":"BRI","channelType":"AGGREGATOR","isActive":true,"priority":1,"cardTypes":["CREDIT","DEBIT"],"merchantIdTag":"TEST_TAG","partnerBaseURL":"https://","prioritizedBIN":["444000"],"partnerProcessor":"MPGS","supportedUseCase":{"allowBypass3ds":false,"allowForeignCard":true,"allowRecurringPayment":true,"allowedCountryRiskLevel3ds":["LOW"],"allowedCountryRiskLevelNon3ds":["LOW"],"allowedECICodes":["02","05"]},"acquirerMerchantId":"TEST000001","principalAvailable":["VISA"]}]}}`),
			modifierMock:    func() { /* Empty Function */ },
			wantStatusCode:  http.StatusOK,
			wantRespBody:    `{"code":"00", "data":{"updated":true}}`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.modifierMock()

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/merchants/%s/payment-methods/%s/config", test.merchantID, test.paymentMethodID), bytes.NewBuffer(test.requestBody))

			router := chi.NewRouter()
			router.Patch("/merchants/{id}/payment-methods/{paymentMethodId}/config", New(svc).SetupConfig)

			router.ServeHTTP(rec, req)

			require.Equal(t, test.wantStatusCode, rec.Result().StatusCode)
			assert.JSONEq(t, test.wantRespBody, rec.Body.String())

			svc.AssertExpectations(t)
		})
	}
}
