package internalXbController

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	chi "github.com/go-chi/chi/v5"
	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	xbModel "github.com/paper-indonesia/pivot-backoffice/internal/model/xb"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetPayoutById(t *testing.T) {
	tests := []struct {
		name           string
		id             string
		wantStatusCode int
		wantRespBody   string
		reqSetting     func(r *http.Request)
		mockSetup      func(d *serviceMocks.IXbPayoutService)
	}{
		{
			name:           "ERROR: invalid merchant info",
			id:             "48e0d7dd-c10f-4032-a70f-64357ee34939",
			wantStatusCode: http.StatusUnauthorized,
			wantRespBody:   `{"code":"merchant_not_found","error":{"details":[{"field":"","message":"Invalid Merchant request"}],"traceId":"","type":"API_ERROR"},"message":"Merchant not found"}`,
			reqSetting: func(req *http.Request) {
				*req = *req.WithContext(context.WithValue(req.Context(), constant.CtxMerchantInfo, "invalid"))
				req.Header.Set("Content-Type", "application/json")
			},
		},
		{
			name:           "ERROR: invalid id",
			id:             "invalid",
			wantStatusCode: http.StatusBadRequest,
			wantRespBody: `{
				"code":"field_required",
				"error":{
					"details": [
						{
							"field": "id",
							"message": "Make sure id value is fulfilled"
						}
					],
					"traceId": "",
					"type": "API_ERROR"
				},
				"message":"Mandatory field is missing"
			}`,
			reqSetting: func(req *http.Request) {
				*req = *req.WithContext(context.WithValue(req.Context(), constant.CtxMerchantInfo, &merchantModel.MerchantAuthTokenClaims{
					MerchantId: "12345",
				}))
				req.Header.Set("Content-Type", "application/json")
			},
		},
		{
			name:           "ERROR: error when request get payout id to xb core processor",
			id:             "48e0d7dd-c10f-4032-a70f-64357ee34939",
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody: `{
				"code":"general_error",
				"error":{
					"details": [
						{
							"field": "",
							"message": "Please contact our representative team"
						}
					],
					"traceId": "",
					"type": "API_ERROR"
				},
				"message":"General error"
			}`,
			reqSetting: func(req *http.Request) {
				*req = *req.WithContext(context.WithValue(req.Context(), constant.CtxMerchantInfo, &merchantModel.MerchantAuthTokenClaims{
					MerchantId: "12345",
				}))
				req.Header.Set("Content-Type", "application/json")
			},
			mockSetup: func(d *serviceMocks.IXbPayoutService) {
				d.On("GetPayoutById", mock.Anything, mock.Anything).Return(nil, assert.AnError)
			},
		},
		{
			name:           "SUCCESS: get payout id",
			id:             "48e0d7dd-c10f-4032-a70f-64357ee34939",
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code":"00", "data":{"beneficiaryData":{"accountNumber":"", "accountType":"", "address":"", "bankCode":"", "bankName":"", "city":"", "contactCountryCode":"", "contactNumber":"", "countryCode":"", "email":"", "name":"", "payoutMethod":"", "postcode":"", "state":""}, "createdAt":"0001-01-01T00:00:00Z", "destinationAmount":"0", "destinationCurrency":"", "destinationFxRate":"0", "fee":"0", "fxRate":"0", "merchantId":"", "referenceId":"", "remark":"", "routingCode":"", "routingValue":"", "senderData":{"accountType":"", "address":"", "bankAccountNumber":"", "city":"", "contactCountryCode":"", "contactNumber":"", "countryCode":"", "dob":"", "identificationNumber":"", "identificationType":"", "name":"", "postcode":"", "sourceOfIncome":"", "state":""}, "sourceCurrency":"", "status":"", "totalAmount":"0", "uuid":""}, "message":"Success"}`,
			reqSetting: func(req *http.Request) {
				*req = *req.WithContext(context.WithValue(req.Context(), constant.CtxMerchantInfo, &merchantModel.MerchantAuthTokenClaims{
					MerchantId: "12345",
				}))
				req.Header.Set("Content-Type", "application/json")
			},
			mockSetup: func(d *serviceMocks.IXbPayoutService) {
				d.On("GetPayoutById", mock.Anything, mock.Anything).Return(&xbModel.GetPayoutResponse{}, nil)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			conf := &config.Config{
				Environment: "development",
			}
			mockXbPayoutSvc := serviceMocks.NewIXbPayoutService(t)

			if tc.mockSetup != nil {
				tc.mockSetup(mockXbPayoutSvc)
			}

			c := New(conf, WithXbPayoutService(mockXbPayoutSvc))

			router := chi.NewRouter()
			router.Get("/payouts/{id}", c.GetPayoutById)

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/payouts/"+tc.id, nil)

			tc.reqSetting(req)

			router.ServeHTTP(rec, req)
			assert.Equal(t, tc.wantStatusCode, rec.Result().StatusCode)
			assert.JSONEq(t, tc.wantRespBody, rec.Body.String())

			mockXbPayoutSvc.AssertExpectations(t)
		})
	}

}
