package payment_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	c "github.com/paper-indonesia/pivot-backoffice/constant"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	. "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/payment"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetInvestigationList(t *testing.T) {
	service := serviceMocks.NewIPaymentService(t)
	handler := New(nil, validator.New(), nil, WithPaymentService(service))

	router := chi.NewRouter()
	router.Get("/cases", handler.GetInvestigationList)

	userTokenClaims := &user.UserTokenClaims{
		UUID:       uuid.NewString(),
		MerchantId: uuid.NewString(),
	}

	tests := []struct {
		name            string
		userTokenClaims *user.UserTokenClaims
		queryParams     string
		setupMock       func()
		wantStatusCode  int
	}{
		{
			name:           "ERROR: User not found",
			wantStatusCode: http.StatusUnauthorized,
		},
		{
			name:            "ERROR: Invalid fromDate format",
			userTokenClaims: userTokenClaims,
			queryParams:     "?fromDate=invalid",
			wantStatusCode:  http.StatusBadRequest,
		},
		{
			name:            "ERROR: Invalid page format",
			userTokenClaims: userTokenClaims,
			queryParams:     "?page=invalid",
			wantStatusCode:  http.StatusBadRequest,
		},
		{
			name:            "ERROR: Service error",
			userTokenClaims: userTokenClaims,
			queryParams:     "?page=1&limit=10",
			setupMock: func() {
				service.On("GetInvestigatedPayments", mock.Anything, mock.Anything).
					Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
			wantStatusCode: http.StatusInternalServerError,
		},
		{
			name:            "SUCCESS: Get investigation list",
			userTokenClaims: userTokenClaims,
			queryParams:     "?page=1&limit=10",
			setupMock: func() {
				service.On("GetInvestigatedPayments", mock.Anything, mock.Anything).
					Return(&commonModel.PaginationResponse{
						Data: []*paymentModel.InvestigatedPaymentResponse{
							{
								PaymentReferenceID:  "PAY-123",
								Amount:              decimal.NewFromInt(100000),
								Currency:            "IDR",
								MerchantName:        "Test Merchant",
								PaymentMethod:       "QRIS",
								InvestigationStatus: "INVESTIGATION_IN_PROCESS",
							},
						},
						Meta: commonModel.Meta{
							Page:       1,
							PerPage:    10,
							TotalItems: 1,
							TotalPages: 1,
						},
					}, nil)
			},
			wantStatusCode: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupMock != nil {
				tt.setupMock()
			}

			req := httptest.NewRequest(http.MethodGet, "/cases"+tt.queryParams, nil)

			if tt.userTokenClaims != nil {
				ctx := context.WithValue(req.Context(), c.CtxUserInfoKey, tt.userTokenClaims)
				req = req.WithContext(ctx)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatusCode, w.Code)

			service.AssertExpectations(t)
		})
	}
}
