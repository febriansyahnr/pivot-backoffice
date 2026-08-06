package payment

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	mockService "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestFilterPaymentHistory(t *testing.T) {
	var (
		validUserClaims = &user.UserTokenClaims{
			UUID:       uuid.NewString(),
			MerchantId: uuid.NewString(),
		}
		mockPaymentService mockService.IPaymentService
		paymentController  = PaymentController{
			config: &config.Config{
				ServiceName: "unit-test",
			},
			paymentService: &mockPaymentService,
		}
	)

	testCases := []struct {
		name           string
		params         string
		callMock       func()
		expectedStatus int
		userClaim      *user.UserTokenClaims
	}{
		{
			name:           "when request does't have auth header, then should return 401",
			callMock:       func() {},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "when user request data with invalid param, then should return error ",
			callMock:       func() {},
			params:         "page=1&perPage=invalid_type",
			userClaim:      validUserClaims,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "when the user enters an invalid date range",
			callMock:       func() {},
			params:         "page=1&perPage=10&startDate=2025-01-01T17:00:00Z&endDate=2025-01-31T16:59:59",
			userClaim:      validUserClaims,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "when failed to get the payments data",
			callMock: func() {
				mockPaymentService.On("FilterPaymentHistory", mock.Anything, paymentModel.FilterPaymentHistoryOption{
					MerchantID: validUserClaims.MerchantId,
					Sort:       "ASC",
					SortBy:     "createdAt",
					Page:       1,
					PerPage:    10,
				}).Return(nil, errors.New("invalid request")).Once()
			},
			userClaim:      validUserClaims,
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name: "when get the payments data",
			callMock: func() {
				mockPaymentService.On("FilterPaymentHistory", mock.Anything, paymentModel.FilterPaymentHistoryOption{
					MerchantID: validUserClaims.MerchantId,
					Sort:       "ASC",
					SortBy:     "createdAt",
					Page:       1,
					PerPage:    10,
				}).Return(&commonModel.PaginationResponse{
					Data: []paymentModel.PaymentHistoryItem{},
					Meta: commonModel.Meta{
						Page:       1,
						PerPage:    10,
						TotalItems: 0,
					},
				}, nil).Once()
			},
			userClaim:      validUserClaims,
			expectedStatus: http.StatusOK,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.callMock()

			req := httptest.NewRequest(http.MethodGet, "/api/v1/payments?"+tc.params, nil)
			if tc.userClaim != nil {
				req = req.WithContext(context.WithValue(req.Context(), constant.CtxUserInfoKey, tc.userClaim))
			}
			rr := httptest.NewRecorder()

			handler := http.HandlerFunc(paymentController.FilterPaymentHistory)
			handler.ServeHTTP(rr, req)

			if rr.Body.String() != "" {
				t.Logf("Handler response body: %s", rr.Body.String())
			}

			assert.Equal(t, tc.expectedStatus, rr.Code)
		})
	}
}

func TestParseFilterParam(t *testing.T) {
	tests := []struct {
		name        string
		queryParams string
		expectedOpt paymentModel.FilterPaymentHistoryOption
		expectedErr bool
	}{
		{
			name:        "default params",
			queryParams: "",
			expectedOpt: paymentModel.FilterPaymentHistoryOption{
				Page:          1,
				PerPage:       10,
				Sort:          "ASC",
				SortBy:        "createdAt",
				StartDate:     time.Time{},
				EndDate:       time.Time{},
				Status:        "",
				ReferenceID:   "",
				PaymentMethod: "",
			},
			expectedErr: false,
		},
		{
			name:        "valid sort param",
			queryParams: "sortBy=updatedAt&sort=ASC",
			expectedOpt: paymentModel.FilterPaymentHistoryOption{
				Page:          1,
				PerPage:       10,
				Sort:          "ASC",
				SortBy:        "updatedAt",
				StartDate:     time.Time{},
				EndDate:       time.Time{},
				Status:        "",
				ReferenceID:   "",
				PaymentMethod: "",
			},
			expectedErr: false,
		},
		{
			name:        "valid page and perPage",
			queryParams: "page=2&perPage=5",
			expectedOpt: paymentModel.FilterPaymentHistoryOption{
				Page:          2,
				PerPage:       5,
				Sort:          "ASC",
				SortBy:        "createdAt",
				StartDate:     time.Time{},
				EndDate:       time.Time{},
				Status:        "",
				ReferenceID:   "",
				PaymentMethod: "",
			},
			expectedErr: false,
		},
		{
			name:        "invalid page format",
			queryParams: "page=abc",
			expectedOpt: paymentModel.FilterPaymentHistoryOption{},
			expectedErr: true,
		},
		{
			name:        "invalid perPage format",
			queryParams: "perPage=abc",
			expectedOpt: paymentModel.FilterPaymentHistoryOption{},
			expectedErr: true,
		},
		{
			name:        "valid date range",
			queryParams: "startDate=2023-09-19T15:04:05Z&endDate=2023-09-20T15:04:05Z",
			expectedOpt: paymentModel.FilterPaymentHistoryOption{
				Page:          1,
				PerPage:       10,
				Sort:          "ASC",
				SortBy:        "createdAt",
				StartDate:     time.Date(2023, 9, 19, 15, 4, 5, 0, time.UTC),
				EndDate:       time.Date(2023, 9, 20, 15, 4, 5, 0, time.UTC),
				Status:        "",
				ReferenceID:   "",
				PaymentMethod: "",
			},
			expectedErr: false,
		},
		{
			name:        "valid payment date range",
			queryParams: "startDate=2023-09-19T15:04:05Z&endDate=2023-09-20T15:04:05Z&paymentStartDate=2023-09-19T15:04:05Z&paymentEndDate=2023-09-20T15:04:05Z",
			expectedOpt: paymentModel.FilterPaymentHistoryOption{
				Page:             1,
				PerPage:          10,
				Sort:             "ASC",
				SortBy:           "createdAt",
				StartDate:        time.Date(2023, 9, 19, 15, 4, 5, 0, time.UTC),
				EndDate:          time.Date(2023, 9, 20, 15, 4, 5, 0, time.UTC),
				PaymentStartDate: time.Date(2023, 9, 19, 15, 4, 5, 0, time.UTC),
				PaymentEndDate:   time.Date(2023, 9, 20, 15, 4, 5, 0, time.UTC),
				Status:           "",
				ReferenceID:      "",
				PaymentMethod:    "",
			},
			expectedErr: false,
		},
		{
			name:        "invalid startDate format",
			queryParams: "startDate=invalid",
			expectedOpt: paymentModel.FilterPaymentHistoryOption{},
			expectedErr: true,
		},
		{
			name:        "invalid endDate format",
			queryParams: "endDate=invalid",
			expectedOpt: paymentModel.FilterPaymentHistoryOption{},
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/payments?"+tt.queryParams, nil)

			ctrl := &PaymentController{}
			opt, err := ctrl.ParseFilterParam(req)

			if tt.expectedErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedOpt, opt)
		})
	}
}
