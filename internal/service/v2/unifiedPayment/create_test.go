package unifiedPaymentService_test

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	c "github.com/paper-indonesia/pivot-backoffice/constant"
	paymentConstant "github.com/paper-indonesia/pivot-backoffice/constant/payment"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	paymentMethodModel "github.com/paper-indonesia/pivot-backoffice/internal/model/paymentMethod"
	qrisModel "github.com/paper-indonesia/pivot-backoffice/internal/model/qris"
	shortLinkModel "github.com/paper-indonesia/pivot-backoffice/internal/model/shortLink"
	unifiedPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/unifiedPayment"
	. "github.com/paper-indonesia/pivot-backoffice/internal/service/v2/unifiedPayment"
	gcpMock "github.com/paper-indonesia/pivot-backoffice/mocks/pdk/gcp"
	jwtMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/jwt"
	mockRabbitMq "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/rabbitmqExt"
	mockRedis "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/redisExt"
	repositoryMock "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	serviceMock "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	pkgErr "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pdk/v2/gcp"
	mockLogger "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreateSession(t *testing.T) {
	cfg := &config.Config{
		Environment: "test",
		PaymentUIConfig: config.PaymentUIConfig{
			PaymentLinkURL: "link.here",
		},
		UnifiedPaymentConfig: config.UnifiedPaymentConfig{
			ExpiryConfig: &config.UnifiedPaymentExpiryConfig{
				Mode: paymentConstant.UnifiedPaymentExpiryModeOff,
			},
		},
		AutoSplitPayment: config.AutoSplitPaymentConfig{
			ProcessorLimitDefault: 2000000000,
		},
	}
	jwt := jwtMocks.NewIJwt(t)
	log, _ := mockLogger.NewZapLogger(mockLogger.Config{})
	redisExt := mockRedis.NewIRedisExt(t)
	rmq := mockRabbitMq.NewRabbitMQExt(t)
	paymentRepo := repositoryMock.NewIPaymentRepository(t)
	paymentMethodSvc := serviceMock.NewIPaymentMethodService(t)
	accountTrxRepo := repositoryMock.NewIAccountTransactionRepository(t)
	merchantRepo := repositoryMock.NewIMerchantRepository(t)
	qrisSvc := serviceMock.NewIQrisService(t)
	paymentSvc := serviceMock.NewIPaymentService(t)
	shortLinkSvc := serviceMock.NewIShortLinkService(t)
	internalUnifiedPaymentSvc := serviceMock.NewIInternalUnifiedPaymentService(t)
	recurringContractRepo := repositoryMock.NewIRecurringContractRepository(t)

	secretManagerClient := gcpMock.NewIGCPSecretManager(t)
	gcp.SetGlobalSecretManagerClient(secretManagerClient)

	paymentMethodCard := &unifiedPaymentModel.CreateUnifiedPaymentSessionRequest{
		PaymentMethod: &unifiedPaymentModel.PaymentMethod{
			Type: c.UnifiedPaymentMethodCard,
		},
		AutoConfirm: false,
		Mode:        c.UnifiedPaymentModeAPI,
	}

	parentMerchantId := uuid.NewString()

	testCases := []struct {
		name      string
		wantErr   bool
		setupMock func()
		request   *unifiedPaymentModel.CreateUnifiedPaymentSessionRequest
	}{
		{
			name:    "ERROR: Find merchant got error database",
			wantErr: true,
			setupMock: func() {
				merchantRepo.On("FindMerchantByID",
					c.ValueCtxMockType(),
					c.StringMockType(),
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
		},
		{
			name:    "ERROR: Merchant not found",
			wantErr: true,
			setupMock: func() {
				merchantRepo.On("FindMerchantByID",
					c.ValueCtxMockType(),
					c.StringMockType(),
				).Once().Return(nil, nil)
			},
		},
		{
			name:    "ERROR: Find parent merchant got error database",
			wantErr: true,
			setupMock: func() {
				merchantRepo.On("FindMerchantByID",
					c.ValueCtxMockType(),
					c.StringMockType(),
				).Once().Return(&merchantModel.Merchant{
					ParentID: sql.NullString{
						Valid:  true,
						String: parentMerchantId,
					},
				}, nil)

				merchantRepo.On("FindMerchantByID",
					c.ValueCtxMockType(),
					c.StringMockType(),
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
		},
		{
			name:    "ERROR: Payment is not activated for sub merchant",
			wantErr: true,
			setupMock: func() {
				merchantRepo.On("FindMerchantByID",
					c.ValueCtxMockType(),
					c.StringMockType(),
				).Once().Return(&merchantModel.Merchant{
					ParentID: sql.NullString{
						Valid:  true,
						String: parentMerchantId,
					},
				}, nil)

				merchantRepo.On("FindMerchantByID",
					c.ValueCtxMockType(),
					c.StringMockType(),
				).Once().Return(&merchantModel.Merchant{
					KYCStatus: sql.NullString{
						Valid:  false,
						String: c.KYCStatusNotRequired,
					},
				}, nil)

				paymentMethodSvc.On("GetActivePaymentMethodDetailForPaymentRequest", c.ValueCtxMockType(), mock.Anything).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
		},
		{
			name: "ERROR: Prepare recurring payment request",
			request: &unifiedPaymentModel.CreateUnifiedPaymentSessionRequest{
				PaymentMethod: &unifiedPaymentModel.PaymentMethod{
					Type: c.UnifiedPaymentMethodCard,
				},
				ExpiryAt:    time.Now().Add(time.Hour),
				RecurringID: "718f931d-e6b0-49c2-ac0a-553aecada667",
			},
			setupMock: func() {
				merchantRepo.On(
					"FindMerchantByID", c.ValueCtxMockType(), c.StringMockType(),
				).Return(&merchantModel.Merchant{}, nil)
				internalUnifiedPaymentSvc.On(
					"PrepareRecurringPaymentRequest", mock.Anything, mock.Anything,
				).Once().Return(assert.AnError)
			},
			wantErr: true,
		},
		{
			name:    "ERROR: Get payment method error database",
			wantErr: true,
			setupMock: func() {
				paymentMethodSvc.On("GetActivePaymentMethodDetailForPaymentRequest", c.ValueCtxMockType(), mock.Anything).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
		},
		{
			name:    "ERROR: Payment Method Not Found (VA)",
			wantErr: true,
			setupMock: func() {
				paymentMethodSvc.On("GetActivePaymentMethodDetailForPaymentRequest", c.ValueCtxMockType(), mock.Anything).Once().Return(nil, nil)
			},
		},
		{
			name:    "ERROR: Payment Method Not Found (EWallet)",
			wantErr: true,
			setupMock: func() {
				paymentMethodSvc.On("GetActivePaymentMethodDetailForPaymentRequest", c.ValueCtxMockType(), mock.Anything).Once().Return(nil, nil)
			},
			request: &unifiedPaymentModel.CreateUnifiedPaymentSessionRequest{
				PaymentMethod: &unifiedPaymentModel.PaymentMethod{
					Type: c.UnifiedPaymentMethodEWallet,
				},
				PaymentMethodOptions: unifiedPaymentModel.PaymentMethodOptions{
					Ewallet: &unifiedPaymentModel.PaymentMethodOptionEwallet{
						Channel: "DANA", // NOSONAR
					},
				},
			},
		},
		{
			name:    "ERROR: Payment method is not activated",
			wantErr: true,
			setupMock: func() {
				paymentMethodSvc.On("GetActivePaymentMethodDetailForPaymentRequest", c.ValueCtxMockType(), mock.Anything).Once().Return(nil, nil)
			},
		},
		{
			name: "ERROR: Recurring payments are not allowed",
			request: &unifiedPaymentModel.CreateUnifiedPaymentSessionRequest{
				PaymentMethod: &unifiedPaymentModel.PaymentMethod{
					Type: c.UnifiedPaymentMethodCard,
				},
				ExpiryAt:                   time.Now().Add(time.Hour),
				RecurringID:                "718f931d-e6b0-49c2-ac0a-553aecada667",
				InitiateFirstAuthorization: true,
			},
			setupMock: func() {
				paymentMethodSvc.On(
					"GetActivePaymentMethodDetailForPaymentRequest", c.ValueCtxMockType(), mock.Anything,
				).Once().Return(&paymentModel.PaymentMethodWithPivot{PaymentMethod: paymentModel.PaymentMethod{UUID: uuid.NewString()}}, nil)
				internalUnifiedPaymentSvc.On(
					"PrepareRecurringPaymentRequest", mock.Anything, mock.Anything,
				).Run(func(args mock.Arguments) {
					request := args.Get(1).(*unifiedPaymentModel.CreateUnifiedPaymentSessionRequest)
					request.RecurringStatus = constant.RecurringContractStatusCreated
					request.CleanupPreparedRecurringPaymentLock = func(context.Context) {}
				}).Return(nil)
			},
			wantErr: true,
		},
		{
			name:    "ERROR: Got error validate merchant reference",
			wantErr: true,
			setupMock: func() {
				paymentMethodSvc.On(
					"GetActivePaymentMethodDetailForPaymentRequest", c.ValueCtxMockType(), mock.Anything,
				).Return(&paymentModel.PaymentMethodWithPivot{
					PaymentMethod: paymentModel.PaymentMethod{UUID: uuid.NewString()},
					MerchantConfigObj: &paymentModel.PaymentMethodMerchantConfigObject{
						PartnerConfig: &paymentMethodModel.SetupPaymentMethodPartnerConfigRequest{
							Card: &paymentMethodModel.SetupPaymentMethodPartnerConfigForCardRequest{
								Items: []paymentMethodModel.SetupPaymentMethodPartnerConfigForCardObj{
									{SupportedUseCase: &paymentMethodModel.CardSupportedUseCase{}, IsActive: true},
									{RecurringType: constant.CardTransactionTypeCIT, SupportedUseCase: &paymentMethodModel.CardSupportedUseCase{}, IsActive: true},
									{RecurringType: constant.CardTransactionTypeMIT, SupportedUseCase: &paymentMethodModel.CardSupportedUseCase{}, IsActive: true},
								},
							},
						},
					},
				}, nil)
				paymentRepo.On(
					"GetPaymentByMerchantAndReferenceId", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType(),
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
		},

		{
			name:    "ERROR: Got error validate merchant reference due to processing status",
			wantErr: true,
			setupMock: func() {
				paymentRepo.On("GetPaymentByMerchantAndReferenceId",
					c.ValueCtxMockType(),
					c.StringMockType(),
					c.StringMockType(),
				).Once().Return(&paymentModel.Payment{Status: c.UnifiedPaymentSessionStatusProcessing}, nil)
			},
		},
		{
			name:    "ERROR: Got error generate payment token",
			wantErr: true,
			setupMock: func() {
				paymentRepo.On("GetPaymentByMerchantAndReferenceId",
					c.ValueCtxMockType(),
					c.StringMockType(),
					c.StringMockType(),
				).Return(nil, nil)

				jwt.On("GeneratePaymentToken",
					c.StringMockType(),
					c.TimeMockType(),
				).Once().Return("", c.ErrSomeErrorForUnitTest)
			},
		},
		{
			name:    "ERROR: Got error generate payment token when set redis",
			wantErr: true,
			setupMock: func() {
				jwt.On("GeneratePaymentToken",
					c.StringMockType(),
					c.TimeMockType(),
				).Return("token", nil)

				redisResp := &redis.StatusCmd{}
				redisResp.SetErr(c.ErrSomeErrorForUnitTest)
				redisExt.On(
					"Set", c.ValueCtxMockType(), c.StringMockType(), c.BoolMockType(), c.DurationMockType(),
				).Once().Return(redisResp)
			},
		},
		{
			name:    "ERROR: Begin transaction",
			wantErr: true,
			setupMock: func() {
				redisExt.On(
					"Set", c.ValueCtxMockType(), c.StringMockType(), c.BoolMockType(), c.DurationMockType(),
				).Return(&redis.StatusCmd{})

				paymentRepo.On("BeginTransaction", c.ValueCtxMockType()).Once().Return(nil, c.ErrSomeErrorForUnitTest)

				redisExt.On("Del", c.ValueCtxMockType(), c.StringMockType()).Return(nil)
			},
		},
		{
			name:    "ERROR: Get latest secret value json",
			request: paymentMethodCard,
			setupMock: func() {
				paymentRepo.On("BeginTransaction", c.ValueCtxMockType()).Return(context.WithValue(context.Background(), c.CtxTest, ""), nil)
				secretManagerClient.On(
					"GetLatestSecretValueJSON", mock.Anything, mock.Anything, mock.Anything, mock.Anything,
				).Once().Return(0, assert.AnError)
				paymentRepo.On("RollbackTransaction", c.ValueCtxMockType()).Once().Return(nil)
			},
			wantErr: true,
		},
		{
			name:    "ERROR: Encrypt RSA key pair",
			request: paymentMethodCard,
			setupMock: func() {
				secretManagerClient.On(
					"GetLatestSecretValueJSON", mock.Anything, mock.Anything, mock.Anything, mock.Anything,
				).Once().Return(1, nil)
				paymentRepo.On("RollbackTransaction", c.ValueCtxMockType()).Once().Return(nil)
			},
			wantErr: true,
		},
		{
			name:    "ERROR: Create payment but got failed rollback",
			wantErr: true,
			setupMock: func() {
				paymentRepo.On("CreatePayment", c.ValueCtxMockType(), mock.AnythingOfType("*paymentModel.PaymentDTO")).Once().Return(c.ErrSomeErrorForUnitTest)
				paymentRepo.On("RollbackTransaction", c.ValueCtxMockType()).Return(c.ErrSomeErrorForUnitTest)
			},
		},
		{
			name:    "ERROR: Create payment with success rollback",
			wantErr: true,
			setupMock: func() {
				paymentRepo.On("CreatePayment", c.ValueCtxMockType(), mock.AnythingOfType("*paymentModel.PaymentDTO")).Once().Return(c.ErrSomeErrorForUnitTest)
				paymentRepo.On("RollbackTransaction", c.ValueCtxMockType()).Return(nil)
			},
		},
		{
			name: "ERROR: Update recurring payment contract status",
			request: &unifiedPaymentModel.CreateUnifiedPaymentSessionRequest{
				PaymentMethod: &unifiedPaymentModel.PaymentMethod{
					Type: c.UnifiedPaymentMethodCard,
				},
				ExpiryAt:                   time.Now().Add(time.Hour),
				RecurringID:                "718f931d-e6b0-49c2-ac0a-553aecada667",
				InitiateFirstAuthorization: true,
			},
			setupMock: func() {
				paymentRepo.On("CreatePayment", mock.Anything, mock.Anything).Return(nil)
				paymentSvc.On(
					"RecordPaymentStatusHistory", mock.Anything, mock.Anything, mock.Anything, mock.Anything,
				).Return()
				recurringContractRepo.On(
					"UpdateRecurringContractStatus", mock.Anything, mock.Anything, constant.RecurringContractStatusPendInitialAuth, constant.UserSystemType,
				).Once().Return(assert.AnError)
			},
			wantErr: true,
		},
		{
			name:    "ERROR: Commit transaction",
			wantErr: true,
			setupMock: func() {
				paymentRepo.On("CommitTransaction", c.ValueCtxMockType()).Once().Return(c.ErrSomeErrorForUnitTest)
			},
		},
		{
			name:    "ERROR: Publish expire message",
			wantErr: false,
			setupMock: func() {
				paymentRepo.On("CommitTransaction", c.ValueCtxMockType()).Return(nil)
				rmq.On(
					"PublishWithDelay", c.ValueCtxMockType(), c.StringMockType(), mock.Anything, mock.Anything,
				).Once().Return(c.ErrSomeErrorForUnitTest)
			},
		},
		{
			name: "ERROR: From Merchant dashboard",
			request: &unifiedPaymentModel.CreateUnifiedPaymentSessionRequest{
				CreatedFrom: constant.SourceMerchantPortal,
			},
			wantErr: true,
			setupMock: func() {
				shortLinkSvc.On("Create", c.ValueCtxMockType(), mock.Anything).Return(nil, errors.New("error")).Once()
			},
		},
		{
			name: "SUCCESS: From Merchant dashboard",
			request: &unifiedPaymentModel.CreateUnifiedPaymentSessionRequest{
				CreatedFrom: constant.SourceMerchantPortal,
			},
			wantErr: false,
			setupMock: func() {
				shortLinkSvc.On("Create", c.ValueCtxMockType(), mock.Anything).Return(&shortLinkModel.ShortLink{
					ShortLinkURL: "short-link-url",
				}, nil).Once()
				rmq.On(
					"PublishWithDelay", c.ValueCtxMockType(), c.StringMockType(), mock.Anything, mock.Anything,
				).Return(nil).Once()
			},
		},
		{
			name: "SUCCESS: Virtual Terminal",
			request: &unifiedPaymentModel.CreateUnifiedPaymentSessionRequest{
				PaymentMethod: &unifiedPaymentModel.PaymentMethod{
					Type: c.UnifiedPaymentMethodCard,
				},
				PaymentMethodOptions: unifiedPaymentModel.PaymentMethodOptions{
					Card: &unifiedPaymentModel.PaymentMethodOptionCard{
						ProcessingConfig: &unifiedPaymentModel.PaymentMethodOptionCardProcessingConfig{
							BankMerchantId: "TEST00001",
						},
					},
				},
				Mode:     c.UnifiedPaymentModeRedirect,
				ExpiryAt: time.Now().Add(time.Hour).UTC(),
				VirtualTerminal: &unifiedPaymentModel.VirtualTerminal{
					BatchID: "de134b32-7f02-4bac-84b9-376689432117",
				},
			},
			setupMock: func() { /* No Body Function */ },
			wantErr:   false,
		},
		{
			name:    "SUCCESS: Without autoConfirm",
			wantErr: false,
			setupMock: func() {
				rmq.On(
					"PublishWithDelay", c.ValueCtxMockType(), c.StringMockType(), mock.Anything, mock.Anything,
				).Return(nil)
			},
		},
		{
			name:    "SUCCESS: QRIS payment method with Qris config sets acquirer",
			wantErr: false,
			request: &unifiedPaymentModel.CreateUnifiedPaymentSessionRequest{
				PaymentMethod: &unifiedPaymentModel.PaymentMethod{
					Type: c.UnifiedPaymentMethodQris,
				},
				PaymentMethodOptions: unifiedPaymentModel.PaymentMethodOptions{},
				ExpiryAt:             time.Now().Add(15 * time.Minute),
			},
			setupMock: func() {
				merchantRepo.On("FindMerchantByID",
					c.ValueCtxMockType(),
					c.StringMockType(),
				).Return(&merchantModel.Merchant{}, nil)

				paymentMethodWithQris := &paymentModel.PaymentMethodWithPivot{
					PaymentMethod: paymentModel.PaymentMethod{
						UUID:     uuid.NewString(),
						Acquirer: "QRIS_ACQUIRER_TEST",
					},
					MerchantConfigObj: &paymentModel.PaymentMethodMerchantConfigObject{
						PartnerConfig: &paymentMethodModel.SetupPaymentMethodPartnerConfigRequest{
							Qris: &paymentMethodModel.SetupPaymentMethodPartnerConfigForQrisRequest{
								AcquirerMerchantID: "test-merchant-id",
							},
						},
					},
				}

				paymentMethodSvc.On("GetActivePaymentMethodDetailForPaymentRequest",
					c.ValueCtxMockType(),
					mock.Anything,
				).Return(paymentMethodWithQris, nil)

				qrisSvc.On("FindQrRegistrationByExternalIDAndAcquirer",
					c.ValueCtxMockType(),
					c.StringMockType(),
					c.StringMockType(),
				).Return(&qrisModel.Registration{}, nil)

				paymentRepo.On("GetPaymentByMerchantAndReferenceId",
					c.ValueCtxMockType(),
					c.StringMockType(),
					c.StringMockType(),
				).Return(nil, nil)

				jwt.On("GeneratePaymentToken",
					c.StringMockType(),
					c.TimeMockType(),
				).Return("token", nil)

				redisExt.On(
					"Set", c.ValueCtxMockType(), c.StringMockType(), c.BoolMockType(), c.DurationMockType(),
				).Return(&redis.StatusCmd{})

				paymentRepo.On("BeginTransaction", c.ValueCtxMockType()).Return(context.WithValue(context.Background(), c.CtxTest, ""), nil)

				paymentRepo.On("CreatePayment", c.ValueCtxMockType(), mock.AnythingOfType("*paymentModel.PaymentDTO")).Return(nil)

				paymentSvc.On("RecordPaymentStatusHistory",
					c.ValueCtxMockType(),
					c.StringMockType(),
					c.StringMockType(),
					c.StringMockType(),
				).Return()

				paymentRepo.On("CommitTransaction", c.ValueCtxMockType()).Return(nil)

				rmq.On(
					"PublishWithDelay", c.ValueCtxMockType(), c.StringMockType(), mock.Anything, mock.Anything,
				).Return(nil)
			},
		},
		{
			name:    "SUCCESS: QRIS payment method with nil Qris config (should not panic)",
			wantErr: false,
			request: &unifiedPaymentModel.CreateUnifiedPaymentSessionRequest{
				PaymentMethod: &unifiedPaymentModel.PaymentMethod{
					Type: c.UnifiedPaymentMethodQris,
				},
				PaymentMethodOptions: unifiedPaymentModel.PaymentMethodOptions{},
				ExpiryAt:             time.Now().Add(15 * time.Minute),
			},
			setupMock: func() {
				merchantRepo.On("FindMerchantByID",
					c.ValueCtxMockType(),
					c.StringMockType(),
				).Return(&merchantModel.Merchant{}, nil)

				paymentMethodWithNilQris := &paymentModel.PaymentMethodWithPivot{
					PaymentMethod: paymentModel.PaymentMethod{
						UUID:     uuid.NewString(),
						Acquirer: "QRIS_ACQUIRER_TEST",
					},
					MerchantConfigObj: &paymentModel.PaymentMethodMerchantConfigObject{
						PartnerConfig: &paymentMethodModel.SetupPaymentMethodPartnerConfigRequest{
							Qris: nil, // This tests the nil case
						},
					},
				}

				paymentMethodSvc.On("GetActivePaymentMethodDetailForPaymentRequest",
					c.ValueCtxMockType(),
					mock.Anything,
				).Return(paymentMethodWithNilQris, nil)

				qrisSvc.On("FindQrRegistrationByExternalIDAndAcquirer",
					c.ValueCtxMockType(),
					c.StringMockType(),
					c.StringMockType(),
				).Return(&qrisModel.Registration{}, nil)

				paymentRepo.On("GetPaymentByMerchantAndReferenceId",
					c.ValueCtxMockType(),
					c.StringMockType(),
					c.StringMockType(),
				).Return(nil, nil)

				jwt.On("GeneratePaymentToken",
					c.StringMockType(),
					c.TimeMockType(),
				).Return("token", nil)

				redisExt.On(
					"Set", c.ValueCtxMockType(), c.StringMockType(), c.BoolMockType(), c.DurationMockType(),
				).Return(&redis.StatusCmd{})

				paymentRepo.On("BeginTransaction", c.ValueCtxMockType()).Return(context.WithValue(context.Background(), c.CtxTest, ""), nil)

				paymentRepo.On("CreatePayment", c.ValueCtxMockType(), mock.AnythingOfType("*paymentModel.PaymentDTO")).Return(nil)

				paymentSvc.On("RecordPaymentStatusHistory",
					c.ValueCtxMockType(),
					c.StringMockType(),
					c.StringMockType(),
					c.StringMockType(),
				).Return()

				paymentRepo.On("CommitTransaction", c.ValueCtxMockType()).Return(nil)

				rmq.On(
					"PublishWithDelay", c.ValueCtxMockType(), c.StringMockType(), mock.Anything, mock.Anything,
				).Return(nil)
			},
		},
		{
			name:    "SUCCESS: Redirect mode should set PaymentUrl for DANA e-wallet",
			wantErr: false,
			request: &unifiedPaymentModel.CreateUnifiedPaymentSessionRequest{
				PaymentMethod: &unifiedPaymentModel.PaymentMethod{
					Type: c.UnifiedPaymentMethodEWallet,
				},
				PaymentMethodOptions: unifiedPaymentModel.PaymentMethodOptions{
					Ewallet: &unifiedPaymentModel.PaymentMethodOptionEwallet{
						Channel: c.UnifiedPaymentEWalletDanaAcquirer,
					},
				},
				Mode:        c.UnifiedPaymentModeRedirect,
				AutoConfirm: false,
				PaymentURL:  "https://payment.redirect.url",
			},
			setupMock: func() {
				merchantRepo.On("FindMerchantByID",
					c.ValueCtxMockType(),
					c.StringMockType(),
				).Return(&merchantModel.Merchant{}, nil)

				paymentMethodWithEwallet := &paymentModel.PaymentMethodWithPivot{
					PaymentMethod: paymentModel.PaymentMethod{
						UUID:     uuid.NewString(),
						Acquirer: c.UnifiedPaymentEWalletDanaAcquirer,
					},
					MerchantConfigObj: &paymentModel.PaymentMethodMerchantConfigObject{
						PartnerConfig: &paymentMethodModel.SetupPaymentMethodPartnerConfigRequest{},
					},
				}

				paymentMethodSvc.On("GetActivePaymentMethodDetailForPaymentRequest",
					c.ValueCtxMockType(),
					mock.Anything,
				).Return(paymentMethodWithEwallet, nil)

				paymentRepo.On("GetPaymentByMerchantAndReferenceId",
					c.ValueCtxMockType(),
					c.StringMockType(),
					c.StringMockType(),
				).Return(nil, nil)

				jwt.On("GeneratePaymentToken",
					c.StringMockType(),
					c.TimeMockType(),
				).Return("token", nil)

				redisExt.On(
					"Set", c.ValueCtxMockType(), c.StringMockType(), c.BoolMockType(), c.DurationMockType(),
				).Return(&redis.StatusCmd{})

				paymentRepo.On("BeginTransaction", c.ValueCtxMockType()).Return(context.WithValue(context.Background(), c.CtxTest, ""), nil)

				paymentRepo.On("CreatePayment", c.ValueCtxMockType(), mock.AnythingOfType("*paymentModel.PaymentDTO")).Return(nil)

				paymentSvc.On("RecordPaymentStatusHistory",
					c.ValueCtxMockType(),
					c.StringMockType(),
					c.StringMockType(),
					c.StringMockType(),
				).Return()

				paymentRepo.On("CommitTransaction", c.ValueCtxMockType()).Return(nil)

				rmq.On(
					"PublishWithDelay", c.ValueCtxMockType(), c.StringMockType(), mock.Anything, mock.Anything,
				).Return(nil)
			},
		},
		{
			name:    "SUCCESS: Redirect mode should set PaymentUrl for VA",
			wantErr: false,
			request: &unifiedPaymentModel.CreateUnifiedPaymentSessionRequest{
				PaymentMethod: &unifiedPaymentModel.PaymentMethod{
					Type: c.UnifiedPaymentMethodVA,
				},
				PaymentMethodOptions: unifiedPaymentModel.PaymentMethodOptions{
					VirtualAccount: &unifiedPaymentModel.PaymentMethodOptionVirtualAccount{
						Channel: "PERMATA",
					},
				},
				Mode:        c.UnifiedPaymentModeRedirect,
				AutoConfirm: false,
				PaymentURL:  "https://payment.redirect.url",
			},
			setupMock: func() {
				merchantRepo.On("FindMerchantByID",
					c.ValueCtxMockType(),
					c.StringMockType(),
				).Return(&merchantModel.Merchant{}, nil)

				paymentMethodWithVA := &paymentModel.PaymentMethodWithPivot{
					PaymentMethod: paymentModel.PaymentMethod{
						UUID:     uuid.NewString(),
						Acquirer: "PERMATA",
					},
					MerchantConfigObj: &paymentModel.PaymentMethodMerchantConfigObject{
						PartnerConfig: &paymentMethodModel.SetupPaymentMethodPartnerConfigRequest{},
					},
				}

				paymentMethodSvc.On("GetActivePaymentMethodDetailForPaymentRequest",
					c.ValueCtxMockType(),
					mock.Anything,
				).Return(paymentMethodWithVA, nil)

				paymentRepo.On("GetPaymentByMerchantAndReferenceId",
					c.ValueCtxMockType(),
					c.StringMockType(),
					c.StringMockType(),
				).Return(nil, nil)

				jwt.On("GeneratePaymentToken",
					c.StringMockType(),
					c.TimeMockType(),
				).Return("token", nil)

				redisExt.On(
					"Set", c.ValueCtxMockType(), c.StringMockType(), c.BoolMockType(), c.DurationMockType(),
				).Return(&redis.StatusCmd{})

				paymentRepo.On("BeginTransaction", c.ValueCtxMockType()).Return(context.WithValue(context.Background(), c.CtxTest, ""), nil)

				paymentRepo.On("CreatePayment", c.ValueCtxMockType(), mock.AnythingOfType("*paymentModel.PaymentDTO")).Return(nil)

				paymentSvc.On("RecordPaymentStatusHistory",
					c.ValueCtxMockType(),
					c.StringMockType(),
					c.StringMockType(),
					c.StringMockType(),
				).Return()

				paymentRepo.On("CommitTransaction", c.ValueCtxMockType()).Return(nil)

				rmq.On(
					"PublishWithDelay", c.ValueCtxMockType(), c.StringMockType(), mock.Anything, mock.Anything,
				).Return(nil)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMock()

			svc := New(
				cfg, log, paymentRepo, nil, accountTrxRepo,
				WithJWTExt(jwt),
				WithRabbitMQClient(rmq),
				WithQRISService(qrisSvc),
				WithRedisClient(redisExt),
				WithPaymentService(paymentSvc),
				WithMerchantRepo(merchantRepo),
				WithPaymentMethodService(paymentMethodSvc),
				WithRecurringContractRepository(recurringContractRepo),
			)
			WithShortLinkService(svc, shortLinkSvc)
			WithInternalUnifiedPaymentService(svc, internalUnifiedPaymentSvc)

			if tc.request == nil {
				tc.request = &unifiedPaymentModel.CreateUnifiedPaymentSessionRequest{
					PaymentMethod: &unifiedPaymentModel.PaymentMethod{
						Type: c.UnifiedPaymentMethodVA,
					},
					PaymentMethodOptions: unifiedPaymentModel.PaymentMethodOptions{
						VirtualAccount: &unifiedPaymentModel.PaymentMethodOptionVirtualAccount{
							Channel: "PERMATA",
						},
					},
					ExpiryAt: time.Now().Add(15 * time.Minute),
				}
			}

			result, err := svc.CreateSession(context.Background(), tc.request)
			if tc.wantErr {
				assert.Error(t, err)
				if strings.Contains(tc.name, "Payment Method Not Found") {
					assert.Equal(t, err, pkgErr.New(response.HttpErrUnprocessableContent, c.ErrPaymentMethodNotFound))
				}
			} else {
				assert.NoError(t, err)
				if result == nil {
					return
				}
				// For QRIS tests, verify the result is properly created
				if strings.Contains(tc.name, "QRIS") && result != nil {
					assert.NotEmpty(t, result.ID)
					assert.Equal(t, tc.request.PaymentMethod.Type, result.PaymentMethod.Type)
				}
				// For Redirect mode tests, verify PaymentUrl is set correctly
				if strings.Contains(tc.name, "Redirect mode should set PaymentUrl") && result != nil {
					assert.Equal(t, tc.request.PaymentURL, result.PaymentUrl, "PaymentUrl should be set for redirect mode")
				}

				// For AutoSplit tests, verify autoSplitDetails is always present
				if strings.Contains(tc.name, "AutoSplit") && result != nil {
					assert.NotNil(t, result.AutoSplitDetails, "autoSplitDetails must always be present when autoSplit is true")
					if tc.request.AutoConfirm {
						assert.Equal(t, constant.AutoSplitStatusProcessing, result.AutoSplitDetails.Status)
						assert.Equal(t, 1, result.AutoSplitDetails.NumberOfCharges)
						assert.Equal(t, 1, result.AutoSplitDetails.NumberOfInProcessCharges)
						assert.NotNil(t, result.AutoSplitDetails.TotalInProcessChargeAmount)
						assert.Equal(t, len(result.ChargeDetails), len(result.AutoSplitDetails.ChargesDetails))
						if len(result.AutoSplitDetails.ChargesDetails) > 0 {
							assert.NotNil(t, result.AutoSplitDetails.ChargesDetails[0].SafeToRetry)
							assert.False(t, *result.AutoSplitDetails.ChargesDetails[0].SafeToRetry)
						}
					} else {
						assert.Equal(t, constant.AutoSplitStatusProcessing, result.AutoSplitDetails.Status)
						assert.Equal(t, 0, result.AutoSplitDetails.NumberOfCharges)
						assert.Equal(t, 0, result.AutoSplitDetails.NumberOfInProcessCharges)
						assert.Equal(t, 0, result.AutoSplitDetails.NumberOfSuccessfulCharges)
						assert.Equal(t, 0, result.AutoSplitDetails.NumberOfFailedCharges)
						assert.Nil(t, result.AutoSplitDetails.TotalInProcessChargeAmount)
						assert.Nil(t, result.AutoSplitDetails.TotalSuccessfulChargeAmount)
						assert.Nil(t, result.AutoSplitDetails.TotalFailedChargeAmount)
						assert.Empty(t, result.AutoSplitDetails.ChargesDetails)
					}
				}

				// For Recurring
				if result.RecurringID != "" {
					assert.Equal(t, tc.request.RecurringID, result.RecurringID)
					assert.Equal(t, constant.UnifiedPaymentSessionStatusProcessing, result.Status)
					assert.Equal(t, "", result.PaymentUrl)
				}
			}
			recurringContractRepo.AssertExpectations(t)
			internalUnifiedPaymentSvc.AssertExpectations(t)
		})
	}
}
