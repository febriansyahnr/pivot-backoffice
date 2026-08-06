package crmCreditcardController

import (
	"bytes"
	"github.com/stretchr/testify/mock"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	creditcardModel "github.com/paper-indonesia/pivot-backoffice/internal/model/creditcard"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
)

func TestVoid(t *testing.T) {
	svc := serviceMocks.NewICreditCardService(t)

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
			name:        "ERROR: Bad Request - Failed Validation",
			requestBody: []byte(`{"merchantId": ""}`),
			modifierMock: func() {
				// empty modifier
			},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40", "errors":"invalid UUID length: 0"}`,
		},
		{
			name:        "ERROR: Service error",
			requestBody: []byte(`{"merchantId": "123e4567-e89b-12d3-a456-426614174000", "referenceId": "REF123"}`),
			modifierMock: func() {
				svc.On("Void", constant.ValueCtxMockType(), mock.Anything).
					Once().Return(nil, constant.ErrSomeErrorForUnitTest)
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   `{"code":"99", "errors":"some error"}`,
		},
		{
			name:        "SUCCESS",
			requestBody: []byte(`{"merchantId": "123e4567-e89b-12d3-a456-426614174000", "referenceId": "REF123"}`),
			modifierMock: func() {
				mockResponse := &creditcardModel.VoidResponse{}
				svc.On("Void", constant.ValueCtxMockType(), mock.Anything).
					Return(mockResponse, nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code":"00", "data":{"card_brand":"", "created_at":"", "grand_total_amount":"0", "status":""}}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.modifierMock()

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/crm/v1/creditcard/void", bytes.NewBuffer(test.requestBody))

			router := chi.NewRouter()
			router.Post("/crm/v1/creditcard/void", New(&config.Config{}, &config.Secret{}, svc).Void)

			router.ServeHTTP(rec, req)

			require.Equal(t, test.wantStatusCode, rec.Result().StatusCode)
			assert.JSONEq(t, test.wantRespBody, rec.Body.String())
		})
	}
}
