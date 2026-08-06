package unifiedPaymentService

import (
	"context"
	"errors"
	"fmt"
	"mime/multipart"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	paymentConst "github.com/paper-indonesia/pivot-backoffice/constant/payment"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	snapCoreQRModel "github.com/paper-indonesia/pivot-backoffice/internal/model/snapCore/qr"
	unifiedPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/unifiedPayment"
	fdsMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/fds"
	gcsMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/gcs"
	redisExtMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/redisExt"
	repositoryMock "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	pkgErr "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/fds"
	gcsPkg "github.com/paper-indonesia/pivot-backoffice/pkg/gcs"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	pdkLogger "github.com/paper-indonesia/pdk/v2/logger"
	redisSdk "github.com/redis/go-redis/v9"
	"github.com/shopspring/decimal"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestIsValidStatusForInvestigation(t *testing.T) {
	svc := &UnifiedPaymentService{}

	tests := []struct {
		name     string
		status   string
		expected bool
	}{
		{
			name:     "ChargeStatusProcessing is valid",
			status:   constant.ChargeStatusProcessing,
			expected: true,
		},
		{
			name:     "ChargeStatusWaitingForUserAction is valid",
			status:   constant.ChargeStatusWaitingForUserAction,
			expected: true,
		},
		{
			name:     "ChargeStatusExpired is valid",
			status:   constant.ChargeStatusExpired,
			expected: true,
		},
		{
			name:     "UnifiedPaymentSessionStatusProcessing is valid",
			status:   constant.UnifiedPaymentSessionStatusProcessing,
			expected: true,
		},
		{
			name:     "UnifiedPaymentSessionStatusRequireAction is valid",
			status:   constant.UnifiedPaymentSessionStatusRequireAction,
			expected: true,
		},
		{
			name:     "ChargeStatusSuccess is not valid",
			status:   constant.ChargeStatusSuccess,
			expected: false,
		},
		{
			name:     "ChargeStatusFailed is not valid",
			status:   constant.ChargeStatusFailed,
			expected: false,
		},
		{
			name:     "Unknown status is not valid",
			status:   "UNKNOWN",
			expected: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := svc.isValidStatusForInvestigation(tc.status)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestIsValidPaymentMethodForInvestigation(t *testing.T) {
	svc := &UnifiedPaymentService{}

	tests := []struct {
		name              string
		paymentMethodType string
		expected          bool
	}{
		{
			name:              "QRIS is valid",
			paymentMethodType: paymentConst.PAYMENT_METHOD_QRIS,
			expected:          true,
		},
		{
			name:              "VIRTUAL_ACCOUNT is valid",
			paymentMethodType: paymentConst.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
			expected:          true,
		},
		{
			name:              "CREDIT_CARD is not valid",
			paymentMethodType: paymentConst.PAYMENT_METHOD_CREDIT_CARD,
			expected:          false,
		},
		{
			name:              "EWALLET is not valid",
			paymentMethodType: paymentConst.PAYMENT_METHOD_EWALLET,
			expected:          false,
		},
		{
			name:              "Unknown is not valid",
			paymentMethodType: "UNKNOWN",
			expected:          false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := svc.isValidPaymentMethodForInvestigation(tc.paymentMethodType)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestIsClientError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "Nil error returns false",
			err:      nil,
			expected: false,
		},
		{
			name:     "Regular error returns false",
			err:      errors.New("some error"),
			expected: false,
		},
		{
			name:     "HttpErrRequest is client error",
			err:      pkgErr.New(response.HttpErrRequest, errors.New("bad request")),
			expected: true,
		},
		{
			name:     "HttpErrInternal is not client error",
			err:      pkgErr.New(response.HttpErrInternal, errors.New("internal error")),
			expected: false,
		},
		{
			name:     "HttpErrDatabase is not client error",
			err:      pkgErr.New(response.HttpErrDatabase, errors.New("database error")),
			expected: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := isClientError(tc.err)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestUploadProofOfPayment(t *testing.T) {

	log, _ := pdkLogger.NewZapLogger(pdkLogger.Config{})
	gcs := gcsMock.NewGCSService(t)
	merchantSvc := serviceMocks.NewIMerchantService(t)
	fdsVelocityCheck := fdsMock.NewVelocityChecker(t)
	paymentRepo := repositoryMock.NewIPaymentRepository(t)
	accountTransactionRepo := repositoryMock.NewIAccountTransactionRepository(t)
	snapCoreRepo := repositoryMock.NewISnapCoreRepository(t)
	orchestratorSvc := serviceMocks.NewIOrchestratorService(t)
	paymentSvc := serviceMocks.NewIPaymentService(t)
	merchantRepo := repositoryMock.NewIMerchantRepository(t)
	redis := redisExtMocks.NewIRedisExt(t)
	redisMutex := redisExtMocks.NewIMutexer(t)

	// Shared mocks for the distributed lock acquired by ProcessNotification (invoked
	// indirectly via processNotificationFromInquiry). Non-static payments use SetNX;
	// static payments still use redsync Mutex. Uses Maybe() so cases that return
	// before reaching the lock are not required to match.
	redisSetNXResult := &redisSdk.BoolCmd{}
	redisSetNXResult.SetVal(true)
	redis.On("SetNX", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe().Return(redisSetNXResult)
	redis.On("NewMutex", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe().Return(redisMutex)
	redisMutex.On("LockContext", mock.Anything).Maybe().Return(nil)
	redisMutex.On("UnlockContext", mock.Anything).Maybe().Return(true, nil)

	svc := &UnifiedPaymentService{
		config:                 &config.Config{Environment: "test"},
		logger:                 log,
		storage:                gcs,
		paymentRepo:            paymentRepo,
		accountTransactionRepo: accountTransactionRepo,
		merchantSvc:            merchantSvc,
		snapCoreRepo:           snapCoreRepo,
		orchestratorSvc:        orchestratorSvc,
		paymentSvc:             paymentSvc,
		fdsVelocityCheck:       fdsVelocityCheck,
		merchantRepo:           merchantRepo,
		redis:                  redis,
	}

	proofFile := &multipart.FileHeader{Filename: "proof.png"}
	request := &unifiedPaymentModel.UploadProofOfPaymentRequest{
		PaymentID:      "3d238f37-cff2-4720-8a1a-c83f21a60bd9", // NOSONAR
		MerchantID:     "ef6e61f1-19ae-4604-a8fa-56e097e0fdf4", // NOSONAR
		ProofOfPayment: proofFile,
		FileExtension:  "png",
		Reason:         "customer paid",
	}
	newPaymentResult := func() *paymentModel.Payment {
		return &paymentModel.Payment{
			UUID:        request.PaymentID,
			MerchantID:  request.MerchantID,
			Status:      constant.ChargeStatusProcessing,
			Currency:    "IDR",
			Amount:      decimal.NewFromInt(100),
			TotalAmount: decimal.NewFromInt(100),
			SnapCoreId:  util.ValueToPtr("snap-core-123"),
			PaymentMethod: paymentModel.PaymentMethod{
				Type: paymentConst.PAYMENT_METHOD_QRIS,
			},
		}
	}

	newPaymentSession := func() *paymentModel.Payment {
		return &paymentModel.Payment{
			UUID:        request.PaymentID,
			MerchantID:  request.MerchantID,
			Status:      constant.ChargeStatusProcessing,
			Type:        constant.UnifiedPaymentTypeSingle,
			Currency:    "IDR",
			Amount:      decimal.NewFromInt(100),
			TotalAmount: decimal.NewFromInt(100),
			PaymentMethod: paymentModel.PaymentMethod{
				Type: paymentConst.PAYMENT_METHOD_QRIS,
			},
		}
	}

	basePaymentResult := newPaymentResult()
	velocityKey := fmt.Sprintf(constant.FDSVelocityMerchantUploadPoPKeyFmt, request.MerchantID)

	tests := []struct {
		name       string
		request    *unifiedPaymentModel.UploadProofOfPaymentRequest
		setupMocks func()
		wantErr    error
		wantResult *unifiedPaymentModel.UploadProofOfPaymentResponse
	}{
		{
			name: "ERROR:Get payment by id",
			request: &unifiedPaymentModel.UploadProofOfPaymentRequest{
				PaymentID:  "payment-123",  // NOSONAR
				MerchantID: "merchant-123", // NOSONAR
			},
			setupMocks: func() {
				paymentRepo.On("GetPaymentById", mock.Anything, "payment-123").Once().Return(nil, assert.AnError)
			},
			wantErr: pkgErr.New(response.HttpErrDatabase, assert.AnError),
		},
		{
			name: "ERROR:Payment not found",
			request: &unifiedPaymentModel.UploadProofOfPaymentRequest{
				PaymentID:  "payment-not-found", // NOSONAR
				MerchantID: "merchant-123",      // NOSONAR
			},
			setupMocks: func() {
				paymentRepo.On("GetPaymentById", mock.Anything, "payment-not-found").Once().Return(nil, nil)
			},
			wantErr: pkgErr.New(response.HttpErrNotFound, constant.ErrPaymentNotFound),
		},
		{
			name: "ERROR:Merchant mismatch",
			request: &unifiedPaymentModel.UploadProofOfPaymentRequest{
				PaymentID:  "payment-123",  // NOSONAR
				MerchantID: "merchant-123", // NOSONAR
			},
			setupMocks: func() {
				payment := &paymentModel.Payment{
					UUID:       "payment-123",    // NOSONAR
					MerchantID: "merchant-other", // NOSONAR
					Status:     constant.ChargeStatusProcessing,
				}
				paymentRepo.On("GetPaymentById", mock.Anything, "payment-123").Once().Return(payment, nil)
			},
			wantErr: pkgErr.New(response.HttpErrNotFound, constant.ErrPaymentNotFound),
		},
		{
			name: "ERROR:Payment already in final status",
			request: &unifiedPaymentModel.UploadProofOfPaymentRequest{
				PaymentID:  "payment-123",  // NOSONAR
				MerchantID: "merchant-123", // NOSONAR
			},
			setupMocks: func() {
				payment := &paymentModel.Payment{
					UUID:       "payment-123",  // NOSONAR
					MerchantID: "merchant-123", // NOSONAR
					Status:     constant.ChargeStatusSuccess,
					PaymentMethod: paymentModel.PaymentMethod{
						Type: paymentConst.PAYMENT_METHOD_QRIS,
					},
				}
				paymentRepo.On("GetPaymentById", mock.Anything, "payment-123").Once().Return(payment, nil)
			},
			wantErr: pkgErr.New(response.HttpErrRequest, constant.ErrPaymentAlreadyInFinalStatus),
		},
		{
			name: "ERROR:Invalid payment method",
			request: &unifiedPaymentModel.UploadProofOfPaymentRequest{
				PaymentID:  "payment-123",  // NOSONAR
				MerchantID: "merchant-123", // NOSONAR
			},
			setupMocks: func() {
				payment := &paymentModel.Payment{
					UUID:       "payment-123",  // NOSONAR
					MerchantID: "merchant-123", // NOSONAR
					Status:     constant.ChargeStatusProcessing,
					PaymentMethod: paymentModel.PaymentMethod{
						Type: paymentConst.PAYMENT_METHOD_CREDIT_CARD,
					},
				}
				paymentRepo.On("GetPaymentById", mock.Anything, "payment-123").Once().Return(payment, nil)
			},
			wantErr: pkgErr.New(response.HttpErrRequest, constant.ErrPaymentMethodNotAllowed),
		},
		{
			name: "ERROR:Check enabled payment investigation",
			request: &unifiedPaymentModel.UploadProofOfPaymentRequest{
				PaymentID:  "payment-123",  // NOSONAR
				MerchantID: "merchant-123", // NOSONAR
			},
			setupMocks: func() {
				payment := &paymentModel.Payment{
					UUID:       "payment-123",  // NOSONAR
					MerchantID: "merchant-123", // NOSONAR
					Status:     constant.ChargeStatusProcessing,
					SnapCoreId: util.ValueToPtr("snap-core-123"),
					PaymentMethod: paymentModel.PaymentMethod{
						Type: paymentConst.PAYMENT_METHOD_QRIS,
					},
				}
				paymentRepo.On("GetPaymentById", mock.Anything, "payment-123").Once().Return(payment, nil)
				merchantRepo.On("IsInvestigationFlowEnabled", mock.Anything, "merchant-123").Once().Return(false, assert.AnError)
			},
			wantErr: pkgErr.New(response.HttpErrDatabase, constant.ErrInternalServerForUser),
		},
		{
			name: "ERROR:Investigation not enabled",
			request: &unifiedPaymentModel.UploadProofOfPaymentRequest{
				PaymentID:  "payment-123",  // NOSONAR
				MerchantID: "merchant-123", // NOSONAR
			},
			setupMocks: func() {
				payment := &paymentModel.Payment{
					UUID:       "payment-123",  // NOSONAR
					MerchantID: "merchant-123", // NOSONAR
					Status:     constant.ChargeStatusProcessing,
					SnapCoreId: util.ValueToPtr("snap-core-123"),
					PaymentMethod: paymentModel.PaymentMethod{
						Type: paymentConst.PAYMENT_METHOD_QRIS,
					},
				}
				paymentRepo.On("GetPaymentById", mock.Anything, "payment-123").Once().Return(payment, nil)
				merchantRepo.On("IsInvestigationFlowEnabled", mock.Anything, "merchant-123").Once().Return(false, nil)
			},
			wantErr: pkgErr.New(response.HttpErrRequest, constant.ErrInvestigationNotEnabled),
		},
		{
			name: "ERROR:Get merchant FDS configuration",
			setupMocks: func() {
				paymentResult := newPaymentResult()
				paymentRepo.On("GetPaymentById", mock.Anything, request.PaymentID).Once().Return(paymentResult, nil)
				merchantSvc.On("GetFDSConfig", mock.Anything, request.MerchantID).Once().Return(nil, assert.AnError)
				merchantRepo.On("IsInvestigationFlowEnabled", mock.Anything, request.MerchantID).Return(true, nil)
			},
			wantErr: assert.AnError,
		},
		{
			name: "ERROR:Bank inquiry confirms success",
			setupMocks: func() {
				paymentResult := newPaymentResult()
				paymentSession := newPaymentSession()
				paymentRepo.On("GetPaymentById", mock.Anything, request.PaymentID).Once().Return(paymentResult, nil)
				merchantSvc.On("GetFDSConfig", mock.Anything, request.MerchantID).Once().Return(&merchant.GetFDSConfigResponse{
					MerchantID: request.MerchantID,
					FDSConfig: merchant.FDSConfig{
						ProofOfPayment: &merchant.FDSFeatureProofOfPayment{
							Velocity: merchant.FDSRuleVelocityConfig{
								Window:    merchant.FDSWindowConfig{Interval: 1, Unit: constant.WindowUnitSecond},
								Threshold: merchant.FDSThresholdConfig{Count: 1},
							},
						},
					},
				}, nil)
				snapCoreRepo.On("InquiryStatusQris", mock.Anything, &snapCoreQRModel.InquiryStatusQrMpmRequest{QrisUUID: "snap-core-123", SkipPublish: true}).Once().Return(&snapCoreQRModel.QrisInquiryStatusResponse{
					Data: &snapCoreQRModel.QrisInquiryStatusResponseData{
						ResponseCode: "00",
						Status:       constant.InquiryStatusSuccess,
					},
				}, nil)
				orchestratorSvc.On("FindByReference", mock.Anything, paymentResult.UUID, constant.TypePayment).Once().Return(nil, nil)
				paymentSvc.On("GetDetailByID", mock.Anything, paymentSession.UUID).Once().Return(paymentSession, nil)
				paymentRepo.On("BeginTransaction", mock.Anything).Once().Return(context.Background(), nil)
				paymentSvc.On("DeterminePaymentFee", mock.Anything, paymentSession).Once().Return(nil)
				paymentRepo.On("UpdatePaymentStatus", mock.Anything, paymentSession.UUID, paymentSession.MerchantID, constant.UnifiedPaymentSessionStatusPaid, mock.Anything).Once().Return(nil)
				accountTransactionRepo.On("FindByID", mock.Anything, mock.Anything).Once().Return(nil, nil)
				paymentSvc.On("PostCreateLedger", mock.Anything, paymentSession, mock.Anything).Once().Return(nil)
				paymentRepo.On("CommitTransaction", mock.Anything).Once().Return(nil)
				accountTransactionRepo.On("FindByReference", mock.Anything, paymentSession.UUID, constant.TypePayment).Once().Return(nil, nil)
				merchantRepo.On("FindMerchantByID", constant.ValueCtxMockType(), constant.StringMockType()).Return(nil, nil)
			},
			wantErr: pkgErr.New(response.HttpErrRequest, constant.ErrBankConfirmedSuccess),
		},
		{
			name: "ERROR:Bank inquiry confirms failed",
			setupMocks: func() {
				paymentResult := newPaymentResult()
				paymentSession := newPaymentSession()
				paymentRepo.On("GetPaymentById", mock.Anything, request.PaymentID).Once().Return(paymentResult, nil)
				merchantSvc.On("GetFDSConfig", mock.Anything, request.MerchantID).Once().Return(&merchant.GetFDSConfigResponse{
					MerchantID: request.MerchantID,
					FDSConfig: merchant.FDSConfig{
						ProofOfPayment: &merchant.FDSFeatureProofOfPayment{
							Velocity: merchant.FDSRuleVelocityConfig{
								Window:    merchant.FDSWindowConfig{Interval: 1, Unit: constant.WindowUnitSecond},
								Threshold: merchant.FDSThresholdConfig{Count: 1},
							},
						},
					},
				}, nil)
				snapCoreRepo.On("InquiryStatusQris", mock.Anything, &snapCoreQRModel.InquiryStatusQrMpmRequest{QrisUUID: "snap-core-123", SkipPublish: true}).Once().Return(&snapCoreQRModel.QrisInquiryStatusResponse{
					Data: &snapCoreQRModel.QrisInquiryStatusResponseData{
						ResponseCode: "00",
						Status:       constant.InquiryStatusFailed,
					},
				}, nil)
				orchestratorSvc.On("FindByReference", mock.Anything, paymentResult.UUID, constant.TypePayment).Once().Return(nil, nil)
				paymentSvc.On("GetDetailByID", mock.Anything, paymentSession.UUID).Once().Return(paymentSession, nil)
				paymentRepo.On("BeginTransaction", mock.Anything).Once().Return(context.Background(), nil)
				paymentRepo.On("UpdatePaymentStatus", mock.Anything, paymentSession.UUID, paymentSession.MerchantID, constant.UnifiedPaymentSessionStatusCancelled, mock.Anything).Once().Return(nil)
				accountTransactionRepo.On("FindByID", mock.Anything, mock.Anything).Once().Return(nil, nil)
				paymentRepo.On("CommitTransaction", mock.Anything).Once().Return(nil)
				accountTransactionRepo.On("FindByReference", mock.Anything, paymentSession.UUID, constant.TypePayment).Once().Return(nil, nil)
				merchantRepo.On("FindMerchantByID", constant.ValueCtxMockType(), constant.StringMockType()).Return(nil, nil)
			},
			wantErr: pkgErr.New(response.HttpErrRequest, constant.ErrBankConfirmedFailed),
		},
		{
			name: "SUCCESS:Bank inquiry expired continues flow",
			setupMocks: func() {
				paymentResult := newPaymentResult()
				paymentSession := newPaymentSession()
				paymentRepo.On("GetPaymentById", mock.Anything, request.PaymentID).Once().Return(paymentResult, nil)
				merchantSvc.On("GetFDSConfig", mock.Anything, request.MerchantID).Once().Return(&merchant.GetFDSConfigResponse{
					MerchantID: request.MerchantID,
					FDSConfig: merchant.FDSConfig{
						ProofOfPayment: &merchant.FDSFeatureProofOfPayment{
							Velocity: merchant.FDSRuleVelocityConfig{
								Window:    merchant.FDSWindowConfig{Interval: 1, Unit: constant.WindowUnitSecond},
								Threshold: merchant.FDSThresholdConfig{Count: 1},
							},
						},
					},
				}, nil)
				snapCoreRepo.On("InquiryStatusQris", mock.Anything, &snapCoreQRModel.InquiryStatusQrMpmRequest{QrisUUID: "snap-core-123", SkipPublish: true}).Once().Return(&snapCoreQRModel.QrisInquiryStatusResponse{
					Data: &snapCoreQRModel.QrisInquiryStatusResponseData{
						ResponseCode: "00",
						Status:       constant.InquiryStatusExpired,
					},
				}, nil)
				fdsVelocityCheck.On("Allow", mock.Anything, velocityKey, mock.Anything).Once().Return(&fds.VelocityResult{Allowed: true}, nil)
				gcs.On(
					"UploadFileFromMultipartToBucket", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
				).Once().Return(&gcsPkg.UploadMultipart{Bucket: "bucket", ObjectName: "object"}, nil)
				paymentRepo.On("UpdatePaymentForInvestigation", mock.Anything, mock.Anything).Once().Return(nil)
				paymentSvc.On("RecordPaymentStatusHistory", mock.Anything, paymentResult.UUID, constant.StatusHistoryActorUser, constant.PaymentStatusHistoryInvestigationInProcess).Once().Return()
				orchestratorSvc.On("FindByReference", mock.Anything, paymentResult.UUID, constant.TypePayment).Once().Return(nil, nil)
				paymentSvc.On("GetDetailByID", mock.Anything, paymentSession.UUID).Once().Return(paymentSession, nil)
				paymentRepo.On("BeginTransaction", mock.Anything).Once().Return(context.Background(), nil)
				paymentSvc.On("DeterminePaymentFee", mock.Anything, paymentSession).Once().Return(nil)
				paymentRepo.On("UpdatePaymentStatus", mock.Anything, paymentSession.UUID, paymentSession.MerchantID, constant.UnifiedPaymentSessionStatusPaid, mock.Anything).Once().Return(nil)
				accountTransactionRepo.On("FindByID", mock.Anything, mock.Anything).Once().Return(nil, nil)
				paymentSvc.On("PostCreateLedger", mock.Anything, paymentSession, mock.Anything).Once().Return(nil)
				paymentRepo.On("CommitTransaction", mock.Anything).Once().Return(nil)
				accountTransactionRepo.On("FindByReference", mock.Anything, paymentSession.UUID, constant.TypePayment).Once().Return(nil, nil)
			},
			wantResult: &unifiedPaymentModel.UploadProofOfPaymentResponse{
				PaymentID:           basePaymentResult.UUID,
				Status:              constant.ChargeStatusSuccess,
				InvestigationStatus: paymentConst.InvestigationStatusInProcess,
				CreatedAt:           basePaymentResult.CreatedAt,
			},
		},
		{
			name: "ERROR:Bank inquiry client error",
			setupMocks: func() {
				paymentResult := newPaymentResult()
				paymentRepo.On("GetPaymentById", mock.Anything, request.PaymentID).Once().Return(paymentResult, nil)
				merchantSvc.On("GetFDSConfig", mock.Anything, request.MerchantID).Once().Return(&merchant.GetFDSConfigResponse{
					MerchantID: request.MerchantID,
					FDSConfig: merchant.FDSConfig{
						ProofOfPayment: &merchant.FDSFeatureProofOfPayment{
							Velocity: merchant.FDSRuleVelocityConfig{
								Window:    merchant.FDSWindowConfig{Interval: 1, Unit: constant.WindowUnitSecond},
								Threshold: merchant.FDSThresholdConfig{Count: 1},
							},
						},
					},
				}, nil)
				snapCoreRepo.On("InquiryStatusQris", mock.Anything, &snapCoreQRModel.InquiryStatusQrMpmRequest{QrisUUID: "snap-core-123", SkipPublish: true}).Once().Return(
					nil, pkgErr.New(response.HttpErrRequest, assert.AnError),
				)
			},
			wantErr: pkgErr.New(response.HttpErrRequest, constant.ErrBankInquiryFailed),
		},
		{
			name: "SUCCESS:Bank inquiry server error continues flow",
			setupMocks: func() {
				paymentResult := newPaymentResult()
				paymentSession := newPaymentSession()
				paymentRepo.On("GetPaymentById", mock.Anything, request.PaymentID).Once().Return(paymentResult, nil)
				merchantSvc.On("GetFDSConfig", mock.Anything, request.MerchantID).Once().Return(&merchant.GetFDSConfigResponse{
					MerchantID: request.MerchantID,
					FDSConfig: merchant.FDSConfig{
						ProofOfPayment: &merchant.FDSFeatureProofOfPayment{
							Velocity: merchant.FDSRuleVelocityConfig{
								Window:    merchant.FDSWindowConfig{Interval: 1, Unit: constant.WindowUnitSecond},
								Threshold: merchant.FDSThresholdConfig{Count: 1},
							},
						},
					},
				}, nil)
				snapCoreRepo.On("InquiryStatusQris", mock.Anything, &snapCoreQRModel.InquiryStatusQrMpmRequest{QrisUUID: "snap-core-123", SkipPublish: true}).Once().Return(nil, pkgErr.New(response.HttpErrBadGateway, assert.AnError))
				fdsVelocityCheck.On("Allow", mock.Anything, velocityKey, mock.Anything).Once().Return(&fds.VelocityResult{Allowed: true}, nil)
				gcs.On(
					"UploadFileFromMultipartToBucket", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
				).Once().Return(&gcsPkg.UploadMultipart{Bucket: "bucket", ObjectName: "object"}, nil)
				paymentRepo.On("UpdatePaymentForInvestigation", mock.Anything, mock.Anything).Once().Return(nil)
				paymentSvc.On("RecordPaymentStatusHistory", mock.Anything, paymentResult.UUID, constant.StatusHistoryActorUser, constant.PaymentStatusHistoryInvestigationInProcess).Once().Return()
				orchestratorSvc.On("FindByReference", mock.Anything, paymentResult.UUID, constant.TypePayment).Once().Return(nil, nil)
				paymentSvc.On("GetDetailByID", mock.Anything, paymentSession.UUID).Once().Return(paymentSession, nil)
				paymentRepo.On("BeginTransaction", mock.Anything).Once().Return(context.Background(), nil)
				paymentSvc.On("DeterminePaymentFee", mock.Anything, paymentSession).Once().Return(nil)
				paymentRepo.On("UpdatePaymentStatus", mock.Anything, paymentSession.UUID, paymentSession.MerchantID, constant.UnifiedPaymentSessionStatusPaid, mock.Anything).Once().Return(nil)
				accountTransactionRepo.On("FindByID", mock.Anything, mock.Anything).Once().Return(nil, nil)
				paymentSvc.On("PostCreateLedger", mock.Anything, paymentSession, mock.Anything).Once().Return(nil)
				paymentRepo.On("CommitTransaction", mock.Anything).Once().Return(nil)
				accountTransactionRepo.On("FindByReference", mock.Anything, paymentSession.UUID, constant.TypePayment).Once().Return(nil, nil)
			},
			wantResult: &unifiedPaymentModel.UploadProofOfPaymentResponse{
				PaymentID:           basePaymentResult.UUID,
				Status:              constant.ChargeStatusSuccess,
				InvestigationStatus: paymentConst.InvestigationStatusInProcess,
				CreatedAt:           basePaymentResult.CreatedAt,
			},
		},
		{
			name: "SUCCESS:Bank inquiry pending",
			setupMocks: func() {
				paymentResult := newPaymentResult()
				paymentSession := newPaymentSession()
				paymentRepo.On("GetPaymentById", mock.Anything, request.PaymentID).Once().Return(paymentResult, nil)
				merchantSvc.On("GetFDSConfig", mock.Anything, request.MerchantID).Once().Return(&merchant.GetFDSConfigResponse{
					MerchantID: request.MerchantID,
					FDSConfig: merchant.FDSConfig{
						ProofOfPayment: &merchant.FDSFeatureProofOfPayment{
							Velocity: merchant.FDSRuleVelocityConfig{
								Window:    merchant.FDSWindowConfig{Interval: 1, Unit: constant.WindowUnitSecond},
								Threshold: merchant.FDSThresholdConfig{Count: 1},
							},
						},
					},
				}, nil)
				snapCoreRepo.On("InquiryStatusQris", mock.Anything, &snapCoreQRModel.InquiryStatusQrMpmRequest{QrisUUID: "snap-core-123", SkipPublish: true}).Once().Return(&snapCoreQRModel.QrisInquiryStatusResponse{
					Data: &snapCoreQRModel.QrisInquiryStatusResponseData{
						ResponseCode: "00",
						Status:       constant.InquiryStatusPending,
					},
				}, nil)
				fdsVelocityCheck.On("Allow", mock.Anything, velocityKey, mock.Anything).Once().Return(&fds.VelocityResult{Allowed: true}, nil)
				gcs.On(
					"UploadFileFromMultipartToBucket", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
				).Once().Return(&gcsPkg.UploadMultipart{Bucket: "bucket", ObjectName: "object"}, nil)
				paymentRepo.On("UpdatePaymentForInvestigation", mock.Anything, mock.Anything).Once().Return(nil)
				paymentSvc.On("RecordPaymentStatusHistory", mock.Anything, paymentResult.UUID, constant.StatusHistoryActorUser, constant.PaymentStatusHistoryInvestigationInProcess).Once().Return()
				orchestratorSvc.On("FindByReference", mock.Anything, paymentResult.UUID, constant.TypePayment).Once().Return(nil, nil)
				paymentSvc.On("GetDetailByID", mock.Anything, paymentSession.UUID).Once().Return(paymentSession, nil)
				paymentRepo.On("BeginTransaction", mock.Anything).Once().Return(context.Background(), nil)
				paymentSvc.On("DeterminePaymentFee", mock.Anything, paymentSession).Once().Return(nil)
				paymentRepo.On("UpdatePaymentStatus", mock.Anything, paymentSession.UUID, paymentSession.MerchantID, constant.UnifiedPaymentSessionStatusPaid, mock.Anything).Once().Return(nil)
				accountTransactionRepo.On("FindByID", mock.Anything, mock.Anything).Once().Return(nil, nil)
				paymentSvc.On("PostCreateLedger", mock.Anything, paymentSession, mock.Anything).Once().Return(nil)
				paymentRepo.On("CommitTransaction", mock.Anything).Once().Return(nil)
				accountTransactionRepo.On("FindByReference", mock.Anything, paymentSession.UUID, constant.TypePayment).Once().Return(nil, nil)
			},
			wantResult: &unifiedPaymentModel.UploadProofOfPaymentResponse{
				PaymentID:           basePaymentResult.UUID,
				Status:              constant.ChargeStatusSuccess,
				InvestigationStatus: paymentConst.InvestigationStatusInProcess,
				CreatedAt:           basePaymentResult.CreatedAt,
			},
		},
		{
			name: "ERROR:FDS velocity check",
			setupMocks: func() {
				paymentResult := newPaymentResult()
				paymentRepo.On("GetPaymentById", mock.Anything, request.PaymentID).Once().Return(paymentResult, nil)
				merchantSvc.On("GetFDSConfig", mock.Anything, request.MerchantID).Once().Return(&merchant.GetFDSConfigResponse{
					MerchantID: request.MerchantID,
					FDSConfig: merchant.FDSConfig{
						ProofOfPayment: &merchant.FDSFeatureProofOfPayment{},
					},
				}, nil)
				snapCoreRepo.On("InquiryStatusQris", mock.Anything, &snapCoreQRModel.InquiryStatusQrMpmRequest{QrisUUID: "snap-core-123", SkipPublish: true}).Once().Return(nil, nil)
				orchestratorSvc.On("FindByReference", mock.Anything, paymentResult.UUID, constant.TypePayment).Once().Return(nil, nil)
				fdsVelocityCheck.On("Allow", mock.Anything, velocityKey, mock.Anything).Once().Return(nil, assert.AnError)
			},
			wantErr: pkgErr.New(response.HttpErrDatabase, constant.ErrInternalServerForUser),
		},
		{
			name: "ERROR:Request not allowed",
			setupMocks: func() {
				paymentResult := newPaymentResult()
				paymentRepo.On("GetPaymentById", mock.Anything, request.PaymentID).Once().Return(paymentResult, nil)
				merchantSvc.On("GetFDSConfig", mock.Anything, request.MerchantID).Once().Return(&merchant.GetFDSConfigResponse{
					MerchantID: request.MerchantID,
					FDSConfig: merchant.FDSConfig{
						ProofOfPayment: &merchant.FDSFeatureProofOfPayment{
							Velocity: merchant.FDSRuleVelocityConfig{
								Window:    merchant.FDSWindowConfig{Interval: 1, Unit: constant.WindowUnitSecond},
								Threshold: merchant.FDSThresholdConfig{Count: 1},
							},
						},
					},
				}, nil)
				snapCoreRepo.On("InquiryStatusQris", mock.Anything, &snapCoreQRModel.InquiryStatusQrMpmRequest{QrisUUID: "snap-core-123", SkipPublish: true}).Once().Return(nil, nil)
				orchestratorSvc.On("FindByReference", mock.Anything, paymentResult.UUID, constant.TypePayment).Once().Return(nil, nil)
				fdsVelocityCheck.On(
					"Allow", mock.Anything, velocityKey, mock.Anything,
				).Once().Return(&fds.VelocityResult{}, nil)
			},
			wantErr: pkgErr.New(response.HttpErrTooManyRequest, constant.ErrProofOfPaymentRateLimitExceeded),
		},
		{
			name: "ERROR:Failed upload proof of payment",
			setupMocks: func() {
				paymentResult := newPaymentResult()
				paymentRepo.On("GetPaymentById", mock.Anything, request.PaymentID).Once().Return(paymentResult, nil)
				merchantSvc.On("GetFDSConfig", mock.Anything, request.MerchantID).Once().Return(&merchant.GetFDSConfigResponse{
					MerchantID: request.MerchantID,
					FDSConfig: merchant.FDSConfig{
						ProofOfPayment: &merchant.FDSFeatureProofOfPayment{
							Velocity: merchant.FDSRuleVelocityConfig{
								Window:    merchant.FDSWindowConfig{Interval: 1, Unit: constant.WindowUnitSecond},
								Threshold: merchant.FDSThresholdConfig{Count: 1},
							},
						},
					},
				}, nil)
				snapCoreRepo.On("InquiryStatusQris", mock.Anything, &snapCoreQRModel.InquiryStatusQrMpmRequest{QrisUUID: "snap-core-123", SkipPublish: true}).Once().Return(nil, nil)
				orchestratorSvc.On("FindByReference", mock.Anything, paymentResult.UUID, constant.TypePayment).Once().Return(nil, nil)
				fdsVelocityCheck.On("Allow", mock.Anything, velocityKey, mock.Anything).Once().Return(&fds.VelocityResult{Allowed: true}, nil)
				gcs.On(
					"UploadFileFromMultipartToBucket", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
				).Once().Return(nil, constant.ErrSomeErrorForUnitTest)
				fdsVelocityCheck.On("Rollback", mock.Anything, velocityKey, paymentResult.UUID).Once().Return(assert.AnError)
			},
			wantErr: pkgErr.New(response.HttpErrDatabase, constant.ErrSomeErrorForUnitTest),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.setupMocks != nil {
				tc.setupMocks()
			}
			if tc.request == nil {
				tc.request = request
			}
			result, err := svc.UploadProofOfPayment(t.Context(), tc.request)
			assert.Equal(t, tc.wantErr, err)
			if tc.wantResult == nil {
				assert.Nil(t, result)
			} else {
				if assert.NotNil(t, result) {
					assert.Equal(t, tc.wantResult.PaymentID, result.PaymentID)
					assert.Equal(t, tc.wantResult.Status, result.Status)
					assert.Equal(t, tc.wantResult.InvestigationStatus, result.InvestigationStatus)
					assert.Equal(t, tc.wantResult.CreatedAt, result.CreatedAt)
					assert.False(t, result.UpdatedAt.IsZero())
				}
			}
		})
	}
}
