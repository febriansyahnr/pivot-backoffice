package paymentService_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	paymentConst "github.com/paper-indonesia/pivot-backoffice/constant/payment"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	. "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/payment"
	pdkLogMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pdk/logger"
	gcsMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/gcs"
	repoMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	pkgErr "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"

	"github.com/google/uuid"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetInvestigatedPayments(t *testing.T) {
	mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
	ctx := context.Background()

	validFilter := &paymentModel.GetInvestigatedPaymentsFilterRequest{
		Page:  1,
		Limit: 20,
	}

	tests := []struct {
		name           string
		filter         *paymentModel.GetInvestigatedPaymentsFilterRequest
		mockSetup      func(mockRepo *repoMocks.IPaymentRepository)
		wantErr        bool
		wantErrContain string
	}{
		{
			name:   "SUCCESS: Get investigated payments",
			filter: validFilter,
			mockSetup: func(mockRepo *repoMocks.IPaymentRepository) {
				dtos := []*paymentModel.InvestigatedPaymentDTO{
					{
						UUID:              uuid.NewString(),
						Amount:            decimal.NewFromInt(50000),
						Currency:          "IDR",
						MerchantID:        uuid.NewString(),
						MerchantName:      "Test Merchant",
						PaymentMethodType: "QRIS",
						Status:            "SUCCESS",
						ReasonType:        paymentConst.InvestigationStatusInProcess,
						UpdatedAt:         time.Now(),
						StartedAt:         util.ValueToPtr(time.Date(2026, 1, 14, 10, 0, 0, 0, time.UTC)),
					},
				}
				mockRepo.On("GetInvestigatedPayments", mock.Anything, validFilter).
					Once().Return(&commonModel.PaginationResponse{
					Data: dtos,
					Meta: commonModel.Meta{Page: 1, PerPage: 20, TotalItems: 1, TotalPages: 1},
				}, nil)
			},
			wantErr: false,
		},
		{
			name:   "SUCCESS: Get investigated payments - empty result",
			filter: validFilter,
			mockSetup: func(mockRepo *repoMocks.IPaymentRepository) {
				mockRepo.On("GetInvestigatedPayments", mock.Anything, validFilter).
					Once().Return(&commonModel.PaginationResponse{
					Data: []*paymentModel.InvestigatedPaymentDTO{},
					Meta: commonModel.Meta{Page: 1, PerPage: 20, TotalItems: 0, TotalPages: 0},
				}, nil)
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: Get investigated payments with status filter",
			filter: &paymentModel.GetInvestigatedPaymentsFilterRequest{
				Page:                1,
				Limit:               20,
				InvestigationStatus: paymentConst.InvestigationStatusInProcess,
			},
			mockSetup: func(mockRepo *repoMocks.IPaymentRepository) {
				mockRepo.On("GetInvestigatedPayments", mock.Anything, mock.Anything).
					Once().Return(&commonModel.PaginationResponse{
					Data: []*paymentModel.InvestigatedPaymentDTO{},
					Meta: commonModel.Meta{Page: 1, PerPage: 20, TotalItems: 0, TotalPages: 0},
				}, nil)
			},
			wantErr: false,
		},
		{
			name:   "ERROR: Repository error",
			filter: validFilter,
			mockSetup: func(mockRepo *repoMocks.IPaymentRepository) {
				mockRepo.On("GetInvestigatedPayments", mock.Anything, validFilter).
					Once().Return(nil, errors.New("database error"))
			},
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockPaymentRepo := repoMocks.NewIPaymentRepository(t)
			test.mockSetup(mockPaymentRepo)

			svc := New(mockPaymentRepo, mockLogger, nil, nil, nil, nil, nil)
			result, err := svc.GetInvestigatedPayments(ctx, test.filter)

			if test.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}

			mockPaymentRepo.AssertExpectations(t)
		})
	}
}

