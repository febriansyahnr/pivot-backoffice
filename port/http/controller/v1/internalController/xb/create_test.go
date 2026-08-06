package internalXbController_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/config"
	c "github.com/paper-indonesia/pivot-backoffice/constant"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	xbModel "github.com/paper-indonesia/pivot-backoffice/internal/model/xb"
	serviceMock "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	. "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/internalController/xb"
	"github.com/paper-indonesia/pdk/v2/logger"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCreatePayoutSession(t *testing.T) {
	cfg := &config.Config{}
	xbPayoutSvc := serviceMock.NewIXbPayoutService(t)
	merchantSvc := serviceMock.NewIMerchantService(t)
	logger := logger.NewSlogger(logger.Config{})

	svc := New(cfg, WithXbPayoutService(xbPayoutSvc), WithMerchantService(merchantSvc), WithLogger(logger))

	validRequestID := func(req *http.Request) {
		*req = *req.WithContext(context.WithValue(req.Context(), c.CtxMerchantInfo, &merchantModel.MerchantAuthTokenClaims{
			MerchantId: "12345",
		}))
		req.Header.Set("Content-Type", "application/json")
	}

	merchantID := uuid.NewString()

	validPayload := xbModel.CreatePayoutSessionRequest{
		SenderData: &xbModel.CreateSenderRequest{
			MerchantId:           merchantID,
			Name:                 "John",
			CountryCode:          "ID",
			State:                "Jatim",
			City:                 "Malang",
			Address:              "Jl. Jalan",
			Postcode:             "12345",
			AccountType:          "Individual",
			IdentificationType:   "Registration ID",
			IdentificationNumber: "1234567890",
		},
		BeneficiaryData: &xbModel.CreateBeneficiaryRequest{
			MerchantId:           merchantID,
			Name:                 "Doe",
			CountryCode:          "US",
			State:                "New York",
			City:                 "New York",
			Address:              "Walking St.",
			Postcode:             "12345",
			AccountType:          "Individual",
			IdentificationType:   "Registration ID",
			IdentificationNumber: "1234567890",
			AccountNumber:        "350810001",
			BankName:             "Bank of America",
			BankCode:             "545343545345355",
		},
		ReferenceId:         "ref-001",
		SourceCurrency:      "IDR",
		DestinationCurrency: "USD",
		DestinationAmount:   decimal.NewFromFloat(100),
		PurposeCode:         "IR001",
		Remark:              "remark001",
		RoutingValue:        "USBKUS44IMT",
	}

	tests := []struct {
		name             string
		mockSetup        func()
		setupBody        func(*testing.T) []byte
		reqSetting       func(r *http.Request)
		expectedStatus   int
		expectedRespBody string
	}{
		{
			name: "ERROR: Invalid merchant info",
			setupBody: func(t *testing.T) []byte {
				return nil
			},
			mockSetup: func() {
				// empty modifier
			},
			expectedStatus:   http.StatusUnauthorized,
			expectedRespBody: `{"code":"merchant_not_found","error":{"details":[{"field":"","message":"Invalid Merchant request"}],"traceId":"","type":"API_ERROR"},"message":"Merchant not found"}`,
		},
		{
			name: "ERROR: Find merchant service",
			setupBody: func(t *testing.T) []byte {
				return nil
			},
			mockSetup: func() {
				merchantSvc.On("FindMerchantByID",
					c.ValueCtxMockType(),
					c.StringMockType(),
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
			reqSetting:       validRequestID,
			expectedStatus:   http.StatusBadRequest,
			expectedRespBody: `{"code":"merchant_not_found","error":{"details":[{"field":"","message":"Invalid Merchant request"}],"traceId":"","type":"API_ERROR"},"message":"Merchant not found"}`,
		},
		{
			name: "ERROR: Invalid request body",
			setupBody: func(t *testing.T) []byte {
				return []byte("{invalid json}")
			},
			mockSetup: func() {
				merchantSvc.On("FindMerchantByID",
					c.ValueCtxMockType(),
					c.StringMockType(),
				).Return(&merchantModel.Merchant{}, nil)
			},
			reqSetting:       validRequestID,
			expectedStatus:   http.StatusBadRequest,
			expectedRespBody: `{"code":"field_format_invalid","error":{"details":[{"field":"","message":"Please check the format of the field"}],"traceId":"","type":"API_ERROR"},"message":"Format Field is invalid"}`,
		},
		{
			name: "ERROR: Missing required request",
			setupBody: func(t *testing.T) []byte {
				payload := xbModel.CreatePayoutSessionRequest{
					SourceCurrency:      "IDR",
					DestinationCurrency: "USD",
					DestinationAmount:   decimal.NewFromFloat(100),
					PurposeCode:         "IR001",
					Remark:              "remark001",
					RoutingValue:        "USBKUS44IMT",
					SenderID:            uuid.NewString(),
					BeneficiaryID:       uuid.NewString(),
				}
				payloadRequestByte, err := json.Marshal(payload)
				assert.NoError(t, err)
				return payloadRequestByte
			},
			mockSetup: func() {
				// empty modifier
			},
			reqSetting:       validRequestID,
			expectedStatus:   http.StatusBadRequest,
			expectedRespBody: `{"code": "field_required","message": "Mandatory field is missing","error": {"type": "API_ERROR","details": [{"field": "referenceId","message": "Make sure referenceId value is fulfilled"}],"traceId": ""}}`,
		},
		{
			name: "ERROR: CreateSession service error",
			setupBody: func(t *testing.T) []byte {
				payloadRequestByte, err := json.Marshal(validPayload)
				assert.NoError(t, err)
				return payloadRequestByte
			},
			mockSetup: func() {
				xbPayoutSvc.On("CreateSession",
					c.ValueCtxMockType(),
					mock.AnythingOfType("*xbModel.CreatePayoutSessionRequest"),
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
			reqSetting:       validRequestID,
			expectedStatus:   http.StatusInternalServerError,
			expectedRespBody: `{"code":"general_error","error":{"details":[{"field":"","message":"Please contact our representative team"}],"traceId":"","type":"API_ERROR"},"message":"General error"}`,
		},
		{
			name: "SUCCESS",
			setupBody: func(t *testing.T) []byte {
				payloadRequestByte, err := json.Marshal(validPayload)
				assert.NoError(t, err)
				return payloadRequestByte
			},
			mockSetup: func() {
				xbPayoutSvc.On("CreateSession",
					c.ValueCtxMockType(),
					mock.AnythingOfType("*xbModel.CreatePayoutSessionRequest"),
				).Once().Return(&xbModel.CreatePayoutSessionResponse{}, nil)
			},
			reqSetting:       validRequestID,
			expectedStatus:   http.StatusOK,
			expectedRespBody: `{"code":"00", "data":{"beneficiaryData":{"accountNumber":"", "accountType":"", "address":"", "bankCode":"", "bankName":"", "city":"", "contactCountryCode":"", "contactNumber":"", "countryCode":"", "email":"", "name":"", "payoutMethod":"", "postcode":"", "state":""}, "beneficiaryId":"", "createdAt":"0001-01-01T00:00:00Z", "destinationAmount":"0", "destinationCurrency":"", "destinationFxRate":"0", "expiredAt":"0001-01-01T00:00:00Z", "fee":"0", "fxRate":"0", "merchantId":"", "referenceId":"", "remark":"", "routingCode":"", "routingValue":"", "senderData":{"accountType":"", "address":"", "bankAccountNumber":"", "city":"", "contactCountryCode":"", "contactNumber":"", "countryCode":"", "dob":"", "identificationNumber":"", "identificationType":"", "name":"", "postcode":"", "sourceOfIncome":"", "state":""}, "senderId":"", "sourceCurrency":"", "status":"", "totalAmount":"0", "uuid":""}, "message":"Success"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/open-api/v1/xb/create-payout-session", bytes.NewBuffer(tt.setupBody(t)))

			if tt.reqSetting != nil {
				tt.reqSetting(req)
			}

			router := chi.NewRouter()
			router.Post("/open-api/v1/xb/create-payout-session", svc.CreatePayoutSession)

			router.ServeHTTP(rec, req)

			require.Equal(t, tt.expectedStatus, rec.Result().StatusCode)
			assert.JSONEq(t, tt.expectedRespBody, rec.Body.String())
		})
	}
}
