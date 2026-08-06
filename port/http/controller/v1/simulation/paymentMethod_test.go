package simulationController

import (
	paymentConstant "github.com/paper-indonesia/pivot-backoffice/constant/payment"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/validatorExt"
)

func TestGetPaymentMethodForPayment(t *testing.T) {
	paymentSvc := serviceMocks.NewIPaymentService(t)
	paymentMethodSvc := serviceMocks.NewIPaymentMethodService(t)
	handler := New(validatorExt.New(), WithPaymentService(paymentSvc), WithPaymentMethodService(paymentMethodSvc))

	testCases := []struct {
		name           string
		setupMock      func()
		wantStatusCode int
		wantRespBody   string
	}{
		{
			name: "ERROR: FindPaymentForSimulationByID service error",
			setupMock: func() {
				paymentMethodSvc.On("FindPaymentMethodByCategory", constant.ValueCtxMockType(), constant.StringMockType()).
					Once().Return(nil, constant.ErrSomeErrorForUnitTest)
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   `{"code":"99","data":null,"error":{"details":[],"traceId":"","type":"UNKNOWN"},"message":"some error"}`,
		},
		{
			name: "SUCCESS",
			setupMock: func() {
				paymentMethodSvc.On("FindPaymentMethodByCategory", constant.ValueCtxMockType(), constant.StringMockType()).
					Once().Return([]*paymentModel.PaymentMethod{
					{
						Type: paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
					},
					{
						Type: paymentConstant.PAYMENT_METHOD_BANK_TRANSFER,
					},
					{
						Type: paymentConstant.PAYMENT_METHOD_CREDIT_CARD,
					},
					{
						Type: paymentConstant.PAYMENT_METHOD_QRIS,
					},
				}, nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody: `{"code":"00","data":{
				"bankTransfer":[{"acquirer":"", "bankName":null, "category":"", "createdAt":"0001-01-01T00:00:00Z", "description":null, "logo":null, "name":"", "type":"BANK_TRANSFER", "updatedAt":"0001-01-01T00:00:00Z", "uuid":""}],
				"creditCard":[{"acquirer":"", "bankName":null, "category":"", "createdAt":"0001-01-01T00:00:00Z", "description":null, "logo":null, "name":"", "type":"CREDIT_CARD", "updatedAt":"0001-01-01T00:00:00Z", "uuid":""}],
				"qris":[{"acquirer":"", "bankName":null, "category":"", "createdAt":"0001-01-01T00:00:00Z", "description":null, "logo":null, "name":"", "type":"QRIS", "updatedAt":"0001-01-01T00:00:00Z", "uuid":""}],
				"virtualAccount":[{"acquirer":"", "bankName":null, "category":"", "createdAt":"0001-01-01T00:00:00Z", "description":null, "logo":null, "name":"", "type":"VIRTUAL_ACCOUNT", "updatedAt":"0001-01-01T00:00:00Z", "uuid":""}]
			},"message":"OK"}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/simulations/payment-methods/payment", nil)

			router := chi.NewRouter()
			router.Get("/simulations/payment-methods/payment", handler.GetPaymentMethodForPayment)
			tc.setupMock()

			router.ServeHTTP(rec, req)
			require.Equal(t, tc.wantStatusCode, rec.Result().StatusCode)
			assert.JSONEq(t, tc.wantRespBody, rec.Body.String())
		})

	}
}
