package recurringContractHandler_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	model "github.com/paper-indonesia/pivot-backoffice/internal/model/recurringContract"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	. "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/internalController/recurringContract"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCancel(t *testing.T) {
	service := serviceMocks.NewIRecurringContractService(t)

	handler := New(nil, service)

	router := chi.NewRouter()
	router.Post("/recurring/{uuid}/cancel", handler.Cancel)

	merchantAuth := &merchant.MerchantAuthTokenClaims{
		MerchantId: "4bf4e5be-59ef-4be8-aa89-33dc11615f70",
	}
	recurringID := "7a48d5ef-5857-44b1-a5e1-f8771d75d1a7"
	submerchantID := "7367dde7-7088-4c33-8d16-95e2d2edbbd3"
	successResponseBody := `{"code":"00","message":"Success","data":{"recurringId":"7a48d5ef-5857-44b1-a5e1-f8771d75d1a7","status":"INACTIVE"}}`

	tests := []struct {
		name             string
		merchantAuth     *merchant.MerchantAuthTokenClaims
		recurringID      string
		submerchantID    string
		setupMock        func()
		wantStatusCode   int
		wantResponseBody string
	}{
		{
			name:             "ERROR:Merchant auth not found", // NOSONAR
			wantStatusCode:   http.StatusUnauthorized,
			wantResponseBody: `{"code":"merchant_not_found","message":"Merchant not found","error":{"type":"API_ERROR","details":[{"field":"","message":"Invalid Merchant request"}],"traceId":""}}`,
		},
		{
			name:             "ERROR:Invalid recurring id", // NOSONAR
			merchantAuth:     merchantAuth,
			recurringID:      "ABC",
			wantStatusCode:   http.StatusBadRequest,
			wantResponseBody: `{"code":"field_format_invalid","message":"Format Field is invalid","error":{"type":"API_ERROR","details":[{"field":"recurringId","message":"Make sure recurringId format is correct"}],"traceId":""}}`,
		},
		{
			name:         "ERROR:Some error", // NOSONAR
			merchantAuth: merchantAuth,
			recurringID:  recurringID,
			setupMock: func() {
				service.On(
					"Cancel", mock.Anything, mock.MatchedBy(func(r model.CancelRecurringContractRequest) bool {
						return r.MerchantID == merchantAuth.MerchantId && r.RecurringID == recurringID
					}),
				).Once().Return(assert.AnError)
			},
			wantStatusCode:   http.StatusInternalServerError,
			wantResponseBody: `{"code":"general_error","message":"General error","error":{"type":"API_ERROR","details":[{"field":"","message":"Please contact our representative team"}],"traceId":""}}`,
		},
		{
			name:         "SUCCESS", // NOSONAR
			merchantAuth: merchantAuth,
			recurringID:  recurringID,
			setupMock: func() {
				service.On(
					"Cancel", mock.Anything, mock.MatchedBy(func(r model.CancelRecurringContractRequest) bool {
						return r.MerchantID == merchantAuth.MerchantId && r.RecurringID == recurringID
					}),
				).Once().Return(nil)
			},
			wantStatusCode:   http.StatusOK,
			wantResponseBody: successResponseBody,
		},
		{
			name:          "SUCCESS:On-behalf of sub-merchant", // NOSONAR
			merchantAuth:  merchantAuth,
			recurringID:   recurringID,
			submerchantID: submerchantID,
			setupMock: func() {
				service.On(
					"Cancel", mock.Anything, mock.MatchedBy(func(r model.CancelRecurringContractRequest) bool {
						return r.MerchantID == submerchantID && r.RecurringID == recurringID && r.UpdatedBy == merchantAuth.MerchantId
					}),
				).Once().Return(nil)
			},
			wantStatusCode:   http.StatusOK,
			wantResponseBody: successResponseBody,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/recurring/"+test.recurringID+"/cancel", nil)

			if test.submerchantID != "" {
				req.Header.Set(constant.HeaderXSubMerchantID, test.submerchantID)
			}
			if test.merchantAuth != nil {
				req = req.WithContext(context.WithValue(req.Context(), constant.CtxMerchantInfo, test.merchantAuth))
			}

			if test.setupMock != nil {
				test.setupMock()
			}

			router.ServeHTTP(rec, req)

			assert.Equal(t, test.wantStatusCode, rec.Result().StatusCode)
			if !assert.JSONEq(t, test.wantResponseBody, rec.Body.String()) {
				t.Log("Actual Response Body:", rec.Body.String())
			}

			service.AssertExpectations(t)
		})
	}
}
