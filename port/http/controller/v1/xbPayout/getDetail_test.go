package xbPayoutController_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/config"
	c "github.com/paper-indonesia/pivot-backoffice/constant"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	xbModel "github.com/paper-indonesia/pivot-backoffice/internal/model/xb"
	serviceMock "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	. "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/xbPayout"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetDetail(t *testing.T) {
	cfg := &config.Config{}
	xbPayoutSvc := serviceMock.NewIXbPayoutService(t)
	disbursementSvc := serviceMock.NewIDisbursementService(t)

	ctrl := New(cfg, WithXbPayoutService(xbPayoutSvc), WithDisbursementService(disbursementSvc))

	validPayoutID := uuid.NewString()
	validMerchantID := uuid.NewString()
	validRequestID := func(req *http.Request) {
		*req = *req.WithContext(context.WithValue(req.Context(), c.CtxUserInfoKey, &userModel.UserTokenClaims{
			MerchantId: validMerchantID,
		}))
		req.Header.Set("Content-Type", "application/json")
	}

	tests := []struct {
		name             string
		payoutID         string
		mockSetup        func()
		reqSetting       func(r *http.Request)
		expectedStatus   int
		expectedRespBody string
	}{
		{
			name:     "ERROR: Invalid user info",
			payoutID: "1",
			mockSetup: func() {
				// empty modifier
			},
			expectedStatus:   http.StatusUnauthorized,
			expectedRespBody: `{"code":"41", "data":null, "error":{"details":[], "traceId":"", "type":"API_ERROR"}, "message":"user not found"}`,
		},
		{
			name:     "ERROR: Invalid payout ID",
			payoutID: "1",
			mockSetup: func() {
				// empty modifier
			},
			reqSetting:       validRequestID,
			expectedStatus:   http.StatusBadRequest,
			expectedRespBody: `{"code":"40", "data":null, "error":{"details":[], "traceId":"", "type":"API_ERROR"}, "message":"id is required"}`,
		},
		{
			name:     "ERROR: Disbursement service error",
			payoutID: validPayoutID,
			mockSetup: func() {
				disbursementSvc.On("FindByID", c.ValueCtxMockType(), c.StringMockType()).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
			reqSetting:       validRequestID,
			expectedStatus:   http.StatusInternalServerError,
			expectedRespBody: `{"code":"99", "data":null, "error":{"details":[], "traceId":"", "type":"UNKNOWN"}, "message":"some error"}`,
		},
		{
			name:     "ERROR: Merchant is not match",
			payoutID: validPayoutID,
			mockSetup: func() {
				disbursementSvc.On("FindByID", c.ValueCtxMockType(), c.StringMockType()).Once().Return(&disbursementModel.DisbursementWithTransaction{}, nil)
			},
			reqSetting:       validRequestID,
			expectedStatus:   http.StatusBadRequest,
			expectedRespBody: `{"code":"40", "data":null, "error":{"details":[], "traceId":"", "type":"API_ERROR"}, "message":"merchant id is not match"}`,
		},
		{
			name:     "SUCCESS",
			payoutID: validPayoutID,
			mockSetup: func() {
				fee := decimal.NewFromFloat(0)
				reasonType := c.XbDisbursementReasonTypeSuccess
				reasonDesc := c.XbDisbursementReasonDescSuccess
				remark := "this is remark"
				disbursementSvc.On("FindByID", c.ValueCtxMockType(), c.StringMockType()).Once().Return(&disbursementModel.DisbursementWithTransaction{
					Disbursement: disbursementModel.Disbursement{
						MerchantID:        validMerchantID,
						Fee:               &fee,
						ReasonType:        &reasonType,
						ReasonDescription: &reasonDesc,
						Remark:            &remark,
						MetadataObj: disbursementModel.Metadata{
							XbDetail: &xbModel.XbPayoutMetadata{
								SourceAmount: decimal.NewFromFloat(1_000_000.0),
								TotalAmount:  decimal.NewFromFloat(1_000_000.0),
							},
						},
					},
				}, nil)
			},
			reqSetting:       validRequestID,
			expectedStatus:   http.StatusOK,
			expectedRespBody: `{"code":"00", "data":{"beneficiaryData":{"accountNumber":"", "accountType":"", "address":"", "bankCode":"", "bankName":"", "city":"", "contactCountryCode":"", "contactNumber":"", "countryCode":"", "email":"", "name":"", "payoutMethod":"", "postcode":"", "state":""}, "beneficiaryId":"", "createdAt":"0001-01-01T00:00:00Z", "destinationAmount":"0", "destinationCurrency":"", "destinationFxRate":"0", "expiredAt":"0001-01-01T00:00:00Z", "fee":"0", "fxRate":"0", "merchantId":"` + validMerchantID + `", "purposeCode":"", "referenceId":"", "remark":"this is remark", "routingCode":"", "routingValue":"", "senderData":{"accountType":"", "address":"", "bankAccountNumber":"", "city":"", "contactCountryCode":"", "contactNumber":"", "countryCode":"", "dob":"", "identificationNumber":"", "identificationType":"", "name":"", "postcode":"", "sourceOfIncome":"", "state":""}, "sourceAmount":"1000000", "sourceCurrency":"", "status":"SUCCESS", "statusDescription":"Payout success and has been received by beneficiary", "totalAmount":"1000000", "updatedAt":"0001-01-01T00:00:00Z", "uuid":""}, "message":"OK"}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockSetup()

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/xb/payout/%s", tc.payoutID), nil)

			if tc.reqSetting != nil {
				tc.reqSetting(req)
			}

			router := chi.NewRouter()
			router.Get("/api/v1/xb/payout/{id}", ctrl.GetDetail)

			router.ServeHTTP(rec, req)

			require.Equal(t, tc.expectedStatus, rec.Result().StatusCode)
			assert.JSONEq(t, tc.expectedRespBody, rec.Body.String())
		})
	}
}
