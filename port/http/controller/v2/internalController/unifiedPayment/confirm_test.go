package v2InternalUnifiedPaymentController_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	unifiedPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/unifiedPayment"
	loggerMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pdk/logger"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	pkgErr "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	. "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v2/internalController/unifiedPayment"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestConfirm(t *testing.T) {
	unifiedPaymentSvc := serviceMocks.NewIUnifiedPaymentService(t)
	loggerMock := loggerMocks.NewILogger(t)
	loggerMock.On("Warn", mock.Anything, mock.Anything, mock.Anything).Maybe()
	cfg := &config.Config{
		UnifiedPaymentConfig: config.UnifiedPaymentConfig{
			VirtualAccountConfig: &config.UnifiedPaymentVirtualAccountConfig{},
		},
	}
	controller := New(cfg, nil, WithUnifiedPaymentService(unifiedPaymentSvc), WithLogger(loggerMock))

	validRequest := &unifiedPaymentModel.ConfirmUnifiedPaymentSessionRequest{}

	tests := []struct {
		name          string
		paymentID     string
		merchantClaim *merchant.MerchantAuthTokenClaims
		setupMock     func()
		requestHeader map[string]string
		requestBody   func() []byte
		wantStatus    int
		wantResponse  string
	}{
		{
			name:         "ERROR: Merchant not found",
			paymentID:    uuid.NewString(),
			wantStatus:   http.StatusUnauthorized,
			wantResponse: `{"code":"merchant_not_found","message":"Merchant not found","error":{"type":"API_ERROR","details":[{"field":"","message":"Invalid Merchant request"}],"traceId":""}}`,
		},
		{
			name:      "ERROR: Invalid UUID",
			paymentID: "invalid-uuid-format",
			merchantClaim: &merchant.MerchantAuthTokenClaims{
				MerchantId: "123456",
			},
			requestBody: func() []byte {
				return []byte(`{}`)
			},
			wantStatus:   http.StatusBadRequest,
			wantResponse: `{"code":"field_format_invalid","message":"Format Field is invalid","error":{"type":"API_ERROR","details":[{"field":"","message":"invalid request payload"}],"traceId":""}}`,
		},
		{
			name:      "ERROR: Invalid Payload",
			paymentID: uuid.NewString(),
			merchantClaim: &merchant.MerchantAuthTokenClaims{
				MerchantId: "123456",
			},
			requestBody: func() []byte {
				return []byte(`{"missing-payload"}`)
			},
			wantStatus:   http.StatusBadRequest,
			wantResponse: `{"code":"field_format_invalid","message":"Format Field is invalid","error":{"type":"API_ERROR","details":[{"field":"","message":"invalid request payload"}],"traceId":""}}`,
		},
		{
			name:      "ERROR: Field Format Invalid",
			paymentID: uuid.NewString(),
			merchantClaim: &merchant.MerchantAuthTokenClaims{
				MerchantId: "123456",
			},
			requestBody: func() []byte {
				return []byte(`{"paymentMethod": {"type": "invalid"}}`)
			},
			wantStatus:   http.StatusBadRequest,
			wantResponse: `{"code":"field_format_invalid","message":"Format Field is invalid","error":{"type":"API_ERROR","details":[{"field":"","message":"Key: 'ConfirmUnifiedPaymentSessionRequest.PaymentMethod.Type' Error:Field validation for 'Type' failed on the 'oneof' tag"}],"traceId":""}}`,
		},
		{
			name:      "ERROR: Payment session detail not found",
			paymentID: uuid.NewString(),
			merchantClaim: &merchant.MerchantAuthTokenClaims{
				MerchantId: "123456", // NOSONAR
			},
			setupMock: func() {
				unifiedPaymentSvc.On("GetSessionDetail", mock.Anything, mock.Anything).Once().Return(nil, pkgErr.New(response.HttpErrUnprocessableContent, constant.ErrPaymentNotFound))
			},
			requestBody: func() []byte {
				request := &unifiedPaymentModel.ConfirmUnifiedPaymentSessionRequest{
					PaymentMethod: &unifiedPaymentModel.PaymentMethod{
						Type: constant.UnifiedPaymentMethodVA,
					},
				}
				raw, _ := json.Marshal(request)
				return raw
			},
			wantStatus:   http.StatusUnprocessableEntity,
			wantResponse: `{"code":"unprocessable_entity","message":"Unprocessable entity","error":{"type":"API_ERROR","details":[{"field":"","message":"payment not found"}],"traceId":""}}`,
		},
		{
			name:      "ERROR: Empty VA payload",
			paymentID: uuid.NewString(),
			merchantClaim: &merchant.MerchantAuthTokenClaims{
				MerchantId: "123456",
			},
			setupMock: func() {
				unifiedPaymentSvc.On("GetSessionDetail", mock.Anything, mock.Anything).
					Return(&unifiedPaymentModel.UnifiedPaymentSessionResponse{
						Amount: unifiedPaymentModel.Amount{
							Currency: "IDR",
							Value:    10000,
						},
					}, nil).Once()
			},
			requestBody: func() []byte {
				request := &unifiedPaymentModel.ConfirmUnifiedPaymentSessionRequest{
					PaymentMethod: &unifiedPaymentModel.PaymentMethod{
						Type: constant.UnifiedPaymentMethodVA,
					},
				}
				reqBody, _ := json.Marshal(request)
				return reqBody
			},
			wantStatus:   http.StatusUnprocessableEntity,
			wantResponse: `{"code":"unprocessable_entity", "error":{"details":[{"field":"", "message":"payment method options for virtual account can not be empty"}], "traceId":"", "type":"API_ERROR"}, "message":"Unprocessable entity"}`,
		},
		{
			name:      "ERROR: Service returns error",
			paymentID: uuid.NewString(),
			merchantClaim: &merchant.MerchantAuthTokenClaims{
				MerchantId: "123456",
			},
			requestBody: func() []byte {
				reqBody, _ := json.Marshal(validRequest)
				return reqBody
			},
			setupMock: func() {
				unifiedPaymentSvc.On("GetSessionDetail", mock.Anything, mock.Anything).
					Return(&unifiedPaymentModel.UnifiedPaymentSessionResponse{
						Amount: unifiedPaymentModel.Amount{
							Currency: "IDR",
							Value:    10000,
						},
					}, nil).Once()
				unifiedPaymentSvc.On("ConfirmSession", mock.Anything, mock.Anything).
					Return(nil, errors.New("service error")).Once()
			},
			wantStatus:   http.StatusInternalServerError,
			wantResponse: `{"code":"general_error", "error":{"details":[{"field":"", "message":"Please contact our representative team"}], "traceId":"", "type":"API_ERROR"}, "message":"General error"}`,
		},
		{
			name:      "SUCCESS",
			paymentID: uuid.NewString(),
			merchantClaim: &merchant.MerchantAuthTokenClaims{
				MerchantId: "123456",
			},
			requestBody: func() []byte {
				reqBody, _ := json.Marshal(validRequest)
				return reqBody
			},
			setupMock: func() {
				unifiedPaymentSvc.On("GetSessionDetail", mock.Anything, mock.Anything).
					Return(&unifiedPaymentModel.UnifiedPaymentSessionResponse{
						Amount: unifiedPaymentModel.Amount{
							Currency: "IDR",
							Value:    10000,
						},
					}, nil).Once()
				unifiedPaymentSvc.On("ConfirmSession", mock.Anything, mock.Anything).
					Return(nil, nil).Once()
			},
			wantStatus:   http.StatusOK,
			wantResponse: `{"code":"00", "data":null, "message":"Success"}`,
		},
		{
			name:      "SUCCESS: With sub-merchant ID",
			paymentID: uuid.NewString(),
			merchantClaim: &merchant.MerchantAuthTokenClaims{
				MerchantId: "parent-merchant-123",
			},
			requestHeader: map[string]string{
				constant.HeaderXSubMerchantID: "sub-merchant-456",
			},
			requestBody: func() []byte {
				reqBody, _ := json.Marshal(validRequest)
				return reqBody
			},
			setupMock: func() {
				unifiedPaymentSvc.On("GetSessionDetail", mock.Anything, mock.Anything).
					Return(&unifiedPaymentModel.UnifiedPaymentSessionResponse{
						Amount: unifiedPaymentModel.Amount{
							Currency: "IDR",
							Value:    10000,
						},
						RecurringID: "4d5fe827-0063-44de-97e8-7baa35022990",
					}, nil).Once()
				unifiedPaymentSvc.On("ConfirmSession", mock.Anything, mock.Anything).Return(nil, nil).Once()
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
			req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/payments/%s/confirm", test.paymentID), bytes.NewBuffer(reqBody))
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
			router.Post("/payments/{uuid}/confirm", controller.Confirm)
			router.ServeHTTP(rec, req)

			assert.Equal(t, test.wantStatus, rec.Result().StatusCode)
			if !assert.JSONEq(t, test.wantResponse, rec.Body.String()) {
				t.Log("Result:", rec.Body.String())
			}
		})
	}
}

