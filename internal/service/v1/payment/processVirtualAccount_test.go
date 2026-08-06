package paymentService

import (
	"bytes"
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	paymentConstant "github.com/paper-indonesia/pivot-backoffice/constant/payment"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	snapCoreModel "github.com/paper-indonesia/pivot-backoffice/internal/model/snapCore/virtualAccount"
	rabbitMqMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/rabbitmqExt"
	redisExtMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/redisExt"
	repositoryMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/paper-indonesia/pdk/v2/logger"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestProcessVirtualAccountPayment(t *testing.T) {
	paymentRepo := repositoryMocks.NewIPaymentRepository(t)
	accountTransactionRepo := repositoryMocks.NewIAccountTransactionRepository(t)
	internalMethod := serviceMocks.NewIPaymentInternalDirectFunc(t)
	rmq := rabbitMqMocks.NewRabbitMQExt(t)
	redis := redisExtMocks.NewIRedisExt(t)
	redisMutex := redisExtMocks.NewIMutexer(t)

	// Shared mocks for the distributed lock acquired when vaType is CLOSED_DYNAMIC.
	// Uses Maybe() so cases that return before reaching the lock are not required to match.
	redis.On("NewMutex", mock.Anything, mock.Anything).Maybe().Return(redisMutex)
	redisMutex.On("LockContext", mock.Anything).Maybe().Return(nil)
	redisMutex.On("UnlockContext", mock.Anything).Maybe().Return(true, nil)

	accountTransactionRepo.On(
		"GetTransactionByReferenceIdAndProcessorId", mock.Anything, mock.Anything, mock.Anything,
	).Maybe().Return(nil, nil)

	buf := new(bytes.Buffer)
	log := logger.NewSlogger(logger.Config{}, logger.WithSlogOutput(buf))
	defer buf.Reset()

	paymentAmount, _ := decimal.NewFromString("1000000.00")

	// Helper to build a fresh CLOSED_DYNAMIC payment (IsClosedAmount=true, IsSingleUse=true).
	newClosedDynamicPayment := func() *paymentModel.Payment {
		return &paymentModel.Payment{
			UUID:        uuid.NewString(),
			Currency:    "IDR",
			Amount:      paymentAmount,
			TotalAmount: paymentAmount,
			MerchantID:  uuid.NewString(),
			Metadata: &map[string]any{
				"snapCore": snapCoreModel.CreateVirtualAccountResponseData{
					IsClosedAmount: true,
					IsSingleUse:    true,
				},
			},
			PaymentMethod: paymentModel.PaymentMethod{
				Acquirer: constant.BANK_ACQUIRER_PERMATA,
			},
		}
	}

	// Helper to build a fresh OPEN_STATIC payment (IsClosedAmount=false, IsSingleUse=false).
	newOpenStaticPayment := func() *paymentModel.Payment {
		return &paymentModel.Payment{
			UUID:        uuid.NewString(),
			Currency:    "IDR",
			Amount:      paymentAmount,
			TotalAmount: paymentAmount,
			MerchantID:  uuid.NewString(),
			Metadata: &map[string]any{
				"snapCore": snapCoreModel.CreateVirtualAccountResponseData{
					IsClosedAmount: false,
					IsSingleUse:    false,
				},
			},
			PaymentMethod: paymentModel.PaymentMethod{
				Acquirer: constant.BANK_ACQUIRER_PERMATA,
			},
		}
	}

	validPaidRequest := &paymentModel.VirtualAccountPaymentNotificationRequest{
		Acquirer: constant.BANK_ACQUIRER_PERMATA,
		Number:   VaNumber,
		Status:   paymentConstant.VirtualAccountStatusPaid,
		PaidAmount: commonModel.Amount{
			Currency: "IDR",
			Value:    "1000000.00",
		},
	}

	testCases := []struct {
		name      string
		wantErr   bool
		request   *paymentModel.VirtualAccountPaymentNotificationRequest
		setupMock func()
	}{
		{
			name:    "SUCCESS: Non-paid status returns nil without processing",
			request: &paymentModel.VirtualAccountPaymentNotificationRequest{Status: paymentConstant.VirtualAccountStatusExpired},
			wantErr: false,
			setupMock: func() {
			},
		},
		{
			name:    "ERROR: BeginTransaction fails",
			request: validPaidRequest,
			wantErr: true,
			setupMock: func() {
				paymentRepo.On("BeginTransaction", mock.Anything).
					Once().Return(context.Background(), constant.ErrSomeErrorForUnitTest)
			},
		},
		{
			name:    "ERROR: Payment not found in GetAndUpdateVirtualAccountPayment",
			request: validPaidRequest,
			wantErr: true,
			setupMock: func() {
				paymentRepo.On("BeginTransaction", mock.Anything).
					Once().Return(context.Background(), nil)
				paymentRepo.On("GetActivePaymentByProcessorReferenceNumber", mock.Anything, mock.Anything).
					Once().Return(nil, nil)
				paymentRepo.On("RollbackTransaction", mock.Anything).
					Once().Return(nil)
			},
		},
		{
			name:    "ERROR: CLOSED_DYNAMIC UpdatePaymentStatus fails in payVirtualAccountPayment",
			request: validPaidRequest,
			wantErr: true,
			setupMock: func() {
				paymentRepo.On("BeginTransaction", mock.Anything).
					Once().Return(context.Background(), nil)
				paymentRepo.On("GetActivePaymentByProcessorReferenceNumber", mock.Anything, mock.Anything).
					Once().Return(newClosedDynamicPayment(), nil)
				paymentRepo.On("UpdatePaymentStatus", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Once().Return(constant.ErrSomeErrorForUnitTest)
				paymentRepo.On("RollbackTransaction", mock.Anything).
					Once().Return(nil)
			},
		},
		{
			name:    "ERROR: Nil metadata defaults to CLOSED_DYNAMIC, lock acquired, DeterminePaymentFee fails",
			request: validPaidRequest,
			wantErr: true,
			setupMock: func() {
				paymentRepo.On("BeginTransaction", mock.Anything).
					Once().Return(context.Background(), nil)
				paymentRepo.On("GetActivePaymentByProcessorReferenceNumber", mock.Anything, mock.Anything).
					Once().Return(&paymentModel.Payment{
					UUID:        uuid.NewString(),
					Currency:    "IDR",
					Amount:      paymentAmount,
					TotalAmount: paymentAmount,
					MerchantID:  uuid.NewString(),
					Metadata:    nil,
					PaymentMethod: paymentModel.PaymentMethod{
						Acquirer: constant.BANK_ACQUIRER_PERMATA,
					},
				}, nil)
				internalMethod.On("DeterminePaymentFee", mock.Anything, mock.Anything).
					Once().Return(constant.ErrSomeErrorForUnitTest)
				paymentRepo.On("RollbackTransaction", mock.Anything).
					Once().Return(nil)
			},
		},
		{
			name:    "ERROR: CLOSED_DYNAMIC lock acquired then DeterminePaymentFee fails",
			request: validPaidRequest,
			wantErr: true,
			setupMock: func() {
				paymentRepo.On("BeginTransaction", mock.Anything).
					Once().Return(context.Background(), nil)
				paymentRepo.On("GetActivePaymentByProcessorReferenceNumber", mock.Anything, mock.Anything).
					Once().Return(newClosedDynamicPayment(), nil)
				paymentRepo.On("UpdatePaymentStatus", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Once().Return(nil)
				internalMethod.On("DeterminePaymentFee", mock.Anything, mock.Anything).
					Once().Return(constant.ErrSomeErrorForUnitTest)
				paymentRepo.On("RollbackTransaction", mock.Anything).
					Once().Return(nil)
			},
		},
		{
			name:    "ERROR: OPEN_STATIC DeterminePaymentFee fails without lock",
			request: validPaidRequest,
			wantErr: true,
			setupMock: func() {
				paymentRepo.On("BeginTransaction", mock.Anything).
					Once().Return(context.Background(), nil)
				paymentRepo.On("GetActivePaymentByProcessorReferenceNumber", mock.Anything, mock.Anything).
					Once().Return(newOpenStaticPayment(), nil)
				internalMethod.On("DeterminePaymentFee", mock.Anything, mock.Anything).
					Once().Return(constant.ErrSomeErrorForUnitTest)
				paymentRepo.On("RollbackTransaction", mock.Anything).
					Once().Return(nil)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			buf.Reset()
			tc.setupMock()
			paymentSvc := New(paymentRepo, log, nil, nil, nil, nil, nil,
				WithInternalDirectFunc(internalMethod),
				WithAccountTransactionRepository(accountTransactionRepo),
				WithRabbitMQClient(rmq),
				WithRedisClient(redis),
			)

			err := paymentSvc.ProcessVirtualAccountPayment(context.Background(), tc.request)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			paymentRepo.AssertExpectations(t)
		})
	}
}
