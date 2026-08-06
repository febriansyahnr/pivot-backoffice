package unifiedPaymentService

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	paymentConst "github.com/paper-indonesia/pivot-backoffice/constant/payment"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	orchestratorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	snapCoreQRModel "github.com/paper-indonesia/pivot-backoffice/internal/model/snapCore/qr"
	snapCoreVAModel "github.com/paper-indonesia/pivot-backoffice/internal/model/snapCore/virtualAccount"
	unifiedPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/unifiedPayment"
	redisMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/redisExt"
	repositoryMock "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	mockLogger "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/redis/go-redis/v9"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestIsFinalStatus(t *testing.T) {
	svc := &UnifiedPaymentService{}

	tests := []struct {
		name     string
		status   string
		expected bool
	}{
		{
			name:     "InquiryStatusSuccess is final",
			status:   constant.InquiryStatusSuccess,
			expected: true,
		},
		{
			name:     "InquiryStatusFailed is final",
			status:   constant.InquiryStatusFailed,
			expected: true,
		},
		{
			name:     "InquiryStatusExpired is final",
			status:   constant.InquiryStatusExpired,
			expected: true,
		},
		{
			name:     "ChargeStatusSuccess is final",
			status:   constant.ChargeStatusSuccess,
			expected: true,
		},
		{
			name:     "ChargeStatusFailed is final",
			status:   constant.ChargeStatusFailed,
			expected: true,
		},
		{
			name:     "ChargeStatusExpired is final",
			status:   constant.ChargeStatusExpired,
			expected: true,
		},
		{
			name:     "InquiryStatusPending is not final",
			status:   constant.InquiryStatusPending,
			expected: false,
		},
		{
			name:     "ChargeStatusProcessing is not final",
			status:   constant.ChargeStatusProcessing,
			expected: false,
		},
		{
			name:     "ChargeStatusWaitingForUserAction is not final",
			status:   constant.ChargeStatusWaitingForUserAction,
			expected: false,
		},
		{
			name:     "Unknown status is not final",
			status:   "UNKNOWN",
			expected: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := svc.isFinalStatus(tc.status)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestMapInquiryStatusToChargeStatus(t *testing.T) {
	svc := &UnifiedPaymentService{}

	tests := []struct {
		name           string
		inquiryStatus  string
		expectedCharge string
	}{
		{
			name:           "SUCCESS maps to ChargeStatusSuccess",
			inquiryStatus:  constant.InquiryStatusSuccess,
			expectedCharge: constant.ChargeStatusSuccess,
		},
		{
			name:           "FAILED maps to ChargeStatusFailed",
			inquiryStatus:  constant.InquiryStatusFailed,
			expectedCharge: constant.ChargeStatusFailed,
		},
		{
			name:           "EXPIRED maps to ChargeStatusExpired",
			inquiryStatus:  constant.InquiryStatusExpired,
			expectedCharge: constant.ChargeStatusExpired,
		},
		{
			name:           "PENDING maps to ChargeStatusProcessing",
			inquiryStatus:  constant.InquiryStatusPending,
			expectedCharge: constant.ChargeStatusProcessing,
		},
		{
			name:           "Unknown status returns as is",
			inquiryStatus:  "CUSTOM_STATUS",
			expectedCharge: "CUSTOM_STATUS",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := svc.mapInquiryStatusToChargeStatus(tc.inquiryStatus)
			assert.Equal(t, tc.expectedCharge, result)
		})
	}
}

func TestMapQrisInquiryResponse(t *testing.T) {
	svc := &UnifiedPaymentService{}

	tests := []struct {
		name     string
		response *snapCoreQRModel.QrisInquiryStatusResponse
		expected *unifiedPaymentModel.InquiryResult
	}{
		{
			name:     "Nil response returns nil",
			response: nil,
			expected: nil,
		},
		{
			name: "Nil data returns nil",
			response: &snapCoreQRModel.QrisInquiryStatusResponse{
				Data: nil,
			},
			expected: nil,
		},
		{
			name: "Success status with amount",
			response: &snapCoreQRModel.QrisInquiryStatusResponse{
				Data: &snapCoreQRModel.QrisInquiryStatusResponseData{
					ResponseCode:        "00",
					ResponseMessage:     "Success",
					UUID:                "qris-uuid-123",
					TransactionID:       "trx-123",
					AcquirerReferenceNo: "acq-ref-123",
					Status:              constant.InquiryStatusSuccess,
					Amount: &snapCoreQRModel.InquiryStatusAmount{
						Value:    "100000.00",
						Currency: "IDR",
					},
				},
			},
			expected: &unifiedPaymentModel.InquiryResult{
				Status:                 constant.InquiryStatusSuccess,
				ResponseCode:           "00",
				ResponseMessage:        "Success",
				ProcessorID:            "qris-uuid-123",
				ProcessorTransactionID: "trx-123",
				ProcessorReferenceNo:   "acq-ref-123",
				Amount: &unifiedPaymentModel.Amount{
					Value:    100000.00,
					Currency: "IDR",
				},
				UpdatedStatus: true,
			},
		},
		{
			name: "Pending status without amount",
			response: &snapCoreQRModel.QrisInquiryStatusResponse{
				Data: &snapCoreQRModel.QrisInquiryStatusResponseData{
					ResponseCode:    "01",
					ResponseMessage: "Pending",
					UUID:            "qris-uuid-456",
					Status:          constant.InquiryStatusPending,
				},
			},
			expected: &unifiedPaymentModel.InquiryResult{
				Status:          constant.InquiryStatusPending,
				ResponseCode:    "01",
				ResponseMessage: "Pending",
				ProcessorID:     "qris-uuid-456",
				UpdatedStatus:   false,
			},
		},
		{
			name: "Invalid amount value returns nil amount",
			response: &snapCoreQRModel.QrisInquiryStatusResponse{
				Data: &snapCoreQRModel.QrisInquiryStatusResponseData{
					ResponseCode: "00",
					Status:       constant.InquiryStatusSuccess,
					Amount: &snapCoreQRModel.InquiryStatusAmount{
						Value:    "invalid",
						Currency: "IDR",
					},
				},
			},
			expected: &unifiedPaymentModel.InquiryResult{
				Status:        constant.InquiryStatusSuccess,
				ResponseCode:  "00",
				UpdatedStatus: true,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := svc.mapQrisInquiryResponse(tc.response)
			if tc.expected == nil {
				assert.Nil(t, result)
			} else {
				assert.Equal(t, tc.expected.Status, result.Status)
				assert.Equal(t, tc.expected.ResponseCode, result.ResponseCode)
				assert.Equal(t, tc.expected.ResponseMessage, result.ResponseMessage)
				assert.Equal(t, tc.expected.ProcessorID, result.ProcessorID)
				assert.Equal(t, tc.expected.ProcessorTransactionID, result.ProcessorTransactionID)
				assert.Equal(t, tc.expected.ProcessorReferenceNo, result.ProcessorReferenceNo)
				assert.Equal(t, tc.expected.UpdatedStatus, result.UpdatedStatus)
				if tc.expected.Amount != nil {
					assert.NotNil(t, result.Amount)
					assert.Equal(t, tc.expected.Amount.Value, result.Amount.Value)
					assert.Equal(t, tc.expected.Amount.Currency, result.Amount.Currency)
				} else {
					assert.Nil(t, result.Amount)
				}
			}
		})
	}
}

func TestMapVAInquiryResponse(t *testing.T) {
	svc := &UnifiedPaymentService{}

	tests := []struct {
		name     string
		response *snapCoreVAModel.InquiryStatusVAResponse
		validate func(t *testing.T, result *unifiedPaymentModel.InquiryResult)
	}{
		{
			name:     "Nil response returns nil",
			response: nil,
			validate: func(t *testing.T, result *unifiedPaymentModel.InquiryResult) {
				assert.Nil(t, result)
			},
		},
		{
			name: "Paid VA returns SUCCESS status",
			response: &snapCoreVAModel.InquiryStatusVAResponse{
				Data: snapCoreVAModel.InquiryStatusVAResponseData{
					ResponseCode:    "2002400",
					ResponseMessage: "Success",
					VirtualAccountData: &snapCoreVAModel.InquiryStatusVAData{
						PaidAmount: &commonModel.Amount{
							Value:    "150000.00",
							Currency: "IDR",
						},
						PaymentRequestId: "payment-req-123",
						ReferenceNo:      "ref-123",
						TrxDateTime:      "2025-01-20T10:30:00Z",
					},
				},
			},
			validate: func(t *testing.T, result *unifiedPaymentModel.InquiryResult) {
				assert.Equal(t, constant.InquiryStatusSuccess, result.Status)
				assert.Equal(t, "2002400", result.ResponseCode)
				assert.Equal(t, "Success", result.ResponseMessage)
				assert.NotNil(t, result.Amount)
				assert.Equal(t, 150000.00, result.Amount.Value)
				assert.Equal(t, "IDR", result.Amount.Currency)
				assert.Equal(t, "payment-req-123", result.ProcessorTransactionID)
				assert.Equal(t, "ref-123", result.ProcessorReferenceNo)
				assert.NotNil(t, result.TrxDatetime)
				assert.True(t, result.UpdatedStatus)
			},
		},
		{
			name: "Not found VA returns PENDING status",
			response: &snapCoreVAModel.InquiryStatusVAResponse{
				Data: snapCoreVAModel.InquiryStatusVAResponseData{
					ResponseCode:    "4042412",
					ResponseMessage: "Not Found",
				},
			},
			validate: func(t *testing.T, result *unifiedPaymentModel.InquiryResult) {
				assert.Equal(t, constant.InquiryStatusPending, result.Status)
				assert.Equal(t, "4042412", result.ResponseCode)
				assert.False(t, result.UpdatedStatus)
			},
		},
		{
			name: "Conflict VA returns SUCCESS status",
			response: &snapCoreVAModel.InquiryStatusVAResponse{
				Data: snapCoreVAModel.InquiryStatusVAResponseData{
					ResponseCode:    "4092400",
					ResponseMessage: "Conflict",
				},
			},
			validate: func(t *testing.T, result *unifiedPaymentModel.InquiryResult) {
				assert.Equal(t, constant.InquiryStatusSuccess, result.Status)
				assert.True(t, result.UpdatedStatus)
			},
		},
		{
			name: "Other response code returns PENDING status",
			response: &snapCoreVAModel.InquiryStatusVAResponse{
				Data: snapCoreVAModel.InquiryStatusVAResponseData{
					ResponseCode:    "5002400",
					ResponseMessage: "Processing",
				},
			},
			validate: func(t *testing.T, result *unifiedPaymentModel.InquiryResult) {
				assert.Equal(t, constant.InquiryStatusPending, result.Status)
				assert.False(t, result.UpdatedStatus)
			},
		},
		{
			name: "Invalid TrxDateTime format is ignored",
			response: &snapCoreVAModel.InquiryStatusVAResponse{
				Data: snapCoreVAModel.InquiryStatusVAResponseData{
					ResponseCode: "2002400",
					VirtualAccountData: &snapCoreVAModel.InquiryStatusVAData{
						TrxDateTime: "invalid-date",
					},
				},
			},
			validate: func(t *testing.T, result *unifiedPaymentModel.InquiryResult) {
				assert.Nil(t, result.TrxDatetime)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := svc.mapVAInquiryResponse(tc.response)
			tc.validate(t, result)
		})
	}
}

func TestGetVANumberFromPayment(t *testing.T) {
	svc := &UnifiedPaymentService{}

	tests := []struct {
		name     string
		payment  *paymentModel.Payment
		expected string
	}{
		{
			name: "ProcessorReferenceNumber takes priority",
			payment: &paymentModel.Payment{
				ProcessorReferenceNumber: strPtr("1234567890"),
				Metadata: &map[string]any{
					"snapCore": map[string]any{
						"number": "9876543210",
					},
				},
			},
			expected: "1234567890",
		},
		{
			name: "Falls back to metadata snapCore number",
			payment: &paymentModel.Payment{
				ProcessorReferenceNumber: nil,
				Metadata: &map[string]any{
					"snapCore": map[string]any{
						"number": "9876543210",
					},
				},
			},
			expected: "9876543210",
		},
		{
			name: "Empty ProcessorReferenceNumber falls back to metadata",
			payment: &paymentModel.Payment{
				ProcessorReferenceNumber: strPtr(""),
				Metadata: &map[string]any{
					"snapCore": map[string]any{
						"number": "9876543210",
					},
				},
			},
			expected: "9876543210",
		},
		{
			name: "Returns empty when no VA number found",
			payment: &paymentModel.Payment{
				ProcessorReferenceNumber: nil,
				Metadata:                 nil,
			},
			expected: "",
		},
		{
			name: "Returns empty when metadata has no snapCore",
			payment: &paymentModel.Payment{
				Metadata: &map[string]any{
					"other": "data",
				},
			},
			expected: "",
		},
		{
			name: "Returns empty when snapCore has no number",
			payment: &paymentModel.Payment{
				Metadata: &map[string]any{
					"snapCore": map[string]any{
						"uuid": "some-uuid",
					},
				},
			},
			expected: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := svc.getVANumberFromPayment(tc.payment)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestBuildNotificationRequestFromInquiry(t *testing.T) {
	svc := &UnifiedPaymentService{}
	trxTime := time.Date(2025, 1, 20, 10, 30, 0, 0, time.UTC)

	tests := []struct {
		name          string
		payment       *paymentModel.Payment
		inquiryResult *unifiedPaymentModel.InquiryResult
		chargeID      string
		validate      func(t *testing.T, req *unifiedPaymentModel.PaymentNotificationRequest)
	}{
		{
			name: "Build request with all inquiry result fields",
			payment: &paymentModel.Payment{
				UUID:   "payment-uuid-123",
				Amount: decimal.NewFromFloat(100000),
				PaymentMethod: paymentModel.PaymentMethod{
					Type: paymentConst.PAYMENT_METHOD_QRIS,
				},
			},
			inquiryResult: &unifiedPaymentModel.InquiryResult{
				Status:                 constant.InquiryStatusSuccess,
				ProcessorID:            "processor-123",
				ProcessorTransactionID: "trx-123",
				ProcessorReferenceNo:   "ref-123",
				Amount: &unifiedPaymentModel.Amount{
					Value:    150000,
					Currency: "IDR",
				},
				TrxDatetime: &trxTime,
			},
			chargeID: "charge-uuid-123",
			validate: func(t *testing.T, req *unifiedPaymentModel.PaymentNotificationRequest) {
				assert.Equal(t, "payment-uuid-123", req.PaymentSessionID)
				assert.Equal(t, paymentConst.PAYMENT_METHOD_QRIS, req.PaymentMethodType)
				assert.Equal(t, "charge-uuid-123", req.ChargeID)
				assert.Equal(t, constant.ChargeStatusSuccess, req.ChargeStatus)
				assert.Equal(t, constant.SnapCoreProcessor, req.Processor)
				assert.Equal(t, "processor-123", req.ProcessorID)
				assert.Equal(t, "trx-123", req.ProcessorTransactionID)
				assert.Equal(t, "ref-123", req.ProcessorReferenceNumber)
				assert.Equal(t, 150000.0, req.Amount.Value)
				assert.Equal(t, "IDR", req.Amount.Currency)
				assert.Equal(t, trxTime, req.TrxDatetime)
			},
		},
		{
			name: "Uses payment amount when inquiry result has no amount",
			payment: &paymentModel.Payment{
				UUID:   "payment-uuid-456",
				Amount: decimal.NewFromFloat(200000),
				PaymentMethod: paymentModel.PaymentMethod{
					Type: paymentConst.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
				},
			},
			inquiryResult: &unifiedPaymentModel.InquiryResult{
				Status: constant.InquiryStatusSuccess,
				Amount: nil,
			},
			chargeID: "charge-uuid-456",
			validate: func(t *testing.T, req *unifiedPaymentModel.PaymentNotificationRequest) {
				assert.Equal(t, 200000.0, req.Amount.Value)
				assert.Equal(t, constant.CurrencyIDR, req.Amount.Currency)
			},
		},
		{
			name: "Uses current time when inquiry result has no TrxDatetime",
			payment: &paymentModel.Payment{
				UUID:   "payment-uuid-789",
				Amount: decimal.NewFromFloat(50000),
				PaymentMethod: paymentModel.PaymentMethod{
					Type: paymentConst.PAYMENT_METHOD_QRIS,
				},
			},
			inquiryResult: &unifiedPaymentModel.InquiryResult{
				Status:      constant.InquiryStatusSuccess,
				TrxDatetime: nil,
			},
			chargeID: "",
			validate: func(t *testing.T, req *unifiedPaymentModel.PaymentNotificationRequest) {
				assert.False(t, req.TrxDatetime.IsZero())
			},
		},
		{
			name: "Maps PENDING status to PROCESSING charge status",
			payment: &paymentModel.Payment{
				UUID:   "payment-uuid-pending",
				Amount: decimal.NewFromFloat(100000),
				PaymentMethod: paymentModel.PaymentMethod{
					Type: paymentConst.PAYMENT_METHOD_QRIS,
				},
			},
			inquiryResult: &unifiedPaymentModel.InquiryResult{
				Status: constant.InquiryStatusPending,
			},
			chargeID: "charge-pending",
			validate: func(t *testing.T, req *unifiedPaymentModel.PaymentNotificationRequest) {
				assert.Equal(t, constant.ChargeStatusProcessing, req.ChargeStatus)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := svc.buildNotificationRequestFromInquiry(tc.payment, tc.inquiryResult, tc.chargeID)
			tc.validate(t, result)
		})
	}
}

func TestIsWithinCooldown(t *testing.T) {
	ctx := context.Background()
	log, _ := mockLogger.NewZapLogger(mockLogger.Config{})

	tests := []struct {
		name        string
		setupMock   func(*redisMock.IRedisExt)
		redisNil    bool
		paymentUUID string
		expected    bool
	}{
		{
			name:        "Returns false when redis is nil",
			redisNil:    true,
			paymentUUID: "payment-123",
			expected:    false,
		},
		{
			name: "Returns true when cooldown key exists",
			setupMock: func(m *redisMock.IRedisExt) {
				m.On("Exists", mock.Anything, "backend-portal:inquiry:cooldown:payment-123").
					Return(redis.NewIntResult(1, nil))
			},
			paymentUUID: "payment-123",
			expected:    true,
		},
		{
			name: "Returns false when cooldown key does not exist",
			setupMock: func(m *redisMock.IRedisExt) {
				m.On("Exists", mock.Anything, "backend-portal:inquiry:cooldown:payment-456").
					Return(redis.NewIntResult(0, nil))
			},
			paymentUUID: "payment-456",
			expected:    false,
		},
		{
			name: "Returns false on redis error",
			setupMock: func(m *redisMock.IRedisExt) {
				m.On("Exists", mock.Anything, "backend-portal:inquiry:cooldown:payment-789").
					Return(redis.NewIntResult(0, errors.New("redis error")))
			},
			paymentUUID: "payment-789",
			expected:    false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var svc *UnifiedPaymentService
			if tc.redisNil {
				svc = &UnifiedPaymentService{
					logger: log,
					redis:  nil,
				}
			} else {
				redisMockInstance := redisMock.NewIRedisExt(t)
				tc.setupMock(redisMockInstance)
				svc = &UnifiedPaymentService{
					logger: log,
					redis:  redisMockInstance,
				}
			}

			result := svc.isWithinCooldown(ctx, tc.paymentUUID)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestSetCooldown(t *testing.T) {
	ctx := context.Background()
	log, _ := mockLogger.NewZapLogger(mockLogger.Config{})

	tests := []struct {
		name            string
		setupMock       func(*redisMock.IRedisExt)
		redisNil        bool
		paymentUUID     string
		cooldownSeconds int
	}{
		{
			name:            "Does nothing when redis is nil",
			redisNil:        true,
			paymentUUID:     "payment-123",
			cooldownSeconds: 30,
		},
		{
			name: "Sets cooldown key with TTL",
			setupMock: func(m *redisMock.IRedisExt) {
				m.On("Set", mock.Anything, "backend-portal:inquiry:cooldown:payment-123", mock.AnythingOfType("string"), 30*time.Second).
					Return(redis.NewStatusResult("OK", nil))
			},
			paymentUUID:     "payment-123",
			cooldownSeconds: 30,
		},
		{
			name: "Logs warning on redis error but does not panic",
			setupMock: func(m *redisMock.IRedisExt) {
				m.On("Set", mock.Anything, "backend-portal:inquiry:cooldown:payment-456", mock.AnythingOfType("string"), 60*time.Second).
					Return(redis.NewStatusResult("", errors.New("redis error")))
			},
			paymentUUID:     "payment-456",
			cooldownSeconds: 60,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var svc *UnifiedPaymentService
			if tc.redisNil {
				svc = &UnifiedPaymentService{
					logger: log,
					redis:  nil,
				}
			} else {
				redisMockInstance := redisMock.NewIRedisExt(t)
				tc.setupMock(redisMockInstance)
				svc = &UnifiedPaymentService{
					logger: log,
					redis:  redisMockInstance,
				}
			}

			svc.setCooldown(ctx, tc.paymentUUID, tc.cooldownSeconds)
		})
	}
}

func TestPerformQrisInquiry(t *testing.T) {
	ctx := context.Background()
	log, _ := mockLogger.NewZapLogger(mockLogger.Config{})
	cfg := &config.Config{Environment: "test"}

	tests := []struct {
		name      string
		payment   *paymentModel.Payment
		setupMock func(*repositoryMock.ISnapCoreRepository, *repositoryMock.IAccountTransactionRepository)
		validate  func(t *testing.T, result *unifiedPaymentModel.InquiryResult, err error)
	}{
		{
			name: "Returns nil when no snap core ID found",
			payment: &paymentModel.Payment{
				UUID:       "payment-123",
				SnapCoreId: nil,
			},
			setupMock: func(snapCore *repositoryMock.ISnapCoreRepository, accTrx *repositoryMock.IAccountTransactionRepository) {
				accTrx.On("FindByReference", mock.Anything, "payment-123", constant.TypePayment).
					Return(nil, nil)
			},
			validate: func(t *testing.T, result *unifiedPaymentModel.InquiryResult, err error) {
				assert.Nil(t, result)
				assert.Nil(t, err)
			},
		},
		{
			name: "Returns error when snap core inquiry fails",
			payment: &paymentModel.Payment{
				UUID:       "payment-456",
				SnapCoreId: strPtr("snap-core-456"),
			},
			setupMock: func(snapCore *repositoryMock.ISnapCoreRepository, accTrx *repositoryMock.IAccountTransactionRepository) {
				snapCore.On("InquiryStatusQris", mock.Anything, &snapCoreQRModel.InquiryStatusQrMpmRequest{QrisUUID: "snap-core-456", SkipPublish: true}).
					Return(nil, errors.New("API error"))
			},
			validate: func(t *testing.T, result *unifiedPaymentModel.InquiryResult, err error) {
				assert.Nil(t, result)
				assert.Error(t, err)
			},
		},
		{
			name: "Returns nil when snap core returns empty response",
			payment: &paymentModel.Payment{
				UUID:       "payment-789",
				SnapCoreId: strPtr("snap-core-789"),
			},
			setupMock: func(snapCore *repositoryMock.ISnapCoreRepository, accTrx *repositoryMock.IAccountTransactionRepository) {
				snapCore.On("InquiryStatusQris", mock.Anything, &snapCoreQRModel.InquiryStatusQrMpmRequest{QrisUUID: "snap-core-789", SkipPublish: true}).
					Return(&snapCoreQRModel.QrisInquiryStatusResponse{Data: nil}, nil)
			},
			validate: func(t *testing.T, result *unifiedPaymentModel.InquiryResult, err error) {
				assert.Nil(t, result)
				assert.Nil(t, err)
			},
		},
		{
			name: "Returns mapped result on success",
			payment: &paymentModel.Payment{
				UUID:       "payment-success",
				SnapCoreId: strPtr("snap-core-success"),
			},
			setupMock: func(snapCore *repositoryMock.ISnapCoreRepository, accTrx *repositoryMock.IAccountTransactionRepository) {
				snapCore.On("InquiryStatusQris", mock.Anything, &snapCoreQRModel.InquiryStatusQrMpmRequest{QrisUUID: "snap-core-success", SkipPublish: true}).
					Return(&snapCoreQRModel.QrisInquiryStatusResponse{
						Data: &snapCoreQRModel.QrisInquiryStatusResponseData{
							ResponseCode: "00",
							Status:       constant.InquiryStatusSuccess,
							UUID:         "processor-uuid",
						},
					}, nil)
			},
			validate: func(t *testing.T, result *unifiedPaymentModel.InquiryResult, err error) {
				assert.NotNil(t, result)
				assert.Nil(t, err)
				assert.Equal(t, constant.InquiryStatusSuccess, result.Status)
				assert.Equal(t, "processor-uuid", result.ProcessorID)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			snapCoreMock := repositoryMock.NewISnapCoreRepository(t)
			accTrxMock := repositoryMock.NewIAccountTransactionRepository(t)
			tc.setupMock(snapCoreMock, accTrxMock)

			svc := &UnifiedPaymentService{
				config:                 cfg,
				logger:                 log,
				snapCoreRepo:           snapCoreMock,
				accountTransactionRepo: accTrxMock,
			}

			result, err := svc.performQrisInquiry(ctx, tc.payment)
			tc.validate(t, result, err)
		})
	}
}

func TestPerformVAInquiry(t *testing.T) {
	ctx := context.Background()
	log, _ := mockLogger.NewZapLogger(mockLogger.Config{})
	cfg := &config.Config{Environment: "test"}

	tests := []struct {
		name         string
		payment      *paymentModel.Payment
		ParamRequest unifiedPaymentModel.PerformInquiryRequest
		setupMock    func(*repositoryMock.ISnapCoreRepository)
		validate     func(t *testing.T, result *unifiedPaymentModel.InquiryResult, err error)
	}{
		{
			name: "Returns nil when no VA number found",
			payment: &paymentModel.Payment{
				UUID:                     "payment-123",
				ProcessorReferenceNumber: nil,
				Metadata:                 nil,
			},
			ParamRequest: unifiedPaymentModel.PerformInquiryRequest{},
			setupMock:    func(snapCore *repositoryMock.ISnapCoreRepository) {},
			validate: func(t *testing.T, result *unifiedPaymentModel.InquiryResult, err error) {
				assert.Nil(t, result)
				assert.Nil(t, err)
			},
		},
		{
			name: "Returns error when snap core inquiry fails",
			payment: &paymentModel.Payment{
				UUID:                     "payment-456",
				ProcessorReferenceNumber: strPtr("1234567890"),
			},
			ParamRequest: unifiedPaymentModel.PerformInquiryRequest{},
			setupMock: func(snapCore *repositoryMock.ISnapCoreRepository) {
				snapCore.On("InquiryStatusVirtualAccount", mock.Anything, mock.MatchedBy(func(req *snapCoreVAModel.InquiryStatusVARequest) bool {
					return req.VirtualAccount == "1234567890" && req.SkipPublish == true
				})).Return(nil, errors.New("API error"))
			},
			validate: func(t *testing.T, result *unifiedPaymentModel.InquiryResult, err error) {
				assert.Nil(t, result)
				assert.Error(t, err)
			},
		},
		{
			name: "Returns nil when snap core returns nil response",
			payment: &paymentModel.Payment{
				UUID:                     "payment-789",
				ProcessorReferenceNumber: strPtr("9876543210"),
			},
			ParamRequest: unifiedPaymentModel.PerformInquiryRequest{},
			setupMock: func(snapCore *repositoryMock.ISnapCoreRepository) {
				snapCore.On("InquiryStatusVirtualAccount", mock.Anything, mock.Anything).
					Return(nil, nil)
			},
			validate: func(t *testing.T, result *unifiedPaymentModel.InquiryResult, err error) {
				assert.Nil(t, result)
				assert.Nil(t, err)
			},
		},
		{
			name: "Returns mapped result on success",
			payment: &paymentModel.Payment{
				UUID:                     "payment-success",
				ProcessorReferenceNumber: strPtr("1111222233"),
			},
			ParamRequest: unifiedPaymentModel.PerformInquiryRequest{
				LedgerID: uuid.NewString(),
			},
			setupMock: func(snapCore *repositoryMock.ISnapCoreRepository) {
				snapCore.On("InquiryStatusVirtualAccount", mock.Anything, mock.Anything).
					Return(&snapCoreVAModel.InquiryStatusVAResponse{
						Data: snapCoreVAModel.InquiryStatusVAResponseData{
							ResponseCode:    "2002400",
							ResponseMessage: "Success",
							VirtualAccountData: &snapCoreVAModel.InquiryStatusVAData{
								PaymentRequestId: "pay-req-123",
								ReferenceNo:      "ref-123",
							},
						},
					}, nil)
			},
			validate: func(t *testing.T, result *unifiedPaymentModel.InquiryResult, err error) {
				assert.NotNil(t, result)
				assert.Nil(t, err)
				assert.Equal(t, constant.InquiryStatusSuccess, result.Status)
				assert.Equal(t, "pay-req-123", result.ProcessorTransactionID)
				assert.Equal(t, "ref-123", result.ProcessorReferenceNo)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			snapCoreMock := repositoryMock.NewISnapCoreRepository(t)
			tc.setupMock(snapCoreMock)

			svc := &UnifiedPaymentService{
				config:       cfg,
				logger:       log,
				snapCoreRepo: snapCoreMock,
			}

			result, err := svc.performVAInquiry(ctx, tc.payment, tc.ParamRequest)
			tc.validate(t, result, err)
		})
	}
}

func TestIsInquiryEligible(t *testing.T) {
	ctx := context.Background()
	log, _ := mockLogger.NewZapLogger(mockLogger.Config{})
	cfg := &config.Config{Environment: "test"}

	tests := []struct {
		name     string
		payment  *paymentModel.Payment
		expected bool
	}{
		{
			name: "Returns true for PROCESSING status with QRIS payment method",
			payment: &paymentModel.Payment{
				UUID:   "payment-123",
				Status: constant.UnifiedPaymentSessionStatusProcessing,
				PaymentMethod: paymentModel.PaymentMethod{
					Type: paymentConst.PAYMENT_METHOD_QRIS,
				},
			},
			expected: true,
		},
		{
			name: "Returns true for REQUIRE_ACTION status with VIRTUAL_ACCOUNT payment method",
			payment: &paymentModel.Payment{
				UUID:   "payment-456",
				Status: constant.UnifiedPaymentSessionStatusRequireAction,
				PaymentMethod: paymentModel.PaymentMethod{
					Type: paymentConst.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
				},
			},
			expected: true,
		},
		{
			name: "Returns true for ChargeStatusProcessing status",
			payment: &paymentModel.Payment{
				UUID:   "payment-789",
				Status: constant.ChargeStatusProcessing,
				PaymentMethod: paymentModel.PaymentMethod{
					Type: paymentConst.PAYMENT_METHOD_QRIS,
				},
			},
			expected: true,
		},
		{
			name: "Returns true for ChargeStatusWaitingForUserAction status",
			payment: &paymentModel.Payment{
				UUID:   "payment-waiting",
				Status: constant.ChargeStatusWaitingForUserAction,
				PaymentMethod: paymentModel.PaymentMethod{
					Type: paymentConst.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
				},
			},
			expected: true,
		},
		{
			name: "Returns false for SUCCESS status",
			payment: &paymentModel.Payment{
				UUID:   "payment-success",
				Status: constant.ChargeStatusSuccess,
				PaymentMethod: paymentModel.PaymentMethod{
					Type: paymentConst.PAYMENT_METHOD_QRIS,
				},
			},
			expected: false,
		},
		{
			name: "Returns false for EXPIRED status",
			payment: &paymentModel.Payment{
				UUID:   "payment-expired",
				Status: constant.ChargeStatusExpired,
				PaymentMethod: paymentModel.PaymentMethod{
					Type: paymentConst.PAYMENT_METHOD_QRIS,
				},
			},
			expected: false,
		},
		{
			name: "Returns false for FAILED status",
			payment: &paymentModel.Payment{
				UUID:   "payment-failed",
				Status: constant.ChargeStatusFailed,
				PaymentMethod: paymentModel.PaymentMethod{
					Type: paymentConst.PAYMENT_METHOD_QRIS,
				},
			},
			expected: false,
		},
		{
			name: "Returns false for CREDIT_CARD payment method",
			payment: &paymentModel.Payment{
				UUID:   "payment-card",
				Status: constant.ChargeStatusProcessing,
				PaymentMethod: paymentModel.PaymentMethod{
					Type: paymentConst.PAYMENT_METHOD_CREDIT_CARD,
				},
			},
			expected: false,
		},
		{
			name: "Returns false for EWALLET payment method",
			payment: &paymentModel.Payment{
				UUID:   "payment-ewallet",
				Status: constant.ChargeStatusProcessing,
				PaymentMethod: paymentModel.PaymentMethod{
					Type: paymentConst.PAYMENT_METHOD_EWALLET,
				},
			},
			expected: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc := &UnifiedPaymentService{
				config: cfg,
				logger: log,
			}

			result := svc.isInquiryEligible(ctx, tc.payment)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestGetSnapCoreIdForQris(t *testing.T) {
	ctx := context.Background()
	log, _ := mockLogger.NewZapLogger(mockLogger.Config{})
	cfg := &config.Config{Environment: "test"}

	tests := []struct {
		name      string
		payment   *paymentModel.Payment
		setupMock func(*repositoryMock.IAccountTransactionRepository)
		expected  string
	}{
		{
			name: "Returns SnapCoreId from payment when available",
			payment: &paymentModel.Payment{
				UUID:       "payment-123",
				SnapCoreId: strPtr("snap-core-from-payment"),
			},
			setupMock: func(accTrx *repositoryMock.IAccountTransactionRepository) {},
			expected:  "snap-core-from-payment",
		},
		{
			name: "Falls back to accountTransaction.ProcessorReferenceId",
			payment: &paymentModel.Payment{
				UUID:       "payment-456",
				SnapCoreId: nil,
			},
			setupMock: func(accTrx *repositoryMock.IAccountTransactionRepository) {
				accTrx.On("FindByReference", mock.Anything, "payment-456", constant.TypePayment).
					Return(&orchestratorModel.AccountTransactionWithUseCase{
						ProcessorReferenceId: "snap-core-from-trx",
					}, nil)
			},
			expected: "snap-core-from-trx",
		},
		{
			name: "Returns empty when SnapCoreId is empty string",
			payment: &paymentModel.Payment{
				UUID:       "payment-789",
				SnapCoreId: strPtr(""),
			},
			setupMock: func(accTrx *repositoryMock.IAccountTransactionRepository) {
				accTrx.On("FindByReference", mock.Anything, "payment-789", constant.TypePayment).
					Return(nil, nil)
			},
			expected: "",
		},
		{
			name: "Returns empty when accountTransaction is nil",
			payment: &paymentModel.Payment{
				UUID:       "payment-nil-trx",
				SnapCoreId: nil,
			},
			setupMock: func(accTrx *repositoryMock.IAccountTransactionRepository) {
				accTrx.On("FindByReference", mock.Anything, "payment-nil-trx", constant.TypePayment).
					Return(nil, nil)
			},
			expected: "",
		},
		{
			name: "Returns empty when accountTransaction has dash ProcessorReferenceId",
			payment: &paymentModel.Payment{
				UUID:       "payment-dash",
				SnapCoreId: nil,
			},
			setupMock: func(accTrx *repositoryMock.IAccountTransactionRepository) {
				accTrx.On("FindByReference", mock.Anything, "payment-dash", constant.TypePayment).
					Return(&orchestratorModel.AccountTransactionWithUseCase{
						ProcessorReferenceId: "-",
					}, nil)
			},
			expected: "",
		},
		{
			name: "Returns empty when accountTransaction lookup fails",
			payment: &paymentModel.Payment{
				UUID:       "payment-error",
				SnapCoreId: nil,
			},
			setupMock: func(accTrx *repositoryMock.IAccountTransactionRepository) {
				accTrx.On("FindByReference", mock.Anything, "payment-error", constant.TypePayment).
					Return(nil, errors.New("db error"))
			},
			expected: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			accTrxMock := repositoryMock.NewIAccountTransactionRepository(t)
			tc.setupMock(accTrxMock)

			svc := &UnifiedPaymentService{
				config:                 cfg,
				logger:                 log,
				accountTransactionRepo: accTrxMock,
			}

			result := svc.getSnapCoreIdForQris(ctx, tc.payment)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestPerformPaymentInquiry(t *testing.T) {
	ctx := context.Background()
	log, _ := mockLogger.NewZapLogger(mockLogger.Config{})
	cfg := &config.Config{Environment: "test"}

	tests := []struct {
		name         string
		payment      *paymentModel.Payment
		ParamRequest unifiedPaymentModel.PerformInquiryRequest
		setupMock    func(*repositoryMock.ISnapCoreRepository, *repositoryMock.IAccountTransactionRepository, *redisMock.IRedisExt)
		redisNil     bool
		validate     func(t *testing.T, result *unifiedPaymentModel.InquiryResult, err error)
	}{
		{
			name:         "Returns nil for nil payment",
			payment:      nil,
			ParamRequest: unifiedPaymentModel.PerformInquiryRequest{},
			setupMock: func(snapCore *repositoryMock.ISnapCoreRepository, accTrx *repositoryMock.IAccountTransactionRepository, redisMock *redisMock.IRedisExt) {
			},
			redisNil: true,
			validate: func(t *testing.T, result *unifiedPaymentModel.InquiryResult, err error) {
				assert.Nil(t, result)
				assert.Nil(t, err)
			},
		},
		{
			name: "Returns nil for ineligible payment status",
			payment: &paymentModel.Payment{
				UUID:   "payment-success",
				Status: constant.ChargeStatusSuccess,
				PaymentMethod: paymentModel.PaymentMethod{
					Type: paymentConst.PAYMENT_METHOD_QRIS,
				},
			},
			ParamRequest: unifiedPaymentModel.PerformInquiryRequest{},
			setupMock: func(snapCore *repositoryMock.ISnapCoreRepository, accTrx *repositoryMock.IAccountTransactionRepository, redisMock *redisMock.IRedisExt) {
			},
			redisNil: true,
			validate: func(t *testing.T, result *unifiedPaymentModel.InquiryResult, err error) {
				assert.Nil(t, result)
				assert.Nil(t, err)
			},
		},
		{
			name: "Returns nil when within cooldown period",
			payment: &paymentModel.Payment{
				UUID:   "payment-cooldown",
				Status: constant.ChargeStatusProcessing,
				PaymentMethod: paymentModel.PaymentMethod{
					Type: paymentConst.PAYMENT_METHOD_QRIS,
				},
			},
			ParamRequest: unifiedPaymentModel.PerformInquiryRequest{},
			setupMock: func(snapCore *repositoryMock.ISnapCoreRepository, accTrx *repositoryMock.IAccountTransactionRepository, redisSvc *redisMock.IRedisExt) {
				redisSvc.On("Exists", mock.Anything, "backend-portal:inquiry:cooldown:payment-cooldown").
					Return(redis.NewIntResult(1, nil))
			},
			redisNil: false,
			validate: func(t *testing.T, result *unifiedPaymentModel.InquiryResult, err error) {
				assert.Nil(t, result)
				assert.Nil(t, err)
			},
		},
		{
			name: "Successfully performs QRIS inquiry",
			payment: &paymentModel.Payment{
				UUID:       "payment-qris",
				Status:     constant.ChargeStatusProcessing,
				SnapCoreId: strPtr("snap-core-qris"),
				PaymentMethod: paymentModel.PaymentMethod{
					Type: paymentConst.PAYMENT_METHOD_QRIS,
				},
			},
			ParamRequest: unifiedPaymentModel.PerformInquiryRequest{},
			setupMock: func(snapCore *repositoryMock.ISnapCoreRepository, accTrx *repositoryMock.IAccountTransactionRepository, redisSvc *redisMock.IRedisExt) {
				redisSvc.On("Exists", mock.Anything, "backend-portal:inquiry:cooldown:payment-qris").
					Return(redis.NewIntResult(0, nil))
				snapCore.On("InquiryStatusQris", mock.Anything, &snapCoreQRModel.InquiryStatusQrMpmRequest{QrisUUID: "snap-core-qris", SkipPublish: true}).
					Return(&snapCoreQRModel.QrisInquiryStatusResponse{
						Data: &snapCoreQRModel.QrisInquiryStatusResponseData{
							ResponseCode: "00",
							Status:       constant.InquiryStatusSuccess,
							UUID:         "processor-uuid",
						},
					}, nil)
				redisSvc.On("Set", mock.Anything, "backend-portal:inquiry:cooldown:payment-qris", mock.AnythingOfType("string"), 30*time.Second).
					Return(redis.NewStatusResult("OK", nil))
			},
			redisNil: false,
			validate: func(t *testing.T, result *unifiedPaymentModel.InquiryResult, err error) {
				assert.NotNil(t, result)
				assert.Nil(t, err)
				assert.Equal(t, constant.InquiryStatusSuccess, result.Status)
				assert.NotNil(t, result.LastInquiryAt)
			},
		},
		{
			name: "Successfully performs VA inquiry",
			payment: &paymentModel.Payment{
				UUID:                     "payment-va",
				Status:                   constant.ChargeStatusProcessing,
				ProcessorReferenceNumber: strPtr("1234567890"),
				PaymentMethod: paymentModel.PaymentMethod{
					Type: paymentConst.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
				},
			},
			ParamRequest: unifiedPaymentModel.PerformInquiryRequest{
				LedgerID: uuid.NewString(),
			},
			setupMock: func(snapCore *repositoryMock.ISnapCoreRepository, accTrx *repositoryMock.IAccountTransactionRepository, redisSvc *redisMock.IRedisExt) {
				redisSvc.On("Exists", mock.Anything, "backend-portal:inquiry:cooldown:payment-va").
					Return(redis.NewIntResult(0, nil))
				snapCore.On("InquiryStatusVirtualAccount", mock.Anything, mock.Anything).
					Return(&snapCoreVAModel.InquiryStatusVAResponse{
						Data: snapCoreVAModel.InquiryStatusVAResponseData{
							ResponseCode:    "2002400",
							ResponseMessage: "Success",
						},
					}, nil)
				redisSvc.On("Set", mock.Anything, "backend-portal:inquiry:cooldown:payment-va", mock.AnythingOfType("string"), 30*time.Second).
					Return(redis.NewStatusResult("OK", nil))
			},
			redisNil: false,
			validate: func(t *testing.T, result *unifiedPaymentModel.InquiryResult, err error) {
				assert.NotNil(t, result)
				assert.Nil(t, err)
				assert.Equal(t, constant.InquiryStatusSuccess, result.Status)
				assert.NotNil(t, result.LastInquiryAt)
			},
		},
		{
			name: "Returns nil for unsupported payment method",
			payment: &paymentModel.Payment{
				UUID:   "payment-card",
				Status: constant.ChargeStatusProcessing,
				PaymentMethod: paymentModel.PaymentMethod{
					Type: paymentConst.PAYMENT_METHOD_CREDIT_CARD,
				},
			},
			ParamRequest: unifiedPaymentModel.PerformInquiryRequest{},
			setupMock: func(snapCore *repositoryMock.ISnapCoreRepository, accTrx *repositoryMock.IAccountTransactionRepository, redisSvc *redisMock.IRedisExt) {
			},
			redisNil: true,
			validate: func(t *testing.T, result *unifiedPaymentModel.InquiryResult, err error) {
				assert.Nil(t, result)
				assert.Nil(t, err)
			},
		},
		{
			name: "Sets cooldown and returns nil on inquiry error",
			payment: &paymentModel.Payment{
				UUID:       "payment-error",
				Status:     constant.ChargeStatusProcessing,
				SnapCoreId: strPtr("snap-core-error"),
				PaymentMethod: paymentModel.PaymentMethod{
					Type: paymentConst.PAYMENT_METHOD_QRIS,
				},
			},
			ParamRequest: unifiedPaymentModel.PerformInquiryRequest{},
			setupMock: func(snapCore *repositoryMock.ISnapCoreRepository, accTrx *repositoryMock.IAccountTransactionRepository, redisSvc *redisMock.IRedisExt) {
				redisSvc.On("Exists", mock.Anything, "backend-portal:inquiry:cooldown:payment-error").
					Return(redis.NewIntResult(0, nil))
				snapCore.On("InquiryStatusQris", mock.Anything, &snapCoreQRModel.InquiryStatusQrMpmRequest{QrisUUID: "snap-core-error", SkipPublish: true}).
					Return(nil, errors.New("API error"))
				redisSvc.On("Set", mock.Anything, "backend-portal:inquiry:cooldown:payment-error", mock.AnythingOfType("string"), 30*time.Second).
					Return(redis.NewStatusResult("OK", nil))
			},
			redisNil: false,
			validate: func(t *testing.T, result *unifiedPaymentModel.InquiryResult, err error) {
				assert.Nil(t, result)
				assert.Nil(t, err)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			snapCoreMock := repositoryMock.NewISnapCoreRepository(t)
			accTrxMock := repositoryMock.NewIAccountTransactionRepository(t)

			var svc *UnifiedPaymentService
			if tc.redisNil {
				tc.setupMock(snapCoreMock, accTrxMock, nil)
				svc = &UnifiedPaymentService{
					config:                 cfg,
					logger:                 log,
					snapCoreRepo:           snapCoreMock,
					accountTransactionRepo: accTrxMock,
					redis:                  nil,
				}
			} else {
				redisMockInstance := redisMock.NewIRedisExt(t)
				tc.setupMock(snapCoreMock, accTrxMock, redisMockInstance)
				svc = &UnifiedPaymentService{
					config:                 cfg,
					logger:                 log,
					snapCoreRepo:           snapCoreMock,
					accountTransactionRepo: accTrxMock,
					redis:                  redisMockInstance,
				}
			}

			result, err := svc.performPaymentInquiry(ctx, tc.payment, tc.ParamRequest)
			tc.validate(t, result, err)
		})
	}
}

func strPtr(s string) *string {
	return &s
}
