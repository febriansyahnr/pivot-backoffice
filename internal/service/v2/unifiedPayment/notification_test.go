package unifiedPaymentService

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/jmoiron/sqlx/types"
	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	c "github.com/paper-indonesia/pivot-backoffice/constant"
	paymentConstant "github.com/paper-indonesia/pivot-backoffice/constant/payment"
	customerModel "github.com/paper-indonesia/pivot-backoffice/internal/model/customer"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	orchestraModel "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	paymentMethodModel "github.com/paper-indonesia/pivot-backoffice/internal/model/paymentMethod"
	unifiedPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/unifiedPayment"
	rabbitMQMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/rabbitmqExt"
	redisExtMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/redisExt"
	repositoryMock "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	serviceMock "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/rabbitMqExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"

	mockLogger "github.com/paper-indonesia/pdk/v2/logger"
	redisSdk "github.com/redis/go-redis/v9"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestProcessNotification(t *testing.T) {
	cfg := &config.Config{
		Environment: "test",
		SlackConfig: config.SlackConfig{
			PaymentNotifWebhookURL: "https://hooks.slack.com/services/test",
		},
	}
	log, _ := mockLogger.NewZapLogger(mockLogger.Config{})
	paymentRepo := repositoryMock.NewIPaymentRepository(t)
	paymentMethodRepo := repositoryMock.NewIPaymentMethodRepository(t)
	accountTrxRepo := repositoryMock.NewIAccountTransactionRepository(t)
	merchantRepo := repositoryMock.NewIMerchantRepository(t)
	paymentSvc := serviceMock.NewIPaymentService(t)
	fdsSvc := serviceMock.NewIFdsService(t)
	rabbitMq := rabbitMQMock.NewRabbitMQExt(t)
	redis := redisExtMocks.NewIRedisExt(t)
	redisMutex := redisExtMocks.NewIMutexer(t)
	recurringContractSvc := serviceMock.NewIRecurringContractService(t)
	paymentMethodSvc := serviceMock.NewIPaymentMethodService(t)
	cardFundedPayoutSvc := serviceMock.NewICardFundedPayoutService(t)

	// avoid test acquirer mid
	paymentMethodSvc.On("FindPaymentMethodByIdAndMerchant", mock.Anything, mock.Anything, mock.Anything).
		Return(&paymentModel.PaymentMethodWithPivot{
			MerchantConfigObj: &paymentModel.PaymentMethodMerchantConfigObject{
				PartnerConfig: &paymentMethodModel.SetupPaymentMethodPartnerConfigRequest{
					Card: &paymentMethodModel.SetupPaymentMethodPartnerConfigForCardRequest{
						Items: []paymentMethodModel.SetupPaymentMethodPartnerConfigForCardObj{
							{
								AcquirerMerchantID: "MID123456",
								Acquirer:           "BCA",
							},
						},
					},
				},
			},
		}, nil).Maybe()

	// Shared mocks for the distributed lock acquired by ProcessNotification.
	// Non-static payments use SetNX; static payments still use redsync Mutex (via payStaticPaymentCharge).
	// Uses Maybe() so cases that return before reaching the lock (e.g. GetDetailByID error) are not required to match.
	redisSetNXResult := &redisSdk.BoolCmd{}
	redisSetNXResult.SetVal(true)
	redis.On("SetNX", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe().Return(redisSetNXResult)
	redis.On("NewMutex", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe().Return(redisMutex)
	redisMutex.On("LockContext", mock.Anything).Maybe().Return(nil)
	redisMutex.On("UnlockContext", mock.Anything).Maybe().Return(true, nil)

	paidRequest := &unifiedPaymentModel.PaymentNotificationRequest{
		PaymentSessionID: uuid.NewString(),
		Amount: unifiedPaymentModel.Amount{
			Currency: "IDR",
			Value:    10000.00,
		},
		ChargeStatus: c.ChargeStatusSuccess,
		ChargePaymentMethodDetails: &unifiedPaymentModel.ChargePaymentMethodDetails{
			Card: &unifiedPaymentModel.ChargePaymentMethodDetailCard{
				BinInformations: unifiedPaymentModel.ChargePaymentMethodDetailBinInformation{
					Country: "IDN",
				},
				AuthorizationResult: &unifiedPaymentModel.ChargePaymentMethodDetailCardAuthorizationResult{
					AuthorizationID: "valid-id",
				},
			},
		},
	}

	failRequest := &unifiedPaymentModel.PaymentNotificationRequest{
		PaymentSessionID: uuid.NewString(),
		Amount: unifiedPaymentModel.Amount{
			Currency: "IDR",
			Value:    10000.00,
		},
		ChargeStatus: c.ChargeStatusFailed,
		ChargePaymentMethodDetails: &unifiedPaymentModel.ChargePaymentMethodDetails{
			Card: &unifiedPaymentModel.ChargePaymentMethodDetailCard{
				BinInformations: unifiedPaymentModel.ChargePaymentMethodDetailBinInformation{
					Country: "IDN",
				},
				AuthorizationResult: &unifiedPaymentModel.ChargePaymentMethodDetailCardAuthorizationResult{
					AuthorizationID: "valid-id",
				},
			},
		},
	}

	staticPaymentRequest := &unifiedPaymentModel.PaymentNotificationRequest{
		PaymentSessionID: uuid.NewString(),
		Amount: unifiedPaymentModel.Amount{
			Currency: "IDR",
			Value:    15000.00,
		},
		ChargeStatus:           c.ChargeStatusSuccess,
		PaymentMethodType:      c.UnifiedPaymentMethodVA,
		ProcessorTransactionID: "proc-txn-123",
		Processor:              "TEST_PROCESSOR",
		ProcessorID:            "proc-id-123",
	}

	staticPaymentRequestWithBankRef := &unifiedPaymentModel.PaymentNotificationRequest{
		PaymentSessionID: uuid.NewString(),
		Amount: unifiedPaymentModel.Amount{
			Currency: "IDR",
			Value:    20000.00,
		},
		ChargeStatus:           c.ChargeStatusSuccess,
		PaymentMethodType:      c.UnifiedPaymentMethodVA,
		ProcessorTransactionID: "proc-txn-456",
		Processor:              "TEST_PROCESSOR",
		ProcessorID:            "proc-id-456",
		ChargePaymentMethodDetails: &unifiedPaymentModel.ChargePaymentMethodDetails{
			VirtualAccount: &unifiedPaymentModel.ChargePaymentMethodDetailVirtualAccount{
				Channel:               "BCA",
				VirtualAccountNumber:  "1234567890",
				VirtualAccountName:    "Test VA",
				VirtualAccountTrxType: "CLOSED",
				BankReferenceNo:       "BANKREF987654321",
			},
		},
	}

	testCases := []struct {
		name      string
		wantErr   bool
		setupMock func()
		request   *unifiedPaymentModel.PaymentNotificationRequest
	}{
		{
			name:    "ERROR: Got error database on GetPaymentById",
			wantErr: true,
			request: &unifiedPaymentModel.PaymentNotificationRequest{},
			setupMock: func() {
				paymentSvc.On("GetDetailByID",
					c.ValueCtxMockType(),
					c.StringMockType(),
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
		},
		{
			name:    "ERROR: SetNX lock acquire error",
			wantErr: true,
			request: &unifiedPaymentModel.PaymentNotificationRequest{},
			setupMock: func() {
				paymentSvc.On("GetDetailByID",
					c.ValueCtxMockType(),
					c.StringMockType(),
				).Once().Return(&paymentModel.Payment{
					UUID:   uuid.NewString(),
					Status: c.UnifiedPaymentSessionStatusRequireAction,
				}, nil)

				redisSetNXResult.SetErr(c.ErrSomeErrorForUnitTest)
			},
		},
		{
			name:    "ERROR: SetNX lock already held",
			wantErr: true,
			request: &unifiedPaymentModel.PaymentNotificationRequest{},
			setupMock: func() {
				paymentSvc.On("GetDetailByID",
					c.ValueCtxMockType(),
					c.StringMockType(),
				).Once().Return(&paymentModel.Payment{
					UUID:   uuid.NewString(),
					Status: c.UnifiedPaymentSessionStatusRequireAction,
				}, nil)

				redisSetNXResult.SetVal(false)
			},
		},
		{
			name:    "ERROR: Got error database on FindMerchantByID",
			wantErr: true,
			request: &unifiedPaymentModel.PaymentNotificationRequest{},
			setupMock: func() {
				paymentSvc.On("GetDetailByID",
					c.ValueCtxMockType(),
					c.StringMockType(),
				).Once().Return(&paymentModel.Payment{
					UUID:   uuid.NewString(),
					Status: c.UnifiedPaymentSessionStatusRequireAction,
				}, nil)

				merchantRepo.On("FindMerchantByID",
					c.ValueCtxMockType(),
					c.StringMockType(),
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
		},
		{
			name:    "ERROR: Payment already in final status from paid to success to avoid false alarm alert",
			wantErr: true,
			request: &unifiedPaymentModel.PaymentNotificationRequest{
				ChargeID:          "charge-123",
				ChargeStatus:      c.ChargeStatusSuccess,
				Processor:         "TEST_PROCESSOR",
				PaymentMethodType: c.UnifiedPaymentMethodVA,
			},
			setupMock: func() {
				paymentSvc.On("GetDetailByID",
					c.ValueCtxMockType(),
					c.StringMockType(),
				).Once().Return(&paymentModel.Payment{
					UUID:   uuid.NewString(),
					Status: c.UnifiedPaymentSessionStatusPaid,
				}, nil)

				merchantRepo.On("FindMerchantByID",
					c.ValueCtxMockType(),
					c.StringMockType(),
				).Return(&merchantModel.Merchant{BusinessCountry: sql.NullString{Valid: true, String: "IDN"}}, nil)
			},
		},
		{
			name:    "ERROR: Payment cancelled to failed - no slack alert",
			wantErr: true,
			request: &unifiedPaymentModel.PaymentNotificationRequest{
				ChargeID:          "charge-123",
				ChargeStatus:      c.ChargeStatusFailed,
				Processor:         "TEST_PROCESSOR",
				PaymentMethodType: c.UnifiedPaymentMethodVA,
			},
			setupMock: func() {
				paymentSvc.On("GetDetailByID",
					c.ValueCtxMockType(),
					c.StringMockType(),
				).Once().Return(&paymentModel.Payment{
					UUID:   uuid.NewString(),
					Status: c.UnifiedPaymentSessionStatusCancelled,
				}, nil)

				// Should NOT expect Slack alert to be sent
			},
		},
		{
			name:    "[Paid Flow] ERROR: Request amount is not match",
			wantErr: true,
			request: paidRequest,
			setupMock: func() {
				paymentSvc.On("GetDetailByID",
					c.ValueCtxMockType(),
					c.StringMockType(),
				).Once().Return(&paymentModel.Payment{}, nil)
			},
		},
		{
			name:    "[Paid Flow] ERROR: Begin transaction",
			wantErr: true,
			request: paidRequest,
			setupMock: func() {
				paymentSvc.On("GetDetailByID",
					c.ValueCtxMockType(),
					c.StringMockType(),
				).Once().Return(&paymentModel.Payment{
					Amount: decimal.NewFromFloat(10000.00),
					PaymentMethod: paymentModel.PaymentMethod{
						Type: c.ChannelCreditCard,
					},
				}, nil)

				paymentRepo.On("BeginTransaction", c.ValueCtxMockType()).
					Once().Return(context.WithValue(context.Background(), mySqlExt.CtxSqlTx, &sqlx.Tx{}), c.ErrSomeErrorForUnitTest)
			},
		},
		{
			name:    "[Paid Flow] ERROR: Find merchant but error rollback",
			wantErr: true,
			request: paidRequest,
			setupMock: func() {
				paymentSvc.On("GetDetailByID",
					c.ValueCtxMockType(),
					c.StringMockType(),
				).Once().Return(&paymentModel.Payment{
					Amount: decimal.NewFromFloat(10000.00),
					PaymentMethod: paymentModel.PaymentMethod{
						Type: c.ChannelCreditCard,
					},
				}, nil)

				paymentRepo.On("BeginTransaction", c.ValueCtxMockType()).
					Once().Return(context.WithValue(context.Background(), mySqlExt.CtxSqlTx, &sqlx.Tx{}), nil)

				paymentSvc.On("DeterminePaymentFee", mock.Anything, mock.Anything).Once().Return(c.ErrSomeErrorForUnitTest)

				paymentRepo.On("RollbackTransaction", c.ValueCtxMockType()).
					Once().Return(c.ErrSomeErrorForUnitTest)
			},
		},
		{
			name:    "[Paid Flow] ERROR: DeterminePaymentFee service",
			wantErr: true,
			request: paidRequest,
			setupMock: func() {
				paymentSvc.On("GetDetailByID",
					c.ValueCtxMockType(),
					c.StringMockType(),
				).Once().Return(&paymentModel.Payment{
					Amount: decimal.NewFromFloat(10000.00),
					PaymentMethod: paymentModel.PaymentMethod{
						Type: c.ChannelCreditCard,
					},
				}, nil)

				paymentRepo.On("BeginTransaction", c.ValueCtxMockType()).
					Once().Return(context.WithValue(context.Background(), mySqlExt.CtxSqlTx, &sqlx.Tx{}), nil)
				paymentSvc.On("DeterminePaymentFee", mock.Anything, mock.Anything).Once().Return(c.ErrSomeErrorForUnitTest)
				paymentRepo.On("RollbackTransaction", c.ValueCtxMockType()).Once().Return(nil)
			},
		},
		{
			name:    "[Paid Flow] ERROR: UpdatePaymentStatus service",
			wantErr: true,
			request: paidRequest,
			setupMock: func() {
				paymentSvc.On("GetDetailByID",
					c.ValueCtxMockType(),
					c.StringMockType(),
				).Once().Return(&paymentModel.Payment{
					Amount: decimal.NewFromFloat(10000.00),
					PaymentMethod: paymentModel.PaymentMethod{
						Type: c.ChannelCreditCard,
					},
				}, nil)

				paymentRepo.On("BeginTransaction", c.ValueCtxMockType()).
					Once().Return(context.WithValue(context.Background(), mySqlExt.CtxSqlTx, &sqlx.Tx{}), nil)
				paymentSvc.On("DeterminePaymentFee", mock.Anything, mock.Anything).Once().Return(nil)
				paymentRepo.On("UpdatePaymentStatus", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType(), c.StringMockType(), c.TimeMockType()).
					Once().Return(c.ErrSomeErrorForUnitTest)
				paymentRepo.On("RollbackTransaction", c.ValueCtxMockType()).Once().Return(nil)
			},
		},
		{
			name:    "[Paid Flow] ERROR: Find ledger service",
			wantErr: true,
			request: paidRequest,
			setupMock: func() {
				paymentSvc.On("GetDetailByID",
					c.ValueCtxMockType(),
					c.StringMockType(),
				).Once().Return(&paymentModel.Payment{
					Amount: decimal.NewFromFloat(10000.00),
					PaymentMethod: paymentModel.PaymentMethod{
						Type: c.ChannelCreditCard,
					},
				}, nil)

				paymentRepo.On("BeginTransaction", c.ValueCtxMockType()).
					Once().Return(context.WithValue(context.Background(), mySqlExt.CtxSqlTx, &sqlx.Tx{}), nil)
				paymentSvc.On("DeterminePaymentFee", mock.Anything, mock.Anything).Once().Return(nil)
				paymentRepo.On("UpdatePaymentStatus", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType(), c.StringMockType(), c.TimeMockType()).
					Once().Return(nil)
				accountTrxRepo.On("FindByID", c.ValueCtxMockType(), c.StringMockType()).
					Once().Return(nil, c.ErrSomeErrorForUnitTest)
				paymentRepo.On("RollbackTransaction", c.ValueCtxMockType()).Once().Return(nil)
			},
		},
		{
			name:    "[Paid Flow] ERROR: Update pending ledger service",
			wantErr: true,
			request: paidRequest,
			setupMock: func() {
				paymentSvc.On("GetDetailByID",
					c.ValueCtxMockType(),
					c.StringMockType(),
				).Once().Return(&paymentModel.Payment{
					Amount: decimal.NewFromFloat(10000.00),
					PaymentMethod: paymentModel.PaymentMethod{
						Type: c.ChannelCreditCard,
					},
				}, nil)

				paymentRepo.On("BeginTransaction", c.ValueCtxMockType()).
					Once().Return(context.WithValue(context.Background(), mySqlExt.CtxSqlTx, &sqlx.Tx{}), nil)
				paymentSvc.On("DeterminePaymentFee", mock.Anything, mock.Anything).Once().Return(nil)
				paymentRepo.On("UpdatePaymentStatus", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType(), c.StringMockType(), c.TimeMockType()).
					Once().Return(nil)
				accountTrxRepo.On("FindByID", c.ValueCtxMockType(), c.StringMockType()).
					Once().Return(&orchestraModel.AccountTransactionWithUseCase{
					UUID:            uuid.New(),
					SettlementModel: sql.NullString{Valid: true, String: c.PaymentMethodChannelTypeAggregator},
				}, nil)
				paymentSvc.On("UpdatePendingLedger", c.ValueCtxMockType(), mock.Anything, mock.Anything).
					Once().Return(c.ErrSomeErrorForUnitTest)
				paymentRepo.On("RollbackTransaction", c.ValueCtxMockType()).Once().Return(nil)
			},
		},
		{
			name:    "[Paid Flow] ERROR: Update recurring payment",
			wantErr: true,
			request: paidRequest,
			setupMock: func() {
				paymentSvc.On(
					"GetDetailByID", c.ValueCtxMockType(), c.StringMockType(),
				).Once().Return(&paymentModel.Payment{
					Amount: decimal.NewFromFloat(10000.00),
					PaymentMethod: paymentModel.PaymentMethod{
						Type: c.ChannelCreditCard,
					},
					RecurringContractID: util.ValueToPtr("7929abae-48b3-4f92-b3f4-c010730bb15e"),
					Metadata: &map[string]any{
						"recurringPayment": map[string]any{
							"initiateFirstAuthorization": true,
							"firstAuthorizationMethod":   c.RecurringContractAuthMethodOneDollar,
						},
					},
				}, nil)

				paymentRepo.On("BeginTransaction", c.ValueCtxMockType()).
					Once().Return(context.WithValue(context.Background(), mySqlExt.CtxSqlTx, &sqlx.Tx{}), nil)
				paymentSvc.On("DeterminePaymentFee", mock.Anything, mock.Anything).Once().Return(nil)
				paymentRepo.On("UpdatePaymentStatus", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType(), c.StringMockType(), c.TimeMockType()).
					Once().Return(nil)
				accountTrxRepo.On("FindByID", c.ValueCtxMockType(), c.StringMockType()).
					Once().Return(&orchestraModel.AccountTransactionWithUseCase{
					UUID:            uuid.New(),
					SettlementModel: sql.NullString{Valid: true, String: c.PaymentMethodChannelTypeAggregator},
				}, nil)
				paymentSvc.On("UpdatePendingLedger", c.ValueCtxMockType(), mock.Anything, mock.Anything).
					Once().Return(nil)
				recurringContractSvc.On("UpdateRecurringPayment", mock.Anything, mock.Anything).Once().Return(assert.AnError)
				paymentRepo.On("RollbackTransaction", c.ValueCtxMockType()).Once().Return(nil)
			},
		},
		{
			name:    "[Paid Flow] ERROR: Commit transaction",
			wantErr: true,
			request: paidRequest,
			setupMock: func() {
				paymentSvc.On("GetDetailByID",
					c.ValueCtxMockType(),
					c.StringMockType(),
				).Once().Return(&paymentModel.Payment{
					Amount: decimal.NewFromFloat(10000.00),
					PaymentMethod: paymentModel.PaymentMethod{
						Type: c.ChannelCreditCard,
					},
				}, nil)

				paymentRepo.On("BeginTransaction", c.ValueCtxMockType()).
					Once().Return(context.WithValue(context.Background(), mySqlExt.CtxSqlTx, &sqlx.Tx{}), nil)
				merchantRepo.On("FindMerchantByID", c.ValueCtxMockType(), c.StringMockType()).
					Return(&merchantModel.Merchant{BusinessCountry: sql.NullString{Valid: true, String: "IDN"}}, nil)
				paymentSvc.On("DeterminePaymentFee", mock.Anything, mock.Anything).Once().Return(nil)
				paymentRepo.On("UpdatePaymentStatus", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType(), c.StringMockType(), c.TimeMockType()).
					Once().Return(nil)
				accountTrxRepo.On("FindByID", c.ValueCtxMockType(), c.StringMockType()).
					Once().Return(&orchestraModel.AccountTransactionWithUseCase{
					UUID:            uuid.New(),
					SettlementModel: sql.NullString{Valid: true, String: c.PaymentMethodChannelTypeAggregator},
				}, nil)
				paymentSvc.On("UpdatePendingLedger", c.ValueCtxMockType(), mock.Anything, mock.Anything).
					Once().Return(nil)
				paymentRepo.On("CommitTransaction", c.ValueCtxMockType()).Once().Return(c.ErrSomeErrorForUnitTest)
				paymentRepo.On("RollbackTransaction", c.ValueCtxMockType()).Once().Return(nil)
			},
		},
		{
			name:    "[Paid & Send Callback Flow] Paid success but got error callback on find ledger by reference",
			wantErr: false,
			request: paidRequest,
			setupMock: func() {
				paymentSvc.On("GetDetailByID",
					c.ValueCtxMockType(),
					c.StringMockType(),
				).Once().Return(&paymentModel.Payment{
					Amount: decimal.NewFromFloat(10000.00),
					PaymentMethod: paymentModel.PaymentMethod{
						Type: c.ChannelCreditCard,
					},
				}, nil)

				paymentRepo.On("BeginTransaction", c.ValueCtxMockType()).
					Once().Return(context.WithValue(context.Background(), mySqlExt.CtxSqlTx, &sqlx.Tx{}), nil)
				merchantRepo.On("FindMerchantByID", c.ValueCtxMockType(), c.StringMockType()).
					Return(&merchantModel.Merchant{BusinessCountry: sql.NullString{Valid: true, String: "IDN"}}, nil)
				paymentSvc.On("DeterminePaymentFee", mock.Anything, mock.Anything).Once().Return(nil)
				paymentRepo.On("UpdatePaymentStatus", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType(), c.StringMockType(), c.TimeMockType()).
					Once().Return(nil)
				accountTrxRepo.On("FindByID", c.ValueCtxMockType(), c.StringMockType()).
					Once().Return(&orchestraModel.AccountTransactionWithUseCase{
					UUID:            uuid.New(),
					SettlementModel: sql.NullString{Valid: true, String: c.PaymentMethodChannelTypeAggregator},
				}, nil)
				paymentSvc.On("UpdatePendingLedger", c.ValueCtxMockType(), mock.Anything, mock.Anything).
					Once().Return(nil)
				paymentRepo.On("CommitTransaction", c.ValueCtxMockType()).Once().Return(nil)

				// Flow Callback
				accountTrxRepo.On("FindByReference", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType()).
					Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
		},
		{
			name:    "[Paid & Send Callback Flow] Paid success but got error callback on failed send stomp notification",
			wantErr: false,
			request: paidRequest,
			setupMock: func() {
				paymentSvc.On("GetDetailByID",
					c.ValueCtxMockType(),
					c.StringMockType(),
				).Once().Return(&paymentModel.Payment{
					Amount: decimal.NewFromFloat(10000.00),
					PaymentMethod: paymentModel.PaymentMethod{
						Type: c.ChannelCreditCard,
					},
					Metadata: &map[string]any{
						"mode":        "REDIRECT",
						"autoConfirm": true,
						"redirectUrl": map[string]any{
							"failedUrl":  "https://payment-stg.harsya.com/final-status?token=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1dWlkIjoiYzIyN2UyODEtM2M5Yy00MGNmLTk5OTktNThiZGI3ODJiNWNjIiwiaXNzIjoiYmFja2VuZC1wb3J0YWwiLCJleHAiOjE3NTA0MzE4NDV9.cWRUXD2dNWTAEOs7xwy3XnOcxJQFFfAOnvrtOXG6hhI",
							"successUrl": "https://payment-stg.harsya.com/final-status?token=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1dWlkIjoiYzIyN2UyODEtM2M5Yy00MGNmLTk5OTktNThiZGI3ODJiNWNjIiwiaXNzIjoiYmFja2VuZC1wb3J0YWwiLCJleHAiOjE3NTA0MzE4NDV9.cWRUXD2dNWTAEOs7xwy3XnOcxJQFFfAOnvrtOXG6hhI",
						},
						"paymentMethod": map[string]any{
							"type": "CARD",
						},
						"clientMetadata": map[string]any{
							"okelur": "okelur",
						},
						"clientRedirectUrl": map[string]any{
							"failureReturnUrl":    "https://merchant.com/failure",
							"successReturnUrl":    "https://merchant.com/success",
							"expirationReturnUrl": "https://merchant.com/expiration",
						},
						"isUnifiedPaymentV2":  true,
						"statementDescriptor": "SPP OK",
						"paymentMethodOptions": map[string]any{
							"card": map[string]any{},
						},
					},
				}, nil)

				paymentRepo.On("BeginTransaction", c.ValueCtxMockType()).
					Once().Return(context.WithValue(context.Background(), mySqlExt.CtxSqlTx, &sqlx.Tx{}), nil)
				merchantRepo.On("FindMerchantByID", c.ValueCtxMockType(), c.StringMockType()).
					Return(&merchantModel.Merchant{BusinessCountry: sql.NullString{Valid: true, String: "IDN"}}, nil)
				paymentSvc.On("DeterminePaymentFee", mock.Anything, mock.Anything).Once().Return(nil)
				paymentRepo.On("UpdatePaymentStatus", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType(), c.StringMockType(), c.TimeMockType()).
					Once().Return(nil)
				accountTrxRepo.On("FindByID", c.ValueCtxMockType(), c.StringMockType()).
					Once().Return(&orchestraModel.AccountTransactionWithUseCase{
					UUID:            uuid.New(),
					SettlementModel: sql.NullString{Valid: true, String: c.PaymentMethodChannelTypeAggregator},
				}, nil)
				paymentSvc.On("UpdatePendingLedger", c.ValueCtxMockType(), mock.Anything, mock.Anything).
					Once().Return(nil)
				paymentRepo.On("CommitTransaction", c.ValueCtxMockType()).Once().Return(nil)

				// Flow Callback
				accountTrxRepo.On("FindByReference", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType()).
					Once().Return(&orchestraModel.AccountTransactionWithUseCase{UUID: uuid.New()}, nil)
				rabbitMq.On("PublishMerchantCallback", mock.Anything, mock.Anything).Return(nil)
				rabbitMq.On("PushNotification", c.ValueCtxMockType(), c.PtrPushNotificationMockType()).
					Once().Return(c.ErrSomeErrorForUnitTest)
			},
		},
		{
			name:    "[Paid & Send Callback Flow] Paid success and success to send callback",
			wantErr: false,
			request: paidRequest,
			setupMock: func() {
				paymentSvc.On("GetDetailByID",
					c.ValueCtxMockType(),
					c.StringMockType(),
				).Once().Return(&paymentModel.Payment{
					Amount: decimal.NewFromFloat(10000.00),
					PaymentMethod: paymentModel.PaymentMethod{
						Type: c.ChannelCreditCard,
					},
					Metadata: &map[string]any{
						"mode":        "REDIRECT",
						"autoConfirm": true,
						"redirectUrl": map[string]any{
							"failedUrl":  "https://payment-stg.harsya.com/final-status?token=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1dWlkIjoiYzIyN2UyODEtM2M5Yy00MGNmLTk5OTktNThiZGI3ODJiNWNjIiwiaXNzIjoiYmFja2VuZC1wb3J0YWwiLCJleHAiOjE3NTA0MzE4NDV9.cWRUXD2dNWTAEOs7xwy3XnOcxJQFFfAOnvrtOXG6hhI",
							"successUrl": "https://payment-stg.harsya.com/final-status?token=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1dWlkIjoiYzIyN2UyODEtM2M5Yy00MGNmLTk5OTktNThiZGI3ODJiNWNjIiwiaXNzIjoiYmFja2VuZC1wb3J0YWwiLCJleHAiOjE3NTA0MzE4NDV9.cWRUXD2dNWTAEOs7xwy3XnOcxJQFFfAOnvrtOXG6hhI",
						},
						"paymentMethod": map[string]any{
							"type": "CARD",
						},
						"clientMetadata": map[string]any{
							"okelur": "okelur",
						},
						"clientRedirectUrl": map[string]any{
							"failureReturnUrl":    "https://merchant.com/failure",
							"successReturnUrl":    "https://merchant.com/success",
							"expirationReturnUrl": "https://merchant.com/expiration",
						},
						"isUnifiedPaymentV2":  true,
						"statementDescriptor": "SPP OK",
						"paymentMethodOptions": map[string]any{
							"card": map[string]any{},
						},
					},
				}, nil)

				paymentRepo.On("BeginTransaction", c.ValueCtxMockType()).
					Once().Return(context.WithValue(context.Background(), mySqlExt.CtxSqlTx, &sqlx.Tx{}), nil)
				merchantRepo.On("FindMerchantByID", c.ValueCtxMockType(), c.StringMockType()).
					Return(&merchantModel.Merchant{BusinessCountry: sql.NullString{Valid: true, String: "IDN"}}, nil)
				paymentSvc.On("DeterminePaymentFee", mock.Anything, mock.Anything).Once().Return(nil)
				paymentRepo.On("UpdatePaymentStatus", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType(), c.StringMockType(), c.TimeMockType()).
					Once().Return(nil)
				accountTrxRepo.On("FindByID", c.ValueCtxMockType(), c.StringMockType()).
					Once().Return(&orchestraModel.AccountTransactionWithUseCase{
					UUID:            uuid.New(),
					SettlementModel: sql.NullString{Valid: true, String: c.PaymentMethodChannelTypeAggregator},
				}, nil)
				paymentSvc.On("UpdatePendingLedger", c.ValueCtxMockType(), mock.Anything, mock.Anything).
					Once().Return(nil)
				paymentRepo.On("CommitTransaction", c.ValueCtxMockType()).Once().Return(nil)

				// Flow Callback
				accountTrxRepo.On("FindByReference", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType()).
					Once().Return(&orchestraModel.AccountTransactionWithUseCase{UUID: uuid.New()}, nil)
				rabbitMq.On("PublishMerchantCallback", mock.Anything, mock.Anything).Return(nil)
				rabbitMq.On("PushNotification", c.ValueCtxMockType(), c.PtrPushNotificationMockType()).
					Once().Return(nil)
			},
		},
		{
			name:    "[Handle Failed Flow] ERROR: Begin transaction",
			wantErr: true,
			request: failRequest,
			setupMock: func() {
				paymentSvc.On("GetDetailByID",
					c.ValueCtxMockType(),
					c.StringMockType(),
				).Once().Return(&paymentModel.Payment{
					Amount: decimal.NewFromFloat(10000.00),
					PaymentMethod: paymentModel.PaymentMethod{
						Type: c.ChannelCreditCard,
					},
				}, nil)

				paymentRepo.On("BeginTransaction", c.ValueCtxMockType()).
					Once().Return(context.WithValue(context.Background(), mySqlExt.CtxSqlTx, &sqlx.Tx{}), c.ErrSomeErrorForUnitTest)

				// FDS auto-update should still be called in defer even if main transaction fails
				fdsSvc.On("UpdateTransaction", mock.MatchedBy(func(ctx context.Context) bool {
					return true // Accept any context (including background context)
				}), c.StringMockType(), mock.Anything).
					Once().Return(nil, nil)
			},
		},
		{
			name:    "[Handle Failed Flow] ERROR: UpdatePaymentStatus",
			wantErr: true,
			request: failRequest,
			setupMock: func() {
				paymentSvc.On("GetDetailByID",
					c.ValueCtxMockType(),
					c.StringMockType(),
				).Once().Return(&paymentModel.Payment{
					Amount: decimal.NewFromFloat(10000.00),
					PaymentMethod: paymentModel.PaymentMethod{
						Type: c.ChannelCreditCard,
					},
				}, nil)

				paymentRepo.On("BeginTransaction", c.ValueCtxMockType()).
					Once().Return(context.WithValue(context.Background(), mySqlExt.CtxSqlTx, &sqlx.Tx{}), nil)
				paymentRepo.On("UpdatePaymentStatus", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType(), c.StringMockType(), c.TimeMockType()).
					Once().Return(c.ErrSomeErrorForUnitTest)
				paymentRepo.On("RollbackTransaction", c.ValueCtxMockType()).Once().Return(nil)

				// FDS auto-update should be called in defer
				fdsSvc.On("UpdateTransaction", mock.MatchedBy(func(ctx context.Context) bool {
					return true // Accept any context (including background context)
				}), c.StringMockType(), mock.Anything).
					Once().Return(nil, nil)
			},
		},
		{
			name:    "[Handle Failed Flow] ERROR: FindByID account transaction",
			wantErr: true,
			request: failRequest,
			setupMock: func() {
				paymentSvc.On("GetDetailByID",
					c.ValueCtxMockType(),
					c.StringMockType(),
				).Once().Return(&paymentModel.Payment{
					Amount: decimal.NewFromFloat(10000.00),
					PaymentMethod: paymentModel.PaymentMethod{
						Type: c.ChannelCreditCard,
					},
				}, nil)

				paymentRepo.On("BeginTransaction", c.ValueCtxMockType()).
					Once().Return(context.WithValue(context.Background(), mySqlExt.CtxSqlTx, &sqlx.Tx{}), nil)
				paymentRepo.On("UpdatePaymentStatus", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType(), c.StringMockType(), c.TimeMockType()).
					Once().Return(nil)
				accountTrxRepo.On("FindByID", c.ValueCtxMockType(), c.StringMockType()).
					Once().Return(nil, c.ErrSomeErrorForUnitTest)
				paymentRepo.On("RollbackTransaction", c.ValueCtxMockType()).Once().Return(nil)

				// FDS auto-update should be called in defer
				fdsSvc.On("UpdateTransaction", mock.MatchedBy(func(ctx context.Context) bool {
					return true // Accept any context (including background context)
				}), c.StringMockType(), mock.Anything).
					Once().Return(nil, nil)
			},
		},
		{
			name:    "[Handle Failed Flow] ERROR: UpdatePendingLedger",
			wantErr: true,
			request: failRequest,
			setupMock: func() {
				paymentSvc.On("GetDetailByID",
					c.ValueCtxMockType(),
					c.StringMockType(),
				).Once().Return(&paymentModel.Payment{
					Amount: decimal.NewFromFloat(10000.00),
					PaymentMethod: paymentModel.PaymentMethod{
						Type: c.ChannelCreditCard,
					},
				}, nil)

				paymentRepo.On("BeginTransaction", c.ValueCtxMockType()).
					Once().Return(context.WithValue(context.Background(), mySqlExt.CtxSqlTx, &sqlx.Tx{}), nil)
				paymentRepo.On("UpdatePaymentStatus", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType(), c.StringMockType(), c.TimeMockType()).
					Once().Return(nil)
				accountTrxRepo.On("FindByID", c.ValueCtxMockType(), c.StringMockType()).
					Once().Return(&orchestraModel.AccountTransactionWithUseCase{UUID: uuid.New()}, nil)
				paymentSvc.On("UpdatePendingLedger", c.ValueCtxMockType(), mock.Anything, mock.Anything).
					Once().Return(c.ErrSomeErrorForUnitTest)
				paymentRepo.On("RollbackTransaction", c.ValueCtxMockType()).Once().Return(nil)

				// FDS auto-update should be called in defer
				fdsSvc.On("UpdateTransaction", mock.MatchedBy(func(ctx context.Context) bool {
					return true // Accept any context (including background context)
				}), c.StringMockType(), mock.Anything).
					Once().Return(nil, nil)
			},
		},
		{
			name:    "[Handle Failed Flow] ERROR: CommitTransaction",
			wantErr: true,
			request: failRequest,
			setupMock: func() {
				paymentSvc.On("GetDetailByID",
					c.ValueCtxMockType(),
					c.StringMockType(),
				).Once().Return(&paymentModel.Payment{
					Amount: decimal.NewFromFloat(10000.00),
					PaymentMethod: paymentModel.PaymentMethod{
						Type: c.ChannelCreditCard,
					},
				}, nil)

				paymentRepo.On("BeginTransaction", c.ValueCtxMockType()).
					Once().Return(context.WithValue(context.Background(), mySqlExt.CtxSqlTx, &sqlx.Tx{}), nil)
				paymentRepo.On("UpdatePaymentStatus", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType(), c.StringMockType(), c.TimeMockType()).
					Once().Return(nil)
				accountTrxRepo.On("FindByID", c.ValueCtxMockType(), c.StringMockType()).
					Once().Return(&orchestraModel.AccountTransactionWithUseCase{UUID: uuid.New()}, nil)
				paymentSvc.On("UpdatePendingLedger", c.ValueCtxMockType(), mock.Anything, mock.Anything).
					Once().Return(nil)
				paymentRepo.On("CommitTransaction", c.ValueCtxMockType()).Once().Return(c.ErrSomeErrorForUnitTest)
				paymentRepo.On("RollbackTransaction", c.ValueCtxMockType()).Once().Return(nil)

				// FDS auto-update should be called in defer
				fdsSvc.On("UpdateTransaction", mock.MatchedBy(func(ctx context.Context) bool {
					return true // Accept any context (including background context)
				}), c.StringMockType(), mock.Anything).
					Once().Return(nil, nil)
			},
		},
		{
			name:    "[Handle Failed Flow] SUCCESS: Failed charge processed successfully with callback",
			wantErr: false,
			request: failRequest,
			setupMock: func() {
				paymentSvc.On("GetDetailByID",
					c.ValueCtxMockType(),
					c.StringMockType(),
				).Once().Return(&paymentModel.Payment{
					Amount: decimal.NewFromFloat(10000.00),
					PaymentMethod: paymentModel.PaymentMethod{
						Type: c.ChannelCreditCard,
					},
					Metadata: &map[string]any{
						"mode": "REDIRECT",
						"paymentMethod": map[string]any{
							"type": "CARD",
						},
						"clientRedirectUrl": map[string]any{
							"successReturnUrl":    "https://merchant.com/success",
							"failureReturnUrl":    "https://merchant.com/failure",
							"expirationReturnUrl": "https://merchant.com/expiration",
						},
					},
				}, nil)

				paymentRepo.On("BeginTransaction", c.ValueCtxMockType()).
					Once().Return(context.WithValue(context.Background(), mySqlExt.CtxSqlTx, &sqlx.Tx{}), nil)
				paymentRepo.On("UpdatePaymentStatus", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType(), c.StringMockType(), c.TimeMockType()).
					Once().Return(nil)
				accountTrxRepo.On("FindByID", c.ValueCtxMockType(), c.StringMockType()).
					Once().Return(&orchestraModel.AccountTransactionWithUseCase{UUID: uuid.New()}, nil)
				paymentSvc.On("UpdatePendingLedger", c.ValueCtxMockType(), mock.Anything, mock.Anything).
					Once().Return(nil)
				paymentRepo.On("CommitTransaction", c.ValueCtxMockType()).Once().Return(nil)

				// FDS auto-update for failed CC transaction
				fdsSvc.On("UpdateTransaction", mock.MatchedBy(func(ctx context.Context) bool {
					return true // Accept any context (including background context)
				}), c.StringMockType(), mock.Anything).
					Once().Return(nil, nil)

				// Flow Callback
				accountTrxRepo.On("FindByReference", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType()).
					Once().Return(&orchestraModel.AccountTransactionWithUseCase{UUID: uuid.New()}, nil)
				rabbitMq.On("PublishMerchantCallback", mock.Anything, mock.Anything).Return(nil)
				rabbitMq.On("PushNotification", c.ValueCtxMockType(), c.PtrPushNotificationMockType()).
					Once().Return(nil)
			},
		},
		{
			name:    "[Handle Failed Flow] SUCCESS: Failed CC charge with successful FDS update",
			wantErr: false,
			request: &unifiedPaymentModel.PaymentNotificationRequest{
				PaymentSessionID: uuid.NewString(),
				ChargeID:         "failed-cc-charge-123",
				Amount: unifiedPaymentModel.Amount{
					Currency: "IDR",
					Value:    10000.00,
				},
				ChargeStatus: c.ChargeStatusFailed,
				ChargePaymentMethodDetails: &unifiedPaymentModel.ChargePaymentMethodDetails{
					Card: &unifiedPaymentModel.ChargePaymentMethodDetailCard{
						BinInformations: unifiedPaymentModel.ChargePaymentMethodDetailBinInformation{
							Country: "IDN",
						},
					},
				},
			},
			setupMock: func() {
				paymentSvc.On("GetDetailByID",
					c.ValueCtxMockType(),
					c.StringMockType(),
				).Once().Return(&paymentModel.Payment{
					Amount: decimal.NewFromFloat(10000.00),
					PaymentMethod: paymentModel.PaymentMethod{
						Type: c.ChannelCreditCard,
					},
				}, nil)

				paymentRepo.On("BeginTransaction", c.ValueCtxMockType()).
					Once().Return(context.WithValue(context.Background(), mySqlExt.CtxSqlTx, &sqlx.Tx{}), nil)
				paymentRepo.On("UpdatePaymentStatus", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType(), c.StringMockType(), c.TimeMockType()).
					Once().Return(nil)
				accountTrxRepo.On("FindByID", c.ValueCtxMockType(), c.StringMockType()).
					Once().Return(&orchestraModel.AccountTransactionWithUseCase{UUID: uuid.New()}, nil)
				paymentSvc.On("UpdatePendingLedger", c.ValueCtxMockType(), mock.Anything, mock.Anything).
					Once().Return(nil)
				paymentRepo.On("CommitTransaction", c.ValueCtxMockType()).Once().Return(nil)

				// FDS auto-update should be called with charge ID and nil request
				fdsSvc.On("UpdateTransaction", mock.MatchedBy(func(ctx context.Context) bool {
					return true // Accept any context (including background context)
				}), "failed-cc-charge-123", mock.Anything).
					Once().Return(nil, nil)

				// Flow Callback
				accountTrxRepo.On("FindByReference", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType()).
					Once().Return(nil, nil)
			},
		},
		{
			name:    "[Handle Failed Flow] SUCCESS: Failed CC charge with FDS update error (should not break main flow)",
			wantErr: false,
			request: &unifiedPaymentModel.PaymentNotificationRequest{
				PaymentSessionID: uuid.NewString(),
				ChargeID:         "failed-cc-charge-456",
				Amount: unifiedPaymentModel.Amount{
					Currency: "IDR",
					Value:    10000.00,
				},
				ChargeStatus: c.ChargeStatusFailed,
				ChargePaymentMethodDetails: &unifiedPaymentModel.ChargePaymentMethodDetails{
					Card: &unifiedPaymentModel.ChargePaymentMethodDetailCard{
						BinInformations: unifiedPaymentModel.ChargePaymentMethodDetailBinInformation{
							Country: "IDN",
						},
					},
				},
			},
			setupMock: func() {
				paymentSvc.On("GetDetailByID",
					c.ValueCtxMockType(), c.StringMockType(),
				).Once().Return(&paymentModel.Payment{
					Amount: decimal.NewFromFloat(10000.00),
					PaymentMethod: paymentModel.PaymentMethod{
						Type: c.ChannelCreditCard,
					},
					RecurringContractID: util.ValueToPtr("7929abae-48b3-4f92-b3f4-c010730bb15e"),
					Metadata: &map[string]any{
						"recurringPayment": map[string]any{
							"initiateFirstAuthorization": true,
						},
					},
				}, nil)

				paymentRepo.On("BeginTransaction", c.ValueCtxMockType()).
					Once().Return(context.WithValue(context.Background(), mySqlExt.CtxSqlTx, &sqlx.Tx{}), nil)
				paymentRepo.On("UpdatePaymentStatus", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType(), c.StringMockType(), c.TimeMockType()).
					Once().Return(nil)
				accountTrxRepo.On("FindByID", c.ValueCtxMockType(), c.StringMockType()).
					Once().Return(&orchestraModel.AccountTransactionWithUseCase{UUID: uuid.New()}, nil)
				paymentSvc.On("UpdatePendingLedger", c.ValueCtxMockType(), mock.Anything, mock.Anything).
					Once().Return(nil)
				paymentRepo.On("CommitTransaction", c.ValueCtxMockType()).Once().Return(nil)

				// FDS auto-update fails but should not break the main flow
				fdsSvc.On("UpdateTransaction", mock.MatchedBy(func(ctx context.Context) bool {
					return true // Accept any context (including background context)
				}), "failed-cc-charge-456", mock.Anything).
					Once().Return(nil, c.ErrSomeErrorForUnitTest)

				// Flow Callback
				accountTrxRepo.On("FindByReference", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType()).
					Once().Return(nil, nil)
				// Release the exclusive lock for recurring payment.
				redisResult := &redisSdk.IntCmd{}
				redisResult.SetErr(assert.AnError)
				redis.On(
					"Del", mock.Anything, fmt.Sprintf(constant.RecurringPaymentMutualExclusionKey, constant.RecurringPaymentTypeFirstAuthorization, "7929abae-48b3-4f92-b3f4-c010730bb15e"),
				).Once().Return(redisResult)
			},
		},
		{
			name:    "[Handle Failed Flow] SUCCESS: Failed non-CC charge should skip FDS update",
			wantErr: false,
			request: &unifiedPaymentModel.PaymentNotificationRequest{
				PaymentSessionID: uuid.NewString(),
				ChargeID:         "failed-va-charge-789",
				Amount: unifiedPaymentModel.Amount{
					Currency: "IDR",
					Value:    10000.00,
				},
				ChargeStatus: c.ChargeStatusFailed,
				ChargePaymentMethodDetails: &unifiedPaymentModel.ChargePaymentMethodDetails{
					VirtualAccount: &unifiedPaymentModel.ChargePaymentMethodDetailVirtualAccount{
						Channel: "BCA",
					},
				},
			},
			setupMock: func() {
				paymentSvc.On("GetDetailByID",
					c.ValueCtxMockType(),
					c.StringMockType(),
				).Once().Return(&paymentModel.Payment{
					Amount: decimal.NewFromFloat(10000.00),
					PaymentMethod: paymentModel.PaymentMethod{
						Type: c.ChannelVirtualAccount,
					},
				}, nil)

				paymentRepo.On("BeginTransaction", c.ValueCtxMockType()).
					Once().Return(context.WithValue(context.Background(), mySqlExt.CtxSqlTx, &sqlx.Tx{}), nil)
				paymentRepo.On("UpdatePaymentStatus", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType(), c.StringMockType(), c.TimeMockType()).
					Once().Return(nil)
				accountTrxRepo.On("FindByID", c.ValueCtxMockType(), c.StringMockType()).
					Once().Return(&orchestraModel.AccountTransactionWithUseCase{UUID: uuid.New()}, nil)
				paymentSvc.On("UpdatePendingLedger", c.ValueCtxMockType(), mock.Anything, mock.Anything).
					Once().Return(nil)
				paymentRepo.On("CommitTransaction", c.ValueCtxMockType()).Once().Return(nil)

				// FDS auto-update should NOT be called for non-CC payment methods
				// No fdsSvc mock expectation means it should not be called

				// Flow Callback
				accountTrxRepo.On("FindByReference", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType()).
					Once().Return(nil, nil)
			},
		},
		{
			name:    "[Invalid Status Flow] ERROR: Status not allowed",
			wantErr: true,
			request: &unifiedPaymentModel.PaymentNotificationRequest{
				PaymentSessionID: uuid.NewString(),
				ChargeStatus:     "UNKNOWN_STATUS",
			},
			setupMock: func() {
				paymentSvc.On("GetDetailByID",
					c.ValueCtxMockType(),
					c.StringMockType(),
				).Once().Return(&paymentModel.Payment{}, nil)
			},
		},
		{
			name:    "[Static Payment Flow] ERROR: Amount mismatch for static closed payment",
			wantErr: true,
			request: &unifiedPaymentModel.PaymentNotificationRequest{
				PaymentSessionID: uuid.NewString(),
				Amount: unifiedPaymentModel.Amount{
					Currency: "IDR",
					Value:    20000.00,
				},
				ChargeStatus: c.ChargeStatusSuccess,
			},
			setupMock: func() {
				paymentSession := &paymentModel.Payment{
					UUID:       uuid.NewString(),
					MerchantID: uuid.NewString(),
					Amount:     decimal.NewFromFloat(10000.00),
					PaymentMethod: paymentModel.PaymentMethod{
						Type: c.ChannelVirtualAccount,
					},
					Type: c.UnifiedPaymentTypeMultiple,
				}
				paymentSvc.On("GetDetailByID",
					c.ValueCtxMockType(),
					c.StringMockType(),
				).Once().Return(paymentSession, nil)
			},
		},
		{
			name:    "[Static Payment Flow] ERROR: Begin transaction",
			wantErr: true,
			request: staticPaymentRequest,
			setupMock: func() {
				paymentSession := &paymentModel.Payment{
					UUID:       uuid.NewString(),
					MerchantID: uuid.NewString(),
					Amount:     decimal.NewFromFloat(0),
					PaymentMethod: paymentModel.PaymentMethod{
						Type: c.UnifiedPaymentMethodVA,
					},
					Type: c.UnifiedPaymentTypeMultiple,
				}
				paymentSvc.On("GetDetailByID",
					c.ValueCtxMockType(),
					c.StringMockType(),
				).Once().Return(paymentSession, nil)
				paymentRepo.On("BeginTransaction", c.ValueCtxMockType()).
					Once().Return(context.WithValue(context.Background(), mySqlExt.CtxSqlTx, &sqlx.Tx{}), c.ErrSomeErrorForUnitTest)
			},
		},
		{
			name:    "[Static Payment Flow] ERROR: DeterminePaymentFee",
			wantErr: true,
			request: staticPaymentRequest,
			setupMock: func() {
				paymentSession := &paymentModel.Payment{
					UUID:       uuid.NewString(),
					MerchantID: uuid.NewString(),
					Amount:     decimal.NewFromFloat(0),
					PaymentMethod: paymentModel.PaymentMethod{
						Type: c.UnifiedPaymentMethodVA,
					},
					Type: c.UnifiedPaymentTypeMultiple,
				}
				paymentSvc.On("GetDetailByID",
					c.ValueCtxMockType(),
					c.StringMockType(),
				).Once().Return(paymentSession, nil)
				paymentRepo.On("BeginTransaction", c.ValueCtxMockType()).
					Once().Return(context.WithValue(context.Background(), mySqlExt.CtxSqlTx, &sqlx.Tx{}), nil)
				paymentSvc.On("DeterminePaymentFee", mock.Anything, mock.Anything).Once().Return(c.ErrSomeErrorForUnitTest)
				paymentRepo.On("RollbackTransaction", c.ValueCtxMockType()).Once().Return(nil)
			},
		},
		{
			name:    "[Static Payment Flow] ERROR: PostCreateLedger",
			wantErr: true,
			request: staticPaymentRequest,
			setupMock: func() {
				paymentSession := &paymentModel.Payment{
					UUID:       uuid.NewString(),
					MerchantID: uuid.NewString(),
					Amount:     decimal.NewFromFloat(0),
					PaymentMethod: paymentModel.PaymentMethod{
						Type: c.UnifiedPaymentMethodVA,
					},
					Type: c.UnifiedPaymentTypeMultiple,
				}
				paymentSvc.On("GetDetailByID",
					c.ValueCtxMockType(),
					c.StringMockType(),
				).Once().Return(paymentSession, nil)
				paymentRepo.On("BeginTransaction", c.ValueCtxMockType()).
					Once().Return(context.WithValue(context.Background(), mySqlExt.CtxSqlTx, &sqlx.Tx{}), nil)
				paymentSvc.On("DeterminePaymentFee", mock.Anything, mock.Anything).Once().Return(nil)
				paymentSvc.On("PostCreateLedger", c.ValueCtxMockType(), mock.Anything, mock.Anything).
					Once().Return(c.ErrSomeErrorForUnitTest)
				paymentRepo.On("RollbackTransaction", c.ValueCtxMockType()).Once().Return(nil)
			},
		},
		{
			name:    "[Static Payment Flow] ERROR: GetAggregateTransactionByReference for summary",
			wantErr: true,
			request: staticPaymentRequest,
			setupMock: func() {
				paymentSession := &paymentModel.Payment{
					UUID:       uuid.NewString(),
					MerchantID: uuid.NewString(),
					Amount:     decimal.NewFromFloat(0),
					PaymentMethod: paymentModel.PaymentMethod{
						Type: c.UnifiedPaymentMethodVA,
					},
					Type:     c.UnifiedPaymentTypeMultiple,
					Metadata: &map[string]any{},
				}
				paymentSvc.On("GetDetailByID",
					c.ValueCtxMockType(),
					c.StringMockType(),
				).Once().Return(paymentSession, nil)
				paymentRepo.On("BeginTransaction", c.ValueCtxMockType()).
					Once().Return(context.WithValue(context.Background(), mySqlExt.CtxSqlTx, &sqlx.Tx{}), nil)
				paymentSvc.On("DeterminePaymentFee", mock.Anything, mock.Anything).Once().Return(nil)
				paymentSvc.On("PostCreateLedger", c.ValueCtxMockType(), mock.Anything, mock.Anything).
					Once().Return(nil)
				accountTrxRepo.On("GetAggregateTransactionByReference", c.ValueCtxMockType(), mock.Anything).
					Once().Return(nil, c.ErrSomeErrorForUnitTest)
				paymentRepo.On("RollbackTransaction", c.ValueCtxMockType()).Once().Return(nil)
			},
		},
		{
			name:    "[Static Payment Flow] ERROR: UpdatePaymentData",
			wantErr: true,
			request: staticPaymentRequest,
			setupMock: func() {
				paymentSession := &paymentModel.Payment{
					UUID:       uuid.NewString(),
					MerchantID: uuid.NewString(),
					Amount:     decimal.NewFromFloat(0),
					PaymentMethod: paymentModel.PaymentMethod{
						Type: c.UnifiedPaymentMethodVA,
					},
					Type: c.UnifiedPaymentTypeMultiple,
					Metadata: &map[string]any{
						"summaryTransaction": &unifiedPaymentModel.SummaryTransaction{
							SumPaidAmount:   1000.0,
							CountPaidAmount: 1,
						},
					},
				}
				paymentSvc.On("GetDetailByID",
					c.ValueCtxMockType(),
					c.StringMockType(),
				).Once().Return(paymentSession, nil)
				paymentRepo.On("BeginTransaction", c.ValueCtxMockType()).
					Once().Return(context.WithValue(context.Background(), mySqlExt.CtxSqlTx, &sqlx.Tx{}), nil)
				paymentSvc.On("DeterminePaymentFee", mock.Anything, mock.Anything).Once().Return(nil)
				paymentSvc.On("PostCreateLedger", c.ValueCtxMockType(), mock.Anything, mock.Anything).
					Once().Return(nil)
				paymentRepo.On("UpdatePaymentMetadataById", c.ValueCtxMockType(), mock.Anything, mock.Anything).
					Once().Return(c.ErrSomeErrorForUnitTest)
				paymentRepo.On("RollbackTransaction", c.ValueCtxMockType()).Once().Return(nil)
			},
		},
		{
			name:    "[Static Payment Flow] ERROR: CommitTransaction",
			wantErr: true,
			request: staticPaymentRequest,
			setupMock: func() {
				paymentSession := &paymentModel.Payment{
					UUID:       uuid.NewString(),
					MerchantID: uuid.NewString(),
					Amount:     decimal.NewFromFloat(0),
					PaymentMethod: paymentModel.PaymentMethod{
						Type: c.UnifiedPaymentMethodVA,
					},
					Type: c.UnifiedPaymentTypeMultiple,
					Metadata: &map[string]any{
						"summaryTransaction": &unifiedPaymentModel.SummaryTransaction{
							SumPaidAmount:   1000.0,
							CountPaidAmount: 1,
						},
					},
				}
				paymentSvc.On("GetDetailByID",
					c.ValueCtxMockType(),
					c.StringMockType(),
				).Once().Return(paymentSession, nil)
				paymentRepo.On("BeginTransaction", c.ValueCtxMockType()).
					Once().Return(context.WithValue(context.Background(), mySqlExt.CtxSqlTx, &sqlx.Tx{}), nil)
				paymentSvc.On("DeterminePaymentFee", mock.Anything, mock.Anything).Once().Return(nil)
				paymentSvc.On("PostCreateLedger", c.ValueCtxMockType(), mock.Anything, mock.Anything).
					Once().Return(nil)
				paymentRepo.On("UpdatePaymentMetadataById", c.ValueCtxMockType(), mock.Anything, mock.Anything).
					Once().Return(nil)
				paymentRepo.On("CommitTransaction", c.ValueCtxMockType()).Once().Return(c.ErrSomeErrorForUnitTest)
				paymentRepo.On("RollbackTransaction", c.ValueCtxMockType()).Once().Return(nil)
			},
		},
		{
			name:    "[Paid Flow] ERROR: PostCreateLedger failed",
			wantErr: true,
			request: paidRequest,
			setupMock: func() {
				paymentSvc.On("GetDetailByID",
					c.ValueCtxMockType(),
					c.StringMockType(),
				).Once().Return(&paymentModel.Payment{
					Amount: decimal.NewFromFloat(10000.00),
					PaymentMethod: paymentModel.PaymentMethod{
						Type: c.ChannelCreditCard,
					},
				}, nil)

				paymentRepo.On("BeginTransaction", c.ValueCtxMockType()).
					Once().Return(context.WithValue(context.Background(), mySqlExt.CtxSqlTx, &sqlx.Tx{}), nil)
				merchantRepo.On("FindMerchantByID", c.ValueCtxMockType(), c.StringMockType()).
					Return(&merchantModel.Merchant{BusinessCountry: sql.NullString{Valid: true, String: "IDN"}}, nil)
				paymentSvc.On("DeterminePaymentFee", mock.Anything, mock.Anything).Once().Return(nil)
				paymentRepo.On("UpdatePaymentStatus", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType(), c.StringMockType(), c.TimeMockType()).
					Once().Return(nil)
				accountTrxRepo.On("FindByID", c.ValueCtxMockType(), c.StringMockType()).
					Once().Return(nil, nil)
				paymentSvc.On("PostCreateLedger", c.ValueCtxMockType(), mock.Anything, mock.Anything).
					Once().Return(c.ErrSomeErrorForUnitTest)
				paymentRepo.On("RollbackTransaction", c.ValueCtxMockType()).Once().Return(nil)
			},
		},
		{
			name:    "[Static Payment Flow] SUCCESS: VA payment with BankReferenceNo",
			wantErr: false,
			request: staticPaymentRequestWithBankRef,
			setupMock: func() {
				paymentSession := &paymentModel.Payment{
					UUID:       staticPaymentRequestWithBankRef.PaymentSessionID,
					MerchantID: uuid.NewString(),
					Amount:     decimal.NewFromFloat(0),
					PaymentMethod: paymentModel.PaymentMethod{
						Type: c.UnifiedPaymentMethodVA,
					},
					Type: c.UnifiedPaymentTypeMultiple,
					Metadata: &map[string]any{
						"summaryTransaction": &unifiedPaymentModel.SummaryTransaction{
							SumPaidAmount:   20000.0,
							CountPaidAmount: 1,
						},
					},
				}
				paymentSvc.On("GetDetailByID",
					c.ValueCtxMockType(),
					c.StringMockType(),
				).Once().Return(paymentSession, nil)
				paymentRepo.On("BeginTransaction", c.ValueCtxMockType()).
					Once().Return(context.WithValue(context.Background(), mySqlExt.CtxSqlTx, &sqlx.Tx{}), nil)
				paymentSvc.On("DeterminePaymentFee", mock.Anything, mock.Anything).Once().Return(nil)
				paymentSvc.On("PostCreateLedger", c.ValueCtxMockType(), mock.MatchedBy(func(p *paymentModel.Payment) bool {
					// Verify BankReferenceId is set from request
					return p.BankReferenceId == "BANKREF987654321"
				}), mock.Anything).Once().Return(nil)
				paymentRepo.On("UpdatePaymentMetadataById", c.ValueCtxMockType(), mock.Anything, mock.Anything).
					Once().Return(nil)
				paymentRepo.On("CommitTransaction", c.ValueCtxMockType()).Once().Return(nil)
				accountTrxRepo.On("FindByID", c.ValueCtxMockType(), mock.Anything).
					Once().Return(&orchestraModel.AccountTransactionWithUseCase{UUID: uuid.New()}, nil)
				rabbitMq.On("PushNotification", mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			name:    "[Paid Flow] SUCCESS: Payment card paid",
			request: paidRequest,
			setupMock: func() {
				paymentSvc.On(
					"GetDetailByID", c.ValueCtxMockType(), c.StringMockType(),
				).Once().Return(&paymentModel.Payment{
					MerchantID:  "merchant-id-123",
					ReferenceID: util.ValueToPtr("ref-000001"),
					Amount:      decimal.NewFromFloat(10000.00),
					PaymentMethod: paymentModel.PaymentMethod{
						Type: c.ChannelCreditCard,
					},
					Type: constant.PaymentTypeCardFundedPayout,
					Metadata: &map[string]any{
						"cardFundedPayout": map[string]any{
							"sequence": 1,
						},
					},
				}, nil)
				paymentRepo.On("BeginTransaction", c.ValueCtxMockType()).Once().Return(context.WithValue(context.Background(), mySqlExt.CtxSqlTx, &sqlx.Tx{}), nil)
				merchantRepo.On("FindMerchantByID", c.ValueCtxMockType(), c.StringMockType()).Return(&merchantModel.Merchant{BusinessCountry: sql.NullString{Valid: true, String: "IDN"}}, nil)
				paymentSvc.On("DeterminePaymentFee", mock.Anything, mock.Anything).Once().Return(nil)
				paymentRepo.On("UpdatePaymentStatus", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType(), c.StringMockType(), c.TimeMockType()).Once().Return(nil)
				accountTrxRepo.On("FindByID", c.ValueCtxMockType(), c.StringMockType()).Once().Return(nil, nil)
				paymentSvc.On("PostCreateLedger", c.ValueCtxMockType(), mock.Anything, mock.Anything).Once().Return(nil)
				paymentRepo.On("CommitTransaction", c.ValueCtxMockType()).Once().Return(nil)
				cardFundedPayoutSvc.On("ProcessPendingSubsequentPayments", mock.Anything, "merchant-id-123", "ref-000001").Once().Return(nil)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Reset shared SetNX result to default success before each case.
			// Lock-error cases override this in setupMock.
			redisSetNXResult.SetVal(true)
			redisSetNXResult.SetErr(nil)

			tc.setupMock()

			svc := New(cfg, log, paymentRepo, paymentMethodRepo, accountTrxRepo,
				WithPaymentService(paymentSvc),
				WithMerchantRepo(merchantRepo),
				WithRabbitMQClient(rabbitMq),
				WithRedisClient(redis),
				WithFdsService(fdsSvc),
				WithPaymentMethodService(paymentMethodSvc),
				WithRecurringContractService(recurringContractSvc),
			)
			WithCardFundedPayoutService(svc, cardFundedPayoutSvc)

			err := svc.ProcessNotification(context.Background(), tc.request)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			cardFundedPayoutSvc.AssertExpectations(t)
			recurringContractSvc.AssertExpectations(t)
		})
	}
}

func TestProcessNotificationSaveCardForFutureUse(t *testing.T) {
	cfg := &config.Config{
		Environment: "test",
	}
	log, _ := mockLogger.NewZapLogger(mockLogger.Config{})
	fdsSvc := serviceMock.NewIFdsService(t)
	paymentMethodSvc := serviceMock.NewIPaymentMethodService(t)

	existingCardFingerprint := "existing-fingerprint-123"
	newCardFingerprint := "new-fingerprint-456"

	existingPaymentMethod := &unifiedPaymentModel.CustomerPaymentMethod{
		Token:          uuid.NewString(),
		PaymentMethod:  c.UnifiedPaymentMethodCard,
		PaymentChannel: "VISA",
		Status:         c.StoredPaymentMethodStatusActive,
		Card: &unifiedPaymentModel.CustomerPaymentMethodCard{
			Fingerprint:         existingCardFingerprint,
			Network:             "VISA",
			Last4:               "1234",
			ExpMonth:            "12",
			ExpYear:             "2025",
			CardHolderFirstName: "John",
			CardHolderLastName:  "Doe",
			CardHolderEmail:     "john.doe@example.com",
			CardHolderPhone:     "+6281234567890",
		},
	}

	paymentMethodSvc.On("FindPaymentMethodByIdAndMerchant", mock.Anything, mock.Anything, mock.Anything).
		Return(&paymentModel.PaymentMethodWithPivot{
			MerchantConfigObj: &paymentModel.PaymentMethodMerchantConfigObject{
				PartnerConfig: &paymentMethodModel.SetupPaymentMethodPartnerConfigRequest{
					Card: &paymentMethodModel.SetupPaymentMethodPartnerConfigForCardRequest{
						Items: []paymentMethodModel.SetupPaymentMethodPartnerConfigForCardObj{
							{
								AcquirerMerchantID: "MID123456",
								Acquirer:           "BCA",
							},
						},
					},
				},
			},
		}, nil).Maybe()

	testCases := []struct {
		name    string
		wantErr bool
		test    func(t *testing.T)
	}{
		{
			name:    "SUCCESS: Save new card with single name",
			wantErr: false,
			test: func(t *testing.T) {
				paymentRepo := repositoryMock.NewIPaymentRepository(t)
				paymentMethodRepo := repositoryMock.NewIPaymentMethodRepository(t)
				accountTrxRepo := repositoryMock.NewIAccountTransactionRepository(t)
				customerRepo := repositoryMock.NewICustomerRepository(t)
				merchantRepo := repositoryMock.NewIMerchantRepository(t)
				paymentSvc := serviceMock.NewIPaymentService(t)
				rabbitMq := rabbitMQMock.NewRabbitMQExt(t)
			redis := redisExtMocks.NewIRedisExt(t)
			redisSetNXResult := &redisSdk.BoolCmd{}
			redisSetNXResult.SetVal(true)
			redis.On("SetNX", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe().Return(redisSetNXResult)
			redisMutex := redisExtMocks.NewIMutexer(t)
			redis.On("NewMutex", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe().Return(redisMutex)
			redisMutex.On("LockContext", mock.Anything).Maybe().Return(nil)
			redisMutex.On("UnlockContext", mock.Anything).Maybe().Return(true, nil)

				customerID := uuid.NewString()
				paymentSessionID := uuid.NewString()
				chargeID := uuid.NewString()

				request := &unifiedPaymentModel.PaymentNotificationRequest{
					PaymentSessionID:       paymentSessionID,
					ChargeID:               chargeID,
					ChargeStatus:           c.ChargeStatusSuccess,
					PaymentMethodType:      c.UnifiedPaymentMethodCard,
					ProcessorTransactionID: "proc-txn-123",
					Processor:              "TEST_PROCESSOR",
					ProcessorID:            "proc-id-123",
					Amount: unifiedPaymentModel.Amount{
						Currency: "IDR",
						Value:    10000.00,
					},
					ChargePaymentMethodDetails: &unifiedPaymentModel.ChargePaymentMethodDetails{
						Card: &unifiedPaymentModel.ChargePaymentMethodDetailCard{
							Fingerprint:      newCardFingerprint,
							CardHolderName:   "Madonna",
							First6:           "370000",
							First8:           "37000000",
							Last4:            "9876",
							ExpMonth:         "03",
							ExpYear:          "2024",
							SaveForFutureUse: util.ValueToPtr(true),
							BinInformations: unifiedPaymentModel.ChargePaymentMethodDetailBinInformation{
								Brand:   "AMEX",
								Country: "IDN",
							},
							AuthorizationResult: &unifiedPaymentModel.ChargePaymentMethodDetailCardAuthorizationResult{
								AuthorizationID: "auth-123",
							},
						},
					},
				}

				basePayment := &paymentModel.Payment{
					UUID:       paymentSessionID,
					CustomerID: customerID,
					MerchantID: uuid.NewString(),
					Amount:     decimal.NewFromFloat(10000.00),
					PaymentMethod: paymentModel.PaymentMethod{
						Type: c.ChannelCreditCard,
					},
					Metadata: &map[string]any{
						"mode": "REDIRECT",
						"paymentMethod": map[string]any{
							"type": "CARD",
						},
						"clientRedirectUrl": map[string]any{
							"successReturnUrl":    "https://merchant.com/success",
							"failureReturnUrl":    "https://merchant.com/failure",
							"expirationReturnUrl": "https://merchant.com/expiration",
						},
					},
				}

				// Payment service mocks
				paymentSvc.On("GetDetailByID", c.ValueCtxMockType(), paymentSessionID).
					Return(basePayment, nil)

				// Transaction mocks
				paymentRepo.On("BeginTransaction", c.ValueCtxMockType()).
					Return(context.WithValue(context.Background(), mySqlExt.CtxSqlTx, &sqlx.Tx{}), nil)
				merchantRepo.On("FindMerchantByID", c.ValueCtxMockType(), mock.AnythingOfType("string")).
					Return(&merchantModel.Merchant{BusinessCountry: sql.NullString{Valid: true, String: "IDN"}}, nil)
				paymentSvc.On("DeterminePaymentFee", mock.Anything, mock.Anything).Return(nil)
				paymentRepo.On("UpdatePaymentStatus", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType(), c.StringMockType(), c.TimeMockType()).
					Return(nil)

				// Customer repository mocks for card storage
				customer := &customerModel.Customer{
					UUID: customerID,
					Metadata: map[string]interface{}{
						"paymentMethods": []*unifiedPaymentModel.CustomerPaymentMethod{
							existingPaymentMethod,
						},
					},
				}
				customerRepo.On("FindCustomerById", c.ValueCtxMockType(), customerID).
					Return(customer, nil)
				customerRepo.On("Update", c.ValueCtxMockType(), mock.MatchedBy(func(updatedCustomer *customerModel.Customer) bool {
					paymentMethods, _ := util.ConvertToStruct[[]*unifiedPaymentModel.CustomerPaymentMethod](updatedCustomer.Metadata["paymentMethods"])
					if len(paymentMethods) != 2 {
						return false
					}
					newCard := paymentMethods[0] // New card is prepended
					return newCard.Card.Fingerprint == newCardFingerprint &&
						newCard.Card.CardHolderFirstName == "Madonna" &&
						newCard.Card.CardHolderLastName == "" &&
						newCard.PaymentChannel == "AMEX"
				})).Return(nil)

				// Ledger mocks
				accountTrxRepo.On("FindByID", c.ValueCtxMockType(), chargeID).
					Return(&orchestraModel.AccountTransactionWithUseCase{
						UUID:            uuid.New(),
						SettlementModel: sql.NullString{Valid: true, String: c.PaymentMethodChannelTypeAggregator},
					}, nil)
				paymentSvc.On("UpdatePendingLedger", c.ValueCtxMockType(), mock.Anything, mock.Anything).
					Return(nil)
				paymentRepo.On("CommitTransaction", c.ValueCtxMockType()).Return(nil)

				// Callback mocks
				accountTrxRepo.On("FindByReference", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType()).
					Return(&orchestraModel.AccountTransactionWithUseCase{UUID: uuid.New()}, nil)
				customerRepo.On("GetCustomerById", c.ValueCtxMockType(), customerID, mock.AnythingOfType("string")).
					Return(customer, nil)
				rabbitMq.On("PublishMerchantCallback", mock.Anything, mock.Anything).Return(nil)
				rabbitMq.On("PushNotification", c.ValueCtxMockType(), c.PtrPushNotificationMockType()).
					Return(nil)

				svc := New(cfg, log, paymentRepo, paymentMethodRepo, accountTrxRepo,
					WithPaymentService(paymentSvc), WithMerchantRepo(merchantRepo), WithCustomerRepo(customerRepo), WithRabbitMQClient(rabbitMq), WithRedisClient(redis), WithFdsService(fdsSvc), WithPaymentMethodService(paymentMethodSvc))

				err := svc.ProcessNotification(context.Background(), request)
				assert.NoError(t, err)
			},
		},
		{
			name:    "SUCCESS: Save new card for existing customer without saved cards",
			wantErr: false,
			test: func(t *testing.T) {
				paymentRepo := repositoryMock.NewIPaymentRepository(t)
				paymentMethodRepo := repositoryMock.NewIPaymentMethodRepository(t)
				accountTrxRepo := repositoryMock.NewIAccountTransactionRepository(t)
				customerRepo := repositoryMock.NewICustomerRepository(t)
				merchantRepo := repositoryMock.NewIMerchantRepository(t)
				paymentSvc := serviceMock.NewIPaymentService(t)
				rabbitMq := rabbitMQMock.NewRabbitMQExt(t)
			redis := redisExtMocks.NewIRedisExt(t)
			redisSetNXResult := &redisSdk.BoolCmd{}
			redisSetNXResult.SetVal(true)
			redis.On("SetNX", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe().Return(redisSetNXResult)
			redisMutex := redisExtMocks.NewIMutexer(t)
			redis.On("NewMutex", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe().Return(redisMutex)
			redisMutex.On("LockContext", mock.Anything).Maybe().Return(nil)
			redisMutex.On("UnlockContext", mock.Anything).Maybe().Return(true, nil)

				customerID := uuid.NewString()
				paymentSessionID := uuid.NewString()
				chargeID := uuid.NewString()

				request := &unifiedPaymentModel.PaymentNotificationRequest{
					PaymentSessionID:       paymentSessionID,
					ChargeID:               chargeID,
					ChargeStatus:           c.ChargeStatusSuccess,
					PaymentMethodType:      c.UnifiedPaymentMethodCard,
					ProcessorTransactionID: "proc-txn-123",
					Processor:              "TEST_PROCESSOR",
					ProcessorID:            "proc-id-123",
					Amount: unifiedPaymentModel.Amount{
						Currency: "IDR",
						Value:    10000.00,
					},
					ChargePaymentMethodDetails: &unifiedPaymentModel.ChargePaymentMethodDetails{
						Card: &unifiedPaymentModel.ChargePaymentMethodDetailCard{
							Fingerprint:      newCardFingerprint,
							CardHolderName:   "Madonna",
							First6:           "370000",
							First8:           "37000000",
							Last4:            "9876",
							ExpMonth:         "03",
							ExpYear:          "2024",
							SaveForFutureUse: util.ValueToPtr(true),
							BinInformations: unifiedPaymentModel.ChargePaymentMethodDetailBinInformation{
								Brand:   "AMEX",
								Country: "IDN",
							},
							AuthorizationResult: &unifiedPaymentModel.ChargePaymentMethodDetailCardAuthorizationResult{
								AuthorizationID: "auth-123",
							},
						},
					},
				}

				basePayment := &paymentModel.Payment{
					UUID:       paymentSessionID,
					CustomerID: customerID,
					MerchantID: uuid.NewString(),
					Amount:     decimal.NewFromFloat(10000.00),
					PaymentMethod: paymentModel.PaymentMethod{
						Type: c.ChannelCreditCard,
					},
					Metadata: &map[string]any{
						"mode": "REDIRECT",
						"paymentMethod": map[string]any{
							"type": "CARD",
						},
						"clientRedirectUrl": map[string]any{
							"successReturnUrl":    "https://merchant.com/success",
							"failureReturnUrl":    "https://merchant.com/failure",
							"expirationReturnUrl": "https://merchant.com/expiration",
						},
					},
				}

				// Payment service mocks
				paymentSvc.On("GetDetailByID", c.ValueCtxMockType(), paymentSessionID).
					Return(basePayment, nil)

				// Transaction mocks
				paymentRepo.On("BeginTransaction", c.ValueCtxMockType()).
					Return(context.WithValue(context.Background(), mySqlExt.CtxSqlTx, &sqlx.Tx{}), nil)
				merchantRepo.On("FindMerchantByID", c.ValueCtxMockType(), mock.AnythingOfType("string")).
					Return(&merchantModel.Merchant{BusinessCountry: sql.NullString{Valid: true, String: "IDN"}}, nil)
				paymentSvc.On("DeterminePaymentFee", mock.Anything, mock.Anything).Return(nil)
				paymentRepo.On("UpdatePaymentStatus", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType(), c.StringMockType(), c.TimeMockType()).
					Return(nil)

				// Customer repository mocks for card storage
				customer := &customerModel.Customer{
					UUID:     customerID,
					Metadata: map[string]interface{}{},
				}
				customerRepo.On("FindCustomerById", c.ValueCtxMockType(), customerID).
					Return(customer, nil)
				customerRepo.On("Update", c.ValueCtxMockType(), mock.MatchedBy(func(updatedCustomer *customerModel.Customer) bool {
					paymentMethods, _ := util.ConvertToStruct[[]*unifiedPaymentModel.CustomerPaymentMethod](updatedCustomer.Metadata["paymentMethods"])
					if len(paymentMethods) != 1 {
						return false
					}
					newCard := paymentMethods[0] // New card is prepended
					return newCard.Card.Fingerprint == newCardFingerprint &&
						newCard.Card.CardHolderFirstName == "Madonna" &&
						newCard.Card.CardHolderLastName == "" &&
						newCard.PaymentChannel == "AMEX"
				})).Return(nil)

				// Ledger mocks
				accountTrxRepo.On("FindByID", c.ValueCtxMockType(), chargeID).
					Return(&orchestraModel.AccountTransactionWithUseCase{
						UUID:            uuid.New(),
						SettlementModel: sql.NullString{Valid: true, String: c.PaymentMethodChannelTypeAggregator},
					}, nil)
				paymentSvc.On("UpdatePendingLedger", c.ValueCtxMockType(), mock.Anything, mock.Anything).
					Return(nil)
				paymentRepo.On("CommitTransaction", c.ValueCtxMockType()).Return(nil)

				// Callback mocks
				accountTrxRepo.On("FindByReference", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType()).
					Return(&orchestraModel.AccountTransactionWithUseCase{UUID: uuid.New()}, nil)
				customerRepo.On("GetCustomerById", c.ValueCtxMockType(), customerID, mock.AnythingOfType("string")).
					Return(customer, nil)
				rabbitMq.On("PublishMerchantCallback", mock.Anything, mock.Anything).Return(nil)
				rabbitMq.On("PushNotification", c.ValueCtxMockType(), c.PtrPushNotificationMockType()).
					Return(nil)

				svc := New(cfg, log, paymentRepo, paymentMethodRepo, accountTrxRepo,
					WithPaymentService(paymentSvc), WithMerchantRepo(merchantRepo), WithCustomerRepo(customerRepo), WithRabbitMQClient(rabbitMq), WithRedisClient(redis), WithFdsService(fdsSvc), WithPaymentMethodService(paymentMethodSvc))

				err := svc.ProcessNotification(context.Background(), request)
				assert.NoError(t, err)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.test(t)
		})
	}
}

func TestUpdateVAMethodDetailOnSuccess(t *testing.T) {
	cfg := &config.Config{
		Environment: "test",
	}
	log, _ := mockLogger.NewZapLogger(mockLogger.Config{})
	fdsSvc := serviceMock.NewIFdsService(t)

	testCases := []struct {
		name           string
		paymentLedger  *orchestraModel.AccountTransactionWithUseCase
		methodDetail   *unifiedPaymentModel.ChargePaymentMethodDetails
		expectedResult *unifiedPaymentModel.ChargePaymentMethodDetails
		description    string
	}{
		{
			name:          "SUCCESS: Update VA name from ledger additional info",
			description:   "Should update VirtualAccountName when ledger has valid method detail in additional info",
			paymentLedger: createMockPaymentLedgerWithVAMethodDetail("Updated VA Name", "BCA", "1234567890"),
			methodDetail: &unifiedPaymentModel.ChargePaymentMethodDetails{
				VirtualAccount: &unifiedPaymentModel.ChargePaymentMethodDetailVirtualAccount{
					Channel:              "BCA",
					VirtualAccountNumber: "1234567890",
					VirtualAccountName:   "Original VA Name",
				},
			},
			expectedResult: &unifiedPaymentModel.ChargePaymentMethodDetails{
				VirtualAccount: &unifiedPaymentModel.ChargePaymentMethodDetailVirtualAccount{
					Channel:              "BCA",
					VirtualAccountNumber: "1234567890",
					VirtualAccountName:   "Updated VA Name",
				},
			},
		},
		{
			name:          "SUCCESS: No update when ledger VA name is empty",
			description:   "Should return original method detail when ledger has empty VirtualAccountName",
			paymentLedger: createMockPaymentLedgerWithVAMethodDetail("", "BCA", "1234567890"),
			methodDetail: &unifiedPaymentModel.ChargePaymentMethodDetails{
				VirtualAccount: &unifiedPaymentModel.ChargePaymentMethodDetailVirtualAccount{
					Channel:              "BCA",
					VirtualAccountNumber: "1234567890",
					VirtualAccountName:   "Original VA Name",
				},
			},
			expectedResult: &unifiedPaymentModel.ChargePaymentMethodDetails{
				VirtualAccount: &unifiedPaymentModel.ChargePaymentMethodDetailVirtualAccount{
					Channel:              "BCA",
					VirtualAccountNumber: "1234567890",
					VirtualAccountName:   "Original VA Name",
				},
			},
		},
		{
			name:           "SUCCESS: Return original when methodDetail is nil",
			description:    "Should return nil when input methodDetail is nil",
			paymentLedger:  createMockPaymentLedgerWithVAMethodDetail("Updated VA Name", "BCA", "1234567890"),
			methodDetail:   nil,
			expectedResult: nil,
		},
		{
			name:          "SUCCESS: Return original when paymentLedger is nil",
			description:   "Should return original methodDetail when paymentLedger is nil",
			paymentLedger: nil,
			methodDetail: &unifiedPaymentModel.ChargePaymentMethodDetails{
				VirtualAccount: &unifiedPaymentModel.ChargePaymentMethodDetailVirtualAccount{
					Channel:              "BCA",
					VirtualAccountNumber: "1234567890",
					VirtualAccountName:   "Original VA Name",
				},
			},
			expectedResult: &unifiedPaymentModel.ChargePaymentMethodDetails{
				VirtualAccount: &unifiedPaymentModel.ChargePaymentMethodDetailVirtualAccount{
					Channel:              "BCA",
					VirtualAccountNumber: "1234567890",
					VirtualAccountName:   "Original VA Name",
				},
			},
		},
		{
			name:          "SUCCESS: Return original when methodDetail.VirtualAccount is nil",
			description:   "Should return original methodDetail when VirtualAccount field is nil",
			paymentLedger: createMockPaymentLedgerWithVAMethodDetail("Updated VA Name", "BCA", "1234567890"),
			methodDetail: &unifiedPaymentModel.ChargePaymentMethodDetails{
				Card: &unifiedPaymentModel.ChargePaymentMethodDetailCard{
					Last4: "1234",
				},
			},
			expectedResult: &unifiedPaymentModel.ChargePaymentMethodDetails{
				Card: &unifiedPaymentModel.ChargePaymentMethodDetailCard{
					Last4: "1234",
				},
			},
		},
		{
			name:          "SUCCESS: Return original when ledger has no VirtualAccount in additional info",
			description:   "Should return original methodDetail when ledger additional info has no VirtualAccount",
			paymentLedger: createMockPaymentLedgerWithCardMethodDetail(),
			methodDetail: &unifiedPaymentModel.ChargePaymentMethodDetails{
				VirtualAccount: &unifiedPaymentModel.ChargePaymentMethodDetailVirtualAccount{
					Channel:              "BCA",
					VirtualAccountNumber: "1234567890",
					VirtualAccountName:   "Original VA Name",
				},
			},
			expectedResult: &unifiedPaymentModel.ChargePaymentMethodDetails{
				VirtualAccount: &unifiedPaymentModel.ChargePaymentMethodDetailVirtualAccount{
					Channel:              "BCA",
					VirtualAccountNumber: "1234567890",
					VirtualAccountName:   "Original VA Name",
				},
			},
		},
		{
			name:          "SUCCESS: Return original when ledger has invalid JSON in additional info",
			description:   "Should return original methodDetail when ledger has invalid JSON in AdditionalInfo",
			paymentLedger: createMockPaymentLedgerWithInvalidJSON(),
			methodDetail: &unifiedPaymentModel.ChargePaymentMethodDetails{
				VirtualAccount: &unifiedPaymentModel.ChargePaymentMethodDetailVirtualAccount{
					Channel:              "BCA",
					VirtualAccountNumber: "1234567890",
					VirtualAccountName:   "Original VA Name",
				},
			},
			expectedResult: &unifiedPaymentModel.ChargePaymentMethodDetails{
				VirtualAccount: &unifiedPaymentModel.ChargePaymentMethodDetailVirtualAccount{
					Channel:              "BCA",
					VirtualAccountNumber: "1234567890",
					VirtualAccountName:   "Original VA Name",
				},
			},
		},
		{
			name:          "SUCCESS: Return original when ledger has empty additional info",
			description:   "Should return original methodDetail when ledger has no additional info",
			paymentLedger: createMockPaymentLedgerWithEmptyAdditionalInfo(),
			methodDetail: &unifiedPaymentModel.ChargePaymentMethodDetails{
				VirtualAccount: &unifiedPaymentModel.ChargePaymentMethodDetailVirtualAccount{
					Channel:              "BCA",
					VirtualAccountNumber: "1234567890",
					VirtualAccountName:   "Original VA Name",
				},
			},
			expectedResult: &unifiedPaymentModel.ChargePaymentMethodDetails{
				VirtualAccount: &unifiedPaymentModel.ChargePaymentMethodDetailVirtualAccount{
					Channel:              "BCA",
					VirtualAccountNumber: "1234567890",
					VirtualAccountName:   "Original VA Name",
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			paymentRepo := repositoryMock.NewIPaymentRepository(t)
			paymentMethodRepo := repositoryMock.NewIPaymentMethodRepository(t)
			accountTrxRepo := repositoryMock.NewIAccountTransactionRepository(t)

			svcInterface := New(cfg, log, paymentRepo, paymentMethodRepo, accountTrxRepo, WithFdsService(fdsSvc))
			svc := svcInterface.(*UnifiedPaymentService)
			result := svc.updateVAMethodDetailOnSuccess(tc.paymentLedger, tc.methodDetail)

			if tc.expectedResult == nil {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
				if tc.expectedResult.VirtualAccount != nil && result.VirtualAccount != nil {
					assert.Equal(t, tc.expectedResult.VirtualAccount.VirtualAccountName, result.VirtualAccount.VirtualAccountName)
					assert.Equal(t, tc.expectedResult.VirtualAccount.Channel, result.VirtualAccount.Channel)
					assert.Equal(t, tc.expectedResult.VirtualAccount.VirtualAccountNumber, result.VirtualAccount.VirtualAccountNumber)
				}
				if tc.expectedResult.Card != nil && result.Card != nil {
					assert.Equal(t, tc.expectedResult.Card.Last4, result.Card.Last4)
				}
			}
		})
	}
}

// Helper functions for creating mock data
func createMockPaymentLedgerWithVAMethodDetail(vaName, channel, vaNumber string) *orchestraModel.AccountTransactionWithUseCase {
	methodDetail := map[string]interface{}{
		"virtualAccount": map[string]interface{}{
			"channel":              channel,
			"virtualAccountNumber": vaNumber,
			"virtualAccountName":   vaName,
		},
	}

	additionalInfo := map[string]interface{}{
		"methodDetail": methodDetail,
	}

	jsonData, _ := json.Marshal(additionalInfo)

	return &orchestraModel.AccountTransactionWithUseCase{
		UUID: uuid.New(),
		AdditionalInfo: types.NullJSONText{
			JSONText: jsonData,
			Valid:    true,
		},
	}
}

func createMockPaymentLedgerWithCardMethodDetail() *orchestraModel.AccountTransactionWithUseCase {
	methodDetail := map[string]interface{}{
		"card": map[string]interface{}{
			"last4":          "1234",
			"fingerprint":    "test-fingerprint",
			"cardHolderName": "John Doe",
		},
	}

	additionalInfo := map[string]interface{}{
		"methodDetail": methodDetail,
	}

	jsonData, _ := json.Marshal(additionalInfo)

	return &orchestraModel.AccountTransactionWithUseCase{
		UUID: uuid.New(),
		AdditionalInfo: types.NullJSONText{
			JSONText: jsonData,
			Valid:    true,
		},
	}
}

func createMockPaymentLedgerWithInvalidJSON() *orchestraModel.AccountTransactionWithUseCase {
	return &orchestraModel.AccountTransactionWithUseCase{
		UUID: uuid.New(),
		AdditionalInfo: types.NullJSONText{
			JSONText: []byte("{invalid json"),
			Valid:    true,
		},
	}
}

func createMockPaymentLedgerWithEmptyAdditionalInfo() *orchestraModel.AccountTransactionWithUseCase {
	return &orchestraModel.AccountTransactionWithUseCase{
		UUID: uuid.New(),
		AdditionalInfo: types.NullJSONText{
			Valid: false,
		},
	}
}

func TestHandleProcessCharge(t *testing.T) {
	cfg := &config.Config{
		Environment: "test",
	}
	ctxTrx := context.WithValue(context.Background(), mySqlExt.CtxSqlTx, &sqlx.Tx{})
	log, _ := mockLogger.NewZapLogger(mockLogger.Config{})
	paymentRepo := repositoryMock.NewIPaymentRepository(t)
	paymentMethodRepo := repositoryMock.NewIPaymentMethodRepository(t)
	accountTrxRepo := repositoryMock.NewIAccountTransactionRepository(t)
	customerRepo := repositoryMock.NewICustomerRepository(t)
	merchantRepo := repositoryMock.NewIMerchantRepository(t)
	paymentSvc := serviceMock.NewIPaymentService(t)
	rabbitMq := rabbitMQMock.NewRabbitMQExt(t)

	accountTrxRepo.On("FindByReference", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType()).Return(nil, nil)

	service := UnifiedPaymentService{
		paymentRepo:            paymentRepo,
		paymentMethodRepo:      paymentMethodRepo,
		accountTransactionRepo: accountTrxRepo,
		customerRepo:           customerRepo,
		merchantRepo:           merchantRepo,
		paymentSvc:             paymentSvc,
		rabbitMqExt:            rabbitMq,
		config:                 cfg,
		logger:                 log,
	}

	paymentIDSuccess := uuid.NewString()
	merchantIDSuccess := uuid.NewString()
	chargeIDSuccess := uuid.NewString()

	paymentIDProcessing := uuid.NewString()
	chargeIDProcessing := uuid.NewString()

	paymentIDBeginTxError := uuid.NewString()
	chargeIDBeginTxError := uuid.NewString()

	paymentIDUpdateStatusError := uuid.NewString()
	merchantIDUpdateStatusError := uuid.NewString()
	chargeIDUpdateStatusError := uuid.NewString()

	paymentIDFindByIDError := uuid.NewString()
	merchantIDFindByIDError := uuid.NewString()
	chargeIDFindByIDError := uuid.NewString()

	paymentIDUpdateLedgerError := uuid.NewString()
	merchantIDUpdateLedgerError := uuid.NewString()
	chargeIDUpdateLedgerError := uuid.NewString()

	paymentIDCommitError := uuid.NewString()
	merchantIDCommitError := uuid.NewString()
	chargeIDCommitError := uuid.NewString()

	paymentIDNilLedger := uuid.NewString()
	merchantIDNilLedger := uuid.NewString()
	chargeIDNilLedger := uuid.NewString()

	testCases := []struct {
		name      string
		wantErr   bool
		payment   *paymentModel.Payment
		request   *unifiedPaymentModel.PaymentNotificationRequest
		setupMock func(paymentRepo *repositoryMock.IPaymentRepository, accountTrxRepo *repositoryMock.IAccountTransactionRepository, paymentSvc *serviceMock.IPaymentService, rabbitMq *rabbitMQMock.RabbitMQExt)
	}{
		{
			name:    "SUCCESS: Process charge",
			wantErr: false,
			payment: &paymentModel.Payment{
				UUID:       paymentIDSuccess,
				MerchantID: merchantIDSuccess,
				Status:     c.UnifiedPaymentSessionStatusRequireAction,
			},
			request: &unifiedPaymentModel.PaymentNotificationRequest{
				ChargeID:               chargeIDSuccess,
				PaymentMethodType:      c.UnifiedPaymentMethodCard,
				ProcessorID:            "proc-id-123",
				ProcessorTransactionID: "proc-txn-123",
				ChargePaymentMethodDetails: &unifiedPaymentModel.ChargePaymentMethodDetails{
					Card: &unifiedPaymentModel.ChargePaymentMethodDetailCard{
						BinInformations: unifiedPaymentModel.ChargePaymentMethodDetailBinInformation{
							Country: "IDN",
						},
					},
				},
			},
			setupMock: func(paymentRepo *repositoryMock.IPaymentRepository, accountTrxRepo *repositoryMock.IAccountTransactionRepository, paymentSvc *serviceMock.IPaymentService, rabbitMq *rabbitMQMock.RabbitMQExt) {
				paymentLedger := &orchestraModel.AccountTransactionWithUseCase{
					UUID: uuid.New(),
				}

				paymentRepo.On("BeginTransaction", mock.Anything).Return(ctxTrx, nil).Once()
				paymentRepo.On("UpdatePaymentStatus", mock.Anything, paymentIDSuccess, merchantIDSuccess, c.UnifiedPaymentSessionStatusProcessing, mock.AnythingOfType("time.Time")).Once().Return(nil)
				accountTrxRepo.On("FindByID", mock.Anything, chargeIDSuccess).Once().Return(paymentLedger, nil)
				paymentSvc.On("UpdatePendingLedger", mock.Anything, mock.AnythingOfType("*paymentModel.Payment"), mock.AnythingOfType("orchestrator_model.UpdatePaymentTransactionRequest")).Once().Return(nil)
				paymentRepo.On("CommitTransaction", mock.Anything).Once().Return(nil)
				rabbitMq.On("PublishWithDelay", mock.Anything, rabbitMqExt.PaymentExpirationRoutingKey, mock.AnythingOfType("*paymentModel.ExpiringPayment"), mock.AnythingOfType("time.Duration")).Once().Return(nil)
			},
		},
		{
			name:    "SUCCESS: Payment already in processing state - should return without error",
			wantErr: false,
			payment: &paymentModel.Payment{
				UUID:   paymentIDProcessing,
				Status: c.UnifiedPaymentSessionStatusProcessing,
			},
			request: &unifiedPaymentModel.PaymentNotificationRequest{
				ChargeID: chargeIDProcessing,
			},
			setupMock: func(paymentRepo *repositoryMock.IPaymentRepository, accountTrxRepo *repositoryMock.IAccountTransactionRepository, paymentSvc *serviceMock.IPaymentService, rabbitMq *rabbitMQMock.RabbitMQExt) {
			},
		},
		{
			name:    "ERROR: BeginTransaction fails",
			wantErr: true,
			payment: &paymentModel.Payment{
				UUID:   paymentIDBeginTxError,
				Status: c.UnifiedPaymentSessionStatusRequireAction,
			},
			request: &unifiedPaymentModel.PaymentNotificationRequest{
				ChargeID: chargeIDBeginTxError,
			},
			setupMock: func(paymentRepo *repositoryMock.IPaymentRepository, accountTrxRepo *repositoryMock.IAccountTransactionRepository, paymentSvc *serviceMock.IPaymentService, rabbitMq *rabbitMQMock.RabbitMQExt) {
				paymentRepo.On("BeginTransaction", mock.Anything).Once().Return(nil, assert.AnError)
			},
		},
		{
			name:    "ERROR: UpdatePaymentStatus fails",
			wantErr: true,
			payment: &paymentModel.Payment{
				UUID:       paymentIDUpdateStatusError,
				MerchantID: merchantIDUpdateStatusError,
				Status:     c.UnifiedPaymentSessionStatusRequireAction,
			},
			request: &unifiedPaymentModel.PaymentNotificationRequest{
				ChargeID: chargeIDUpdateStatusError,
			},
			setupMock: func(paymentRepo *repositoryMock.IPaymentRepository, accountTrxRepo *repositoryMock.IAccountTransactionRepository, paymentSvc *serviceMock.IPaymentService, rabbitMq *rabbitMQMock.RabbitMQExt) {
				paymentRepo.On("BeginTransaction", mock.Anything).Once().Return(ctxTrx, nil)
				paymentRepo.On("UpdatePaymentStatus", mock.Anything, paymentIDUpdateStatusError, merchantIDUpdateStatusError, c.UnifiedPaymentSessionStatusProcessing, mock.AnythingOfType("time.Time")).Once().Return(assert.AnError)
				paymentRepo.On("RollbackTransaction", mock.Anything).Once().Return(nil)
			},
		},
		{
			name:    "ERROR: FindByID fails",
			wantErr: true,
			payment: &paymentModel.Payment{
				UUID:       paymentIDFindByIDError,
				MerchantID: merchantIDFindByIDError,
				Status:     c.UnifiedPaymentSessionStatusRequireAction,
			},
			request: &unifiedPaymentModel.PaymentNotificationRequest{
				ChargeID: chargeIDFindByIDError,
			},
			setupMock: func(paymentRepo *repositoryMock.IPaymentRepository, accountTrxRepo *repositoryMock.IAccountTransactionRepository, paymentSvc *serviceMock.IPaymentService, rabbitMq *rabbitMQMock.RabbitMQExt) {
				paymentRepo.On("BeginTransaction", mock.Anything).Once().Return(ctxTrx, nil)
				paymentRepo.On("UpdatePaymentStatus", ctxTrx, paymentIDFindByIDError, merchantIDFindByIDError, c.UnifiedPaymentSessionStatusProcessing, mock.AnythingOfType("time.Time")).Once().Return(nil)
				accountTrxRepo.On("FindByID", ctxTrx, chargeIDFindByIDError).Once().Return(nil, assert.AnError)
				paymentRepo.On("RollbackTransaction", ctxTrx).Once().Return(nil)
			},
		},
		{
			name:    "ERROR: UpdatePendingLedger fails",
			wantErr: true,
			payment: &paymentModel.Payment{
				UUID:       paymentIDUpdateLedgerError,
				MerchantID: merchantIDUpdateLedgerError,
				Status:     c.UnifiedPaymentSessionStatusRequireAction,
			},
			request: &unifiedPaymentModel.PaymentNotificationRequest{
				ChargeID:               chargeIDUpdateLedgerError,
				PaymentMethodType:      c.UnifiedPaymentMethodCard,
				ProcessorID:            "proc-id-123",
				ProcessorTransactionID: "proc-txn-123",
			},
			setupMock: func(paymentRepo *repositoryMock.IPaymentRepository, accountTrxRepo *repositoryMock.IAccountTransactionRepository, paymentSvc *serviceMock.IPaymentService, rabbitMq *rabbitMQMock.RabbitMQExt) {
				paymentLedger := &orchestraModel.AccountTransactionWithUseCase{
					UUID: uuid.New(),
				}

				paymentRepo.On("BeginTransaction", mock.Anything).Once().Return(ctxTrx, nil)
				paymentRepo.On("UpdatePaymentStatus", ctxTrx, paymentIDUpdateLedgerError, merchantIDUpdateLedgerError, c.UnifiedPaymentSessionStatusProcessing, mock.AnythingOfType("time.Time")).Once().Return(nil)
				accountTrxRepo.On("FindByID", ctxTrx, chargeIDUpdateLedgerError).Once().Return(paymentLedger, nil)
				paymentSvc.On("UpdatePendingLedger", ctxTrx, mock.AnythingOfType("*paymentModel.Payment"), mock.AnythingOfType("orchestrator_model.UpdatePaymentTransactionRequest")).Once().Return(assert.AnError)
				paymentRepo.On("RollbackTransaction", ctxTrx).Once().Return(nil)
			},
		},
		{
			name:    "ERROR: CommitTransaction fails",
			wantErr: true,
			payment: &paymentModel.Payment{
				UUID:       paymentIDCommitError,
				MerchantID: merchantIDCommitError,
				Status:     c.UnifiedPaymentSessionStatusRequireAction,
			},
			request: &unifiedPaymentModel.PaymentNotificationRequest{
				ChargeID:               chargeIDCommitError,
				PaymentMethodType:      c.UnifiedPaymentMethodCard,
				ProcessorID:            "proc-id-123",
				ProcessorTransactionID: "proc-txn-123",
			},
			setupMock: func(paymentRepo *repositoryMock.IPaymentRepository, accountTrxRepo *repositoryMock.IAccountTransactionRepository, paymentSvc *serviceMock.IPaymentService, rabbitMq *rabbitMQMock.RabbitMQExt) {
				paymentLedger := &orchestraModel.AccountTransactionWithUseCase{
					UUID: uuid.New(),
				}

				paymentRepo.On("BeginTransaction", mock.Anything).Once().Return(ctxTrx, nil)
				paymentRepo.On("UpdatePaymentStatus", ctxTrx, paymentIDCommitError, merchantIDCommitError, c.UnifiedPaymentSessionStatusProcessing, mock.AnythingOfType("time.Time")).Once().Return(nil)
				accountTrxRepo.On("FindByID", ctxTrx, chargeIDCommitError).Once().Return(paymentLedger, nil)
				paymentSvc.On("UpdatePendingLedger", ctxTrx, mock.AnythingOfType("*paymentModel.Payment"), mock.AnythingOfType("orchestrator_model.UpdatePaymentTransactionRequest")).Once().Return(nil)
				paymentRepo.On("CommitTransaction", ctxTrx).Once().Return(assert.AnError)
				paymentRepo.On("RollbackTransaction", ctxTrx).Once().Return(nil)
			},
		},
		{
			name:    "SUCCESS: PaymentLedger is nil - should skip UpdatePendingLedger",
			wantErr: false,
			payment: &paymentModel.Payment{
				UUID:       paymentIDNilLedger,
				MerchantID: merchantIDNilLedger,
				Status:     c.UnifiedPaymentSessionStatusRequireAction,
			},
			request: &unifiedPaymentModel.PaymentNotificationRequest{
				ChargeID:               chargeIDNilLedger,
				PaymentMethodType:      c.UnifiedPaymentMethodCard,
				ProcessorID:            "proc-id-123",
				ProcessorTransactionID: "proc-txn-123",
				ChargeStatus:           c.ChargeStatusWaitingForCapture,
			},
			setupMock: func(paymentRepo *repositoryMock.IPaymentRepository, accountTrxRepo *repositoryMock.IAccountTransactionRepository, paymentSvc *serviceMock.IPaymentService, rabbitMq *rabbitMQMock.RabbitMQExt) {
				paymentRepo.On("BeginTransaction", mock.Anything).Once().Return(ctxTrx, nil)
				paymentRepo.On("UpdatePaymentStatus", ctxTrx, paymentIDNilLedger, merchantIDNilLedger, c.UnifiedPaymentSessionStatusProcessing, mock.AnythingOfType("time.Time")).Once().Return(nil)
				accountTrxRepo.On("FindByID", ctxTrx, chargeIDNilLedger).Once().Return(nil, nil)
				paymentRepo.On("CommitTransaction", ctxTrx).Once().Return(nil)
				rabbitMq.On("PublishWithDelay", mock.Anything, rabbitMqExt.PaymentExpirationRoutingKey, mock.AnythingOfType("*paymentModel.ExpiringPayment"), mock.AnythingOfType("time.Duration")).Return(nil)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			tc.setupMock(paymentRepo, accountTrxRepo, paymentSvc, rabbitMq)

			err := service.handleProcessCharge(context.Background(), tc.payment, tc.request)

			if tc.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			if tc.payment.Status != c.UnifiedPaymentSessionStatusProcessing && tc.name != "SUCCESS: Payment already in processing state - should return without error" {
				assert.Equal(t, c.UnifiedPaymentSessionStatusProcessing, tc.payment.Status)
			}
		})
	}
}

func TestChangeSettlementModelForCardPayment(t *testing.T) {
	cfg := &config.Config{
		Environment: "test",
	}
	log, _ := mockLogger.NewZapLogger(mockLogger.Config{})
	accountTrxRepo := repositoryMock.NewIAccountTransactionRepository(t)
	orchestratorSvc := serviceMock.NewIOrchestratorService(t)

	service := &UnifiedPaymentService{
		config:                 cfg,
		logger:                 log,
		accountTransactionRepo: accountTrxRepo,
		orchestratorSvc:        orchestratorSvc,
	}

	paymentID := uuid.NewString()
	ledgerID := uuid.NewString()
	ctx := context.Background()

	testCases := []struct {
		name          string
		payment       *paymentModel.Payment
		request       *unifiedPaymentModel.PaymentNotificationRequest
		setupMock     func(*repositoryMock.IAccountTransactionRepository, *serviceMock.IOrchestratorService)
		expectedError bool
	}{
		{
			name:          "SUCCESS: payment is nil - should return nil",
			payment:       nil,
			request:       &unifiedPaymentModel.PaymentNotificationRequest{},
			setupMock:     func(_ *repositoryMock.IAccountTransactionRepository, _ *serviceMock.IOrchestratorService) {},
			expectedError: false,
		},
		{
			name:    "SUCCESS: notification request is nil - should return nil",
			payment: &paymentModel.Payment{UUID: paymentID},
			request: nil,
			setupMock: func(_ *repositoryMock.IAccountTransactionRepository, _ *serviceMock.IOrchestratorService) {
			},
			expectedError: false,
		},
		{
			name: "SUCCESS: payment method is not card - should return nil",
			payment: &paymentModel.Payment{
				UUID: paymentID,
				PaymentMethod: paymentModel.PaymentMethod{
					Type: c.UnifiedPaymentMethodQris,
				},
			},
			request: &unifiedPaymentModel.PaymentNotificationRequest{
				ChargePaymentMethodDetails: &unifiedPaymentModel.ChargePaymentMethodDetails{
					Card: &unifiedPaymentModel.ChargePaymentMethodDetailCard{
						MIDInfo: &unifiedPaymentModel.MIDInfo{
							Type: c.CreditCardMidTypeAggregator,
						},
					},
				},
			},
			setupMock:     func(_ *repositoryMock.IAccountTransactionRepository, _ *serviceMock.IOrchestratorService) {},
			expectedError: false,
		},
		{
			name: "SUCCESS: ChargePaymentMethodDetails is nil - should return nil",
			payment: &paymentModel.Payment{
				UUID: paymentID,
				PaymentMethod: paymentModel.PaymentMethod{
					Type: paymentConstant.PAYMENT_METHOD_CREDIT_CARD,
				},
			},
			request: &unifiedPaymentModel.PaymentNotificationRequest{
				ChargePaymentMethodDetails: nil,
			},
			setupMock:     func(_ *repositoryMock.IAccountTransactionRepository, _ *serviceMock.IOrchestratorService) {},
			expectedError: false,
		},
		{
			name: "SUCCESS: Card is nil - should return nil",
			payment: &paymentModel.Payment{
				UUID: paymentID,
				PaymentMethod: paymentModel.PaymentMethod{
					Type: paymentConstant.PAYMENT_METHOD_CREDIT_CARD,
				},
			},
			request: &unifiedPaymentModel.PaymentNotificationRequest{
				ChargePaymentMethodDetails: &unifiedPaymentModel.ChargePaymentMethodDetails{
					Card: nil,
				},
			},
			setupMock:     func(_ *repositoryMock.IAccountTransactionRepository, _ *serviceMock.IOrchestratorService) {},
			expectedError: false,
		},
		{
			name: "SUCCESS: MIDInfo is nil - should return nil",
			payment: &paymentModel.Payment{
				UUID: paymentID,
				PaymentMethod: paymentModel.PaymentMethod{
					Type: paymentConstant.PAYMENT_METHOD_CREDIT_CARD,
				},
			},
			request: &unifiedPaymentModel.PaymentNotificationRequest{
				ChargePaymentMethodDetails: &unifiedPaymentModel.ChargePaymentMethodDetails{
					Card: &unifiedPaymentModel.ChargePaymentMethodDetailCard{
						MIDInfo: nil,
					},
				},
			},
			setupMock:     func(_ *repositoryMock.IAccountTransactionRepository, _ *serviceMock.IOrchestratorService) {},
			expectedError: false,
		},
		{
			name: "SUCCESS: settlement model already equal - should return without update",
			payment: &paymentModel.Payment{
				UUID: paymentID,
				PaymentMethod: paymentModel.PaymentMethod{
					Type: paymentConstant.PAYMENT_METHOD_CREDIT_CARD,
				},
			},
			request: &unifiedPaymentModel.PaymentNotificationRequest{
				ChargePaymentMethodDetails: &unifiedPaymentModel.ChargePaymentMethodDetails{
					Card: &unifiedPaymentModel.ChargePaymentMethodDetailCard{
						MIDInfo: &unifiedPaymentModel.MIDInfo{
							Type: c.CreditCardMidTypeAggregator,
						},
					},
				},
			},
			setupMock: func(_ *repositoryMock.IAccountTransactionRepository, orchestratorSvc *serviceMock.IOrchestratorService) {
				orchestratorSvc.On("FindByReference", ctx, paymentID, c.TypePayment).Once().Return(&orchestraModel.AccountTransactionWithUseCase{
					UUID: uuid.MustParse(ledgerID),
					SettlementModel: sql.NullString{
						Valid:  true,
						String: c.PaymentMethodChannelTypeAggregator,
					},
				}, nil)
			},
			expectedError: false,
		},
		{
			name: "SUCCESS: settlement model needs update - aggregator to facilitator",
			payment: &paymentModel.Payment{
				UUID: paymentID,
				PaymentMethod: paymentModel.PaymentMethod{
					Type: paymentConstant.PAYMENT_METHOD_CREDIT_CARD,
				},
			},
			request: &unifiedPaymentModel.PaymentNotificationRequest{
				ChargePaymentMethodDetails: &unifiedPaymentModel.ChargePaymentMethodDetails{
					Card: &unifiedPaymentModel.ChargePaymentMethodDetailCard{
						MIDInfo: &unifiedPaymentModel.MIDInfo{
							Type: c.CreditCardMidTypeDirect,
						},
					},
				},
			},
			setupMock: func(accountTrxRepo *repositoryMock.IAccountTransactionRepository, orchestratorSvc *serviceMock.IOrchestratorService) {
				orchestratorSvc.On("FindByReference", ctx, paymentID, c.TypePayment).Once().Return(&orchestraModel.AccountTransactionWithUseCase{
					UUID: uuid.MustParse(ledgerID),
					SettlementModel: sql.NullString{
						Valid:  true,
						String: c.PaymentMethodChannelTypeAggregator,
					},
				}, nil)
				accountTrxRepo.On("UpdatePaymentTransactionStatusAndMetadataByID", ctx, mock.MatchedBy(func(req orchestraModel.UpdatePaymentTransactionRequest) bool {
					return req.LedgerId == ledgerID &&
						req.SettlementModel != nil &&
						c.IsDirectPSP(*req.SettlementModel)
				}), mock.AnythingOfType("orchestrator_model.MetadataPayment[interface {}]")).Once().Return(nil)
			},
			expectedError: false,
		},
		{
			name: "SUCCESS: settlement model needs update - facilitator to aggregator",
			payment: &paymentModel.Payment{
				UUID: paymentID,
				PaymentMethod: paymentModel.PaymentMethod{
					Type: paymentConstant.PAYMENT_METHOD_CREDIT_CARD,
				},
			},
			request: &unifiedPaymentModel.PaymentNotificationRequest{
				ChargePaymentMethodDetails: &unifiedPaymentModel.ChargePaymentMethodDetails{
					Card: &unifiedPaymentModel.ChargePaymentMethodDetailCard{
						MIDInfo: &unifiedPaymentModel.MIDInfo{
							Type: c.CreditCardMidTypeAggregator,
						},
					},
				},
			},
			setupMock: func(accountTrxRepo *repositoryMock.IAccountTransactionRepository, orchestratorSvc *serviceMock.IOrchestratorService) {
				orchestratorSvc.On("FindByReference", ctx, paymentID, c.TypePayment).Once().Return(&orchestraModel.AccountTransactionWithUseCase{
					UUID: uuid.MustParse(ledgerID),
					SettlementModel: sql.NullString{
						Valid:  true,
						String: c.PaymentMethodChannelTypeFacilitator,
					},
				}, nil)
				accountTrxRepo.On("UpdatePaymentTransactionStatusAndMetadataByID", ctx, mock.MatchedBy(func(req orchestraModel.UpdatePaymentTransactionRequest) bool {
					return req.LedgerId == ledgerID &&
						req.SettlementModel != nil &&
						*req.SettlementModel == c.PaymentMethodChannelTypeAggregator
				}), mock.AnythingOfType("orchestrator_model.MetadataPayment[interface {}]")).Once().Return(nil)
			},
			expectedError: false,
		},
		{
			name: "SUCCESS: payment ledger is nil - should use default aggregator",
			payment: &paymentModel.Payment{
				UUID: paymentID,
				PaymentMethod: paymentModel.PaymentMethod{
					Type: paymentConstant.PAYMENT_METHOD_CREDIT_CARD,
				},
			},
			request: &unifiedPaymentModel.PaymentNotificationRequest{
				ChargePaymentMethodDetails: &unifiedPaymentModel.ChargePaymentMethodDetails{
					Card: &unifiedPaymentModel.ChargePaymentMethodDetailCard{
						MIDInfo: &unifiedPaymentModel.MIDInfo{
							Type: c.CreditCardMidTypeDirect,
						},
					},
				},
			},
			setupMock: func(accountTrxRepo *repositoryMock.IAccountTransactionRepository, orchestratorSvc *serviceMock.IOrchestratorService) {
				orchestratorSvc.On("FindByReference", ctx, paymentID, c.TypePayment).Once().Return(nil, nil)
			},
			expectedError: false,
		},
		{
			name: "SUCCESS: settlement model is invalid - should use default aggregator and update",
			payment: &paymentModel.Payment{
				UUID: paymentID,
				PaymentMethod: paymentModel.PaymentMethod{
					Type: paymentConstant.PAYMENT_METHOD_CREDIT_CARD,
				},
			},
			request: &unifiedPaymentModel.PaymentNotificationRequest{
				ChargePaymentMethodDetails: &unifiedPaymentModel.ChargePaymentMethodDetails{
					Card: &unifiedPaymentModel.ChargePaymentMethodDetailCard{
						MIDInfo: &unifiedPaymentModel.MIDInfo{
							Type: c.CreditCardMidTypeDirect,
						},
					},
				},
			},
			setupMock: func(accountTrxRepo *repositoryMock.IAccountTransactionRepository, orchestratorSvc *serviceMock.IOrchestratorService) {
				orchestratorSvc.On("FindByReference", ctx, paymentID, c.TypePayment).Once().Return(&orchestraModel.AccountTransactionWithUseCase{
					UUID: uuid.MustParse(ledgerID),
					SettlementModel: sql.NullString{
						Valid:  false,
						String: "",
					},
				}, nil)
				accountTrxRepo.On("UpdatePaymentTransactionStatusAndMetadataByID", ctx, mock.MatchedBy(func(req orchestraModel.UpdatePaymentTransactionRequest) bool {
					return req.LedgerId == ledgerID &&
						req.SettlementModel != nil &&
						c.IsDirectPSP(*req.SettlementModel)
				}), mock.AnythingOfType("orchestrator_model.MetadataPayment[interface {}]")).Once().Return(nil)
			},
			expectedError: false,
		},
		{
			name: "ERROR: update transaction returns ErrDataNotFound",
			payment: &paymentModel.Payment{
				UUID: paymentID,
				PaymentMethod: paymentModel.PaymentMethod{
					Type: paymentConstant.PAYMENT_METHOD_CREDIT_CARD,
				},
			},
			request: &unifiedPaymentModel.PaymentNotificationRequest{
				ChargePaymentMethodDetails: &unifiedPaymentModel.ChargePaymentMethodDetails{
					Card: &unifiedPaymentModel.ChargePaymentMethodDetailCard{
						MIDInfo: &unifiedPaymentModel.MIDInfo{
							Type: c.CreditCardMidTypeDirect,
						},
					},
				},
			},
			setupMock: func(accountTrxRepo *repositoryMock.IAccountTransactionRepository, orchestratorSvc *serviceMock.IOrchestratorService) {
				orchestratorSvc.On("FindByReference", ctx, paymentID, c.TypePayment).Once().Return(&orchestraModel.AccountTransactionWithUseCase{
					UUID: uuid.MustParse(ledgerID),
					SettlementModel: sql.NullString{
						Valid:  true,
						String: c.PaymentMethodChannelTypeAggregator,
					},
				}, nil)
				accountTrxRepo.On("UpdatePaymentTransactionStatusAndMetadataByID", ctx, mock.Anything, mock.Anything).Once().Return(c.ErrDataNotFound)
			},
			expectedError: true,
		},
		{
			name: "ERROR: update transaction returns database error",
			payment: &paymentModel.Payment{
				UUID: paymentID,
				PaymentMethod: paymentModel.PaymentMethod{
					Type: paymentConstant.PAYMENT_METHOD_CREDIT_CARD,
				},
			},
			request: &unifiedPaymentModel.PaymentNotificationRequest{
				ChargePaymentMethodDetails: &unifiedPaymentModel.ChargePaymentMethodDetails{
					Card: &unifiedPaymentModel.ChargePaymentMethodDetailCard{
						MIDInfo: &unifiedPaymentModel.MIDInfo{
							Type: c.CreditCardMidTypeDirect,
						},
					},
				},
			},
			setupMock: func(accountTrxRepo *repositoryMock.IAccountTransactionRepository, orchestratorSvc *serviceMock.IOrchestratorService) {
				orchestratorSvc.On("FindByReference", ctx, paymentID, c.TypePayment).Once().Return(&orchestraModel.AccountTransactionWithUseCase{
					UUID: uuid.MustParse(ledgerID),
					SettlementModel: sql.NullString{
						Valid:  true,
						String: c.PaymentMethodChannelTypeAggregator,
					},
				}, nil)
				accountTrxRepo.On("UpdatePaymentTransactionStatusAndMetadataByID", ctx, mock.Anything, mock.Anything).Once().Return(assert.AnError)
			},
			expectedError: true,
		},
		{
			name: "ERROR: when failed to get payment ledger",
			payment: &paymentModel.Payment{
				UUID: paymentID,
				PaymentMethod: paymentModel.PaymentMethod{
					Type: paymentConstant.PAYMENT_METHOD_CREDIT_CARD,
				},
			},
			request: &unifiedPaymentModel.PaymentNotificationRequest{
				ChargePaymentMethodDetails: &unifiedPaymentModel.ChargePaymentMethodDetails{
					Card: &unifiedPaymentModel.ChargePaymentMethodDetailCard{
						MIDInfo: &unifiedPaymentModel.MIDInfo{
							Type: c.CreditCardMidTypeDirect,
						},
					},
				},
			},
			setupMock: func(accountTrxRepo *repositoryMock.IAccountTransactionRepository, orchestratorSvc *serviceMock.IOrchestratorService) {
				orchestratorSvc.On("FindByReference", ctx, paymentID, c.TypePayment).Once().Return(nil, assert.AnError)
			},
			expectedError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMock(accountTrxRepo, orchestratorSvc)

			err := service.changeSettlementModelForCardPayment(ctx, tc.payment, tc.request)

			if tc.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidatePaymentFinalStatus(t *testing.T) {
	ctx := context.Background()
	cfg := &config.Config{
		Environment: "test",
	}
	log, _ := mockLogger.NewZapLogger(mockLogger.Config{})

	testCases := []struct {
		name      string
		paymentID string
		status    string
		wantErr   bool
	}{
		{
			name:      "SUCCESS: Payment status is PROCESSING (allowed)",
			paymentID: "payment-456",
			status:    c.UnifiedPaymentSessionStatusProcessing,
			wantErr:   false,
		},
		{
			name:      "SUCCESS: Payment status is REQUIRE_ACTION (allowed)",
			paymentID: "payment-789",
			status:    c.UnifiedPaymentSessionStatusRequireAction,
			wantErr:   false,
		},
		{
			name:      "SUCCESS: Payment status is EXPIRED (allowed to be replaced)",
			paymentID: "payment-999",
			status:    c.UnifiedPaymentSessionStatusExpired,
			wantErr:   false,
		},
		{
			name:      "ERROR: Payment status is PAID (final status)",
			paymentID: "payment-paid-123",
			status:    c.UnifiedPaymentSessionStatusPaid,
			wantErr:   true,
		},
		{
			name:      "ERROR: Payment status is CANCELLED (final status)",
			paymentID: "payment-cancelled-456",
			status:    c.UnifiedPaymentSessionStatusCancelled,
			wantErr:   true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			paymentRepo := repositoryMock.NewIPaymentRepository(t)
			paymentMethodRepo := repositoryMock.NewIPaymentMethodRepository(t)
			accountTrxRepo := repositoryMock.NewIAccountTransactionRepository(t)

			svcInterface := New(cfg, log, paymentRepo, paymentMethodRepo, accountTrxRepo)
			svc := svcInterface.(*UnifiedPaymentService)

			err := svc.validatePaymentFinalStatus(ctx, tc.paymentID, tc.status)

			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSendPaymentFinalStatusConflictAlert(t *testing.T) {
	ctx := context.Background()
	cfg := &config.Config{
		Environment: "test",
		SlackConfig: config.SlackConfig{
			PaymentNotifWebhookURL: "https://hooks.slack.com/services/test",
		},
	}
	log, _ := mockLogger.NewZapLogger(mockLogger.Config{})

	testCases := []struct {
		name              string
		paymentID         string
		previousStatus    string
		afterStatus       string
		chargeID          string
		processor         string
		paymentMethodType string
		setupMock         func(*rabbitMQMock.RabbitMQExt)
		expectPublish     bool
	}{
		{
			name:              "SUCCESS: Send alert for PAID to SUCCESS conflict",
			paymentID:         "payment-123",
			previousStatus:    c.UnifiedPaymentSessionStatusPaid,
			afterStatus:       c.ChargeStatusSuccess,
			chargeID:          "charge-456",
			processor:         "TEST_PROCESSOR",
			paymentMethodType: c.UnifiedPaymentMethodVA,
			expectPublish:     true,
			setupMock: func(rabbitMq *rabbitMQMock.RabbitMQExt) {
				rabbitMq.On("Publish",
					c.ValueCtxMockType(),
					rabbitMqExt.SlackPostWebhookRoutingKey,
					mock.Anything,
					mock.MatchedBy(func(data []byte) bool {
						// Verify that the message contains expected fields
						return len(data) > 0
					}),
				).Once().Return(nil)
			},
		},
		{
			name:              "SUCCESS: Send alert for CANCELLED to SUCCESS conflict",
			paymentID:         "payment-789",
			previousStatus:    c.UnifiedPaymentSessionStatusCancelled,
			afterStatus:       c.ChargeStatusSuccess,
			chargeID:          "charge-101",
			processor:         "SNAP_CORE",
			paymentMethodType: c.UnifiedPaymentMethodCard,
			expectPublish:     true,
			setupMock: func(rabbitMq *rabbitMQMock.RabbitMQExt) {
				rabbitMq.On("Publish",
					c.ValueCtxMockType(),
					rabbitMqExt.SlackPostWebhookRoutingKey,
					mock.Anything,
					mock.Anything,
				).Once().Return(nil)
			},
		},
		{
			name:              "ERROR: RabbitMQ publish fails (should not panic)",
			paymentID:         "payment-999",
			previousStatus:    c.UnifiedPaymentSessionStatusPaid,
			afterStatus:       c.ChargeStatusSuccess,
			chargeID:          "charge-999",
			processor:         "TEST_PROCESSOR",
			paymentMethodType: paymentConstant.PAYMENT_METHOD_QRIS,
			expectPublish:     true,
			setupMock: func(rabbitMq *rabbitMQMock.RabbitMQExt) {
				rabbitMq.On("Publish",
					c.ValueCtxMockType(),
					rabbitMqExt.SlackPostWebhookRoutingKey,
					mock.Anything,
					mock.Anything,
				).Once().Return(c.ErrSomeErrorForUnitTest)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			paymentRepo := repositoryMock.NewIPaymentRepository(t)
			paymentMethodRepo := repositoryMock.NewIPaymentMethodRepository(t)
			accountTrxRepo := repositoryMock.NewIAccountTransactionRepository(t)
			rabbitMq := rabbitMQMock.NewRabbitMQExt(t)
			redis := redisExtMocks.NewIRedisExt(t)
			redisSetNXResult := &redisSdk.BoolCmd{}
			redisSetNXResult.SetVal(true)
			redis.On("SetNX", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe().Return(redisSetNXResult)
			redisMutex := redisExtMocks.NewIMutexer(t)
			redis.On("NewMutex", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe().Return(redisMutex)
			redisMutex.On("LockContext", mock.Anything).Maybe().Return(nil)
			redisMutex.On("UnlockContext", mock.Anything).Maybe().Return(true, nil)

			tc.setupMock(rabbitMq)

			svcInterface := New(cfg, log, paymentRepo, paymentMethodRepo, accountTrxRepo,
				WithRabbitMQClient(rabbitMq))
			svc := svcInterface.(*UnifiedPaymentService)

			// This should not panic even if RabbitMQ fails
			svc.sendPaymentFinalStatusConflictAlert(
				ctx,
				tc.paymentID,
				tc.previousStatus,
				tc.afterStatus,
				tc.chargeID,
				tc.processor,
				tc.paymentMethodType,
			)

			// Verify all expectations were met
			rabbitMq.AssertExpectations(t)
		})
	}
}

func TestHandleAutoSplitPayment(t *testing.T) {
	log, _ := mockLogger.NewZapLogger(mockLogger.Config{})

	const (
		paymentUUID = "payment-session-uuid"
		parentID    = "parent-payment-id"
		merchantID  = "merchant-id"
		pending     = "PENDING"
	)

	// autoSplitMetadata builds a Payment.Metadata payload whose autoSplitPayment
	// entry is parsed back by ToUnifiedPaymentMetadata, driving the branch taken.
	autoSplitMetadata := func(txType string) *map[string]any {
		return &map[string]any{
			"autoSplitPayment": map[string]any{
				"transactionType": txType,
			},
		}
	}

	basePayment := func(metadata *map[string]any) *paymentModel.Payment {
		return &paymentModel.Payment{
			UUID:        paymentUUID,
			ReferenceID: util.ValueToPtr(parentID),
			MerchantID:  merchantID,
			Metadata:    metadata,
		}
	}

	successRequest := &unifiedPaymentModel.PaymentNotificationRequest{ChargeStatus: c.ChargeStatusSuccess}
	failedRequest := &unifiedPaymentModel.PaymentNotificationRequest{ChargeStatus: c.ChargeStatusFailed}
	pendingRequest := &unifiedPaymentModel.PaymentNotificationRequest{ChargeStatus: pending}

	// Expected payloads forwarded to the internal service per branch.
	initiateReq := &paymentModel.ProcessSplitPaymentRequest{
		ParentPaymentID:   paymentUUID,
		ThreeDSCallbackID: "",
		FingerprintID:     "",
	}
	abortReq := &paymentModel.ProcessSplitPaymentRequest{
		ParentPaymentID: parentID,
		MerchantID:      merchantID,
	}
	continueReq := &paymentModel.ProcessSplitPaymentRequest{
		ParentPaymentID: parentID,
		MerchantID:      merchantID,
	}

	tests := []struct {
		name      string
		inErr     error
		request   *unifiedPaymentModel.PaymentNotificationRequest
		payment   *paymentModel.Payment
		setupMock func(internalSvc *serviceMock.IInternalUnifiedPaymentService)
	}{
		{
			name:    "skip: incoming error is non-nil",
			inErr:   c.ErrSomeErrorForUnitTest,
			request: successRequest,
			payment: basePayment(autoSplitMetadata(c.AutoSplitPaymentTypeAuthentication)),
			setupMock: func(internalSvc *serviceMock.IInternalUnifiedPaymentService) {
			},
		},
		{
			name:    "skip: payment has no auto split metadata",
			inErr:   nil,
			request: successRequest,
			payment: basePayment(nil),
			setupMock: func(internalSvc *serviceMock.IInternalUnifiedPaymentService) {
			},
		},
		{
			name:    "auth: non-success charge status returns early",
			inErr:   nil,
			request: failedRequest,
			payment: basePayment(autoSplitMetadata(c.AutoSplitPaymentTypeAuthentication)),
			setupMock: func(internalSvc *serviceMock.IInternalUnifiedPaymentService) {
			},
		},
		{
			name:    "auth: success initiates split payment",
			inErr:   nil,
			request: successRequest,
			payment: basePayment(autoSplitMetadata(c.AutoSplitPaymentTypeAuthentication)),
			setupMock: func(internalSvc *serviceMock.IInternalUnifiedPaymentService) {
				internalSvc.On("InitiateSplitPayment", mock.Anything, initiateReq).Once().Return(nil)
			},
		},
		{
			name:    "auth: success logs error when InitiateSplitPayment fails",
			inErr:   nil,
			request: successRequest,
			payment: basePayment(autoSplitMetadata(c.AutoSplitPaymentTypeAuthentication)),
			setupMock: func(internalSvc *serviceMock.IInternalUnifiedPaymentService) {
				internalSvc.On("InitiateSplitPayment", mock.Anything, initiateReq).Once().Return(c.ErrSomeErrorForUnitTest)
			},
		},
		{
			name:    "first payment: failed status aborts split and returns",
			inErr:   nil,
			request: failedRequest,
			payment: basePayment(autoSplitMetadata(c.AutoSplitPaymentTypeFirstPayment)),
			setupMock: func(internalSvc *serviceMock.IInternalUnifiedPaymentService) {
				internalSvc.On("AbortSplitPaymentOnCITFailure", mock.Anything, abortReq).Once().Return(nil)
			},
		},
		{
			name:    "first payment: failed status logs error when abort fails",
			inErr:   nil,
			request: failedRequest,
			payment: basePayment(autoSplitMetadata(c.AutoSplitPaymentTypeFirstPayment)),
			setupMock: func(internalSvc *serviceMock.IInternalUnifiedPaymentService) {
				internalSvc.On("AbortSplitPaymentOnCITFailure", mock.Anything, abortReq).Once().Return(c.ErrSomeErrorForUnitTest)
			},
		},
		{
			name:    "first payment: pending status returns early without abort or continue",
			inErr:   nil,
			request: pendingRequest,
			payment: basePayment(autoSplitMetadata(c.AutoSplitPaymentTypeFirstPayment)),
			setupMock: func(internalSvc *serviceMock.IInternalUnifiedPaymentService) {
			},
		},
		{
			name:    "first payment: success continues and evaluates outcome",
			inErr:   nil,
			request: successRequest,
			payment: basePayment(autoSplitMetadata(c.AutoSplitPaymentTypeFirstPayment)),
			setupMock: func(internalSvc *serviceMock.IInternalUnifiedPaymentService) {
				internalSvc.On("ContinueSplitPaymentExecution", mock.Anything, continueReq).Once().Return(nil)
				internalSvc.On("EvaluateSplitPaymentOutcome", mock.Anything, mock.Anything).Once().Return(nil)
			},
		},
		{
			name:    "first payment: success logs error when continue fails but still evaluates",
			inErr:   nil,
			request: successRequest,
			payment: basePayment(autoSplitMetadata(c.AutoSplitPaymentTypeFirstPayment)),
			setupMock: func(internalSvc *serviceMock.IInternalUnifiedPaymentService) {
				internalSvc.On("ContinueSplitPaymentExecution", mock.Anything, continueReq).Once().Return(c.ErrSomeErrorForUnitTest)
				internalSvc.On("EvaluateSplitPaymentOutcome", mock.Anything, mock.Anything).Once().Return(nil)
			},
		},
		{
			name:    "first payment: success logs error when evaluate fails",
			inErr:   nil,
			request: successRequest,
			payment: basePayment(autoSplitMetadata(c.AutoSplitPaymentTypeFirstPayment)),
			setupMock: func(internalSvc *serviceMock.IInternalUnifiedPaymentService) {
				internalSvc.On("ContinueSplitPaymentExecution", mock.Anything, continueReq).Once().Return(nil)
				internalSvc.On("EvaluateSplitPaymentOutcome", mock.Anything, mock.Anything).Once().Return(c.ErrSomeErrorForUnitTest)
			},
		},
		{
			name:    "subsequent payment: only evaluates outcome",
			inErr:   nil,
			request: successRequest,
			payment: basePayment(autoSplitMetadata(c.AutoSplitPaymentTypeSubsequentPayment)),
			setupMock: func(internalSvc *serviceMock.IInternalUnifiedPaymentService) {
				internalSvc.On("EvaluateSplitPaymentOutcome", mock.Anything, mock.Anything).Once().Return(nil)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			internalSvc := serviceMock.NewIInternalUnifiedPaymentService(t)
			svc := &UnifiedPaymentService{
				logger:                    log,
				internalUnifiedPaymentSvc: internalSvc,
			}

			tc.setupMock(internalSvc)

			svc.handleAutoSplitPayment(context.Background(), tc.request, tc.payment, tc.inErr)

			internalSvc.AssertExpectations(t)
		})
	}
}

func TestHandleWaitingForAuthentication(t *testing.T) {
	log, _ := mockLogger.NewZapLogger(mockLogger.Config{})

	const paymentUUID = "payment-uuid"

	// metadataWithMode builds a Payment.Metadata that parses to the given mode.
	metadataWithMode := func(mode string) *map[string]any {
		return &map[string]any{"mode": mode}
	}

	// unmarshalableMetadata causes json.Marshal to fail so ToUnifiedPaymentMetadata returns nil.
	unmarshalableMetadata := &map[string]any{"bad": make(chan int)}

	creditCardPayment := func(metadata *map[string]any) *paymentModel.Payment {
		return &paymentModel.Payment{
			UUID: paymentUUID,
			PaymentMethod: paymentModel.PaymentMethod{
				Type: c.ChannelCreditCard,
			},
			Metadata: metadata,
		}
	}

	tests := []struct {
		name      string
		payment   *paymentModel.Payment
		setupMock func(paymentRepo *repositoryMock.IPaymentRepository)
		wantErr   bool
		errIs     error
	}{
		{
			name: "skip: non credit card payment method",
			payment: &paymentModel.Payment{
				UUID: paymentUUID,
				PaymentMethod: paymentModel.PaymentMethod{
					Type: c.UnifiedPaymentMethodVA,
				},
			},
			setupMock: func(paymentRepo *repositoryMock.IPaymentRepository) {
			},
			wantErr: false,
		},
		{
			name:    "skip: empty unified payment metadata",
			payment: creditCardPayment(unmarshalableMetadata),
			setupMock: func(paymentRepo *repositoryMock.IPaymentRepository) {
			},
			wantErr: false,
		},
		{
			name:    "skip: non API mode",
			payment: creditCardPayment(metadataWithMode("REDIRECT")),
			setupMock: func(paymentRepo *repositoryMock.IPaymentRepository) {
			},
			wantErr: false,
		},
		{
			name:    "success: API mode updates payment data",
			payment: creditCardPayment(metadataWithMode(c.UnifiedPaymentModeAPI)),
			setupMock: func(paymentRepo *repositoryMock.IPaymentRepository) {
				paymentRepo.On("UpdatePaymentData", mock.Anything, mock.Anything).Once().Return(nil)
			},
			wantErr: false,
		},
		{
			name:    "error: UpdatePaymentData fails wraps database error",
			payment: creditCardPayment(metadataWithMode(c.UnifiedPaymentModeAPI)),
			setupMock: func(paymentRepo *repositoryMock.IPaymentRepository) {
				paymentRepo.On("UpdatePaymentData", mock.Anything, mock.Anything).Once().Return(c.ErrSomeErrorForUnitTest)
			},
			wantErr: true,
			errIs:   c.ErrSomeErrorForUnitTest,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			paymentRepo := repositoryMock.NewIPaymentRepository(t)
			svc := &UnifiedPaymentService{
				logger:      log,
				paymentRepo: paymentRepo,
			}

			tc.setupMock(paymentRepo)

			err := svc.handleWaitingForAuthentication(context.Background(), tc.payment)

			if tc.wantErr {
				assert.Error(t, err)
				if tc.errIs != nil {
					assert.True(t, errors.Is(err, tc.errIs), "expected error to wrap %v, got %v", tc.errIs, err)
				}
			} else {
				assert.NoError(t, err)
			}

			paymentRepo.AssertExpectations(t)
		})
	}
}
