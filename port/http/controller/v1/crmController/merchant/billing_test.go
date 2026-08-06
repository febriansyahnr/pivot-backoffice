package merchant_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/validatorExt"
	. "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/crmController/merchant"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetBillingFees(t *testing.T) {
	service := serviceMocks.NewIMerchantService(t)

	route := chi.NewRouter()
	route.Get("/{merchantId}/billing/fees", New(service, nil, validatorExt.New(), nil).GetBillingFees)

	merchantId := "32f86ccf-2698-4cda-8ccf-4a5354fd00bc"
	params := `startDate=2025-06-16&endDate=2025-06-16&status=unpaid`

	tests := []struct {
		name           string
		merchantId     string
		params         string
		setupMock      func()
		wantStatusCode int
		wantRespBody   string
	}{
		{
			name:           "ERROR:Invalid merchant id",
			merchantId:     "XXX",
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40","errors":{"MerchantId":"Key: 'BillingFeeRequest.MerchantId' Error:Field validation for 'MerchantId' failed on the 'uuid' tag"}}`,
		},
		{
			name:           "ERROR:Invalid date range input",
			merchantId:     merchantId,
			params:         "startDate=2025-06-16&endDate=2025-06-15",
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40","errors":"invalid date range"}`,
		},
		{
			name:       "ERROR:Some error", // NOSONAR
			merchantId: merchantId,
			setupMock: func() {
				service.On(
					"GetBillingFees", mock.Anything, mock.Anything,
				).Once().Return(nil, assert.AnError)
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   `{"code":"99","errors":"assert.AnError general error for testing"}`,
		},
		{
			name:       "SUCCESS", // NOSONAR
			merchantId: merchantId,
			setupMock: func() {
				service.On(
					"GetBillingFees", mock.Anything, mock.Anything,
				).Once().Return(&merchant.BillingFeeResponse{Details: map[string][]merchant.BillingFeeDetailResponse{}}, nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code":"00","data":{"merchantId":"","merchantName":"","total":0,"totalFeeAmount":0,"details":{},"subMerchants":null}}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			if test.params == "" {
				test.params = params
			}
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/%s/billing/fees?%s", test.merchantId, test.params), nil)

			if test.setupMock != nil {
				test.setupMock()
			}

			route.ServeHTTP(rec, req)
			assert.Equal(t, test.wantStatusCode, rec.Result().StatusCode)
			if !assert.JSONEq(t, test.wantRespBody, rec.Body.String()) {
				t.Log("Actual Response:", rec.Body.String())
			}
		})
	}
}

func TestPayBillingFees(t *testing.T) {
	service := serviceMocks.NewIMerchantService(t)

	route := chi.NewRouter()
	route.Post("/{merchantId}/billing/fees/pay", New(service, nil, validatorExt.New(), nil).PayBillingFees)

	merchantId := "df5666bf-c18c-43ef-b81e-af95354e3819"
	validRequestBody := `{"startDate": "2025-06-16","endDate": "2025-06-16","username": "john"}`

	tests := []struct {
		name           string
		merchantId     string
		reqBody        string
		setupMock      func()
		wantStatusCode int
		wantRespBody   string
	}{
		{
			name:           "ERROR:Invalid request body",
			reqBody:        "A",
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40","errors":"invalid character 'A' looking for beginning of value"}`,
		},
		{
			name:           "ERROR:Invalid merchant id",
			merchantId:     "XXX",
			reqBody:        validRequestBody,
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40","errors":{"MerchantId":"Key: 'PayBillingFeeRequest.MerchantId' Error:Field validation for 'MerchantId' failed on the 'uuid' tag"}}`,
		},
		{
			name:           "ERROR:Invalid date range",
			merchantId:     merchantId,
			reqBody:        `{"startDate": "2025-06-18","endDate": "2025-06-16","username": "john"}`,
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40","errors":"invalid date range"}`,
		},
		{
			name:       "ERROR:Some error", // NOSONAR
			merchantId: merchantId,
			reqBody:    validRequestBody,
			setupMock: func() {
				service.On("PayBillingFees", mock.Anything, mock.Anything).Once().Return(nil, assert.AnError)
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   `{"code":"99","errors":"assert.AnError general error for testing"}`,
		},
		{
			name:       "SUCCESS", // NOSONAR
			merchantId: merchantId,
			reqBody:    validRequestBody,
			setupMock: func() {
				service.On("PayBillingFees", mock.Anything, mock.Anything).Once().Return(&merchant.BillingFeeResponse{Details: map[string][]merchant.BillingFeeDetailResponse{}}, nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code":"00","data":{"merchantId":"","merchantName":"","total":0,"totalFeeAmount":0,"details":{},"subMerchants":null}}`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/%s/billing/fees/pay", test.merchantId), strings.NewReader(test.reqBody))

			if test.setupMock != nil {
				test.setupMock()
			}

			route.ServeHTTP(rec, req)
			assert.Equal(t, test.wantStatusCode, rec.Result().StatusCode)
			if !assert.JSONEq(t, test.wantRespBody, rec.Body.String()) {
				t.Log("Actual Response:", rec.Body.String())
			}
		})
	}
}
