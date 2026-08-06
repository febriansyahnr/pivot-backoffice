package paymentMethodController

import (
	"context"
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	paymentConstant "github.com/paper-indonesia/pivot-backoffice/constant/payment"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	mocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestFindPaymentMethodByCategory(t *testing.T) {
	paymentMethods := []*paymentModel.PaymentMethod{
		{
			UUID:      uuid.NewString(),
			Type:      paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
			Category:  paymentConstant.PAYMENT_METHOD_CATEGORY_DISBURSEMENT_TOPUP,
			Name:      "VA Permata",
			Acquirer:  constant.BANK_ACQUIRER_PERMATA,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			UUID:      uuid.NewString(),
			Type:      paymentConstant.PAYMENT_METHOD_QRIS,
			Category:  paymentConstant.PAYMENT_METHOD_CATEGORY_DISBURSEMENT_TOPUP,
			Name:      "QRIS",
			Acquirer:  "bnc",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			UUID:      uuid.NewString(),
			Type:      paymentConstant.PAYMENT_METHOD_BANK_TRANSFER,
			Category:  paymentConstant.PAYMENT_METHOD_CATEGORY_DISBURSEMENT_TOPUP,
			Name:      "BT Permata",
			Acquirer:  constant.BANK_ACQUIRER_PERMATA,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			UUID:      uuid.NewString(),
			Type:      paymentConstant.PAYMENT_METHOD_CREDIT_CARD,
			Category:  paymentConstant.PAYMENT_METHOD_CATEGORY_DISBURSEMENT_TOPUP,
			Name:      "CC Permata",
			Acquirer:  constant.BANK_ACQUIRER_PERMATA,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	category := paymentConstant.PAYMENT_METHOD_CATEGORY_DISBURSEMENT_TOPUP

	testCase := []struct {
		name           string
		category       string
		mockSetup      func(paymentMethodSvc *mocks.IPaymentMethodService)
		expectedStatus int
	}{
		{
			name:     "SUCCESS",
			category: category,
			mockSetup: func(paymentMethodSvc *mocks.IPaymentMethodService) {
				paymentMethodSvc.On("FindPaymentMethodByCategory",
					mock.Anything,
					mock.AnythingOfType("string"),
				).Return(paymentMethods, nil)
			},
			expectedStatus: 200,
		},
		{
			name:     "ERROR: Service Error",
			category: category,
			mockSetup: func(paymentMethodSvc *mocks.IPaymentMethodService) {
				paymentMethodSvc.On("FindPaymentMethodByCategory",
					mock.Anything,
					mock.AnythingOfType("string"),
				).Return(nil, errors.New("service error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:     "ERROR: Category is required",
			category: "",
			mockSetup: func(paymentMethodSvc *mocks.IPaymentMethodService) {
				// empty mock setup
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range testCase {
		t.Run(tt.name, func(t *testing.T) {
			paymentMethodSvc := mocks.NewIPaymentMethodService(t)

			tt.mockSetup(paymentMethodSvc)

			svc := New(paymentMethodSvc)

			// Create the HTTP request for the test case
			req := httptest.NewRequest(http.MethodGet, "/internal/v1/payment-methods/", nil)

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("category", tt.category)

			rr := httptest.NewRecorder()

			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			// Create the handler and serve the request
			handler := http.HandlerFunc(svc.FindPaymentMethodByCategory)
			handler.ServeHTTP(rr, req)

			if rr.Body.String() != "" {
				t.Logf("Handler response body: %s", rr.Body.String())
			}

			// Assertions
			assert.Equal(t, tt.expectedStatus, rr.Code)
			paymentMethodSvc.AssertExpectations(t)
		})
	}
}
