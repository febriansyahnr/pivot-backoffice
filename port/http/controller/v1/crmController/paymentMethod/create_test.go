package crmPaymentMethodController

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	paymentMethodModel "github.com/paper-indonesia/pivot-backoffice/internal/model/paymentMethod"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
)

func TestCreate(t *testing.T) {
	svc := serviceMocks.NewIPaymentMethodService(t)

	tests := []struct {
		name           string
		requestBody    []byte
		modifierMock   func()
		wantStatusCode int
		wantRespBody   string
	}{
		{
			name:        "ERROR: Bad Request - Invalid JSON",
			requestBody: []byte("{invalid JSON"),
			modifierMock: func() {
				// empty modifier
			},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40", "errors":"invalid character 'i' looking for beginning of object key string"}`,
		},
		{
			name:        "ERROR: Bad Request - Failed Validation - Missing type",
			requestBody: []byte(`{"category":"PAYMENT","name":"Test","acquirer":"TEST","activationMethod":"INSTANT","description":"desc","logo":"logo.png"}`),
			modifierMock: func() {
				// empty modifier
			},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40", "errors":{"Type":"Key: 'CreatePaymentMethodRequest.Type' Error:Field validation for 'Type' failed on the 'required' tag"}}`,
		},
		{
			name:        "ERROR: Bad Request - Failed Validation - Invalid type value",
			requestBody: []byte(`{"type":"INVALID_TYPE","category":"PAYMENT","name":"Test","acquirer":"TEST","activationMethod":"INSTANT","description":"desc","logo":"logo.png"}`),
			modifierMock: func() {
				// empty modifier
			},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40", "errors":{"Type":"Key: 'CreatePaymentMethodRequest.Type' Error:Field validation for 'Type' failed on the 'oneof' tag"}}`,
		},
		{
			name:        "ERROR: Bad Request - Failed Validation - Missing category",
			requestBody: []byte(`{"type":"VIRTUAL_ACCOUNT","name":"Test","acquirer":"TEST","activationMethod":"INSTANT","description":"desc","logo":"logo.png"}`),
			modifierMock: func() {
				// empty modifier
			},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40", "errors":{"Category":"Key: 'CreatePaymentMethodRequest.Category' Error:Field validation for 'Category' failed on the 'required' tag"}}`,
		},
		{
			name:        "ERROR: Bad Request - Failed Validation - Invalid category value",
			requestBody: []byte(`{"type":"VIRTUAL_ACCOUNT","category":"INVALID_CATEGORY","name":"Test","acquirer":"TEST","activationMethod":"INSTANT","description":"desc","logo":"logo.png"}`),
			modifierMock: func() {
				// empty modifier
			},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40", "errors":{"Category":"Key: 'CreatePaymentMethodRequest.Category' Error:Field validation for 'Category' failed on the 'oneof' tag"}}`,
		},
		{
			name:        "ERROR: Bad Request - Failed Validation - Missing name",
			requestBody: []byte(`{"type":"VIRTUAL_ACCOUNT","category":"PAYMENT","acquirer":"TEST","activationMethod":"INSTANT","description":"desc","logo":"logo.png"}`),
			modifierMock: func() {
				// empty modifier
			},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40", "errors":{"Name":"Key: 'CreatePaymentMethodRequest.Name' Error:Field validation for 'Name' failed on the 'required' tag"}}`,
		},
		{
			name:        "ERROR: Bad Request - Failed Validation - Missing acquirer",
			requestBody: []byte(`{"type":"VIRTUAL_ACCOUNT","category":"PAYMENT","name":"Test","activationMethod":"INSTANT","description":"desc","logo":"logo.png"}`),
			modifierMock: func() {
				// empty modifier
			},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40", "errors":{"Acquirer":"Key: 'CreatePaymentMethodRequest.Acquirer' Error:Field validation for 'Acquirer' failed on the 'required' tag"}}`,
		},
		{
			name:        "ERROR: Bad Request - Failed Validation - Missing activationMethod",
			requestBody: []byte(`{"type":"VIRTUAL_ACCOUNT","category":"PAYMENT","name":"Test","acquirer":"TEST","description":"desc","logo":"logo.png"}`),
			modifierMock: func() {
				// empty modifier
			},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40", "errors":{"ActivationMethod":"Key: 'CreatePaymentMethodRequest.ActivationMethod' Error:Field validation for 'ActivationMethod' failed on the 'required' tag"}}`,
		},
		{
			name:        "ERROR: Bad Request - Failed Validation - Invalid activationMethod value",
			requestBody: []byte(`{"type":"VIRTUAL_ACCOUNT","category":"PAYMENT","name":"Test","acquirer":"TEST","activationMethod":"INVALID","description":"desc","logo":"logo.png"}`),
			modifierMock: func() {
				// empty modifier
			},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40", "errors":{"ActivationMethod":"Key: 'CreatePaymentMethodRequest.ActivationMethod' Error:Field validation for 'ActivationMethod' failed on the 'oneof' tag"}}`,
		},
		{
			name:        "ERROR: Service error",
			requestBody: []byte(`{"type":"VIRTUAL_ACCOUNT","category":"PAYMENT","name":"Test","acquirer":"BCA","activationMethod":"INSTANT","description":"desc","logo":"logo.png"}`),
			modifierMock: func() {
				svc.On("Create", constant.ValueCtxMockType(), &paymentMethodModel.CreatePaymentMethodRequest{
					Type:             "VIRTUAL_ACCOUNT",
					Category:         "PAYMENT",
					Name:             "Test",
					Acquirer:         "BCA",
					ActivationMethod: "INSTANT",
					Logo:             util.ValueToPtr("logo.png"),
					Description:      util.ValueToPtr("desc"),
				}).
					Once().Return(constant.ErrSomeErrorForUnitTest)
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   `{"code":"99", "errors":"some error"}`,
		},
		{
			name:        "SUCCESS - Create VIRTUAL_ACCOUNT payment method with minimal fields",
			requestBody: []byte(`{"type":"VIRTUAL_ACCOUNT","category":"PAYMENT","name":"Test","acquirer":"BCA","activationMethod":"INSTANT","description":"desc","logo":"logo.png"}`),
			modifierMock: func() {
				svc.On("Create", constant.ValueCtxMockType(), &paymentMethodModel.CreatePaymentMethodRequest{
					Type:             "VIRTUAL_ACCOUNT",
					Category:         "PAYMENT",
					Name:             "Test",
					Acquirer:         "BCA",
					ActivationMethod: "INSTANT",
					Logo:             util.ValueToPtr("logo.png"),
					Description:      util.ValueToPtr("desc"),
				}).
					Return(nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code":"00", "data":{"created":true}}`,
		},
		{
			name:        "SUCCESS - Create EWALLET payment method with optional fields",
			requestBody: []byte(`{"type":"EWALLET","subtype":"GOPAY","category":"PAYMENT","name":"GoPay","description":"desc","logo":"logo.png","acquirer":"GOJEK","bankName":"GoPay","instructions":"Scan","processor":"GOPAY","activationMethod":"API","countryOfOperation":"ID","supportedCurrency":"ID"}`),
			modifierMock: func() {
				svc.On("Create", constant.ValueCtxMockType(), mock.AnythingOfType("*paymentMethodModel.CreatePaymentMethodRequest")).
					Return(nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code":"00", "data":{"created":true}}`,
		},
		{
			name:        "SUCCESS - Create QRIS payment method",
			requestBody: []byte(`{"type":"QRIS","category":"PAYMENT","name":"QRIS","acquirer":"BCA","activationMethod":"MANUAL","description":"desc","logo":"logo.png"}`),
			modifierMock: func() {
				svc.On("Create", constant.ValueCtxMockType(), mock.AnythingOfType("*paymentMethodModel.CreatePaymentMethodRequest")).
					Return(nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code":"00", "data":{"created":true}}`,
		},
		{
			name:        "SUCCESS - Create CREDIT_CARD payment method",
			requestBody: []byte(`{"type":"CREDIT_CARD","category":"PAYMENT","name":"CC","acquirer":"MASTER","activationMethod":"INSTANT","description":"desc","logo":"logo.png"}`),
			modifierMock: func() {
				svc.On("Create", constant.ValueCtxMockType(), mock.AnythingOfType("*paymentMethodModel.CreatePaymentMethodRequest")).
					Return(nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code":"00", "data":{"created":true}}`,
		},
		{
			name:        "SUCCESS - Create MERCHANT_TOP_UP category",
			requestBody: []byte(`{"type":"VIRTUAL_ACCOUNT","category":"MERCHANT_TOP_UP","name":"TopUp","acquirer":"BCA","activationMethod":"INSTANT","description":"desc","logo":"logo.png"}`),
			modifierMock: func() {
				svc.On("Create", constant.ValueCtxMockType(), mock.AnythingOfType("*paymentMethodModel.CreatePaymentMethodRequest")).
					Return(nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code":"00", "data":{"created":true}}`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.modifierMock()

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/crm/v1/payment-methods", bytes.NewBuffer(test.requestBody))

			router := chi.NewRouter()
			router.Post("/crm/v1/payment-methods", New(svc).Create)

			router.ServeHTTP(rec, req)

			require.Equal(t, test.wantStatusCode, rec.Result().StatusCode)
			assert.JSONEq(t, test.wantRespBody, rec.Body.String())
		})
	}
}