func TestConfirmValidation(t *testing.T) {
	minAmount := float64(10000)
	maxAmount := float64(1000000)

	unifiedPaymentSvc := serviceMocks.NewIUnifiedPaymentService(t)
	loggerMock := loggerMocks.NewILogger(t)
	loggerMock.On("Warn", mock.Anything, mock.Anything, mock.Anything).Maybe()
	cfg := &config.Config{
		UnifiedPaymentConfig: config.UnifiedPaymentConfig{
			VirtualAccountConfig: &config.UnifiedPaymentVirtualAccountConfig{
				MinAmount: &minAmount,
				MaxAmount: &maxAmount,
			},
		},
	}
	controller := New(cfg, nil, WithUnifiedPaymentService(unifiedPaymentSvc), WithLogger(loggerMock))

	tests := []struct {
		name          string
		paymentID     string
		merchantClaim *merchant.MerchantAuthTokenClaims
		setupMock     func()
		requestBody   func() []byte
		requestHeader map[string]string
		wantStatus    int
		wantResponse  string
	}{
		{
			name:      "ERROR: Payment type MULTIPLE not allowed for EWALLET",
			paymentID: uuid.NewString(),
			merchantClaim: &merchant.MerchantAuthTokenClaims{
				MerchantId: "123456",
			},
			setupMock: func() {
				unifiedPaymentSvc.On("GetSessionDetail", mock.Anything, mock.Anything).
					Return(&unifiedPaymentModel.UnifiedPaymentSessionResponse{
						Amount: unifiedPaymentModel.Amount{
							Currency: "IDR",
							Value:    50000,
						},
						PaymentType: constant.UnifiedPaymentTypeMultiple,
					}, nil).Once()
			},
			requestBody: func() []byte {
				request := &unifiedPaymentModel.ConfirmUnifiedPaymentSessionRequest{
					PaymentMethod: &unifiedPaymentModel.PaymentMethod{
						Type: constant.UnifiedPaymentMethodEWallet,
					},
				}
				reqBody, _ := json.Marshal(request)
				return reqBody
			},
			wantStatus:   http.StatusBadRequest,
			wantResponse: `{"code":"field_format_invalid","message":"Format Field is invalid","error":{"type":"API_ERROR","details":[{"field":"","message":"paymentType 'MULTIPLE' is not allowed for this payment method type"}],"traceId":""}}`,
		},
		{
			name:      "ERROR: VA amount is zero",
			paymentID: uuid.NewString(),
			merchantClaim: &merchant.MerchantAuthTokenClaims{
				MerchantId: "123456",
			},
			setupMock: func() {
				unifiedPaymentSvc.On("GetSessionDetail", mock.Anything, mock.Anything).
					Return(&unifiedPaymentModel.UnifiedPaymentSessionResponse{
						Amount: unifiedPaymentModel.Amount{
							Currency: "IDR",
							Value:    0,
						},
					}, nil).Once()
			},
			requestBody: func() []byte {
				request := &unifiedPaymentModel.ConfirmUnifiedPaymentSessionRequest{
					PaymentMethod: &unifiedPaymentModel.PaymentMethod{
						Type: constant.UnifiedPaymentMethodVA,
					},
					PaymentMethodOptions: &unifiedPaymentModel.PaymentMethodOptions{
						VirtualAccount: &unifiedPaymentModel.PaymentMethodOptionVirtualAccount{
							Channel: "BCA",
						},
					},
				}
				reqBody, _ := json.Marshal(request)
				return reqBody
			},
			wantStatus:   http.StatusBadRequest,
			wantResponse: `{"code":"field_format_invalid","message":"Format Field is invalid","error":{"type":"API_ERROR","details":[{"field":"","message":"amount is required"}],"traceId":""}}`,
		},
		{
			name:      "ERROR: VA amount below minimum",
			paymentID: uuid.NewString(),
			merchantClaim: &merchant.MerchantAuthTokenClaims{
				MerchantId: "123456",
			},
			setupMock: func() {
				unifiedPaymentSvc.On("GetSessionDetail", mock.Anything, mock.Anything).
					Return(&unifiedPaymentModel.UnifiedPaymentSessionResponse{
						Amount: unifiedPaymentModel.Amount{
							Currency: "IDR",
							Value:    5000,
						},
					}, nil).Once()
			},
			requestBody: func() []byte {
				request := &unifiedPaymentModel.ConfirmUnifiedPaymentSessionRequest{
					PaymentMethod: &unifiedPaymentModel.PaymentMethod{
						Type: constant.UnifiedPaymentMethodVA,
					},
					PaymentMethodOptions: &unifiedPaymentModel.PaymentMethodOptions{
						VirtualAccount: &unifiedPaymentModel.PaymentMethodOptionVirtualAccount{
							Channel: "BCA",
						},
					},
				}
				reqBody, _ := json.Marshal(request)
				return reqBody
			},
			wantStatus:   http.StatusBadRequest,
			wantResponse: `{"code":"field_format_invalid","message":"Format Field is invalid","error":{"type":"API_ERROR","details":[{"field":"","message":"amount value below the minimum"}],"traceId":""}}`,
		},
		{
			name:      "ERROR: VA amount above maximum",
			paymentID: uuid.NewString(),
			merchantClaim: &merchant.MerchantAuthTokenClaims{
				MerchantId: "123456",
			},
			setupMock: func() {
				unifiedPaymentSvc.On("GetSessionDetail", mock.Anything, mock.Anything).
					Return(&unifiedPaymentModel.UnifiedPaymentSessionResponse{
						Amount: unifiedPaymentModel.Amount{
							Currency: "IDR",
							Value:    2000000,
						},
					}, nil).Once()
			},
			requestBody: func() []byte {
				request := &unifiedPaymentModel.ConfirmUnifiedPaymentSessionRequest{
					PaymentMethod: &unifiedPaymentModel.PaymentMethod{
						Type: constant.UnifiedPaymentMethodVA,
					},
					PaymentMethodOptions: &unifiedPaymentModel.PaymentMethodOptions{
						VirtualAccount: &unifiedPaymentModel.PaymentMethodOptionVirtualAccount{
							Channel: "BCA",
						},
					},
				}
				reqBody, _ := json.Marshal(request)
				return reqBody
			},
			wantStatus:   http.StatusBadRequest,
			wantResponse: `{"code":"field_format_invalid","message":"Format Field is invalid","error":{"type":"API_ERROR","details":[{"field":"","message":"amount value above the maximum"}],"traceId":""}}`,
		},
		{
			name:      "SUCCESS: VA with valid amount",
			paymentID: uuid.NewString(),
			merchantClaim: &merchant.MerchantAuthTokenClaims{
				MerchantId: "123456",
			},
			setupMock: func() {
				unifiedPaymentSvc.On("GetSessionDetail", mock.Anything, mock.Anything).
					Return(&unifiedPaymentModel.UnifiedPaymentSessionResponse{
						Amount: unifiedPaymentModel.Amount{
							Currency: "IDR",
							Value:    50000,
						},
					}, nil).Once()
				unifiedPaymentSvc.On("ConfirmSession", mock.Anything, mock.Anything).
					Return(nil, nil).Once()
			},
			requestBody: func() []byte {
				request := &unifiedPaymentModel.ConfirmUnifiedPaymentSessionRequest{
					PaymentMethod: &unifiedPaymentModel.PaymentMethod{
						Type: constant.UnifiedPaymentMethodVA,
					},
					PaymentMethodOptions: &unifiedPaymentModel.PaymentMethodOptions{
						VirtualAccount: &unifiedPaymentModel.PaymentMethodOptionVirtualAccount{
							Channel: "BCA",
						},
					},
				}
				reqBody, _ := json.Marshal(request)
				return reqBody
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
			req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/payments/%s/confirm", test.paymentID), bytes.NewBuffer(reqBody))
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
			router.Post("/payments/{uuid}/confirm", controller.Confirm)
			router.ServeHTTP(rec, req)

			assert.Equal(t, test.wantStatus, rec.Result().StatusCode)
			if !assert.JSONEq(t, test.wantResponse, rec.Body.String()) {
				t.Log("Result:", rec.Body.String())
			}
		})
	}
}
