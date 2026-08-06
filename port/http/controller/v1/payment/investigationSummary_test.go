package payment_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	c "github.com/paper-indonesia/pivot-backoffice/constant"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	s "github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	. "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/payment"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetInvestigationSummary(t *testing.T) {
	service := serviceMocks.NewIPaymentService(t)
	handler := New(nil, validator.New(), nil, WithPaymentService(service))

	router := chi.NewRouter()
	router.Get("/cases/summary", handler.GetInvestigationSummary)

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
		wantRespBody    string
	}{
		{
			name:           "ERROR: User not found",
			wantStatusCode: http.StatusUnauthorized,
			wantRespBody:   c.WrapErrApiRespForTest(41, s.ErrTypeAPI, "user not found"),
		},
		{
			name:            "ERROR: Missing startDate",
			userTokenClaims: userTokenClaims,
			queryParams:     "?endDate=2024-01-31T23:59:59Z",
			wantStatusCode:  http.StatusBadRequest,
		},
		{
			name:            "ERROR: Missing endDate",
			userTokenClaims: userTokenClaims,
			queryParams:     "?startDate=2024-01-01T00:00:00Z",
			wantStatusCode:  http.StatusBadRequest,
		},
		{
			name:            "ERROR: Invalid startDate format",
			userTokenClaims: userTokenClaims,
			queryParams:     "?startDate=invalid&endDate=2024-01-31T23:59:59Z",
			wantStatusCode:  http.StatusBadRequest,
		},
		{
			name:            "ERROR: Invalid endDate format",
			userTokenClaims: userTokenClaims,
			queryParams:     "?startDate=2024-01-01T00:00:00Z&endDate=invalid",
			wantStatusCode:  http.StatusBadRequest,
		},
		{
			name:            "ERROR: Service error",
			userTokenClaims: userTokenClaims,
			queryParams:     "?startDate=2026-02-01T00:00:00Z&endDate=2026-02-13T23:59:59Z",
			setupMock: func() {
				service.On("GetInvestigationSummary", mock.Anything, mock.Anything).
					Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
			wantStatusCode: http.StatusInternalServerError,
		},
		{
			name:            "SUCCESS: Get investigation summary",
			userTokenClaims: userTokenClaims,
			queryParams:     "?startDate=2026-02-01T00:00:00Z&endDate=2026-02-13T23:59:59Z",
			setupMock: func() {
				service.On("GetInvestigationSummary", mock.Anything, mock.Anything).
					Return(&paymentModel.InvestigationSummaryResponse{
						OnInvestigation: paymentModel.InvestigationSummaryItem{
							TotalAmount: "5000000.00",
							Currency:    "IDR",
						},
						Success: paymentModel.InvestigationSummaryItem{
							TotalAmount: "10000000.00",
							Currency:    "IDR",
						},
						Failed: paymentModel.InvestigationSummaryItem{
							TotalAmount: "2000000.00",
							Currency:    "IDR",
						},
					}, nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code":"00","message":"OK","data":{"onInvestigation":{"totalAmount":"5000000.00","currency":"IDR"},"success":{"totalAmount":"10000000.00","currency":"IDR"},"failed":{"totalAmount":"2000000.00","currency":"IDR"}}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupMock != nil {
				tt.setupMock()
			}

			req := httptest.NewRequest(http.MethodGet, "/cases/summary"+tt.queryParams, nil)

			if tt.userTokenClaims != nil {
				ctx := context.WithValue(req.Context(), c.CtxUserInfoKey, tt.userTokenClaims)
				req = req.WithContext(ctx)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatusCode, w.Code)
			if tt.wantRespBody != "" {
				assert.JSONEq(t, tt.wantRespBody, w.Body.String())
			}

			service.AssertExpectations(t)
		})
	}
}
