package internalXbController_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestConfirmPayoutSession(t *testing.T) {
	cfg := &config.Config{}
	xbPayoutSvc := serviceMock.NewIXbPayoutService(t)
	logger := logger.NewSlogger(logger.Config{})
	svc := New(cfg, WithXbPayoutService(xbPayoutSvc), WithLogger(logger))

	validRequestID := func(req *http.Request) {
		*req = *req.WithContext(context.WithValue(req.Context(), c.CtxMerchantInfo, &merchantModel.MerchantAuthTokenClaims{
			MerchantId: "12345",
		}))
		req.Header.Set("Content-Type", "application/json")
	}

	validPayoutId := uuid.NewString()
	validPayload := xbModel.ConfirmPayoutRequest{
		PayoutId:   validPayoutId,
		MerchantId: uuid.NewString(),
		ApprovedBy: "System",
	}

	tests := []struct {
		name             string
		mockSetup        func()
		setupBody        func(*testing.T) []byte
		reqSetting       func(r *http.Request)
		payoutId         string
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
			name: "ERROR: Invalid id",
			setupBody: func(t *testing.T) []byte {
				return nil
			},
			mockSetup: func() {
				// empty modifier
			},
			reqSetting:       validRequestID,
			payoutId:         "invalid",
			expectedStatus:   http.StatusBadRequest,
			expectedRespBody: `{"code":"field_required","error":{"details":[{"field":"id","message":"Make sure id value is fulfilled"}],"traceId":"","type":"API_ERROR"},"message":"Mandatory field is missing"}`,
		},
		{
			name: "ERROR: Confirm service error",
			setupBody: func(t *testing.T) []byte {
				payloadRequestByte, err := json.Marshal(validPayload)
				assert.NoError(t, err)
				return payloadRequestByte
			},
			mockSetup: func() {
				xbPayoutSvc.On("Confirm",
					c.ValueCtxMockType(),
					mock.AnythingOfType("*xbModel.ConfirmPayoutRequest"),
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
			reqSetting:       validRequestID,
			payoutId:         validPayoutId,
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
				xbPayoutSvc.On("Confirm",
					c.ValueCtxMockType(),
					mock.AnythingOfType("*xbModel.ConfirmPayoutRequest"),
				).Once().Return(&xbModel.ConfirmPayoutResponse{}, nil)
			},
			reqSetting:       validRequestID,
			payoutId:         validPayoutId,
			expectedStatus:   http.StatusOK,
			expectedRespBody: `{"code":"00", "data":{"beneficiaryData":{"accountNumber":"", "accountType":"", "address":"", "bankCode":"", "bankName":"", "city":"", "contactCountryCode":"", "contactNumber":"", "countryCode":"", "email":"", "name":"", "payoutMethod":"", "postcode":"", "state":""}, "beneficiaryId":"", "createdAt":"0001-01-01T00:00:00Z", "destinationAmount":"0", "destinationCurrency":"", "destinationFxRate":"0", "fee":"0", "fxRate":"0", "merchantId":"", "referenceId":"", "remark":"", "senderData":{"accountType":"", "address":"", "bankAccountNumber":"", "city":"", "contactCountryCode":"", "contactNumber":"", "countryCode":"", "dob":"", "identificationNumber":"", "identificationType":"", "name":"", "postcode":"", "sourceOfIncome":"", "state":""}, "senderId":"", "sourceCurrency":"", "status":"", "totalAmount":"0", "uuid":""}, "message":"Success"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/open-api/v1/xb/create-payout-session/%s/confirm", tt.payoutId), bytes.NewBuffer(tt.setupBody(t)))

			if tt.reqSetting != nil {
				tt.reqSetting(req)
			}

			router := chi.NewRouter()
			router.Post("/open-api/v1/xb/create-payout-session/{id}/confirm", svc.ConfirmPayoutSession)

			router.ServeHTTP(rec, req)

			require.Equal(t, tt.expectedStatus, rec.Result().StatusCode)
			assert.JSONEq(t, tt.expectedRespBody, rec.Body.String())
		})
	}
}