func TestUpdateInvestigationStatus(t *testing.T) {
	mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
	ctx := context.Background()

	validPaymentID := uuid.NewString()
	validNotes := "Bank confirmation received"

	tests := []struct {
		name           string
		paymentID      string
		request        *paymentModel.UpdateInvestigationRequest
		mockSetup      func(mockRepo *repoMocks.IPaymentRepository)
		wantErr        bool
		wantErrContain error
	}{
		{
			name:      "SUCCESS: Update to INVESTIGATION_SUCCESS",
			paymentID: validPaymentID,
			request: &paymentModel.UpdateInvestigationRequest{
				InvestigationStatus: paymentConst.InvestigationStatusSuccess,
				Notes:               &validNotes,
			},
			mockSetup: func(mockRepo *repoMocks.IPaymentRepository) {
				inProcessStatus := paymentConst.InvestigationStatusInProcess
				mockRepo.On("GetPaymentById", mock.Anything, validPaymentID).
					Once().Return(&paymentModel.Payment{
					UUID:       validPaymentID,
					ReasonType: &inProcessStatus,
				}, nil)
				mockRepo.On("UpdateInvestigationStatus", mock.Anything, mock.MatchedBy(func(req paymentModel.UpdateInvestigationStatusRequest) bool {
					return req.PaymentID == validPaymentID && req.Status == paymentConst.InvestigationStatusSuccess && req.Notes != nil && *req.Notes == validNotes
				})).Once().Return(nil)
			},
			wantErr: false,
		},
		{
			name:      "SUCCESS: Update to INVESTIGATION_FAILED",
			paymentID: validPaymentID,
			request: &paymentModel.UpdateInvestigationRequest{
				InvestigationStatus: paymentConst.InvestigationStatusFailed,
				Notes:               &validNotes,
			},
			mockSetup: func(mockRepo *repoMocks.IPaymentRepository) {
				inProcessStatus := paymentConst.InvestigationStatusInProcess
				mockRepo.On("GetPaymentById", mock.Anything, validPaymentID).
					Once().Return(&paymentModel.Payment{
					UUID:       validPaymentID,
					ReasonType: &inProcessStatus,
				}, nil)
				mockRepo.On("UpdateInvestigationStatus", mock.Anything, mock.MatchedBy(func(req paymentModel.UpdateInvestigationStatusRequest) bool {
					return req.PaymentID == validPaymentID && req.Status == paymentConst.InvestigationStatusFailed && req.Notes != nil && *req.Notes == validNotes
				})).Once().Return(nil)
			},
			wantErr: false,
		},
		{
			name:      "SUCCESS: Update without notes",
			paymentID: validPaymentID,
			request: &paymentModel.UpdateInvestigationRequest{
				InvestigationStatus: paymentConst.InvestigationStatusSuccess,
				Notes:               nil,
			},
			mockSetup: func(mockRepo *repoMocks.IPaymentRepository) {
				inProcessStatus := paymentConst.InvestigationStatusInProcess
				mockRepo.On("GetPaymentById", mock.Anything, validPaymentID).
					Once().Return(&paymentModel.Payment{
					UUID:       validPaymentID,
					ReasonType: &inProcessStatus,
				}, nil)
				mockRepo.On("UpdateInvestigationStatus", mock.Anything, mock.MatchedBy(func(req paymentModel.UpdateInvestigationStatusRequest) bool {
					return req.PaymentID == validPaymentID && req.Status == paymentConst.InvestigationStatusSuccess && req.Notes == nil
				})).Once().Return(nil)
			},
			wantErr: false,
		},
		{
			name:      "ERROR: Payment not found",
			paymentID: validPaymentID,
			request: &paymentModel.UpdateInvestigationRequest{
				InvestigationStatus: paymentConst.InvestigationStatusSuccess,
			},
			mockSetup: func(mockRepo *repoMocks.IPaymentRepository) {
				mockRepo.On("GetPaymentById", mock.Anything, validPaymentID).
					Once().Return(nil, nil)
			},
			wantErr:        true,
			wantErrContain: constant.ErrPaymentNotFound,
		},
		{
			name:      "ERROR: Payment not under investigation",
			paymentID: validPaymentID,
			request: &paymentModel.UpdateInvestigationRequest{
				InvestigationStatus: paymentConst.InvestigationStatusSuccess,
			},
			mockSetup: func(mockRepo *repoMocks.IPaymentRepository) {
				mockRepo.On("GetPaymentById", mock.Anything, validPaymentID).
					Once().Return(&paymentModel.Payment{
					UUID:       validPaymentID,
					ReasonType: nil,
				}, nil)
			},
			wantErr:        true,
			wantErrContain: constant.ErrInvestigationNotFound,
		},
		{
			name:      "ERROR: Investigation already finalized (SUCCESS)",
			paymentID: validPaymentID,
			request: &paymentModel.UpdateInvestigationRequest{
				InvestigationStatus: paymentConst.InvestigationStatusSuccess,
			},
			mockSetup: func(mockRepo *repoMocks.IPaymentRepository) {
				successStatus := paymentConst.InvestigationStatusSuccess
				mockRepo.On("GetPaymentById", mock.Anything, validPaymentID).
					Once().Return(&paymentModel.Payment{
					UUID:       validPaymentID,
					ReasonType: &successStatus,
				}, nil)
			},
			wantErr:        true,
			wantErrContain: constant.ErrInvestigationAlreadyFinalized,
		},
		{
			name:      "ERROR: Investigation already finalized (FAILED)",
			paymentID: validPaymentID,
			request: &paymentModel.UpdateInvestigationRequest{
				InvestigationStatus: paymentConst.InvestigationStatusSuccess,
			},
			mockSetup: func(mockRepo *repoMocks.IPaymentRepository) {
				failedStatus := paymentConst.InvestigationStatusFailed
				mockRepo.On("GetPaymentById", mock.Anything, validPaymentID).
					Once().Return(&paymentModel.Payment{
					UUID:       validPaymentID,
					ReasonType: &failedStatus,
				}, nil)
			},
			wantErr:        true,
			wantErrContain: constant.ErrInvestigationAlreadyFinalized,
		},
		{
			name:      "ERROR: Invalid investigation status",
			paymentID: validPaymentID,
			request: &paymentModel.UpdateInvestigationRequest{
				InvestigationStatus: "INVALID_STATUS",
			},
			mockSetup: func(mockRepo *repoMocks.IPaymentRepository) {
				inProcessStatus := paymentConst.InvestigationStatusInProcess
				mockRepo.On("GetPaymentById", mock.Anything, validPaymentID).
					Once().Return(&paymentModel.Payment{
					UUID:       validPaymentID,
					ReasonType: &inProcessStatus,
				}, nil)
			},
			wantErr:        true,
			wantErrContain: constant.ErrInvalidInvestigationStatus,
		},
		{
			name:      "ERROR: Notes exceed 200 characters",
			paymentID: validPaymentID,
			request: &paymentModel.UpdateInvestigationRequest{
				InvestigationStatus: paymentConst.InvestigationStatusSuccess,
				Notes: func() *string {
					s := "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip."
					return &s
				}(),
			},
			mockSetup: func(mockRepo *repoMocks.IPaymentRepository) {
				inProcessStatus := paymentConst.InvestigationStatusInProcess
				mockRepo.On("GetPaymentById", mock.Anything, validPaymentID).
					Once().Return(&paymentModel.Payment{
					UUID:       validPaymentID,
					ReasonType: &inProcessStatus,
				}, nil)
			},
			wantErr:        true,
			wantErrContain: constant.ErrInvestigationNotesExceedLimit,
		},
		{
			name:      "ERROR: Database error when getting payment",
			paymentID: validPaymentID,
			request: &paymentModel.UpdateInvestigationRequest{
				InvestigationStatus: paymentConst.InvestigationStatusSuccess,
			},
			mockSetup: func(mockRepo *repoMocks.IPaymentRepository) {
				mockRepo.On("GetPaymentById", mock.Anything, validPaymentID).
					Once().Return(nil, errors.New("database error"))
			},
			wantErr: true,
		},
		{
			name:      "ERROR: Database error when updating",
			paymentID: validPaymentID,
			request: &paymentModel.UpdateInvestigationRequest{
				InvestigationStatus: paymentConst.InvestigationStatusSuccess,
			},
			mockSetup: func(mockRepo *repoMocks.IPaymentRepository) {
				inProcessStatus := paymentConst.InvestigationStatusInProcess
				mockRepo.On("GetPaymentById", mock.Anything, validPaymentID).
					Once().Return(&paymentModel.Payment{
					UUID:       validPaymentID,
					ReasonType: &inProcessStatus,
				}, nil)
				mockRepo.On("UpdateInvestigationStatus", mock.Anything, mock.MatchedBy(func(req paymentModel.UpdateInvestigationStatusRequest) bool {
					return req.PaymentID == validPaymentID && req.Status == paymentConst.InvestigationStatusSuccess && req.Notes == nil
				})).Once().Return(errors.New("database error"))
			},
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockPaymentRepo := repoMocks.NewIPaymentRepository(t)
			test.mockSetup(mockPaymentRepo)

			svc := New(mockPaymentRepo, mockLogger, nil, nil, nil, nil, nil)
			result, err := svc.UpdateInvestigationStatus(ctx, test.paymentID, test.request)

			if test.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
				if test.wantErrContain != nil {
					assert.ErrorContains(t, err, test.wantErrContain.Error())
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, test.paymentID, result.PaymentReferenceID)
				assert.Equal(t, test.request.InvestigationStatus, result.InvestigationStatus)
				assert.NotNil(t, result.CompletedAt)
			}

			mockPaymentRepo.AssertExpectations(t)
		})
	}
}

