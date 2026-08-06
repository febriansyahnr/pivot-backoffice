package internalPaymentController

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	paymentConstant "github.com/paper-indonesia/pivot-backoffice/constant/payment"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	mockRabbitMq "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/rabbitmqExt"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/validatorExt"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreateQr(t *testing.T) {
	validMerchantClaim := &merchant.MerchantAuthTokenClaims{
		MerchantId: uuid.NewString(),
	}

	testCases := []struct {
		name           string
		mockSetup      func(mockService *serviceMocks.IPaymentService, rabbitMqMock *mockRabbitMq.RabbitMQExt)
		setupBody      func(*testing.T) []byte
		expectedStatus int
		merchantClaim  *merchant.MerchantAuthTokenClaims
		setHeaders     func(req *http.Request)
	}{
		{
			name: "SUCCESS: Create payments QR static",
			mockSetup: func(mockService *serviceMocks.IPaymentService, rabbitMqMock *mockRabbitMq.RabbitMQExt) {
				mockService.On(
					"CreatePayment",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("paymentModel.PaymentRequest"),
				).Return(&paymentModel.PaymentResponse{}, nil)

				rabbitMqMock.On(
					"PublishActivity",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*string"),
					mock.AnythingOfType("*string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.Anything,
				).Return(nil)
			},

			setupBody: func(t *testing.T) []byte {
				payload := paymentModel.PaymentRequest{
					ReferenceID:   uuid.NewString(),
					PaymentMethod: paymentConstant.PAYMENT_METHOD_QRIS,
					Qris: &paymentModel.PaymentMetadataQris{
						QrType:        constant.QrTypeStatic,
						QrMethodType:  constant.QrMethodTypeMPM,
						SubMerchantId: "000510000928",
					},
				}
				payloadRequestByte, err := json.Marshal(payload)
				assert.NoError(t, err)
				return payloadRequestByte
			},
			expectedStatus: http.StatusOK,
			merchantClaim:  validMerchantClaim,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockService := serviceMocks.NewIPaymentService(t)
			mockMerchant := serviceMocks.NewIMerchantService(t)
			mockValidator := validatorExt.New()
			mockRmq := mockRabbitMq.NewRabbitMQExt(t)

			tc.mockSetup(mockService, mockRmq)
			paymentController := New(mockValidator, mockService, mockMerchant, mockRmq)

			baseUrl := "/api/internal/v1/payments/create"
			chiRouterCtx := chi.NewRouteContext()
			req := httptest.NewRequest(http.MethodPost, baseUrl, bytes.NewBuffer(tc.setupBody(t)))
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiRouterCtx))
			if tc.merchantClaim != nil {
				req = req.WithContext(context.WithValue(req.Context(), constant.CtxMerchantInfo, tc.merchantClaim))
			}

			if tc.setHeaders != nil {
				tc.setHeaders(req)
			}

			httpRecorder := httptest.NewRecorder()
			handler := http.HandlerFunc(paymentController.Create)
			handler.ServeHTTP(httpRecorder, req)

			if httpRecorder.Body.String() != "" {
				t.Logf("response: %s", httpRecorder.Body.String())
			}

			assert.Equal(t, tc.expectedStatus, httpRecorder.Code)
			mockService.AssertExpectations(t)
		})
	}
}

func TestValidateQrisPayload(t *testing.T) {
	validPayload := paymentModel.PaymentRequest{
		ReferenceID:   uuid.NewString(),
		PaymentMethod: paymentConstant.PAYMENT_METHOD_QRIS,
		Qris: &paymentModel.PaymentMetadataQris{
			QrType:        constant.QrTypeStatic,
			QrMethodType:  constant.QrMethodTypeMPM,
			SubMerchantId: "000510000928",
		},
	}

	testCases := []struct {
		name      string
		payload   func() paymentModel.PaymentRequest
		wantError bool
	}{
		{
			name: "Nil qris",
			payload: func() paymentModel.PaymentRequest {
				nilQris := validPayload
				nilQris.Qris = nil
				return nilQris
			},
			wantError: true,
		},
		{
			name: "invalid qris struct",
			payload: func() paymentModel.PaymentRequest {
				qrisPayload := validPayload
				qrisPayload.Qris = &paymentModel.PaymentMetadataQris{}
				return qrisPayload
			},
			wantError: true,
		},
		{
			name: "invalid qris type",
			payload: func() paymentModel.PaymentRequest {
				qrisPayload := validPayload
				qrisPayload.Qris.QrType = "INVALID"
				return qrisPayload
			},
			wantError: true,
		},
		{
			name: "invalid amount qris static",
			payload: func() paymentModel.PaymentRequest {
				qrisPayload := validPayload
				qrisPayload.Qris.QrType = "STATIC"
				qrisPayload.Qris.Amount = &paymentModel.Amount{
					Value:    decimal.New(1000, 0),
					Currency: "IDR",
				}
				return qrisPayload
			},
			wantError: true,
		},
		{
			name: "invalid validity period qris static",
			payload: func() paymentModel.PaymentRequest {
				qrisPayload := validPayload
				qrisPayload.Qris.QrType = "STATIC"
				qrisPayload.Qris.Amount = nil
				qrisPayload.Qris.ValidityPeriod = 100

				return qrisPayload
			},
			wantError: true,
		},
		{
			name: "invalid amount qris dynamic",
			payload: func() paymentModel.PaymentRequest {
				qrisPayload := validPayload
				qrisPayload.Qris.QrType = "DYNAMIC"
				qrisPayload.Qris.Amount = nil

				return qrisPayload
			},
			wantError: true,
		},
		{
			name: "invalid validity period qris dynamic",
			payload: func() paymentModel.PaymentRequest {
				qrisPayload := validPayload
				qrisPayload.Qris.QrType = "DYNAMIC"
				qrisPayload.Qris.Amount = &paymentModel.Amount{
					Value:    decimal.New(1000, 0),
					Currency: "IDR",
				}
				qrisPayload.Qris.ValidityPeriod = 28001

				return qrisPayload
			},
			wantError: true,
		},
		{
			name: "valid qris dynamic without fee amount",
			payload: func() paymentModel.PaymentRequest {
				qrisPayload := validPayload
				qrisPayload.Qris.QrType = "DYNAMIC"
				qrisPayload.Qris.Amount = &paymentModel.Amount{
					Value:    decimal.New(1000, 0),
					Currency: "IDR",
				}
				qrisPayload.Qris.ValidityPeriod = 10800

				return qrisPayload
			},
			wantError: false,
		},
	}

	controller := InternalPaymentController{
		validate: validatorExt.New(),
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := controller.validateQrisPayload(tc.payload())
			if tc.wantError {
				assert.Error(t, err)
				return
			}
		})
	}
}
