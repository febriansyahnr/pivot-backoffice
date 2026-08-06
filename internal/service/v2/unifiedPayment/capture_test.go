package unifiedPaymentService

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/jmoiron/sqlx/types"
	"github.com/paper-indonesia/pivot-backoffice/pkg/rabbitMqExt"
	"github.com/redis/go-redis/v9"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/paper-indonesia/pivot-backoffice/config"
	c "github.com/paper-indonesia/pivot-backoffice/constant"
	paymentConstant "github.com/paper-indonesia/pivot-backoffice/constant/payment"
	creditcardCoreProcessorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/creditcardCoreProcessor"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	orchestratorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	paymentCaptureModel "github.com/paper-indonesia/pivot-backoffice/internal/model/paymentCapture"
	unifiedPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/unifiedPayment"
	rabbitMQMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/rabbitmqExt"
	redisExtMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/redisExt"
	repositoryMock "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	serviceMock "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	mockLogger "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCapture(t *testing.T) {
	cfg := &config.Config{
		Environment: "test",
	}
	log, _ := mockLogger.NewZapLogger(mockLogger.Config{})

	merchantID := "merchant-" + uuid.NewString()
	paymentID := "payment-" + uuid.NewString()
	referenceID := "ref-" + uuid.NewString()
	databaseError := errors.New("database error")
	accountTrxID := uuid.New()

	validPayment := &paymentModel.Payment{
		UUID:        paymentID,
		MerchantID:  merchantID,
		Amount:      decimal.NewFromFloat(100000),
		Currency:    "IDR",
		Status:      c.UnifiedPaymentSessionStatusProcessing,
		ReferenceID: &referenceID,
		PaymentMethod: paymentModel.PaymentMethod{
			Type: paymentConstant.PAYMENT_METHOD_CREDIT_CARD,
		},
		Metadata: &map[string]interface{}{
			"paymentMethodOptions": map[string]interface{}{
				"card": map[string]interface{}{
					"captureMethod": c.UnifiedPaymentCardCaptureMethodManual,
				},
			},
		},
	}

	validAdditionalInfo := map[string]interface{}{
		"chargeStatus": c.ChargeStatusHistoryWaitingForCapture,
	}
	validAdditionalInfoBytes, _ := json.Marshal(validAdditionalInfo)
	validAccountTrx := &orchestratorModel.AccountTransactionWithUseCase{
		UUID:                 accountTrxID,
		Type:                 c.TypePayment,
		Reference:            paymentID,
		Status:               c.StatusPending,
		ProcessorReferenceId: "proc-ref-123",
		Currency:             "IDR",
		Credit:               0,
		AdditionalInfo: types.NullJSONText{
			Valid:    true,
			JSONText: validAdditionalInfoBytes,
		},
	}

	validRequest := &unifiedPaymentModel.CaptureRequest{
		PaymentID:              paymentID,
		MerchantID:             merchantID,
		ReleaseRemainingAmount: false,
		Amount: &unifiedPaymentModel.Amount{
			Currency: "IDR",
			Value:    100000,
		},
	}

	testCases := []struct {
		name      string
		wantErr   bool
		setupMock func(*repositoryMock.IPaymentRepository, *repositoryMock.IAccountTransactionRepository, *repositoryMock.IPaymentCaptureRepository, *rabbitMQMock.RabbitMQExt)
		request   *unifiedPaymentModel.CaptureRequest
	}{
		{
			name:    "ERROR: Database error when getting payment",
			wantErr: true,
			request: validRequest,
			setupMock: func(paymentRepo *repositoryMock.IPaymentRepository, accountTrxRepo *repositoryMock.IAccountTransactionRepository, paymentCaptureRepo *repositoryMock.IPaymentCaptureRepository, rabbitMq *rabbitMQMock.RabbitMQExt) {
				paymentRepo.On("GetPaymentById", c.ValueCtxMockType(), paymentID).Return(nil, databaseError)
			},
		},
		{
			name:    "ERROR: Payment not found",
			wantErr: true,
			request: validRequest,
			setupMock: func(paymentRepo *repositoryMock.IPaymentRepository, accountTrxRepo *repositoryMock.IAccountTransactionRepository, paymentCaptureRepo *repositoryMock.IPaymentCaptureRepository, rabbitMq *rabbitMQMock.RabbitMQExt) {
				paymentRepo.On("GetPaymentById", c.ValueCtxMockType(), paymentID).Return(nil, nil)
			},
		},
		{
			name:    "ERROR: Database error when finding account transaction",
			wantErr: true,
			request: validRequest,
			setupMock: func(paymentRepo *repositoryMock.IPaymentRepository, accountTrxRepo *repositoryMock.IAccountTransactionRepository, paymentCaptureRepo *repositoryMock.IPaymentCaptureRepository, rabbitMq *rabbitMQMock.RabbitMQExt) {
				paymentRepo.On("GetPaymentById", c.ValueCtxMockType(), paymentID).Return(validPayment, nil)
				accountTrxRepo.On("FindByReference", c.ValueCtxMockType(), paymentID, c.TypePayment).Return(nil, databaseError)
			},
		},
		{
			name:    "ERROR: Account transaction not found",
			wantErr: true,
			request: validRequest,
			setupMock: func(paymentRepo *repositoryMock.IPaymentRepository, accountTrxRepo *repositoryMock.IAccountTransactionRepository, paymentCaptureRepo *repositoryMock.IPaymentCaptureRepository, rabbitMq *rabbitMQMock.RabbitMQExt) {
				paymentRepo.On("GetPaymentById", c.ValueCtxMockType(), paymentID).Return(validPayment, nil)
				accountTrxRepo.On("FindByReference", c.ValueCtxMockType(), paymentID, c.TypePayment).Return(nil, nil)
			},
		},
		{
			name:    "ERROR: Merchant ID mismatch",
			wantErr: true,
			request: &unifiedPaymentModel.CaptureRequest{
				PaymentID:              paymentID,
				MerchantID:             "different-merchant",
				ReleaseRemainingAmount: false,
				Amount: &unifiedPaymentModel.Amount{
					Currency: "IDR",
					Value:    100000,
				},
			},
			setupMock: func(paymentRepo *repositoryMock.IPaymentRepository, accountTrxRepo *repositoryMock.IAccountTransactionRepository, paymentCaptureRepo *repositoryMock.IPaymentCaptureRepository, rabbitMq *rabbitMQMock.RabbitMQExt) {
				paymentRepo.On("GetPaymentById", c.ValueCtxMockType(), paymentID).Return(validPayment, nil)
				accountTrxRepo.On("FindByReference", c.ValueCtxMockType(), paymentID, c.TypePayment).Return(validAccountTrx, nil)
			},
		},
		{
			name:    "ERROR: Payment method not credit card",
			wantErr: true,
			request: validRequest,
			setupMock: func(paymentRepo *repositoryMock.IPaymentRepository, accountTrxRepo *repositoryMock.IAccountTransactionRepository, paymentCaptureRepo *repositoryMock.IPaymentCaptureRepository, rabbitMq *rabbitMQMock.RabbitMQExt) {
				payment := *validPayment
				payment.PaymentMethod.Type = c.ChannelVirtualAccount
				paymentRepo.On("GetPaymentById", c.ValueCtxMockType(), paymentID).Return(&payment, nil)
				accountTrxRepo.On("FindByReference", c.ValueCtxMockType(), paymentID, c.TypePayment).Return(validAccountTrx, nil)
			},
		},
		{
			name:    "ERROR: Capture method not manual",
			wantErr: true,
			request: validRequest,
			setupMock: func(paymentRepo *repositoryMock.IPaymentRepository, accountTrxRepo *repositoryMock.IAccountTransactionRepository, paymentCaptureRepo *repositoryMock.IPaymentCaptureRepository, rabbitMq *rabbitMQMock.RabbitMQExt) {
				payment := *validPayment
				metadata := map[string]interface{}{
					"payment_method_options": map[string]interface{}{
						"card": map[string]interface{}{
							"capture_method": c.UnifiedPaymentCardCaptureMethodAutomatic,
						},
					},
				}
				payment.Metadata = &metadata
				paymentRepo.On("GetPaymentById", c.ValueCtxMockType(), paymentID).Return(&payment, nil)
				accountTrxRepo.On("FindByReference", c.ValueCtxMockType(), paymentID, c.TypePayment).Return(validAccountTrx, nil)
			},
		},
		{
			name:    "ERROR: Payment already in final status",
			wantErr: true,
			request: validRequest,
			setupMock: func(paymentRepo *repositoryMock.IPaymentRepository, accountTrxRepo *repositoryMock.IAccountTransactionRepository, paymentCaptureRepo *repositoryMock.IPaymentCaptureRepository, rabbitMq *rabbitMQMock.RabbitMQExt) {
				payment := *validPayment
				payment.Status = c.UnifiedPaymentSessionStatusPaid
				paymentRepo.On("GetPaymentById", c.ValueCtxMockType(), paymentID).Return(&payment, nil)
				accountTrxRepo.On("FindByReference", c.ValueCtxMockType(), paymentID, c.TypePayment).Return(validAccountTrx, nil)
			},
		},
		{
			name:    "ERROR: Charge status not waiting for capture",
			wantErr: true,
			request: validRequest,
			setupMock: func(paymentRepo *repositoryMock.IPaymentRepository, accountTrxRepo *repositoryMock.IAccountTransactionRepository, paymentCaptureRepo *repositoryMock.IPaymentCaptureRepository, rabbitMq *rabbitMQMock.RabbitMQExt) {
				accountTrx := *validAccountTrx
				updateAdditionalInfo := map[string]interface{}{
					"chargeStatus": c.ChargeStatusSuccess,
				}
				updateAdditionalInfoBytes, _ := json.Marshal(updateAdditionalInfo)
				accountTrx.AdditionalInfo = types.NullJSONText{
					Valid:    true,
					JSONText: updateAdditionalInfoBytes,
				}
				paymentRepo.On("GetPaymentById", c.ValueCtxMockType(), paymentID).Return(validPayment, nil)
				accountTrxRepo.On("FindByReference", c.ValueCtxMockType(), paymentID, c.TypePayment).Return(&accountTrx, nil)
			},
		},
		{
			name:    "ERROR: Currency mismatch",
			wantErr: true,
			request: &unifiedPaymentModel.CaptureRequest{
				PaymentID:              paymentID,
				MerchantID:             merchantID,
				ReleaseRemainingAmount: false,
				Amount: &unifiedPaymentModel.Amount{
					Currency: "USD",
					Value:    100000,
				},
			},
			setupMock: func(paymentRepo *repositoryMock.IPaymentRepository, accountTrxRepo *repositoryMock.IAccountTransactionRepository, paymentCaptureRepo *repositoryMock.IPaymentCaptureRepository, rabbitMq *rabbitMQMock.RabbitMQExt) {
				paymentRepo.On("GetPaymentById", c.ValueCtxMockType(), paymentID).Return(validPayment, nil)
				accountTrxRepo.On("FindByReference", c.ValueCtxMockType(), paymentID, c.TypePayment).Return(validAccountTrx, nil)
			},
		},
		{
			name:    "ERROR: Capture amount exceeds authorized amount",
			wantErr: true,
			request: &unifiedPaymentModel.CaptureRequest{
				PaymentID:              paymentID,
				MerchantID:             merchantID,
				ReleaseRemainingAmount: false,
				Amount: &unifiedPaymentModel.Amount{
					Currency: "IDR",
					Value:    200000,
				},
			},
			setupMock: func(paymentRepo *repositoryMock.IPaymentRepository, accountTrxRepo *repositoryMock.IAccountTransactionRepository, paymentCaptureRepo *repositoryMock.IPaymentCaptureRepository, rabbitMq *rabbitMQMock.RabbitMQExt) {
				paymentRepo.On("GetPaymentById", c.ValueCtxMockType(), paymentID).Return(validPayment, nil)
				accountTrxRepo.On("FindByReference", c.ValueCtxMockType(), paymentID, c.TypePayment).Return(validAccountTrx, nil)
			},
		},
		{
			name:    "ERROR: IDR amount with decimal",
			wantErr: true,
			request: &unifiedPaymentModel.CaptureRequest{
				PaymentID:              paymentID,
				MerchantID:             merchantID,
				ReleaseRemainingAmount: false,
				Amount: &unifiedPaymentModel.Amount{
					Currency: "IDR",
					Value:    100000.50,
				},
			},
			setupMock: func(paymentRepo *repositoryMock.IPaymentRepository, accountTrxRepo *repositoryMock.IAccountTransactionRepository, paymentCaptureRepo *repositoryMock.IPaymentCaptureRepository, rabbitMq *rabbitMQMock.RabbitMQExt) {
				paymentRepo.On("GetPaymentById", c.ValueCtxMockType(), paymentID).Return(validPayment, nil)
				accountTrxRepo.On("FindByReference", c.ValueCtxMockType(), paymentID, c.TypePayment).Return(validAccountTrx, nil)
			},
		},
		{
			name:    "ERROR: Database error when inserting payment capture",
			wantErr: true,
			request: validRequest,
			setupMock: func(paymentRepo *repositoryMock.IPaymentRepository, accountTrxRepo *repositoryMock.IAccountTransactionRepository, paymentCaptureRepo *repositoryMock.IPaymentCaptureRepository, rabbitMq *rabbitMQMock.RabbitMQExt) {
				paymentRepo.On("GetPaymentById", c.ValueCtxMockType(), paymentID).Return(validPayment, nil)
				accountTrxRepo.On("FindByReference", c.ValueCtxMockType(), paymentID, c.TypePayment).Return(validAccountTrx, nil)
				paymentCaptureRepo.On("Insert", c.ValueCtxMockType(), mock.AnythingOfType("*paymentCaptureModel.PaymentCapture")).Return(databaseError)
			},
		},
		{
			name:    "SUCCESS: Valid capture request with amount",
			wantErr: false,
			request: validRequest,
			setupMock: func(paymentRepo *repositoryMock.IPaymentRepository, accountTrxRepo *repositoryMock.IAccountTransactionRepository, paymentCaptureRepo *repositoryMock.IPaymentCaptureRepository, rabbitMq *rabbitMQMock.RabbitMQExt) {
				paymentRepo.On("GetPaymentById", c.ValueCtxMockType(), paymentID).Return(validPayment, nil)
				accountTrxRepo.On("FindByReference", c.ValueCtxMockType(), paymentID, c.TypePayment).Return(validAccountTrx, nil)
				paymentCaptureRepo.On("Insert", c.ValueCtxMockType(), mock.AnythingOfType("*paymentCaptureModel.PaymentCapture")).Return(nil)
				rabbitMq.On("Publish", c.ValueCtxMockType(), rabbitMqExt.PaymentCaptureProcessRoutingKey, mock.Anything, mock.AnythingOfType("*unifiedPaymentModel.ProcessCaptureRequest")).Return(nil)
			},
		},
		{
			name:    "SUCCESS: Valid capture request without amount (full capture)",
			wantErr: false,
			request: &unifiedPaymentModel.CaptureRequest{
				PaymentID:              paymentID,
				MerchantID:             merchantID,
				ReleaseRemainingAmount: false,
			},
			setupMock: func(paymentRepo *repositoryMock.IPaymentRepository, accountTrxRepo *repositoryMock.IAccountTransactionRepository, paymentCaptureRepo *repositoryMock.IPaymentCaptureRepository, rabbitMq *rabbitMQMock.RabbitMQExt) {
				paymentRepo.On("GetPaymentById", c.ValueCtxMockType(), paymentID).Return(validPayment, nil)
				accountTrxRepo.On("FindByReference", c.ValueCtxMockType(), paymentID, c.TypePayment).Return(validAccountTrx, nil)
				paymentCaptureRepo.On("Insert", c.ValueCtxMockType(), mock.AnythingOfType("*paymentCaptureModel.PaymentCapture")).Return(nil)
				rabbitMq.On("Publish", c.ValueCtxMockType(), rabbitMqExt.PaymentCaptureProcessRoutingKey, mock.Anything, mock.AnythingOfType("*unifiedPaymentModel.ProcessCaptureRequest")).Return(nil)
			},
		},
		{
			name:    "SUCCESS: Valid capture request with release remaining amount",
			wantErr: false,
			request: &unifiedPaymentModel.CaptureRequest{
				PaymentID:              paymentID,
				MerchantID:             merchantID,
				ReleaseRemainingAmount: true,
				Amount: &unifiedPaymentModel.Amount{
					Currency: "IDR",
					Value:    50000,
				},
			},
			setupMock: func(paymentRepo *repositoryMock.IPaymentRepository, accountTrxRepo *repositoryMock.IAccountTransactionRepository, paymentCaptureRepo *repositoryMock.IPaymentCaptureRepository, rabbitMq *rabbitMQMock.RabbitMQExt) {
				paymentRepo.On("GetPaymentById", c.ValueCtxMockType(), paymentID).Return(validPayment, nil)
				accountTrxRepo.On("FindByReference", c.ValueCtxMockType(), paymentID, c.TypePayment).Return(validAccountTrx, nil)
				paymentCaptureRepo.On("Insert", c.ValueCtxMockType(), mock.AnythingOfType("*paymentCaptureModel.PaymentCapture")).Return(nil)
				rabbitMq.On("Publish", c.ValueCtxMockType(), rabbitMqExt.PaymentCaptureProcessRoutingKey, mock.Anything, mock.AnythingOfType("*unifiedPaymentModel.ProcessCaptureRequest")).Return(nil)
			},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			paymentRepo := repositoryMock.NewIPaymentRepository(t)
			paymentMethodRepo := repositoryMock.NewIPaymentMethodRepository(t)
			accountTrxRepo := repositoryMock.NewIAccountTransactionRepository(t)
			paymentCaptureRepo := repositoryMock.NewIPaymentCaptureRepository(t)
			merchantRepo := repositoryMock.NewIMerchantRepository(t)
			paymentSvc := serviceMock.NewIPaymentService(t)
			fdsSvc := serviceMock.NewIFdsService(t)
			rabbitMq := rabbitMQMock.NewRabbitMQExt(t)
			redisExt := redisExtMocks.NewIRedisExt(t)

			tt.setupMock(paymentRepo, accountTrxRepo, paymentCaptureRepo, rabbitMq)

			svc := New(cfg, log, paymentRepo, paymentMethodRepo, accountTrxRepo,
				WithPaymentCaptureRepository(paymentCaptureRepo),
				WithMerchantRepo(merchantRepo),
				WithPaymentService(paymentSvc),
				WithFdsService(fdsSvc),
				WithRabbitMQClient(rabbitMq),
				WithRedisClient(redisExt))
			result, err := svc.Capture(context.Background(), tt.request)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, paymentID, result.PaymentSessionID)
				assert.Equal(t, referenceID, result.PaymentSessionClientReferenceId)
				assert.Equal(t, c.StatusPending, result.Status)
				assert.Equal(t, tt.request.ReleaseRemainingAmount, result.ReleaseRemainingAmount)
				assert.Equal(t, tt.request.Amount, result.Amount)
			}

			paymentRepo.AssertExpectations(t)
			accountTrxRepo.AssertExpectations(t)
			paymentCaptureRepo.AssertExpectations(t)
			rabbitMq.AssertExpectations(t)
		})
	}
}