func TestGetInvestigationProofOfPayment(t *testing.T) {
	gcs := gcsMocks.NewGCSService(t)
	logger := pdkLogMocks.NewILogger(t)
	paymentRepo := repoMocks.NewIPaymentRepository(t)

	svc := New(paymentRepo, logger, nil, nil, nil, nil, nil, WithGCSService(gcs))

	validPaymentID := uuid.NewString()
	validPath := "investigations/merchant-123/payment-456.png"
	validNotes := "Customer showed payment screenshot"
	validSignedURL := "https://storage.googleapis.com/test-bucket/investigations/merchant-123/payment-456.png?signature=abc123"

	tests := []struct {
		name       string
		request    paymentModel.GetInvestigationProofOfPaymentRequest
		mockSetup  func()
		wantErr    error
		wantResult *paymentModel.GetInvestigationProofOfPaymentResponse
	}{
		{
			name: "ERROR: Database error when getting payment",
			request: paymentModel.GetInvestigationProofOfPaymentRequest{
				PaymentID: validPaymentID,
			},
			mockSetup: func() {
				logger.On(
					"Error", mock.Anything, "Failed to retrieve payment details", mock.Anything,
				).Once().Return()
				paymentRepo.On("GetPaymentById", mock.Anything, validPaymentID).Once().Return(nil, assert.AnError)
			},
			wantErr: pkgErr.New(response.HttpErrDatabase, constant.ErrInternalServerForUser),
		},
		{
			name: "ERROR: Payment not found",
			request: paymentModel.GetInvestigationProofOfPaymentRequest{
				PaymentID: validPaymentID,
			},
			mockSetup: func() {
				paymentRepo.On("GetPaymentById", mock.Anything, validPaymentID).Once().Return(nil, nil)
			},
			wantErr: pkgErr.New(response.HttpErrNotFound, constant.ErrPaymentNotFound),
		},
		{
			name: "ERROR: Payment metadata is nil",
			request: paymentModel.GetInvestigationProofOfPaymentRequest{
				PaymentID: validPaymentID,
			},
			mockSetup: func() {
				paymentRepo.On(
					"GetPaymentById", mock.Anything, validPaymentID,
				).Once().Return(&paymentModel.Payment{UUID: validPaymentID}, nil)
			},
			wantErr: pkgErr.New(
				response.HttpErrUnprocessableContent, errors.New("payment is not under investigation"),
			),
		},
		{
			name: "ERROR: investigationPoP path is empty",
			request: paymentModel.GetInvestigationProofOfPaymentRequest{
				PaymentID: validPaymentID,
			},
			mockSetup: func() {
				paymentRepo.On(
					"GetPaymentById", mock.Anything, validPaymentID,
				).Once().Return(&paymentModel.Payment{
					UUID:     validPaymentID,
					Metadata: &map[string]any{"investigationPoP": map[string]any{}},
				}, nil)
			},
			wantErr: pkgErr.New(
				response.HttpErrUnprocessableContent,
				errors.New("uploaded proof of transaction was not found for this payment"),
			),
		},
		{
			name: "ERROR: GCS failed to generate signed URL",
			request: paymentModel.GetInvestigationProofOfPaymentRequest{
				PaymentID: validPaymentID,
			},
			mockSetup: func() {
				paymentRepo.On(
					"GetPaymentById", mock.Anything, validPaymentID,
				).Return(&paymentModel.Payment{
					UUID: validPaymentID,
					Metadata: &map[string]any{
						"investigationPoP": map[string]any{
							"bucket":        "test-bucket",
							"path":          validPath,
							"merchantNotes": validNotes,
						},
					},
				}, nil)
				gcs.On("CreateSignedURL", mock.Anything, validPath, mock.Anything).Once().Return("", assert.AnError)
				logger.On("Error", mock.Anything, "Failed to generate signed URL", mock.Anything).Once().Return()
			},
			wantErr: pkgErr.New(response.HttpErrInternal, constant.ErrInternalServerForUser),
		},
		{
			name: "SUCCESS: Get proof of payment",
			request: paymentModel.GetInvestigationProofOfPaymentRequest{
				PaymentID: validPaymentID,
			},
			mockSetup: func() {
				gcs.On("CreateSignedURL", mock.Anything, validPath, mock.Anything).Once().Return(validSignedURL, nil)
			},
			wantErr: nil,
			wantResult: &paymentModel.GetInvestigationProofOfPaymentResponse{
				SignedURL:     validSignedURL,
				MerchantNotes: validNotes,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			result, err := svc.GetInvestigationProofOfPayment(t.Context(), tt.request)

			assert.Equal(t, tt.wantErr, err)
			if tt.wantResult != nil && result != nil {
				assert.NotNil(t, result.ExpiresAt)
				tt.wantResult.ExpiresAt = result.ExpiresAt
			}
			assert.Equal(t, tt.wantResult, result)
		})
	}
}
