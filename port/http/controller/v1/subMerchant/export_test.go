package subMerchant

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	mockRabbitMq "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/rabbitmqExt"
	mockSvc "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestExportPaymentHistory(t *testing.T) {
	parentMerchantID := uuid.NewString()
	subMerchantID := uuid.NewString()
	startDate := time.Now().UTC().AddDate(0, 0, -7).Format(time.DateOnly)
	endDate := time.Now().UTC().Format(time.DateOnly)

	testcases := []struct {
		name           string
		url            string
		body           string
		mockSetup      func(merchantSvc *mockSvc.IMerchantService, paymentSvc *mockSvc.IPaymentService)
		expectedStatus int
	}{
		{
			name: "SUCCESS",
			url:  fmt.Sprintf("/api/v1/sub-merchants/%s/payments/export", subMerchantID),
			body: fmt.Sprintf(`{"startDate":"%s","endDate":"%s"}`, startDate, endDate),
			mockSetup: func(merchantSvc *mockSvc.IMerchantService, paymentSvc *mockSvc.IPaymentService) {
				merchantSvc.On("FindMerchantByID", constant.ValueCtxMockType(), subMerchantID).Return(&merchantModel.Merchant{
					UUID:     subMerchantID,
					ParentID: sql.NullString{String: parentMerchantID, Valid: true},
				}, nil).Once()
				paymentSvc.On(
					"Export",
					constant.ValueCtxMockType(),
					mock.MatchedBy(func(req *paymentModel.PaymentDownloadHistoryRequest) bool {
						return req != nil && req.MerchantId == subMerchantID && req.StartDate == startDate && req.EndDate == endDate
					}),
				).Return(&paymentModel.PaymentDownloadHistoryResponse{URL: "https://signed-url"}, nil).Once()
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "ERROR:Invalid sub merchant id format",
			url:            "/api/v1/sub-merchants/not-uuid/payments/export",
			body:           fmt.Sprintf(`{"startDate":"%s","endDate":"%s"}`, startDate, endDate),
			mockSetup:      func(_ *mockSvc.IMerchantService, _ *mockSvc.IPaymentService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "ERROR:Sub merchant is not owned by parent merchant",
			url:  fmt.Sprintf("/api/v1/sub-merchants/%s/payments/export", subMerchantID),
			body: fmt.Sprintf(`{"startDate":"%s","endDate":"%s"}`, startDate, endDate),
			mockSetup: func(merchantSvc *mockSvc.IMerchantService, paymentSvc *mockSvc.IPaymentService) {
				merchantSvc.On("FindMerchantByID", constant.ValueCtxMockType(), subMerchantID).Return(&merchantModel.Merchant{
					UUID:     subMerchantID,
					ParentID: sql.NullString{String: uuid.NewString(), Valid: true},
				}, nil).Once()
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name: "ERROR:Sub merchant not found",
			url:  fmt.Sprintf("/api/v1/sub-merchants/%s/payments/export", subMerchantID),
			body: fmt.Sprintf(`{"startDate":"%s","endDate":"%s"}`, startDate, endDate),
			mockSetup: func(merchantSvc *mockSvc.IMerchantService, paymentSvc *mockSvc.IPaymentService) {
				merchantSvc.On("FindMerchantByID", constant.ValueCtxMockType(), subMerchantID).Return(nil, nil).Once()
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name: "ERROR:Invalid request body",
			url:  fmt.Sprintf("/api/v1/sub-merchants/%s/payments/export", subMerchantID),
			body: fmt.Sprintf(`{"startDate":"%s"}`, startDate),
			mockSetup: func(merchantSvc *mockSvc.IMerchantService, paymentSvc *mockSvc.IPaymentService) {
				merchantSvc.On("FindMerchantByID", constant.ValueCtxMockType(), subMerchantID).Return(&merchantModel.Merchant{
					UUID:     subMerchantID,
					ParentID: sql.NullString{String: parentMerchantID, Valid: true},
				}, nil).Once()
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "ERROR:Payment service error",
			url:  fmt.Sprintf("/api/v1/sub-merchants/%s/payments/export", subMerchantID),
			body: fmt.Sprintf(`{"startDate":"%s","endDate":"%s"}`, startDate, endDate),
			mockSetup: func(merchantSvc *mockSvc.IMerchantService, paymentSvc *mockSvc.IPaymentService) {
				merchantSvc.On("FindMerchantByID", constant.ValueCtxMockType(), subMerchantID).Return(&merchantModel.Merchant{
					UUID:     subMerchantID,
					ParentID: sql.NullString{String: parentMerchantID, Valid: true},
				}, nil).Once()
				paymentSvc.On(
					"Export",
					constant.ValueCtxMockType(),
					mock.Anything,
				).Return(nil, errors.New("export failed")).Once()
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			merchantSvc := mockSvc.NewIMerchantService(t)
			accountSvc := mockSvc.NewIAccountService(t)
			orchestratorSvc := mockSvc.NewIOrchestratorService(t)
			forbiddenSvc := mockSvc.NewIMerchantForbiddenUseCaseService(t)
			paymentSvc := mockSvc.NewIPaymentService(t)
			disbursementSvc := mockSvc.NewIDisbursementService(t)
			rabbitMq := mockRabbitMq.NewRabbitMQExt(t)

			tt.mockSetup(merchantSvc, paymentSvc)

			controller := New(
				merchantSvc,
				accountSvc,
				orchestratorSvc,
				forbiddenSvc,
				validator.New(),
				rabbitMq,
				WithPaymentService(paymentSvc),
				WithDisbursementService(disbursementSvc),
			)

			router := chi.NewRouter()
			router.Post("/api/v1/sub-merchants/{subMerchantId}/payments/export", controller.ExportPaymentHistory)

			req := httptest.NewRequest(http.MethodPost, tt.url, strings.NewReader(tt.body))
			ctx := context.WithValue(req.Context(), constant.CtxUserInfoKey, &userModel.UserTokenClaims{
				MerchantId: parentMerchantID,
				UUID:       uuid.NewString(),
			})
			req = req.WithContext(ctx)

			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
		})
	}
}

func TestExportDisbursementHistory(t *testing.T) {
	parentMerchantID := uuid.NewString()
	subMerchantID := uuid.NewString()
	startDate := time.Now().UTC().AddDate(0, 0, -7).Format(time.DateOnly)
	endDate := time.Now().UTC().Format(time.DateOnly)

	testcases := []struct {
		name           string
		url            string
		body           string
		mockSetup      func(merchantSvc *mockSvc.IMerchantService, disbursementSvc *mockSvc.IDisbursementService)
		expectedStatus int
	}{
		{
			name: "SUCCESS",
			url:  fmt.Sprintf("/api/v1/sub-merchants/%s/disbursements/export", subMerchantID),
			body: fmt.Sprintf(`{"startCreatedAt":"%s","endCreatedAt":"%s"}`, startDate, endDate),
			mockSetup: func(merchantSvc *mockSvc.IMerchantService, disbursementSvc *mockSvc.IDisbursementService) {
				merchantSvc.On("FindMerchantByID", constant.ValueCtxMockType(), subMerchantID).Return(&merchantModel.Merchant{
					UUID:     subMerchantID,
					ParentID: sql.NullString{String: parentMerchantID, Valid: true},
				}, nil).Once()
				disbursementSvc.On(
					"ExportToExcel",
					constant.ValueCtxMockType(),
					mock.MatchedBy(func(req *disbursementModel.GetDisbursementFilterRequest) bool {
						return req != nil && req.MerchantID == subMerchantID && req.StartCreatedAt != nil && req.EndCreatedAt != nil
					}),
				).Return(&disbursementModel.ExportDisbursementListResponse{Url: "https://signed-url"}, nil).Once()
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "SUCCESS:Without dates",
			url:  fmt.Sprintf("/api/v1/sub-merchants/%s/disbursements/export", subMerchantID),
			body: `{}`,
			mockSetup: func(merchantSvc *mockSvc.IMerchantService, disbursementSvc *mockSvc.IDisbursementService) {
				merchantSvc.On("FindMerchantByID", constant.ValueCtxMockType(), subMerchantID).Return(&merchantModel.Merchant{
					UUID:     subMerchantID,
					ParentID: sql.NullString{String: parentMerchantID, Valid: true},
				}, nil).Once()
				disbursementSvc.On(
					"ExportToExcel",
					constant.ValueCtxMockType(),
					mock.MatchedBy(func(req *disbursementModel.GetDisbursementFilterRequest) bool {
						return req != nil && req.MerchantID == subMerchantID
					}),
				).Return(&disbursementModel.ExportDisbursementListResponse{Url: "https://signed-url"}, nil).Once()
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "ERROR:Invalid sub merchant id format",
			url:            "/api/v1/sub-merchants/not-uuid/disbursements/export",
			body:           `{}`,
			mockSetup:      func(_ *mockSvc.IMerchantService, _ *mockSvc.IDisbursementService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "ERROR:Sub merchant is not owned by parent merchant",
			url:  fmt.Sprintf("/api/v1/sub-merchants/%s/disbursements/export", subMerchantID),
			body: `{}`,
			mockSetup: func(merchantSvc *mockSvc.IMerchantService, disbursementSvc *mockSvc.IDisbursementService) {
				merchantSvc.On("FindMerchantByID", constant.ValueCtxMockType(), subMerchantID).Return(&merchantModel.Merchant{
					UUID:     subMerchantID,
					ParentID: sql.NullString{String: uuid.NewString(), Valid: true},
				}, nil).Once()
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name: "ERROR:Sub merchant not found",
			url:  fmt.Sprintf("/api/v1/sub-merchants/%s/disbursements/export", subMerchantID),
			body: `{}`,
			mockSetup: func(merchantSvc *mockSvc.IMerchantService, disbursementSvc *mockSvc.IDisbursementService) {
				merchantSvc.On("FindMerchantByID", constant.ValueCtxMockType(), subMerchantID).Return(nil, nil).Once()
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name: "ERROR:Malformed request body",
			url:  fmt.Sprintf("/api/v1/sub-merchants/%s/disbursements/export", subMerchantID),
			body: `{"startCreatedAt":"invalid-time"`,
			mockSetup: func(merchantSvc *mockSvc.IMerchantService, disbursementSvc *mockSvc.IDisbursementService) {
				merchantSvc.On("FindMerchantByID", constant.ValueCtxMockType(), subMerchantID).Return(&merchantModel.Merchant{
					UUID:     subMerchantID,
					ParentID: sql.NullString{String: parentMerchantID, Valid: true},
				}, nil).Once()
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "ERROR:Invalid date format",
			url:  fmt.Sprintf("/api/v1/sub-merchants/%s/disbursements/export", subMerchantID),
			body: `{"startCreatedAt":"not-a-date"}`,
			mockSetup: func(merchantSvc *mockSvc.IMerchantService, disbursementSvc *mockSvc.IDisbursementService) {
				merchantSvc.On("FindMerchantByID", constant.ValueCtxMockType(), subMerchantID).Return(&merchantModel.Merchant{
					UUID:     subMerchantID,
					ParentID: sql.NullString{String: parentMerchantID, Valid: true},
				}, nil).Once()
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "ERROR:Disbursement service error",
			url:  fmt.Sprintf("/api/v1/sub-merchants/%s/disbursements/export", subMerchantID),
			body: `{}`,
			mockSetup: func(merchantSvc *mockSvc.IMerchantService, disbursementSvc *mockSvc.IDisbursementService) {
				merchantSvc.On("FindMerchantByID", constant.ValueCtxMockType(), subMerchantID).Return(&merchantModel.Merchant{
					UUID:     subMerchantID,
					ParentID: sql.NullString{String: parentMerchantID, Valid: true},
				}, nil).Once()
				disbursementSvc.On("ExportToExcel", constant.ValueCtxMockType(), mock.Anything).Return(nil, errors.New("export failed")).Once()
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			merchantSvc := mockSvc.NewIMerchantService(t)
			accountSvc := mockSvc.NewIAccountService(t)
			orchestratorSvc := mockSvc.NewIOrchestratorService(t)
			forbiddenSvc := mockSvc.NewIMerchantForbiddenUseCaseService(t)
			paymentSvc := mockSvc.NewIPaymentService(t)
			disbursementSvc := mockSvc.NewIDisbursementService(t)
			rabbitMq := mockRabbitMq.NewRabbitMQExt(t)

			tt.mockSetup(merchantSvc, disbursementSvc)

			controller := New(
				merchantSvc,
				accountSvc,
				orchestratorSvc,
				forbiddenSvc,
				validator.New(),
				rabbitMq,
				WithPaymentService(paymentSvc),
				WithDisbursementService(disbursementSvc),
			)

			router := chi.NewRouter()
			router.Post("/api/v1/sub-merchants/{subMerchantId}/disbursements/export", controller.ExportDisbursementHistory)

			req := httptest.NewRequest(http.MethodPost, tt.url, strings.NewReader(tt.body))
			ctx := context.WithValue(req.Context(), constant.CtxUserInfoKey, &userModel.UserTokenClaims{
				MerchantId: parentMerchantID,
				UUID:       uuid.NewString(),
			})
			req = req.WithContext(ctx)

			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
		})
	}
}
