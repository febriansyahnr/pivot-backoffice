package v1InternalRefundController_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	refundModel "github.com/paper-indonesia/pivot-backoffice/internal/model/refund"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	. "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/internalController/refund"
	"github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
)

func TestCreate(t *testing.T) {
	refundSvc := serviceMocks.NewIRefundService(t)
	cfg := &config.Config{}
	logger := logger.NewSlogger(logger.Config{})
	controller := New(cfg, WithRefundService(refundSvc), WithLogger(logger))

	validRequest := &refundModel.CreateRefundRequest{
		ClientReferenceID: "REF001",
		PaymentSessionID:  uuid.NewString(),
		Reason:            constant.RefundReasonOthers,
		Method:            constant.RefundMethodAuto,
		IsFullAmount:      true,
	}

	tests := []struct {
		name          string
		merchantClaim *merchant.MerchantAuthTokenClaims
		setupMock     func()
		requestHeader map[string]string
		requestBody   func() []byte
		wantStatus    int
		wantResponse  string
	}{
		{
			name:         "ERROR: Merchant not found",
			wantStatus:   http.StatusUnauthorized,
			wantResponse: wrapErrOpenApiNonSnap(41, "merchant not found", "ERROR_UNAUTHORIZED"),
		},
		{
			name: "ERROR: Invalid Payload",
			merchantClaim: &merchant.MerchantAuthTokenClaims{
				MerchantId: "123456",
			},
			requestBody: func() []byte {
				return []byte(`{"missing-payload"}`)
			},
			wantStatus:   http.StatusBadRequest,
			wantResponse: `{"code":"field_format_invalid", "error":{"details":[{"field":"", "message":"invalid request payload"}], "traceId":"", "type":"API_ERROR"}, "message":"Format Field is invalid"}`,
		},
		{
			name: "ERROR: Field Format Invalid",
			merchantClaim: &merchant.MerchantAuthTokenClaims{
				MerchantId: "123456",
			},
			requestBody: func() []byte {
				return []byte(`{"paymentMethod": {"type": "invalid"}}`)
			},
			wantStatus:   http.StatusBadRequest,
			wantResponse: `{"code":"field_format_invalid", "error":{"details":[{"field":"", "message":"Key: 'CreateRefundRequest.ClientReferenceID' Error:Field validation for 'ClientReferenceID' failed on the 'required' tag\nKey: 'CreateRefundRequest.PaymentSessionID' Error:Field validation for 'PaymentSessionID' failed on the 'required' tag\nKey: 'CreateRefundRequest.Amount' Error:Field validation for 'Amount' failed on the 'required_if' tag\nKey: 'CreateRefundRequest.Reason' Error:Field validation for 'Reason' failed on the 'required' tag"}], "traceId":"", "type":"API_ERROR"}, "message":"Format Field is invalid"}`,
		},
		{
			name: "ERROR: Service returns error",
			merchantClaim: &merchant.MerchantAuthTokenClaims{
				MerchantId: "123456",
			},
			requestBody: func() []byte {
				reqBody, _ := json.Marshal(validRequest)
				return reqBody
			},
			setupMock: func() {
				refundSvc.On("Create", constant.ValueCtxMockType(), constant.PtrCreateRefundRequest()).
					Return(nil, errors.New("service error")).Once()
			},
			wantStatus:   http.StatusInternalServerError,
			wantResponse: `{"code":"general_error", "error":{"details":[{"field":"", "message":"Please contact our representative team"}], "traceId":"", "type":"API_ERROR"}, "message":"General error"}`,
		},
		{
			name: "SUCCESS",
			merchantClaim: &merchant.MerchantAuthTokenClaims{
				MerchantId: "123456",
			},
			requestBody: func() []byte {
				reqBody, _ := json.Marshal(validRequest)
				return reqBody
			},
			setupMock: func() {
				refundSvc.On("Create", constant.ValueCtxMockType(), constant.PtrCreateRefundRequest()).
					Return(nil, nil)
			},
			wantStatus:   http.StatusOK,
			wantResponse: `{"code":"00", "data":null, "message":"Success"}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.setupMock != nil {
				test.setupMock()
			}

			reqBody := []byte{}
			if test.requestBody != nil {
				reqBody = test.requestBody()
			}
			req := httptest.NewRequest(http.MethodPost, "/refunds", bytes.NewBuffer(reqBody))
			rec := httptest.NewRecorder()

			ctx := req.Context()
			if test.merchantClaim != nil {
				ctx = context.WithValue(ctx, constant.CtxMerchantInfo, test.merchantClaim)
			}
			req = req.WithContext(ctx)

			for key, value := range test.requestHeader {
				req.Header.Set(key, value)
			}

			router := chi.NewRouter()
			router.Post("/refunds", controller.Create)
			router.ServeHTTP(rec, req)

			assert.Equal(t, test.wantStatus, rec.Result().StatusCode)
			assert.JSONEq(t, test.wantResponse, rec.Body.String())
		})
	}
}
