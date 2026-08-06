package recurringContractHandler_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	recurringContractModel "github.com/paper-indonesia/pivot-backoffice/internal/model/recurringContract"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/validatorExt"
	. "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/internalController/recurringContract"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreate(t *testing.T) {
	vld := validatorExt.New()
	service := serviceMocks.NewIRecurringContractService(t)

	handler := New(vld, service)

	router := chi.NewRouter()
	router.Post("/recurring", handler.Create)

	merchantAuth := &merchant.MerchantAuthTokenClaims{
		MerchantId: "4bf4e5be-59ef-4be8-aa89-33dc11615f70",
	}
	submerchantID := "6540a6bc-82aa-43d7-a401-3a02ad41e77c"
	validRequestBody := `{"clientReferenceId":"1","mode":"SELF_MANAGED","plan":{"planId":"123456789","planName":"Platinum Mobile"},"amount":{"currency":"IDR","value":75000},"billingInterval":1,"billingIntervalUnit":"MONTH","endDate":"2026-06-30T16:59:59Z","firstAuthorization":"ONE_DOLLAR","customerId":"cc121f58-d229-4e69-a4fe-11fd47821996"}`

	result := &recurringContractModel.CreateRecurringContractResponse{
		RecurringID: "1ee25b15-5535-4dc9-ad58-5e1de65b29a5",
		CustomerID:  "cc121f58-d229-4e69-a4fe-11fd47821996",
	}
	resultRaw, _ := json.Marshal(result)
	successResponseBody := fmt.Sprintf(`{"code":"00","message":"Success","data":%s}`, string(resultRaw))

	tests := []struct {
		name             string
		merchantAuth     *merchant.MerchantAuthTokenClaims
		requestBody      string
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
			name:             "ERROR:Malformed request body payload", // NOSONAR
			merchantAuth:     merchantAuth,
			requestBody:      "B", // NOSONAR
			wantStatusCode:   http.StatusBadRequest,
			wantResponseBody: `{"code":"api_validation_error","message":"malformed request body payload","error":{"type":"API_ERROR","details":[{"field":"","message":"malformed request body payload"}],"traceId":""}}`,
		},
		{
			name:             "ERROR:Fields validation", // NOSONAR
			merchantAuth:     merchantAuth,
			requestBody:      `{"clientReferenceId":"","mode":"SELF_MANAGED","plan":{"planId":"123456789","planName":"Platinum Mobile"},"amount":{"currency":"IDR","value":75000},"billingInterval":1,"billingIntervalUnit":"MONTH","endDate":"2026-06-30T16:59:59Z","firstAuthorization":"ONE_DOLLAR","customerId":"cc121f58-d229-4e69-a4fe-11fd47821996"}`,
			wantStatusCode:   http.StatusBadRequest,
			wantResponseBody: `{"code":"api_validation_error","message":"The request was invalid, or an error occurred in downstream provider","error":{"type":"API_ERROR","details":[{"field":"clientReferenceID","message":"Make sure clientReferenceID value is fulfilled"}],"traceId":""}}`,
		},
		{
			name:             "ERROR:Deep validation", // NOSONAR
			merchantAuth:     merchantAuth,
			requestBody:      `{"clientReferenceId":"1","mode":"SELF_MANAGED","plan":{"planId":"123456789","planName":"Platinum Mobile"},"amount":{"currency":"IDR","value":75000},"trials":[{"trialStart":2,"trialEnd":2,"type":"FREE"}],"billingInterval":1,"billingIntervalUnit":"MONTH","endDate":"2026-06-30T16:59:59Z","firstAuthorization":"ONE_DOLLAR","customerId":"cc121f58-d229-4e69-a4fe-11fd47821996"}`,
			wantStatusCode:   http.StatusBadRequest,
			wantResponseBody: `{"code":"api_validation_error","message":"Ensure the initial trial value starts from 1","error":{"type":"API_ERROR","details":[{"field":"","message":"Ensure the initial trial value starts from 1"}],"traceId":""}}`,
		},
		{
			name:         "ERROR:Some error", // NOSONAR
			merchantAuth: merchantAuth,
			requestBody:  validRequestBody,
			setupMock: func() {
				service.On("Create", mock.Anything, mock.Anything).Once().Return(nil, assert.AnError)
			},
			wantStatusCode:   http.StatusInternalServerError,
			wantResponseBody: `{"code":"general_error","message":"General error","error":{"type":"API_ERROR","details":[{"field":"","message":"Please contact our representative team"}],"traceId":""}}`,
		},
		{
			name:         "SUCCESS:Without sub merchant id", // NOSONAR
			merchantAuth: merchantAuth,
			requestBody:  validRequestBody,
			setupMock: func() {
				service.On(
					"Create", mock.Anything, mock.MatchedBy(func(r recurringContractModel.CreateRecurringContractRequest) bool {
						return r.MerchantID == merchantAuth.MerchantId
					}),
				).Once().Return(result, nil)
			},
			wantStatusCode:   http.StatusOK,
			wantResponseBody: successResponseBody,
		},
		{
			name:          "SUCCESS:With sub merchant id", // NOSONAR
			merchantAuth:  merchantAuth,
			requestBody:   validRequestBody,
			submerchantID: submerchantID,
			setupMock: func() {
				service.On(
					"Create", mock.Anything, mock.MatchedBy(func(r recurringContractModel.CreateRecurringContractRequest) bool {
						return r.MerchantID == submerchantID
					}),
				).Once().Return(result, nil)
			},
			wantStatusCode:   http.StatusOK,
			wantResponseBody: successResponseBody,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/recurring", strings.NewReader(test.requestBody))

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
		})
	}
}
