package cardFundedPayoutController

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	model "github.com/paper-indonesia/pivot-backoffice/internal/model/cardFundedPayout"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/validatorExt"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetPayoutList(t *testing.T) {
	vld := validatorExt.New()
	service := serviceMocks.NewICardFundedPayoutService(t)

	validUserClaims := &userModel.UserTokenClaims{
		UUID:       uuid.NewString(),
		Name:       "John Doe", // NOSONAR
		MerchantId: uuid.NewString(),
	}

	cfg := &config.Config{ServiceName: "testing"}
	handler := New(cfg, vld, service)

	router := chi.NewRouter()
	router.Get("/", handler.GetPayoutList)

	testCases := []struct {
		name             string
		queryParams      string
		userClaim        *userModel.UserTokenClaims
		setupMock        func()
		wantStatusCode   int
		wantResponseBody string
	}{
		{
			name:        "SUCCESS",
			queryParams: "",
			userClaim:   validUserClaims,
			setupMock: func() {
				service.On(
					"GetPayoutList", mock.Anything, mock.Anything,
				).Once().Return(&commonModel.PaginationResponse{
					Data: []model.GetPayoutListResponse{},
					Meta: commonModel.Meta{
						Page:       1,
						PerPage:    20,
						TotalItems: 0,
						TotalPages: 0,
					},
				}, nil)
			},
			wantStatusCode:   http.StatusOK,
			wantResponseBody: `{"code":"00","message":"OK","data":[],"pagination":{"page":1,"perPage":20,"totalItems":0,"totalPages":0}}`,
		},
		{
			name:             "ERROR: User not in Context",
			queryParams:      "",
			userClaim:        nil,
			wantStatusCode:   http.StatusUnauthorized,
			wantResponseBody: `{"code":"41","message":"user not found","error":{"type":"API_ERROR","details":[],"traceId":""},"data":null}`,
		},
		{
			name:             "ERROR: Invalid page format",
			queryParams:      "?page=invalid",
			userClaim:        validUserClaims,
			wantStatusCode:   http.StatusBadRequest,
			wantResponseBody: `{"code":"40","message":"invalid page format. Use number format instead","error":{"type":"API_ERROR","details":[],"traceId":""},"data":null}`,
		},
		{
			name:             "ERROR: Invalid perPage format",
			queryParams:      "?perPage=invalid",
			userClaim:        validUserClaims,
			wantStatusCode:   http.StatusBadRequest,
			wantResponseBody: `{"code":"40","message":"invalid perPage format. Use number format instead","error":{"type":"API_ERROR","details":[],"traceId":""},"data":null}`,
		},
		{
			name:             "ERROR: Invalid startDate format",
			queryParams:      "?startDate=invalid-date",
			userClaim:        validUserClaims,
			wantStatusCode:   http.StatusBadRequest,
			wantResponseBody: `{"code":"40","message":"invalid startDate format. Use 'YYYY-MM-DDTHH:mm:ssZ' format","error":{"type":"API_ERROR","details":[],"traceId":""},"data":null}`,
		},
		{
			name:             "ERROR: Invalid endDate format",
			queryParams:      "?endDate=invalid-date",
			userClaim:        validUserClaims,
			wantStatusCode:   http.StatusBadRequest,
			wantResponseBody: `{"code":"40","message":"invalid endDate format. Use 'YYYY-MM-DDTHH:mm:ssZ' format","error":{"type":"API_ERROR","details":[],"traceId":""},"data":null}`,
		},
		{
			name:        "ERROR: Service error",
			queryParams: "",
			userClaim:   validUserClaims,
			setupMock: func() {
				service.On("GetPayoutList", mock.Anything, mock.Anything).Once().Return(nil, assert.AnError)
			},
			wantStatusCode:   http.StatusInternalServerError,
			wantResponseBody: `{"code":"99","message":"assert.AnError general error for testing","error":{"type":"UNKNOWN","details":[],"traceId":""},"data":null}`,
		},
		{
			name:        "SUCCESS: With pagination params",
			queryParams: "?page=2&perPage=10",
			userClaim:   validUserClaims,
			setupMock: func() {
				service.On(
					"GetPayoutList", mock.Anything, mock.Anything,
				).Once().Return(&commonModel.PaginationResponse{
					Data: []model.GetPayoutListResponse{},
					Meta: commonModel.Meta{
						Page:       2,
						PerPage:    10,
						TotalItems: 0,
						TotalPages: 0,
					},
				}, nil)
			},
			wantStatusCode:   http.StatusOK,
			wantResponseBody: `{"code":"00","message":"OK","data":[],"pagination":{"page":2,"perPage":10,"totalItems":0,"totalPages":0}}`,
		},
		{
			name:        "SUCCESS: With date range",
			queryParams: "?startDate=2026-03-01T00:00:00Z&endDate=2026-03-15T23:59:59Z",
			userClaim:   validUserClaims,
			setupMock: func() {
				service.On(
					"GetPayoutList", mock.Anything, mock.Anything,
				).Once().Return(&commonModel.PaginationResponse{
					Data: []model.GetPayoutListResponse{},
					Meta: commonModel.Meta{
						Page:       1,
						PerPage:    20,
						TotalItems: 0,
						TotalPages: 0,
					},
				}, nil)
			},
			wantStatusCode:   http.StatusOK,
			wantResponseBody: `{"code":"00","message":"OK","data":[],"pagination":{"page":1,"perPage":20,"totalItems":0,"totalPages":0}}`,
		},
		{
			name:        "SUCCESS: With filters",
			queryParams: "?transactionStatus=success&approval=approved&searchId=test-search",
			userClaim:   validUserClaims,
			setupMock: func() {
				service.On(
					"GetPayoutList", mock.Anything, mock.Anything,
				).Once().Return(&commonModel.PaginationResponse{
					Data: []model.GetPayoutListResponse{},
					Meta: commonModel.Meta{
						Page:       1,
						PerPage:    20,
						TotalItems: 0,
						TotalPages: 0,
					},
				}, nil)
			},
			wantStatusCode:   http.StatusOK,
			wantResponseBody: `{"code":"00","message":"OK","data":[],"pagination":{"page":1,"perPage":20,"totalItems":0,"totalPages":0}}`,
		},
		{
			name:        "SUCCESS: With sorting",
			queryParams: "?sort=ASC&sortBy=referenceId",
			userClaim:   validUserClaims,
			setupMock: func() {
				service.On(
					"GetPayoutList", mock.Anything, mock.Anything,
				).Once().Return(&commonModel.PaginationResponse{
					Data: []model.GetPayoutListResponse{},
					Meta: commonModel.Meta{
						Page:       1,
						PerPage:    20,
						TotalItems: 0,
						TotalPages: 0,
					},
				}, nil)
			},
			wantStatusCode:   http.StatusOK,
			wantResponseBody: `{"code":"00","message":"OK","data":[],"pagination":{"page":1,"perPage":20,"totalItems":0,"totalPages":0}}`,
		},
		{
			name:        "SUCCESS: With data",
			queryParams: "",
			userClaim:   validUserClaims,
			setupMock: func() {
				createdAt := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
				service.On(
					"GetPayoutList", mock.Anything, mock.Anything,
				).Once().Return(&commonModel.PaginationResponse{
					Data: []model.GetPayoutListResponse{
						{
							UUID:              "6e34cfca-de00-4aa1-bc6b-069d878012c8",
							CreatedAt:         createdAt,
							ReferenceID:       "REF-123",
							Amount:            "1500000",
							Fee:               "15000",
							TotalAmount:       "1515000",
							TransactionStatus: "SUCCESS",
							ApprovalStatus:    "APPROVED",
							VendorID:          "vendor-001",
							VendorName:        "PT Sample Vendor",
							Remarks:           "Monthly payment",
							BankName:          "Bank Central Asia",
							AccountNumber:     "1234567890",
							AccountName:       "PT Sample Vendor",
							Card: model.CardInfo{
								LastFour: "4242",
								Brand:    "Visa",
								Channel:  "Credit",
								Name:     "Business Card",
								Issuer:   "BCA",
								Expiry:   "12/25",
							},
						},
					},
					Meta: commonModel.Meta{
						Page:       1,
						PerPage:    20,
						TotalItems: 1,
						TotalPages: 1,
					},
				}, nil)
			},
			wantStatusCode:   http.StatusOK,
			wantResponseBody: `{"code":"00","message":"OK","data":[{"uuid":"6e34cfca-de00-4aa1-bc6b-069d878012c8","createdAt":"2024-01-15T10:30:00Z","referenceId":"REF-123","amount":"1500000","fee":"15000","totalAmount":"1515000","transactionStatus":"SUCCESS","approvalStatus":"APPROVED","vendorId":"vendor-001","vendorName":"PT Sample Vendor","remarks":"Monthly payment","bankName":"Bank Central Asia","accountNumber":"1234567890","accountName":"PT Sample Vendor","card":{"lastFour":"4242","brand":"Visa","type":"Credit","name":"Business Card","issuer":"BCA","expiry":"12/25"}}],"pagination":{"page":1,"perPage":20,"totalItems":1,"totalPages":1}}`,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupMock != nil {
				tt.setupMock()
			}

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/"+tt.queryParams, nil)

			if tt.userClaim != nil {
				req = req.WithContext(context.WithValue(req.Context(), constant.CtxUserInfoKey, tt.userClaim))
			}

			router.ServeHTTP(rec, req)

			assert.Equal(t, tt.wantStatusCode, rec.Result().StatusCode)
			if !assert.JSONEq(t, tt.wantResponseBody, rec.Body.String()) {
				t.Log("Actual:", rec.Body.String())
			}

			service.AssertExpectations(t)
		})
	}
}
