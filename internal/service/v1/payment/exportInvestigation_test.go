package paymentService

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	c "github.com/paper-indonesia/pivot-backoffice/constant"
	paymentConst "github.com/paper-indonesia/pivot-backoffice/constant/payment"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	gcsMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/gcs"
	repoMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestExportInvestigatedPayments(t *testing.T) {
	mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

	merchantID := uuid.NewString()
	traceID := "test-trace-id"
	ctx := context.WithValue(context.Background(), pdkConst.CtxTraceIdKey, traceID)

	fromDate := time.Now().Add(-7 * 24 * time.Hour)
	toDate := time.Now()

	validRequest := &paymentModel.InvestigationDownloadHistoryRequest{
		MerchantId:          merchantID,
		InvestigationStatus: paymentConst.InvestigationStatusInProcess,
		PaymentMethod:       "QRIS",
		FromDate:            &fromDate,
		ToDate:              &toDate,
	}

	tests := []struct {
		name      string
		request   *paymentModel.InvestigationDownloadHistoryRequest
		mockSetup func(mockRepo *repoMocks.IPaymentRepository, mockGCS *gcsMocks.GCSService)
		wantErr   bool
	}{
		{
			name:    "SUCCESS: Export investigated payments with filters",
			request: validRequest,
			mockSetup: func(mockRepo *repoMocks.IPaymentRepository, mockGCS *gcsMocks.GCSService) {
				// Mock repository GetInvestigatedPayments - returns DTOs
				dtos := []*paymentModel.InvestigatedPaymentDTO{
					{
						UUID:              uuid.NewString(),
						ReferenceId:       "PAY-123456",
						Amount:            decimal.NewFromInt(100000),
						Currency:          "IDR",
						MerchantID:        merchantID,
						MerchantName:      "Test Merchant",
						PaymentMethodType: "QRIS",
						PaymentChannel:    "QRIS",
						Status:            "PAID",
						ReasonType:        paymentConst.InvestigationStatusInProcess,
						ReasonDescription: nil,
						UpdatedAt:         time.Now(),
						StartedAt:         &fromDate,
						CompletedAt:       nil,
					},
				}
				mockRepo.On("GetInvestigatedPayments", mock.Anything, mock.MatchedBy(func(filter *paymentModel.GetInvestigatedPaymentsFilterRequest) bool {
					return filter.MerchantID == merchantID &&
						filter.InvestigationStatus == paymentConst.InvestigationStatusInProcess &&
						filter.PaymentMethod == "QRIS" &&
						filter.Limit == 10000
				})).Once().Return(&commonModel.PaginationResponse{
					Data: dtos,
					Meta: commonModel.Meta{Page: 1, PerPage: 10000, TotalItems: 1, TotalPages: 1},
				}, nil)

				// Mock GCS upload
				mockGCS.On("UploadFile", mock.Anything, c.StringMockType(), mock.Anything, c.BoolMockType(), mock.Anything).
					Once().Return(nil, nil)

				// Mock GCS CreateSignedURL
				mockGCS.On("CreateSignedURL", mock.Anything, c.StringMockType(), c.DurationMockType()).
					Once().Return("https://storage.googleapis.com/signed-url", nil)
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: Export investigated payments without filters",
			request: &paymentModel.InvestigationDownloadHistoryRequest{
				MerchantId: merchantID,
			},
			mockSetup: func(mockRepo *repoMocks.IPaymentRepository, mockGCS *gcsMocks.GCSService) {
				// Mock GetInvestigatedPayments - empty result
				mockRepo.On("GetInvestigatedPayments", mock.Anything, mock.Anything).
					Once().Return(&commonModel.PaginationResponse{
					Data: []*paymentModel.InvestigatedPaymentDTO{},
					Meta: commonModel.Meta{Page: 1, PerPage: 10000, TotalItems: 0, TotalPages: 0},
				}, nil)

				// Mock GCS upload
				mockGCS.On("UploadFile", mock.Anything, c.StringMockType(), mock.Anything, c.BoolMockType(), mock.Anything).
					Once().Return(nil, nil)

				// Mock GCS CreateSignedURL
				mockGCS.On("CreateSignedURL", mock.Anything, c.StringMockType(), c.DurationMockType()).
					Once().Return("https://storage.googleapis.com/signed-url-empty", nil)
			},
			wantErr: false,
		},
		{
			name:    "SUCCESS: Export with completed investigation",
			request: validRequest,
			mockSetup: func(mockRepo *repoMocks.IPaymentRepository, mockGCS *gcsMocks.GCSService) {
				completedAt := time.Now()
				notes := "Investigation completed successfully"
				dtos := []*paymentModel.InvestigatedPaymentDTO{
					{
						UUID:              uuid.NewString(),
						ReferenceId:       "PAY-789012",
						Amount:            decimal.NewFromInt(250000),
						Currency:          "IDR",
						MerchantID:        merchantID,
						MerchantName:      "Test Merchant",
						PaymentMethodType: "VIRTUAL_ACCOUNT",
						PaymentChannel:    "BCA",
						Status:            "PAID",
						ReasonType:        paymentConst.InvestigationStatusSuccess,
						ReasonDescription: &notes,
						UpdatedAt:         completedAt,
						StartedAt:         &fromDate,
						CompletedAt:       &completedAt,
					},
				}
				mockRepo.On("GetInvestigatedPayments", mock.Anything, mock.Anything).
					Once().Return(&commonModel.PaginationResponse{
					Data: dtos,
					Meta: commonModel.Meta{Page: 1, PerPage: 10000, TotalItems: 1, TotalPages: 1},
				}, nil)

				mockGCS.On("UploadFile", mock.Anything, c.StringMockType(), mock.Anything, c.BoolMockType(), mock.Anything).
					Once().Return(nil, nil)

				mockGCS.On("CreateSignedURL", mock.Anything, c.StringMockType(), c.DurationMockType()).
					Once().Return("https://storage.googleapis.com/signed-url-completed", nil)
			},
			wantErr: false,
		},
		{
			name:    "ERROR: Failed to get investigated payments",
			request: validRequest,
			mockSetup: func(mockRepo *repoMocks.IPaymentRepository, mockGCS *gcsMocks.GCSService) {
				mockRepo.On("GetInvestigatedPayments", mock.Anything, mock.Anything).
					Once().Return(nil, errors.New("database error"))
			},
			wantErr: true,
		},
		{
			name:    "ERROR: Failed to upload file to GCS",
			request: validRequest,
			mockSetup: func(mockRepo *repoMocks.IPaymentRepository, mockGCS *gcsMocks.GCSService) {
				dtos := []*paymentModel.InvestigatedPaymentDTO{
					{
						UUID:              uuid.NewString(),
						ReferenceId:       "PAY-123456",
						Amount:            decimal.NewFromInt(100000),
						Currency:          "IDR",
						MerchantID:        merchantID,
						MerchantName:      "Test Merchant",
						PaymentMethodType: "QRIS",
						Status:            "PAID",
						ReasonType:        paymentConst.InvestigationStatusInProcess,
						UpdatedAt:         time.Now(),
					},
				}
				mockRepo.On("GetInvestigatedPayments", mock.Anything, mock.Anything).
					Once().Return(&commonModel.PaginationResponse{
					Data: dtos,
					Meta: commonModel.Meta{Page: 1, PerPage: 10000, TotalItems: 1, TotalPages: 1},
				}, nil)

				mockGCS.On("UploadFile", mock.Anything, c.StringMockType(), mock.Anything, c.BoolMockType(), mock.Anything).
					Once().Return(nil, errors.New("upload failed"))
			},
			wantErr: true,
		},
		{
			name:    "ERROR: Failed to create signed URL",
			request: validRequest,
			mockSetup: func(mockRepo *repoMocks.IPaymentRepository, mockGCS *gcsMocks.GCSService) {
				dtos := []*paymentModel.InvestigatedPaymentDTO{
					{
						UUID:              uuid.NewString(),
						ReferenceId:       "PAY-123456",
						Amount:            decimal.NewFromInt(100000),
						Currency:          "IDR",
						MerchantID:        merchantID,
						MerchantName:      "Test Merchant",
						PaymentMethodType: "QRIS",
						Status:            "PAID",
						ReasonType:        paymentConst.InvestigationStatusInProcess,
						UpdatedAt:         time.Now(),
					},
				}
				mockRepo.On("GetInvestigatedPayments", mock.Anything, mock.Anything).
					Once().Return(&commonModel.PaginationResponse{
					Data: dtos,
					Meta: commonModel.Meta{Page: 1, PerPage: 10000, TotalItems: 1, TotalPages: 1},
				}, nil)

				mockGCS.On("UploadFile", mock.Anything, c.StringMockType(), mock.Anything, c.BoolMockType(), mock.Anything).
					Once().Return(nil, nil)

				mockGCS.On("CreateSignedURL", mock.Anything, c.StringMockType(), c.DurationMockType()).
					Once().Return("", errors.New("failed to create signed URL"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPaymentRepo := repoMocks.NewIPaymentRepository(t)
			mockGCS := gcsMocks.NewGCSService(t)

			tt.mockSetup(mockPaymentRepo, mockGCS)

			svc := New(mockPaymentRepo, mockLogger, nil, nil, nil, nil, nil, WithGCSService(mockGCS))
			result, err := svc.ExportInvestigatedPayments(ctx, tt.request)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
				// Verify error contains trace ID
				assert.ErrorContains(t, err, fmt.Sprintf(c.InternalErrorFmt, traceID))
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.NotEmpty(t, result.URL)
				assert.Contains(t, result.URL, "https://storage.googleapis.com")
			}

			mockPaymentRepo.AssertExpectations(t)
			mockGCS.AssertExpectations(t)
		})
	}
}
