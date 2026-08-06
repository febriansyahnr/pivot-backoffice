package xbPayoutController_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	"github.com/shopspring/decimal"

	"github.com/go-chi/chi/v5"
	"github.com/paper-indonesia/pivot-backoffice/config"
	c "github.com/paper-indonesia/pivot-backoffice/constant"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	xbModel "github.com/paper-indonesia/pivot-backoffice/internal/model/xb"
	serviceMock "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	. "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/xbPayout"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCreate(t *testing.T) {
	cfg := &config.Config{}
	xbPayoutSvc := serviceMock.NewIXbPayoutService(t)
	merchantSvc := serviceMock.NewIMerchantService(t)

	ctrl := New(cfg, WithXbPayoutService(xbPayoutSvc), WithMerchantService(merchantSvc))

	validRequestID := func(req *http.Request) {
		*req = *req.WithContext(context.WithValue(req.Context(), c.CtxUserInfoKey, &userModel.UserTokenClaims{
			MerchantId: "12345",
		}))
		req.Header.Set("Content-Type", "application/json")
	}

	validRequest := &xbModel.CreatePayoutSessionRequest{
		SenderData: &xbModel.CreateSenderRequest{
			Name:                 "John Doe",
			CountryCode:          "ID",
			State:                "East Java",
			City:                 "Malang",
			Address:              "St. Jalan",
			Postcode:             "1234",
			AccountType:          "Individual",
			IdentificationType:   "KTP",
			IdentificationNumber: "1234",
		},
		BeneficiaryData: &xbModel.CreateBeneficiaryRequest{
			Name:                 "John Doe",
			CountryCode:          "US",
			State:                "Washington",
			City:                 "Washington",
			Address:              "St. Jalan",
			Postcode:             "1234",
			AccountType:          "Individual",
			IdentificationType:   "KTP",
			IdentificationNumber: "1234",
			AccountNumber:        "1234",
			BankName:             "Bank of America",
		},
		ReferenceId:         "testref1",
		SourceCurrency:      "IDR",
		DestinationCurrency: "USD",
		DestinationAmount:   decimal.NewFromFloat(100),
		PurposeCode:         "IR001",
		RoutingValue:        "MFBBMYKLXXX",
		Remark:              "Payout remark",
	}
	validRequestBody, _ := json.Marshal(validRequest)

	tests := []struct {
		name             string
		requestBody      []byte
		mockSetup        func()
		reqSetting       func(r *http.Request)
		expectedStatus   int
		expectedRespBody string
	}{
		{
			name: "ERROR: Invalid user info",
			mockSetup: func() {
				// empty modifier
			},
			expectedStatus:   http.StatusUnauthorized,
			expectedRespBody: `{"code":"41", "data":null, "error":{"details":[], "traceId":"", "type":"API_ERROR"}, "message":"user not found"}`,
		},
		{
			name: "ERROR: Merchant service error",
			mockSetup: func() {
				merchantSvc.On("FindMerchantByID", c.ValueCtxMockType(), c.StringMockType()).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
			reqSetting:       validRequestID,
			expectedStatus:   http.StatusBadRequest,
			expectedRespBody: `{"code":"40", "data":null, "error":{"details":[], "traceId":"", "type":"API_ERROR"}, "message":"some error"}`,
		},
		{
			name: "ERROR: Merchant not found",
			mockSetup: func() {
				merchantSvc.On("FindMerchantByID", c.ValueCtxMockType(), c.StringMockType()).Once().Return(nil, nil)
			},
			reqSetting:       validRequestID,
			expectedStatus:   http.StatusBadRequest,
			expectedRespBody: `{"code":"40", "data":null, "error":{"details":[], "traceId":"", "type":"API_ERROR"}, "message":"merchant not found"}`,
		},
		{
			name: "ERROR: Bad request",
			mockSetup: func() {
				merchantSvc.On("FindMerchantByID", c.ValueCtxMockType(), c.StringMockType()).Return(&merchant.Merchant{Name: "Test Merchant"}, nil)
			},
			requestBody:      []byte("{invalid JSON"),
			reqSetting:       validRequestID,
			expectedStatus:   http.StatusBadRequest,
			expectedRespBody: `{"code":"40", "data":null, "error":{"details":[], "traceId":"", "type":"API_ERROR"}, "message":"invalid character 'i' looking for beginning of object key string"}`,
		},
		{
			name: "ERROR: Failed validation",
			mockSetup: func() {
				merchantSvc.On("FindMerchantByID", c.ValueCtxMockType(), c.StringMockType()).Return(&merchant.Merchant{Name: "Test Merchant"}, nil)
			},
			requestBody:      []byte(`{"referenceId": "12345abcde"}`),
			reqSetting:       validRequestID,
			expectedStatus:   http.StatusBadRequest,
			expectedRespBody: `{"code":"40", "data":null, "error":{"details":[{"field":"SenderID", "message":"Key: 'CreatePayoutSessionRequest.SenderID' Error:Field validation for 'SenderID' failed on the 'required_without' tag"}, {"field":"BeneficiaryID", "message":"Key: 'CreatePayoutSessionRequest.BeneficiaryID' Error:Field validation for 'BeneficiaryID' failed on the 'required_without' tag"}, {"field":"SourceCurrency", "message":"Key: 'CreatePayoutSessionRequest.SourceCurrency' Error:Field validation for 'SourceCurrency' failed on the 'required' tag"}, {"field":"DestinationCurrency", "message":"Key: 'CreatePayoutSessionRequest.DestinationCurrency' Error:Field validation for 'DestinationCurrency' failed on the 'required' tag"}, {"field":"PurposeCode", "message":"Key: 'CreatePayoutSessionRequest.PurposeCode' Error:Field validation for 'PurposeCode' failed on the 'required' tag"}, {"field":"Remark", "message":"Key: 'CreatePayoutSessionRequest.Remark' Error:Field validation for 'Remark' failed on the 'required' tag"}], "traceId":"", "type":"API_ERROR"}, "message":"invalid validation"}`,
		},
		{
			name: "ERROR: Create service error",
			mockSetup: func() {
				xbPayoutSvc.On("CreateSession",
					c.ValueCtxMockType(),
					mock.AnythingOfType("*xbModel.CreatePayoutSessionRequest"),
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
			reqSetting:       validRequestID,
			requestBody:      validRequestBody,
			expectedStatus:   http.StatusInternalServerError,
			expectedRespBody: `{"code":"99", "data":null, "error":{"details":[], "traceId":"", "type":"UNKNOWN"}, "message":"some error"}`,
		},
		{
			name: "SUCCESS",
			mockSetup: func() {
				xbPayoutSvc.On("CreateSession",
					c.ValueCtxMockType(),
					mock.AnythingOfType("*xbModel.CreatePayoutSessionRequest"),
				).Once().Return(&xbModel.CreatePayoutSessionResponse{}, nil)
			},
			reqSetting:       validRequestID,
			requestBody:      validRequestBody,
			expectedStatus:   http.StatusOK,
			expectedRespBody: `{"code":"00", "data":{"beneficiaryData":{"accountNumber":"", "accountType":"", "address":"", "bankCode":"", "bankName":"", "city":"", "contactCountryCode":"", "contactNumber":"", "countryCode":"", "email":"", "name":"", "payoutMethod":"", "postcode":"", "state":""}, "beneficiaryId":"", "createdAt":"0001-01-01T00:00:00Z", "destinationAmount":"0", "destinationCurrency":"", "destinationFxRate":"0", "expiredAt":"0001-01-01T00:00:00Z", "fee":"0", "fxRate":"0", "merchantId":"", "referenceId":"", "remark":"", "routingCode":"", "routingValue":"", "senderData":{"accountType":"", "address":"", "bankAccountNumber":"", "city":"", "contactCountryCode":"", "contactNumber":"", "countryCode":"", "dob":"", "identificationNumber":"", "identificationType":"", "name":"", "postcode":"", "sourceOfIncome":"", "state":""}, "senderId":"", "sourceCurrency":"", "status":"", "totalAmount":"0", "uuid":""}, "message":"OK"}`,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockSetup()

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/api/v1/xb/payout", bytes.NewBuffer(tc.requestBody))

			if tc.reqSetting != nil {
				tc.reqSetting(req)
			}

			router := chi.NewRouter()
			router.Post("/api/v1/xb/payout", ctrl.CreateSession)

			router.ServeHTTP(rec, req)

			require.Equal(t, tc.expectedStatus, rec.Result().StatusCode)
			assert.JSONEq(t, tc.expectedRespBody, rec.Body.String())
		})
	}
}