func TestProcessCapture(t *testing.T) {
	cfg := &config.Config{
		Environment: "test",
	}
	log, _ := mockLogger.NewZapLogger(mockLogger.Config{})

	merchantID := "merchant-" + uuid.NewString()
	paymentID := "payment-" + uuid.NewString()
	paymentCaptureID := "capture-" + uuid.NewString()
	referenceID := "ref-" + uuid.NewString()
	databaseError := errors.New("database error")
	redisError := errors.New("redis error")
	accountTrxID := uuid.New()

	validPayment := &paymentModel.Payment{
		UUID:        paymentID,
		MerchantID:  merchantID,
		Amount:      decimal.NewFromFloat(100000),
		Currency:    "IDR",
		Status:      c.UnifiedPaymentSessionStatusRequireAction,
		ReferenceID: &referenceID,
		PaymentMethod: paymentModel.PaymentMethod{
			Type: paymentConstant.PAYMENT_METHOD_CREDIT_CARD,
		},
		Metadata: &map[string]interface{}{
			"paymentMethod": map[string]interface{}{
				"type": c.UnifiedPaymentMethodCard,
			},
			"paymentMethodOptions": map[string]interface{}{
				"card": map[string]interface{}{
					"captureMethod": c.UnifiedPaymentCardCaptureMethodManual,
				},
			},
		},
	}

	validAdditionalInfo := map[string]interface{}{
		"chargeStatus": c.ChargeStatusHistoryWaitingForCapture,
	}
	validAdditionalInfoBytes, _ := json.Marshal(validAdditionalInfo)
	validAccountTrx := &orchestratorModel.AccountTransactionWithUseCase{
		UUID:                 accountTrxID,
		Type:                 c.TypePayment,
		Reference:            paymentID,
		Status:               c.StatusPending,
		ProcessorReferenceId: "proc-ref-123",
		Currency:             "IDR",
		Credit:               0,
		AdditionalInfo: types.NullJSONText{
			Valid:    true,
			JSONText: validAdditionalInfoBytes,
		},
	}

	validPaymentCapture := &paymentCaptureModel.PaymentCapture{
		ID:                     paymentCaptureID,
		PaymentID:              paymentID,
		Amount:                 50000,
		Currency:               "IDR",
		Status:                 c.StatusPending,
		ReleaseRemainingAmount: false,
		CreatedAt:              time.Now().UTC(),
		UpdatedAt:              time.Now().UTC(),
	}

	validMerchant := &merchantModel.Merchant{
		UUID: merchantID,
		BusinessCountry: sql.NullString{
			String: "IDN",
			Valid:  true,
		},
	}

	validRequest := &unifiedPaymentModel.ProcessCaptureRequest{
		ID: paymentCaptureID,
	}

	testCases := []struct {
		name      string
		wantErr   bool
		setupMock func(*repositoryMock.IPaymentRepository, *repositoryMock.IAccountTransactionRepository, *repositoryMock.IPaymentCaptureRepository, *repositoryMock.IMerchantRepository, *repositoryMock.ICreditcardCoreProcessorRepository, *serviceMock.IPaymentService, *redisExtMocks.IRedisExt, *redisExtMocks.IMutexer, *rabbitMQMock.RabbitMQExt)
		request   *unifiedPaymentModel.ProcessCaptureRequest
	}{
		{
			name:    "ERROR: Redis SetNX error",
			wantErr: true,
			request: validRequest,
			setupMock: func(paymentRepo *repositoryMock.IPaymentRepository, accountTrxRepo *repositoryMock.IAccountTransactionRepository, paymentCaptureRepo *repositoryMock.IPaymentCaptureRepository, merchantRepo *repositoryMock.IMerchantRepository, creditCardRepo *repositoryMock.ICreditcardCoreProcessorRepository, paymentSvc *serviceMock.IPaymentService, redisExt *redisExtMocks.IRedisExt, mutex *redisExtMocks.IMutexer, rmq *rabbitMQMock.RabbitMQExt) {
				redisExt.On("SetNX", c.ValueCtxMockType(), mock.AnythingOfType("string"), true, 5*time.Minute).Return(redis.NewBoolResult(false, redisError))
			},
		},
		{
			name:    "ERROR: Redis SetNX returns false (already being processed)",
			wantErr: true,
			request: validRequest,
			setupMock: func(paymentRepo *repositoryMock.IPaymentRepository, accountTrxRepo *repositoryMock.IAccountTransactionRepository, paymentCaptureRepo *repositoryMock.IPaymentCaptureRepository, merchantRepo *repositoryMock.IMerchantRepository, creditCardRepo *repositoryMock.ICreditcardCoreProcessorRepository, paymentSvc *serviceMock.IPaymentService, redisExt *redisExtMocks.IRedisExt, mutex *redisExtMocks.IMutexer, rmq *rabbitMQMock.RabbitMQExt) {
				redisExt.On("SetNX", c.ValueCtxMockType(), mock.AnythingOfType("string"), true, 5*time.Minute).Return(redis.NewBoolResult(false, nil))
			},
		},
		{
			name:    "ERROR: Database error when getting payment capture",
			wantErr: true,
			request: validRequest,
			setupMock: func(paymentRepo *repositoryMock.IPaymentRepository, accountTrxRepo *repositoryMock.IAccountTransactionRepository, paymentCaptureRepo *repositoryMock.IPaymentCaptureRepository, merchantRepo *repositoryMock.IMerchantRepository, creditCardRepo *repositoryMock.ICreditcardCoreProcessorRepository, paymentSvc *serviceMock.IPaymentService, redisExt *redisExtMocks.IRedisExt, mutex *redisExtMocks.IMutexer, rmq *rabbitMQMock.RabbitMQExt) {
				redisExt.On("SetNX", c.ValueCtxMockType(), mock.AnythingOfType("string"), true, 5*time.Minute).Return(redis.NewBoolResult(true, nil))
				paymentCaptureRepo.On("GetByID", c.ValueCtxMockType(), paymentCaptureID).Return(nil, databaseError)
			},
		},
		{
			name:    "ERROR: Payment capture not found",
			wantErr: true,
			request: validRequest,
			setupMock: func(paymentRepo *repositoryMock.IPaymentRepository, accountTrxRepo *repositoryMock.IAccountTransactionRepository, paymentCaptureRepo *repositoryMock.IPaymentCaptureRepository, merchantRepo *repositoryMock.IMerchantRepository, creditCardRepo *repositoryMock.ICreditcardCoreProcessorRepository, paymentSvc *serviceMock.IPaymentService, redisExt *redisExtMocks.IRedisExt, mutex *redisExtMocks.IMutexer, rmq *rabbitMQMock.RabbitMQExt) {
				redisExt.On("SetNX", c.ValueCtxMockType(), mock.AnythingOfType("string"), true, 5*time.Minute).Return(redis.NewBoolResult(true, nil))
				paymentCaptureRepo.On("GetByID", c.ValueCtxMockType(), paymentCaptureID).Return(nil, nil)
			},
		},
		{
			name:    "ERROR: Failed to acquire mutex lock",
			wantErr: true,
			request: validRequest,
			setupMock: func(paymentRepo *repositoryMock.IPaymentRepository, accountTrxRepo *repositoryMock.IAccountTransactionRepository, paymentCaptureRepo *repositoryMock.IPaymentCaptureRepository, merchantRepo *repositoryMock.IMerchantRepository, creditCardRepo *repositoryMock.ICreditcardCoreProcessorRepository, paymentSvc *serviceMock.IPaymentService, redisExt *redisExtMocks.IRedisExt, mutex *redisExtMocks.IMutexer, rmq *rabbitMQMock.RabbitMQExt) {
				redisExt.On("SetNX", c.ValueCtxMockType(), mock.AnythingOfType("string"), true, 5*time.Minute).Return(redis.NewBoolResult(true, nil))
				paymentCaptureRepo.On("GetByID", c.ValueCtxMockType(), paymentCaptureID).Return(validPaymentCapture, nil)
				redisExt.On("NewMutex", mock.AnythingOfType("string"), mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(mutex)
				mutex.On("LockContext", c.ValueCtxMockType()).Return(errors.New("lock failed"))
			},
		},
		{
			name:    "ERROR: Database error when getting payment",
			wantErr: true,
			request: validRequest,
			setupMock: func(paymentRepo *repositoryMock.IPaymentRepository, accountTrxRepo *repositoryMock.IAccountTransactionRepository, paymentCaptureRepo *repositoryMock.IPaymentCaptureRepository, merchantRepo *repositoryMock.IMerchantRepository, creditCardRepo *repositoryMock.ICreditcardCoreProcessorRepository, paymentSvc *serviceMock.IPaymentService, redisExt *redisExtMocks.IRedisExt, mutex *redisExtMocks.IMutexer, rmq *rabbitMQMock.RabbitMQExt) {
				redisExt.On("SetNX", c.ValueCtxMockType(), mock.AnythingOfType("string"), true, 5*time.Minute).Return(redis.NewBoolResult(true, nil))
				paymentCaptureRepo.On("GetByID", c.ValueCtxMockType(), paymentCaptureID).Return(validPaymentCapture, nil)
				redisExt.On("NewMutex", mock.AnythingOfType("string"), mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(mutex)
				mutex.On("LockContext", c.ValueCtxMockType()).Return(nil)
				mutex.On("UnlockContext", c.ValueCtxMockType()).Return(true, nil)
				paymentRepo.On("GetPaymentById", c.ValueCtxMockType(), paymentID).Return(nil, databaseError)
			},
		},
		{
			name:    "ERROR: Payment not found",
			wantErr: true,
			request: validRequest,
			setupMock: func(paymentRepo *repositoryMock.IPaymentRepository, accountTrxRepo *repositoryMock.IAccountTransactionRepository, paymentCaptureRepo *repositoryMock.IPaymentCaptureRepository, merchantRepo *repositoryMock.IMerchantRepository, creditCardRepo *repositoryMock.ICreditcardCoreProcessorRepository, paymentSvc *serviceMock.IPaymentService, redisExt *redisExtMocks.IRedisExt, mutex *redisExtMocks.IMutexer, rmq *rabbitMQMock.RabbitMQExt) {
				redisExt.On("SetNX", c.ValueCtxMockType(), mock.AnythingOfType("string"), true, 5*time.Minute).Return(redis.NewBoolResult(true, nil))
				paymentCaptureRepo.On("GetByID", c.ValueCtxMockType(), paymentCaptureID).Return(validPaymentCapture, nil)
				redisExt.On("NewMutex", mock.AnythingOfType("string"), mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(mutex)
				mutex.On("LockContext", c.ValueCtxMockType()).Return(nil)
				mutex.On("UnlockContext", c.ValueCtxMockType()).Return(true, nil)
				paymentRepo.On("GetPaymentById", c.ValueCtxMockType(), paymentID).Return(nil, nil)
			},
		},
		{
			name:    "ERROR: Database error when finding account transaction",
			wantErr: true,
			request: validRequest,
			setupMock: func(paymentRepo *repositoryMock.IPaymentRepository, accountTrxRepo *repositoryMock.IAccountTransactionRepository, paymentCaptureRepo *repositoryMock.IPaymentCaptureRepository, merchantRepo *repositoryMock.IMerchantRepository, creditCardRepo *repositoryMock.ICreditcardCoreProcessorRepository, paymentSvc *serviceMock.IPaymentService, redisExt *redisExtMocks.IRedisExt, mutex *redisExtMocks.IMutexer, rmq *rabbitMQMock.RabbitMQExt) {
				redisExt.On("SetNX", c.ValueCtxMockType(), mock.AnythingOfType("string"), true, 5*time.Minute).Return(redis.NewBoolResult(true, nil))
				paymentCaptureRepo.On("GetByID", c.ValueCtxMockType(), paymentCaptureID).Return(validPaymentCapture, nil)
				redisExt.On("NewMutex", mock.AnythingOfType("string"), mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(mutex)
				mutex.On("LockContext", c.ValueCtxMockType()).Return(nil)
				mutex.On("UnlockContext", c.ValueCtxMockType()).Return(true, nil)
				paymentRepo.On("GetPaymentById", c.ValueCtxMockType(), paymentID).Return(validPayment, nil)
				accountTrxRepo.On("FindByReference", c.ValueCtxMockType(), paymentID, c.TypePayment).Return(nil, databaseError)
			},
		},
		{
			name:    "ERROR: Account transaction not found",
			wantErr: true,
			request: validRequest,
			setupMock: func(paymentRepo *repositoryMock.IPaymentRepository, accountTrxRepo *repositoryMock.IAccountTransactionRepository, paymentCaptureRepo *repositoryMock.IPaymentCaptureRepository, merchantRepo *repositoryMock.IMerchantRepository, creditCardRepo *repositoryMock.ICreditcardCoreProcessorRepository, paymentSvc *serviceMock.IPaymentService, redisExt *redisExtMocks.IRedisExt, mutex *redisExtMocks.IMutexer, rmq *rabbitMQMock.RabbitMQExt) {
				redisExt.On("SetNX", c.ValueCtxMockType(), mock.AnythingOfType("string"), true, 5*time.Minute).Return(redis.NewBoolResult(true, nil))
				paymentCaptureRepo.On("GetByID", c.ValueCtxMockType(), paymentCaptureID).Return(validPaymentCapture, nil)
				redisExt.On("NewMutex", mock.AnythingOfType("string"), mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(mutex)
				mutex.On("LockContext", c.ValueCtxMockType()).Return(nil)
				mutex.On("UnlockContext", c.ValueCtxMockType()).Return(true, nil)
				paymentRepo.On("GetPaymentById", c.ValueCtxMockType(), paymentID).Return(validPayment, nil)
				accountTrxRepo.On("FindByReference", c.ValueCtxMockType(), paymentID, c.TypePayment).Return(nil, nil)
			},
		},
		{
			name:    "ERROR: Capture amount exceeds authorized amount",
			wantErr: true,
			request: validRequest,
			setupMock: func(paymentRepo *repositoryMock.IPaymentRepository, accountTrxRepo *repositoryMock.IAccountTransactionRepository, paymentCaptureRepo *repositoryMock.IPaymentCaptureRepository, merchantRepo *repositoryMock.IMerchantRepository, creditCardRepo *repositoryMock.ICreditcardCoreProcessorRepository, paymentSvc *serviceMock.IPaymentService, redisExt *redisExtMocks.IRedisExt, mutex *redisExtMocks.IMutexer, rmq *rabbitMQMock.RabbitMQExt) {
				redisExt.On("SetNX", c.ValueCtxMockType(), mock.AnythingOfType("string"), true, 5*time.Minute).Return(redis.NewBoolResult(true, nil))

				capture := *validPaymentCapture
				capture.Amount = 200000 // Exceeds authorized amount
				paymentCaptureRepo.On("GetByID", c.ValueCtxMockType(), paymentCaptureID).Return(&capture, nil)

				redisExt.On("NewMutex", mock.AnythingOfType("string"), mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(mutex)
				mutex.On("LockContext", c.ValueCtxMockType()).Return(nil)
				mutex.On("UnlockContext", c.ValueCtxMockType()).Return(true, nil)
				paymentRepo.On("GetPaymentById", c.ValueCtxMockType(), paymentID).Return(validPayment, nil)
				accountTrxRepo.On("FindByReference", c.ValueCtxMockType(), paymentID, c.TypePayment).Return(validAccountTrx, nil)
			},
		},
		{
			name:    "ERROR: Processor capture failed",
			wantErr: true,
			request: validRequest,
			setupMock: func(paymentRepo *repositoryMock.IPaymentRepository, accountTrxRepo *repositoryMock.IAccountTransactionRepository, paymentCaptureRepo *repositoryMock.IPaymentCaptureRepository, merchantRepo *repositoryMock.IMerchantRepository, creditCardRepo *repositoryMock.ICreditcardCoreProcessorRepository, paymentSvc *serviceMock.IPaymentService, redisExt *redisExtMocks.IRedisExt, mutex *redisExtMocks.IMutexer, rmq *rabbitMQMock.RabbitMQExt) {
				redisExt.On("SetNX", c.ValueCtxMockType(), mock.AnythingOfType("string"), true, 5*time.Minute).Return(redis.NewBoolResult(true, nil))
				paymentCaptureRepo.On("GetByID", c.ValueCtxMockType(), paymentCaptureID).Return(validPaymentCapture, nil)
				redisExt.On("NewMutex", mock.AnythingOfType("string"), mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(mutex)
				mutex.On("LockContext", c.ValueCtxMockType()).Return(nil)
				mutex.On("UnlockContext", c.ValueCtxMockType()).Return(true, nil)
				paymentRepo.On("GetPaymentById", c.ValueCtxMockType(), paymentID).Return(validPayment, nil)
				accountTrxRepo.On("FindByReference", c.ValueCtxMockType(), paymentID, c.TypePayment).Return(validAccountTrx, nil)
				creditCardRepo.On("Capture", c.ValueCtxMockType(), mock.AnythingOfType("*creditcardCoreProcessorModel.CaptureRequest")).Return(nil, errors.New("processor error"))
			},
		},
		{
			name:    "ERROR: Processor capture response status not success",
			wantErr: true,
			request: validRequest,
			setupMock: func(paymentRepo *repositoryMock.IPaymentRepository, accountTrxRepo *repositoryMock.IAccountTransactionRepository, paymentCaptureRepo *repositoryMock.IPaymentCaptureRepository, merchantRepo *repositoryMock.IMerchantRepository, creditCardRepo *repositoryMock.ICreditcardCoreProcessorRepository, paymentSvc *serviceMock.IPaymentService, redisExt *redisExtMocks.IRedisExt, mutex *redisExtMocks.IMutexer, rmq *rabbitMQMock.RabbitMQExt) {
				redisExt.On("SetNX", c.ValueCtxMockType(), mock.AnythingOfType("string"), true, 5*time.Minute).Return(redis.NewBoolResult(true, nil))
				paymentCaptureRepo.On("GetByID", c.ValueCtxMockType(), paymentCaptureID).Return(validPaymentCapture, nil)
				redisExt.On("NewMutex", mock.AnythingOfType("string"), mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(mutex)
				mutex.On("LockContext", c.ValueCtxMockType()).Return(nil)
				mutex.On("UnlockContext", c.ValueCtxMockType()).Return(true, nil)
				paymentRepo.On("GetPaymentById", c.ValueCtxMockType(), paymentID).Return(validPayment, nil)
				accountTrxRepo.On("FindByReference", c.ValueCtxMockType(), paymentID, c.TypePayment).Return(validAccountTrx, nil)
				creditCardRepo.On("Capture", c.ValueCtxMockType(), mock.AnythingOfType("*creditcardCoreProcessorModel.CaptureRequest")).Return(&creditcardCoreProcessorModel.CaptureResponseData{
					Status: c.StatusFailed,
				}, nil)
			},
		},
		{
			name:    "ERROR: Database error when beginning transaction",
			wantErr: true,
			request: validRequest,
			setupMock: func(paymentRepo *repositoryMock.IPaymentRepository, accountTrxRepo *repositoryMock.IAccountTransactionRepository, paymentCaptureRepo *repositoryMock.IPaymentCaptureRepository, merchantRepo *repositoryMock.IMerchantRepository, creditCardRepo *repositoryMock.ICreditcardCoreProcessorRepository, paymentSvc *serviceMock.IPaymentService, redisExt *redisExtMocks.IRedisExt, mutex *redisExtMocks.IMutexer, rmq *rabbitMQMock.RabbitMQExt) {
				redisExt.On("SetNX", c.ValueCtxMockType(), mock.AnythingOfType("string"), true, 5*time.Minute).Return(redis.NewBoolResult(true, nil))
				paymentCaptureRepo.On("GetByID", c.ValueCtxMockType(), paymentCaptureID).Return(validPaymentCapture, nil)
				redisExt.On("NewMutex", mock.AnythingOfType("string"), mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(mutex)
				mutex.On("LockContext", c.ValueCtxMockType()).Return(nil)
				mutex.On("UnlockContext", c.ValueCtxMockType()).Return(true, nil)
				paymentRepo.On("GetPaymentById", c.ValueCtxMockType(), paymentID).Return(validPayment, nil)
				accountTrxRepo.On("FindByReference", c.ValueCtxMockType(), paymentID, c.TypePayment).Return(validAccountTrx, nil)
				creditCardRepo.On("Capture", c.ValueCtxMockType(), mock.AnythingOfType("*creditcardCoreProcessorModel.CaptureRequest")).Return(&creditcardCoreProcessorModel.CaptureResponseData{
					Status: c.StatusSuccess,
					ID:     "capture-proc-123",
				}, nil)
				paymentRepo.On("BeginTransaction", c.ValueCtxMockType()).Return(context.WithValue(context.Background(), mySqlExt.CtxSqlTx, &sqlx.Tx{}), databaseError)
			},
		},
		{
			name:    "SUCCESS: Partial capture without release remaining amount",
			wantErr: false,
			request: validRequest,
			setupMock: func(paymentRepo *repositoryMock.IPaymentRepository, accountTrxRepo *repositoryMock.IAccountTransactionRepository, paymentCaptureRepo *repositoryMock.IPaymentCaptureRepository, merchantRepo *repositoryMock.IMerchantRepository, creditCardRepo *repositoryMock.ICreditcardCoreProcessorRepository, paymentSvc *serviceMock.IPaymentService, redisExt *redisExtMocks.IRedisExt, mutex *redisExtMocks.IMutexer, rmq *rabbitMQMock.RabbitMQExt) {
				redisExt.On("SetNX", c.ValueCtxMockType(), mock.AnythingOfType("string"), true, 5*time.Minute).Return(redis.NewBoolResult(true, nil))
				paymentCaptureRepo.On("GetByID", c.ValueCtxMockType(), paymentCaptureID).Return(validPaymentCapture, nil)
				redisExt.On("NewMutex", mock.AnythingOfType("string"), mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(mutex)
				mutex.On("LockContext", c.ValueCtxMockType()).Return(nil)
				mutex.On("UnlockContext", c.ValueCtxMockType()).Return(true, nil)
				paymentRepo.On("GetPaymentById", c.ValueCtxMockType(), paymentID).Return(validPayment, nil)
				accountTrxRepo.On("FindByReference", c.ValueCtxMockType(), paymentID, c.TypePayment).Return(validAccountTrx, nil)
				creditCardRepo.On("Capture", c.ValueCtxMockType(), mock.AnythingOfType("*creditcardCoreProcessorModel.CaptureRequest")).Return(&creditcardCoreProcessorModel.CaptureResponseData{
					Status: c.StatusSuccess,
					ID:     "capture-proc-123",
				}, nil)
				ctxTx := context.WithValue(context.Background(), mySqlExt.CtxSqlTx, &sqlx.Tx{})
				paymentRepo.On("BeginTransaction", c.ValueCtxMockType()).Return(ctxTx, nil)
				accountTrxRepo.On("UpdateCreditDebitByID", ctxTx, accountTrxID.String(), mock.AnythingOfType("*float64"), (*float64)(nil)).Return(nil)
				paymentCaptureRepo.On("Update", ctxTx, mock.AnythingOfType("*paymentCaptureModel.PaymentCapture")).Return(nil)
				paymentRepo.On("CommitTransaction", ctxTx).Return(nil)
				accountTrxRepo.On("FindByID", c.ValueCtxMockType(), c.StringMockType()).
					Once().Return(&orchestratorModel.AccountTransactionWithUseCase{
					UUID:            uuid.New(),
					SettlementModel: sql.NullString{Valid: true, String: c.PaymentMethodChannelTypeAggregator},
				}, nil)
				rmq.On("PublishMerchantCallback", mock.Anything, mock.Anything).Return(nil)
				rmq.On("PushNotification", mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			name:    "SUCCESS: Full capture",
			wantErr: false,
			request: validRequest,
			setupMock: func(paymentRepo *repositoryMock.IPaymentRepository, accountTrxRepo *repositoryMock.IAccountTransactionRepository, paymentCaptureRepo *repositoryMock.IPaymentCaptureRepository, merchantRepo *repositoryMock.IMerchantRepository, creditCardRepo *repositoryMock.ICreditcardCoreProcessorRepository, paymentSvc *serviceMock.IPaymentService, redisExt *redisExtMocks.IRedisExt, mutex *redisExtMocks.IMutexer, rmq *rabbitMQMock.RabbitMQExt) {
				redisExt.On("SetNX", c.ValueCtxMockType(), mock.AnythingOfType("string"), true, 5*time.Minute).Return(redis.NewBoolResult(true, nil))

				fullCapture := *validPaymentCapture
				fullCapture.Amount = 100000 // Full amount
				paymentCaptureRepo.On("GetByID", c.ValueCtxMockType(), paymentCaptureID).Return(&fullCapture, nil)

				redisExt.On("NewMutex", mock.AnythingOfType("string"), mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(mutex)
				mutex.On("LockContext", c.ValueCtxMockType()).Return(nil)
				mutex.On("UnlockContext", c.ValueCtxMockType()).Return(true, nil)
				paymentRepo.On("GetPaymentById", c.ValueCtxMockType(), paymentID).Return(validPayment, nil)
				accountTrxRepo.On("FindByReference", c.ValueCtxMockType(), paymentID, c.TypePayment).Return(validAccountTrx, nil)
				creditCardRepo.On("Capture", c.ValueCtxMockType(), mock.AnythingOfType("*creditcardCoreProcessorModel.CaptureRequest")).Return(&creditcardCoreProcessorModel.CaptureResponseData{
					Status: c.StatusSuccess,
					ID:     "capture-proc-123",
				}, nil)
				ctxTx := context.WithValue(context.Background(), mySqlExt.CtxSqlTx, &sqlx.Tx{})
				paymentRepo.On("BeginTransaction", c.ValueCtxMockType()).Return(ctxTx, nil)
				accountTrxRepo.On("UpdateCreditDebitByID", ctxTx, accountTrxID.String(), mock.AnythingOfType("*float64"), (*float64)(nil)).Return(nil)
				paymentCaptureRepo.On("Update", ctxTx, mock.AnythingOfType("*paymentCaptureModel.PaymentCapture")).Return(nil)

				// For full capture, payment status changes to PAID
				paymentRepo.On("UpdatePaymentData", ctxTx, mock.AnythingOfType("*paymentModel.PaymentDTO")).Return(nil)
				merchantRepo.On("FindMerchantByID", c.ValueCtxMockType(), merchantID).Return(validMerchant, nil)
				paymentSvc.On("DeterminePaymentFee", &ctxTx, mock.AnythingOfType("*paymentModel.Payment")).Return(nil)
				paymentSvc.On("UpdatePendingLedger", ctxTx, mock.AnythingOfType("*paymentModel.Payment"), mock.AnythingOfType("orchestrator_model.UpdatePaymentTransactionRequest")).Return(nil)
				paymentRepo.On("CommitTransaction", ctxTx).Return(nil)

				accountTrxRepo.On("FindByID", c.ValueCtxMockType(), c.StringMockType()).
					Once().Return(&orchestratorModel.AccountTransactionWithUseCase{
					UUID:            uuid.New(),
					SettlementModel: sql.NullString{Valid: true, String: c.PaymentMethodChannelTypeAggregator},
				}, nil)
				rmq.On("PublishMerchantCallback", mock.Anything, mock.Anything).Return(nil)
				rmq.On("PushNotification", c.ValueCtxMockType(), c.PtrPushNotificationMockType()).
					Twice().Return(nil)
			},
		},
		{
			name:    "SUCCESS: Release remaining amount with zero capture",
			wantErr: false,
			request: validRequest,
			setupMock: func(paymentRepo *repositoryMock.IPaymentRepository, accountTrxRepo *repositoryMock.IAccountTransactionRepository, paymentCaptureRepo *repositoryMock.IPaymentCaptureRepository, merchantRepo *repositoryMock.IMerchantRepository, creditCardRepo *repositoryMock.ICreditcardCoreProcessorRepository, paymentSvc *serviceMock.IPaymentService, redisExt *redisExtMocks.IRedisExt, mutex *redisExtMocks.IMutexer, rmq *rabbitMQMock.RabbitMQExt) {
				redisExt.On("SetNX", c.ValueCtxMockType(), mock.AnythingOfType("string"), true, 5*time.Minute).Return(redis.NewBoolResult(true, nil))

				releaseCapture := *validPaymentCapture
				releaseCapture.Amount = 0
				releaseCapture.ReleaseRemainingAmount = true
				paymentCaptureRepo.On("GetByID", c.ValueCtxMockType(), paymentCaptureID).Return(&releaseCapture, nil)

				redisExt.On("NewMutex", mock.AnythingOfType("string"), mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(mutex)
				mutex.On("LockContext", c.ValueCtxMockType()).Return(nil)
				mutex.On("UnlockContext", c.ValueCtxMockType()).Return(true, nil)
				paymentRepo.On("GetPaymentById", c.ValueCtxMockType(), paymentID).Return(validPayment, nil)
				accountTrxRepo.On("FindByReference", c.ValueCtxMockType(), paymentID, c.TypePayment).Return(validAccountTrx, nil)
				creditCardRepo.On("Capture", c.ValueCtxMockType(), mock.AnythingOfType("*creditcardCoreProcessorModel.CaptureRequest")).Return(&creditcardCoreProcessorModel.CaptureResponseData{
					Status: c.StatusSuccess,
					ID:     "capture-proc-123",
				}, nil)
				ctxTx := context.WithValue(context.Background(), mySqlExt.CtxSqlTx, &sqlx.Tx{})
				paymentRepo.On("BeginTransaction", c.ValueCtxMockType()).Return(ctxTx, nil)
				accountTrxRepo.On("UpdateCreditDebitByID", ctxTx, accountTrxID.String(), mock.AnythingOfType("*float64"), (*float64)(nil)).Return(nil)
				paymentCaptureRepo.On("Update", ctxTx, mock.AnythingOfType("*paymentCaptureModel.PaymentCapture")).Return(nil)

				// For zero capture with release, payment status changes to CANCELLED
				paymentRepo.On("UpdatePaymentData", ctxTx, mock.AnythingOfType("*paymentModel.PaymentDTO")).Return(nil)
				merchantRepo.On("FindMerchantByID", c.ValueCtxMockType(), merchantID).Return(validMerchant, nil)
				paymentSvc.On("DeterminePaymentFee", &ctxTx, mock.AnythingOfType("*paymentModel.Payment")).Return(nil)
				paymentSvc.On("UpdatePendingLedger", ctxTx, mock.AnythingOfType("*paymentModel.Payment"), mock.AnythingOfType("orchestrator_model.UpdatePaymentTransactionRequest")).Return(nil)
				paymentRepo.On("CommitTransaction", ctxTx).Return(nil)

				accountTrxRepo.On("FindByID", c.ValueCtxMockType(), c.StringMockType()).
					Once().Return(&orchestratorModel.AccountTransactionWithUseCase{
					UUID:            uuid.New(),
					SettlementModel: sql.NullString{Valid: true, String: c.PaymentMethodChannelTypeAggregator},
				}, nil)
				rmq.On("PublishMerchantCallback", mock.Anything, mock.Anything).Return(nil)
				rmq.On("PushNotification", c.ValueCtxMockType(), c.PtrPushNotificationMockType()).
					Twice().Return(nil)
			},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			paymentRepo := repositoryMock.NewIPaymentRepository(t)
			paymentMethodRepo := repositoryMock.NewIPaymentMethodRepository(t)
			accountTrxRepo := repositoryMock.NewIAccountTransactionRepository(t)
			paymentCaptureRepo := repositoryMock.NewIPaymentCaptureRepository(t)
			merchantRepo := repositoryMock.NewIMerchantRepository(t)
			paymentSvc := serviceMock.NewIPaymentService(t)
			fdsSvc := serviceMock.NewIFdsService(t)
			rabbitMq := rabbitMQMock.NewRabbitMQExt(t)
			redisExt := redisExtMocks.NewIRedisExt(t)
			mutex := redisExtMocks.NewIMutexer(t)
			creditCardRepo := repositoryMock.NewICreditcardCoreProcessorRepository(t)

			tt.setupMock(paymentRepo, accountTrxRepo, paymentCaptureRepo, merchantRepo, creditCardRepo, paymentSvc, redisExt, mutex, rabbitMq)

			svc := New(cfg, log, paymentRepo, paymentMethodRepo, accountTrxRepo,
				WithPaymentCaptureRepository(paymentCaptureRepo),
				WithMerchantRepo(merchantRepo),
				WithPaymentService(paymentSvc),
				WithFdsService(fdsSvc),
				WithRabbitMQClient(rabbitMq),
				WithRedisClient(redisExt),
				WithCreditCardCoreProcessorRepo(creditCardRepo))
			err := svc.ProcessCapture(context.Background(), tt.request)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			paymentRepo.AssertExpectations(t)
			accountTrxRepo.AssertExpectations(t)
			paymentCaptureRepo.AssertExpectations(t)
			redisExt.AssertExpectations(t)
		})
	}
}
