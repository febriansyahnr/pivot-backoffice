package creditcard_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/validatorExt"
	. "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/internalController/creditcard"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRemoveCardByCustomerIDAndTokenID(t *testing.T) {
	service := serviceMocks.NewICreditCardService(t)

	handler := New(nil, validatorExt.New(), nil, nil, Services{CreditcardSvc: service})

	route := chi.NewRouter()
	route.Delete("/stored-card/{merchantId}/{customerId}/{tokenId}", handler.RemoveCardByCustomerIDAndTokenID)

	var (
		merchantId = "6381f181-4b57-44b5-901e-773761af880e"
		customerId = "7629b7be-f4d9-4b90-bb01-fe87540a65a4"
		tokenId    = "7629b7be-f4d9-4b90-bb01-fe87540a65a4"
		paymentId  = "1d891d92-786f-454a-b8f2-75c647112911"
	)

	tests := []struct {
		name             string
		paymentId        string
		setupMock        func()
		wantStatusCode   int
		wantResponseBody string
	}{
		{
			name:             "ERROR:Empty payment id", // NOSONAR
			paymentId:        "",
			setupMock:        func() { /* Empty Function */ },
			wantStatusCode:   http.StatusBadRequest,
			wantResponseBody: `{"code":"42","message":"invalid validation","error":{"type":"API_VALIDATION_ERROR","details":[{"field":"PaymentID","message":"Key: 'RemoveCardTokenizationRequest.PaymentID' Error:Field validation for 'PaymentID' failed on the 'required' tag"}],"traceId":""},"data":null}`,
		},
		{
			name:      "ERROR:Some error", // NOSONAR
			paymentId: paymentId,
			setupMock: func() {
				service.On("RemoveCardTokenization", mock.Anything, mock.Anything).Once().Return(assert.AnError)
			},
			wantStatusCode:   http.StatusInternalServerError,
			wantResponseBody: `{"code":"99","message":"assert.AnError general error for testing","error":{"type":"UNKNOWN","details":[],"traceId":""},"data":null}`,
		},
		{
			name:      "SUCCESS", // NOSONAR
			paymentId: paymentId,
			setupMock: func() {
				service.On("RemoveCardTokenization", mock.Anything, mock.Anything).Once().Return(nil)
			},
			wantStatusCode:   http.StatusOK,
			wantResponseBody: fmt.Sprintf(`{"code":"00","message":"OK","data":{"customerId":"%s", "tokenId":"%s"}}`, customerId, tokenId),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequestWithContext(
				t.Context(), http.MethodDelete, fmt.Sprintf("/stored-card/%s/%s/%s", merchantId, customerId, tokenId), nil,
			)
			req = req.WithContext(context.WithValue(req.Context(), constant.CtxPaymentID, test.paymentId))

			test.setupMock()
			route.ServeHTTP(rec, req)

			assert.Equal(t, test.wantStatusCode, rec.Result().StatusCode)
			if !assert.JSONEq(t, test.wantResponseBody, rec.Body.String()) {
				t.Log("Handler Result:", rec.Body.String())
			}
		})
	}
}
