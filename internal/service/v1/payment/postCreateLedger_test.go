package paymentService_test

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"github.com/stretchr/testify/mock"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	paymentConstant "github.com/paper-indonesia/pivot-backoffice/constant/payment"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	feeModel "github.com/paper-indonesia/pivot-backoffice/internal/model/fee"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	orchestrator_model "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	settlementHold "github.com/paper-indonesia/pivot-backoffice/internal/model/settlementHolds"
	unifiedPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/unifiedPayment"
	. "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/payment"
	pdkLoggerMock "github.com/paper-indonesia/pivot-backoffice/mocks/pdk/logger"
	mockRabbitMq "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/rabbitmqExt"
	repositoryMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	serviceMock "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	pkgErr "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
)

func TestPostCreateLedger(t *testing.T) {
	logger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
	merchantRepo := repositoryMocks.NewIMerchantRepository(t)
	paymentRepo := repositoryMocks.NewIPaymentRepository(t)
	accountTransactionRepo := repositoryMocks.NewIAccountTransactionRepository(t)
	orchestratorSvc := serviceMock.NewIOrchestratorService(t)
	paymentMethodSvc := serviceMock.NewIPaymentMethodService(t)
	settlementHoldSvc := serviceMock.NewISettlementHoldService(t)
	feeSvc := serviceMock.NewIFeeService(t)
	rmq := mockRabbitMq.NewRabbitMQExt(t)

	subMerchantID := uuid.NewString()
	parentMerchantID := uuid.NewString()

	validPayment := &paymentModel.Payment{
		UUID:       uuid.NewString(),
		MerchantID: uuid.NewString(),
	}

	feeAmount := decimal.NewFromFloat(2_000)
	validPaymentWithFee := &paymentModel.Payment{
		UUID:       uuid.NewString(),
		MerchantID: uuid.NewString(),
		Fee:        &feeAmount,
		Metadata: &map[string]any{
			"feeDetail": feeModel.FeeMetadataObject{
				DeductionType: constant.MerchantFeeDeductionTypeDirect,
				AmountType:    constant.MerchantFeeAmountType,
				Amount:        2_000.00,
			},
		},
	}

	zeroFeeAmount := decimal.NewFromFloat(0)
	validStaticPaymentWithFee := &paymentModel.Payment{
		UUID:       uuid.NewString(),
		MerchantID: uuid.NewString(),
		Fee:        &zeroFeeAmount,
		Metadata: &map[string]any{
			"feeDetail": feeModel.FeeMetadataObject{
				DeductionType: constant.MerchantFeeDeductionTypeDirect,
				AmountType:    constant.MerchantFeeAmountType,
				Amount:        2_000.00,
			},
		},
	}

	testCases := []struct {
		name      string
		payment   *paymentModel.Payment
		status    string
		wantErr   bool
		setupMock func()
	}{
		{
			name:    "ERROR: Failed to get merchant by ID",
			wantErr: true,
			payment: &paymentModel.Payment{
				MerchantID: uuid.NewString(),
			},
			setupMock: func() {
				merchantRepo.On("FindMerchantByID", constant.ValueCtxMockType(), constant.StringMockType()).Once().
					Return(nil, constant.ErrSomeErrorForUnitTest)
			},
		},
		{
			name:    "ERROR: Not found data when get merchant by ID",
			wantErr: true,
			payment: &paymentModel.Payment{
				MerchantID: uuid.NewString(),
			},
			setupMock: func() {
				merchantRepo.On("FindMerchantByID", constant.ValueCtxMockType(), constant.StringMockType()).Once().
					Return(nil, nil)
			},
		},
		{
			name:    "ERROR: Failed to get payment method",
			wantErr: true,
			payment: &paymentModel.Payment{
				MerchantID: uuid.NewString(),
			},
			setupMock: func() {
				merchantRepo.On("FindMerchantByID", constant.ValueCtxMockType(), constant.StringMockType()).
					Return(&merchant.Merchant{}, nil)

				paymentMethodSvc.On("FindPaymentMethodByIdAndMerchant", constant.ValueCtxMockType(), constant.StringMockType(), constant.StringMockType()).
					Once().Return(nil, constant.ErrSomeErrorForUnitTest)
			},
		},
		{
			name:    "ERROR: Failed to parse merchantID",
			wantErr: true,
			payment: &paymentModel.Payment{
				MerchantID: "invalid",
			},
			setupMock: func() {
				paymentMethodSvc.On("FindPaymentMethodByIdAndMerchant", constant.ValueCtxMockType(), constant.StringMockType(), constant.StringMockType()).
					Return(&paymentModel.PaymentMethodWithPivot{ChannelType: constant.PaymentMethodChannelTypeAggregator}, nil)
			},
		},
		{
			name:    "ERROR: Post account transaction",
			wantErr: true,
			payment: validPayment,
			setupMock: func() {
				orchestratorSvc.On("PostAccountTransaction",
					constant.ValueCtxMockType(),
					constant.PtrCreateAccTransactionReqMockType(),
				).Once().Return(constant.ErrSomeErrorForUnitTest)
			},
		},
		{
			name:    "ERROR: Get merchant settlement config",
			wantErr: true,
			payment: func() *paymentModel.Payment {
				payment := *validPayment
				payment.Type = constant.TypeVirtualTerminal
				return &payment
			}(),
			status: constant.StatusSuccess,
			setupMock: func() {
				merchantRepo.On(
					"GetSettlementConfig", mock.Anything, mock.MatchedBy(func(p merchantModel.GetSettlementConfigRequest) bool {
						return p.Reference == constant.ReferencePayment &&
							*p.Method == constant.TypeVirtualTerminal
					})).Once().Return(nil, assert.AnError)
			},
		},
		{
			name:    "ERROR: Get settlement hold config for payment Multiple QRIS",
			wantErr: false,
			payment: &paymentModel.Payment{
				UUID:       uuid.NewString(),
				MerchantID: uuid.NewString(),
				PaymentMethod: paymentModel.PaymentMethod{
					Type: paymentConstant.PAYMENT_METHOD_QRIS,
				},
				Type: constant.UnifiedPaymentTypeMultiple,
			},
			status: constant.StatusSuccess,
			setupMock: func() {
				paymentRepo.On("GetPaymentById", mock.Anything, constant.StringMockType()).
					Return(&paymentModel.Payment{}, nil)
				merchantRepo.On("FindMerchantByID", constant.ValueCtxMockType(), constant.StringMockType()).
					Return(&merchant.Merchant{KYCStatus: sql.NullString{
						String: constant.KYCStatusApproved,
						Valid:  true,
					}}, nil)

				merchantRepo.On(
					"GetSettlementConfig", mock.Anything, mock.MatchedBy(func(p merchantModel.GetSettlementConfigRequest) bool {
						return p.Reference == constant.ReferencePayment &&
							*p.Method == paymentConstant.PAYMENT_METHOD_QRIS
					}),
				).Once().Return(nil, nil)

				settlementHoldSvc.On("GetSettlementHoldByPaymentID", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("error")).Once()
				orchestratorSvc.On("PostAccountTransaction",
					constant.ValueCtxMockType(),
					constant.PtrCreateAccTransactionReqMockType(),
				).Once().Return(nil)

				rmq.On("PublishForSettlementProcess",
					constant.ValueCtxMockType(),
					mock.Anything,
				).Once().Return(nil)

			},
		},
		{
			name:    "SUCCESS: On Hold payment Multiple QRIS",
			wantErr: false,
			payment: &paymentModel.Payment{
				UUID:       uuid.NewString(),
				MerchantID: uuid.NewString(),
				PaymentMethod: paymentModel.PaymentMethod{
					Type: paymentConstant.PAYMENT_METHOD_QRIS,
				},
				Type: constant.UnifiedPaymentTypeMultiple,
			},
			status: constant.StatusSuccess,
			setupMock: func() {
				paymentRepo.On("GetPaymentById", mock.Anything, constant.StringMockType()).
					Return(&paymentModel.Payment{}, nil)
				merchantRepo.On("FindMerchantByID", constant.ValueCtxMockType(), constant.StringMockType()).
					Return(&merchant.Merchant{KYCStatus: sql.NullString{
						String: constant.KYCStatusApproved,
						Valid:  true,
					}}, nil)

				merchantRepo.On("GetSettlementConfig", constant.ValueCtxMockType(), mock.Anything).Once().Return(nil, nil)

				settlementHoldSvc.On("GetSettlementHoldByPaymentID", mock.Anything, mock.Anything).Return(&settlementHold.SettlementHold{
					Status: constant.SettlementHoldActionHold,
				}, nil)

				orchestratorSvc.On("PostAccountTransaction",
					constant.ValueCtxMockType(),
					constant.PtrCreateAccTransactionReqMockType(),
				).Once().Return(nil)

				rmq.On("PublishForSettlementProcess",
					constant.ValueCtxMockType(),
					mock.Anything,
				).Once().Return(nil)

			},
		},
		{
			name:    "SUCCESS: With default settlement config for payment VA",
			wantErr: false,
			payment: &paymentModel.Payment{
				UUID:       uuid.NewString(),
				MerchantID: uuid.NewString(),
				PaymentMethod: paymentModel.PaymentMethod{
					Type: paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
				},
			},
			status: constant.StatusSuccess,
			setupMock: func() {
				merchantRepo.On("GetSettlementConfig", constant.ValueCtxMockType(), mock.Anything).Once().Return(nil, nil)

				orchestratorSvc.On("PostAccountTransaction",
					constant.ValueCtxMockType(),
					constant.PtrCreateAccTransactionReqMockType(),
				).Once().Return(nil)

				paymentRepo.On("GetPaymentById", mock.Anything, constant.StringMockType()).
					Return(&paymentModel.Payment{}, nil)
				merchantRepo.On("FindMerchantByID", constant.ValueCtxMockType(), constant.StringMockType()).
					Return(&merchant.Merchant{KYCStatus: sql.NullString{
						String: constant.KYCStatusApproved,
						Valid:  true,
					}}, nil)
			},
		},
		{
			name:    "SUCCESS: With default settlement config for payment QRIS",
			wantErr: false,
			payment: &paymentModel.Payment{
				UUID:       uuid.NewString(),
				MerchantID: uuid.NewString(),
				PaymentMethod: paymentModel.PaymentMethod{
					Type: paymentConstant.PAYMENT_METHOD_QRIS,
				},
			},
			status: constant.StatusSuccess,
			setupMock: func() {
				merchantRepo.On("GetSettlementConfig", constant.ValueCtxMockType(), mock.Anything).Once().Return(nil, nil)

				orchestratorSvc.On("PostAccountTransaction",
					constant.ValueCtxMockType(),
					constant.PtrCreateAccTransactionReqMockType(),
				).Once().Return(nil)

				rmq.On("PublishForSettlementProcess",
					constant.ValueCtxMockType(),
					mock.Anything,
				).Once().Return(nil)
			},
		},
		{
			name:    "SUCCESS: With default settlement config for payment CC",
			wantErr: false,
			payment: &paymentModel.Payment{
				UUID:       uuid.NewString(),
				MerchantID: uuid.NewString(),
				PaymentMethod: paymentModel.PaymentMethod{
					Type: paymentConstant.PAYMENT_METHOD_CREDIT_CARD,
				},
				RecurringPayment: &unifiedPaymentModel.MetadataRecurringPayment{
					InitiateFirstAuthorization: true,
					FirstAuthorizationMethod:   constant.RecurringContractAuthMethodOneDollar,
				},
			},
			status: constant.StatusSuccess,
			setupMock: func() {
				merchantRepo.On("GetSettlementConfig", constant.ValueCtxMockType(), mock.Anything).Once().Return(nil, nil)

				orchestratorSvc.On(
					"PostAccountTransaction", constant.ValueCtxMockType(), constant.PtrCreateAccTransactionReqMockType(),
				).Once().Return(nil)
			},
		},
		{
			name:    "ERROR: With default settlement config (settlement type = INSTANT) + fee calculation",
			wantErr: true,
			payment: validPaymentWithFee,
			status:  constant.StatusSuccess,
			setupMock: func() {
				merchantRepo.On("GetSettlementConfig", constant.ValueCtxMockType(), mock.Anything).Once().Return(nil, nil)

				orchestratorSvc.On("PostAccountTransaction",
					constant.ValueCtxMockType(),
					constant.PtrCreateAccTransactionReqMockType(),
				).Once().Return(nil)

				feeSvc.On("CalculateFee",
					constant.ValueCtxMockType(), constant.PtrGetFeeRequestMockType(), constant.PtrFeeMetadataObjectMockType(),
				).Return(1_000.00, 0.00)

				orchestratorSvc.On("PostAccountTransaction",
					constant.ValueCtxMockType(),
					constant.PtrCreateAccTransactionReqMockType(),
				).Once().Return(constant.ErrSomeErrorForUnitTest)
			},
		},
		{
			name:    "SUCCESS: With settlement config (settlement type = T+1)",
			wantErr: false,
			payment: validPayment,
			status:  constant.StatusSuccess,
			setupMock: func() {
				merchantRepo.On(
					"GetSettlementConfig", constant.ValueCtxMockType(), mock.Anything,
				).Once().Return(&merchant.SettlementConfig{Type: "T+1"}, nil)

				orchestratorSvc.On("PostAccountTransaction",
					constant.ValueCtxMockType(),
					constant.PtrCreateAccTransactionReqMockType(),
				).Once().Return(nil)

				rmq.On("PublishForSettlementProcess",
					constant.ValueCtxMockType(),
					mock.Anything,
				).Once().Return(nil)
			},
		},
		{
			name:    "SUCCESS: With settlement config day based (settlement type = D+1)",
			wantErr: false,
			payment: validPayment,
			status:  constant.StatusSuccess,
			setupMock: func() {
				merchantRepo.On(
					"GetSettlementConfig", constant.ValueCtxMockType(), mock.Anything,
				).Once().Return(&merchant.SettlementConfig{Type: "D+1"}, nil)

				orchestratorSvc.On("PostAccountTransaction",
					constant.ValueCtxMockType(),
					constant.PtrCreateAccTransactionReqMockType(),
				).Once().Return(nil)

				rmq.On("PublishWithDelay",
					constant.ValueCtxMockType(),
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Once().Return(nil)

				accountTransactionRepo.On("UpdateSettlementDetailByIDs",
					constant.ValueCtxMockType(),
					mock.Anything,
					mock.Anything,
				).Once().Return(nil)

			},
		},
		{
			name:    "SUCCESS: With settlement config day based (settlement type = D+1) error update",
			wantErr: false,
			payment: validPayment,
			status:  constant.StatusSuccess,
			setupMock: func() {
				merchantRepo.On(
					"GetSettlementConfig", constant.ValueCtxMockType(), mock.Anything,
				).Once().Return(&merchant.SettlementConfig{Type: "D+1"}, nil)

				orchestratorSvc.On("PostAccountTransaction",
					constant.ValueCtxMockType(),
					constant.PtrCreateAccTransactionReqMockType(),
				).Once().Return(nil)

				rmq.On("PublishWithDelay",
					constant.ValueCtxMockType(),
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Once().Return(nil)

				accountTransactionRepo.On("UpdateSettlementDetailByIDs",
					constant.ValueCtxMockType(),
					mock.Anything,
					mock.Anything,
				).Once().Return(fmt.Errorf("error"))

			},
		},
		{
			name:    "SUCCESS: With settlement config (settlement type = Invalid format)",
			wantErr: false,
			payment: validPayment,
			status:  constant.StatusSuccess,
			setupMock: func() {
				merchantRepo.On(
					"GetSettlementConfig", constant.ValueCtxMockType(), mock.Anything,
				).Once().Return(&merchant.SettlementConfig{Type: "invalid"}, nil)

				orchestratorSvc.On("PostAccountTransaction",
					constant.ValueCtxMockType(),
					constant.PtrCreateAccTransactionReqMockType(),
				).Once().Return(nil)

				rmq.On("PublishForSettlementProcess",
					constant.ValueCtxMockType(),
					mock.Anything,
				).Once().Return(nil)
			},
		},
		{
			name:    "SUCCESS: With default settlement config (settlement type = INSTANT) + fee calculation",
			wantErr: false,
			payment: validPaymentWithFee,
			status:  constant.StatusSuccess,
			setupMock: func() {
				merchantRepo.On("GetSettlementConfig", constant.ValueCtxMockType(), mock.Anything).Once().Return(nil, nil)

				orchestratorSvc.On("PostAccountTransaction",
					constant.ValueCtxMockType(),
					constant.PtrCreateAccTransactionReqMockType(),
				).Twice().Return(nil)
			},
		},
		{
			name:    "SUCCESS: With default settlement config (settlement type = INSTANT) + fee calculation + static payment",
			wantErr: false,
			payment: validStaticPaymentWithFee,
			status:  constant.StatusSuccess,
			setupMock: func() {
				merchantRepo.On("GetSettlementConfig", constant.ValueCtxMockType(), mock.Anything).Once().Return(nil, nil)

				orchestratorSvc.On("PostAccountTransaction",
					constant.ValueCtxMockType(),
					constant.PtrCreateAccTransactionReqMockType(),
				).Twice().Return(nil)
			},
		},
		{
			name:    "SUCCESS: Card Funded Payout Settlement Config",
			wantErr: false,
			payment: &paymentModel.Payment{
				UUID:       uuid.NewString(),
				MerchantID: uuid.NewString(),
				Fee:        &zeroFeeAmount,
				Type:       constant.PaymentTypeCardFundedPayout,
				Metadata: &map[string]any{
					"feeDetail": &feeModel.FeeMetadataObject{
						DeductionType: constant.MerchantFeeDeductionTypeDirect,
						AmountType:    constant.MerchantFeeAmountType,
						Amount:        2_000.00,
					},
					"cardFundedDetail": &disbursementModel.CardFundedDetailMetadata{
						SettlementMethod: constant.SettlementTypeStandard,
					},
				},
			},
			status: constant.StatusSuccess,
			setupMock: func() {
				merchantRepo.On("GetSettlementConfig", constant.ValueCtxMockType(), mock.Anything).Once().Return(&merchant.SettlementConfig{Type: "D+1"}, nil)

				orchestratorSvc.On("PostAccountTransaction",
					constant.ValueCtxMockType(),
					constant.PtrCreateAccTransactionReqMockType(),
				).Twice().Return(nil)

				rmq.On("PublishWithDelay",
					constant.ValueCtxMockType(),
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Once().Return(nil)

				accountTransactionRepo.On("UpdateSettlementDetailByIDs",
					constant.ValueCtxMockType(),
					mock.Anything,
					mock.Anything,
				).Once().Return(nil)

			},
		},
		{
			name:    "SUCCESS: Card Funded Payout with CardFundedPayout field set",
			wantErr: false,
			payment: &paymentModel.Payment{
				UUID:       uuid.NewString(),
				MerchantID: uuid.NewString(),
				Fee:        util.ValueToPtr(decimal.NewFromFloat(2_000.00)),
				Type:       constant.PaymentTypeCardFundedPayout,
				CardFundedPayout: &unifiedPaymentModel.CardFundedPayout{
					SettlementMethod: constant.SettlementTypeStandard,
				},
				PaymentMethod: paymentModel.PaymentMethod{
					Type:     paymentConstant.PAYMENT_METHOD_CREDIT_CARD,
					Acquirer: "VISA",
				},
				Metadata: &map[string]any{
					"feeDetail": &feeModel.FeeMetadataObject{
						DeductionType: constant.MerchantFeeDeductionTypeDirect,
						AmountType:    constant.MerchantFeeAmountType,
						Amount:        2_000.00,
						FinalAmount:   2_000.00,
					},
				},
			},
			status: constant.StatusSuccess,
			setupMock: func() {
				merchantRepo.On("GetSettlementConfig", constant.ValueCtxMockType(), mock.MatchedBy(func(p merchantModel.GetSettlementConfigRequest) bool {
					return p.Reference == constant.ReferencePaymentFundedPayout &&
						p.SettlementMethod == constant.SettlementTypeStandard &&
						*p.Channel == "VISA"
				})).Once().Return(&merchant.SettlementConfig{Type: "D+1"}, nil)

				orchestratorSvc.On("PostAccountTransaction",
					constant.ValueCtxMockType(),
					constant.PtrCreateAccTransactionReqMockType(),
				).Twice().Return(nil)

				rmq.On("PublishWithDelay",
					constant.ValueCtxMockType(),
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Once().Return(nil)

				accountTransactionRepo.On("UpdateSettlementDetailByIDs",
					constant.ValueCtxMockType(),
					mock.Anything,
					mock.Anything,
				).Once().Return(nil)

			},
		},
		{
			name:    "SUCCESS: Sub-merchant without own settlement config falls back to parent config",
			wantErr: false,
			payment: &paymentModel.Payment{
				UUID:       uuid.NewString(),
				MerchantID: subMerchantID,
				PaymentMethod: paymentModel.PaymentMethod{
					Type: paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
				},
				Metadata: &map[string]any{
					"onBehalf": merchantModel.OnBehalfObject{ParentMerchantId: parentMerchantID},
				},
			},
			status: constant.StatusSuccess,
			setupMock: func() {
				// sub-merchant has no own settlement config -> fall back to parent
				merchantRepo.On("GetSettlementConfig", constant.ValueCtxMockType(), mock.MatchedBy(func(p merchantModel.GetSettlementConfigRequest) bool {
					return p.MerchantId == subMerchantID
				})).Once().Return(nil, nil)
				// parent fallback returns its config
				merchantRepo.On("GetSettlementConfig", constant.ValueCtxMockType(), mock.MatchedBy(func(p merchantModel.GetSettlementConfigRequest) bool {
					return p.MerchantId == parentMerchantID
				})).Once().Return(&merchant.SettlementConfig{Type: constant.SettlementTypeInstant}, nil)

				orchestratorSvc.On("PostAccountTransaction",
					constant.ValueCtxMockType(),
					constant.PtrCreateAccTransactionReqMockType(),
				).Twice().Return(nil)
			},
		},
		{
			name:    "SUCCESS: Sub-merchant with no parent settlement config keeps default",
			wantErr: false,
			payment: &paymentModel.Payment{
				UUID:       uuid.NewString(),
				MerchantID: subMerchantID,
				PaymentMethod: paymentModel.PaymentMethod{
					Type: paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
				},
				Metadata: &map[string]any{
					"onBehalf": merchantModel.OnBehalfObject{ParentMerchantId: parentMerchantID},
				},
			},
			status: constant.StatusSuccess,
			setupMock: func() {
				// neither sub-merchant nor parent has a config -> default is kept
				merchantRepo.On("GetSettlementConfig", constant.ValueCtxMockType(), mock.MatchedBy(func(p merchantModel.GetSettlementConfigRequest) bool {
					return p.MerchantId == subMerchantID
				})).Once().Return(nil, nil)
				merchantRepo.On("GetSettlementConfig", constant.ValueCtxMockType(), mock.MatchedBy(func(p merchantModel.GetSettlementConfigRequest) bool {
					return p.MerchantId == parentMerchantID
				})).Once().Return(nil, nil)

				orchestratorSvc.On("PostAccountTransaction",
					constant.ValueCtxMockType(),
					constant.PtrCreateAccTransactionReqMockType(),
				).Twice().Return(nil)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMock()

			cfg := &config.Config{
				PaymentSettlementConfig: config.PaymentSettlementConfig{
					CreditCard: map[string]config.SettlementConfigDetail{
						"other_channel": {
							Type:    "T+5",
							DayType: "ANYDAY",
						},
					},
					VirtualAccount: map[string]config.SettlementConfigDetail{
						"other_channel": {
							Type:    "INSTANT",
							DayType: "ANYDAY",
						},
					},
					Qris: map[string]config.SettlementConfigDetail{
						"other_channel": {
							Type:    "T+1",
							DayType: "ANYDAY",
						},
					},
				},
			}

			paymentSvc := New(paymentRepo, logger, nil, nil, merchantRepo, nil, nil,
				WithOrchestratorService(orchestratorSvc),
				WithRabbitMQClient(rmq),
				WithFeeService(feeSvc),
				WithConfig(cfg),
				WithPaymentMethodService(paymentMethodSvc),
				WithAccountTransactionRepository(accountTransactionRepo),
			)
			WithSettlementHoldService(paymentSvc, settlementHoldSvc)
			ctx := context.WithValue(context.Background(), constant.CtxParentMerchantId, uuid.NewString())
			err := paymentSvc.PostCreateLedger(ctx, tc.payment, &paymentModel.PostCreateLedgerRequest{
				Status:  tc.status,
				Channel: constant.ChannelVirtualAccount,
				Amount: commonModel.Amount{
					Currency: "IDR",
					Value:    "10000.00",
				},
			})

			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			rmq.AssertExpectations(t)
			merchantRepo.AssertExpectations(t)
			orchestratorSvc.AssertExpectations(t)
		})
	}
}

func TestUpdatePendingLedger(t *testing.T) {

	logger := pdkLoggerMock.NewILogger(t)
	feeService := serviceMock.NewIFeeService(t)
	paymentRepo := repositoryMocks.NewIPaymentRepository(t)
	merchantRepo := repositoryMocks.NewIMerchantRepository(t)
	orchestratorService := serviceMock.NewIOrchestratorService(t)
	accountTrxRepo := repositoryMocks.NewIAccountTransactionRepository(t)

	service := New(
		paymentRepo, logger, nil, nil, merchantRepo, nil, nil,
		WithFeeService(feeService),
		WithAccountTransactionRepository(accountTrxRepo),
		WithOrchestratorService(orchestratorService),
		WithConfig(&config.Config{}),
	)

	normalPayment := &paymentModel.Payment{
		Metadata: &map[string]any{},
	}
	recurringPayment := &paymentModel.Payment{
		Metadata:    &map[string]any{},
		ReferenceID: util.ValueToPtr("123401d1-3d21-407c-9c8e-f43e8f503f5f"),
		RecurringPayment: &unifiedPaymentModel.MetadataRecurringPayment{
			InitiateFirstAuthorization: true,
			FirstAuthorizationMethod:   constant.RecurringContractAuthMethodOneDollar,
		},
	}

	tests := []struct {
		name       string
		payment    *paymentModel.Payment
		status     string
		setupMocks func()
		wantErr    error
	}{
		{
			name:   "ERROR:Get merchant fee",
			status: constant.StatusSuccess,
			payment: &paymentModel.Payment{
				UUID:       "85c84247-cdba-42ab-9aa6-4c263e43e5b2",
				MerchantID: "c4811ee2-9b12-40d6-8fa6-aa4ff69ca7b7",
				PaymentMethod: paymentModel.PaymentMethod{
					Type: paymentConstant.PAYMENT_METHOD_CREDIT_CARD,
				},
				Type: constant.TypeVirtualTerminal,
			},
			setupMocks: func() {
				merchantRepo.On(
					"GetSettlementConfig", mock.Anything, mock.MatchedBy(func(p merchantModel.GetSettlementConfigRequest) bool {
						return p.Reference == constant.ReferencePayment &&
							*p.Method == constant.TypeVirtualTerminal
					}),
				).Once().Return(nil, constant.ErrSomeErrorForUnitTest)
				logger.On(
					"Error", constant.ValueCtxMockType(), "Update pending ledger - failed to get merchant settlement config", mock.Anything,
				).Once().Return()
			},
			wantErr: pkgErr.New(response.HttpErrDatabase, constant.ErrSomeErrorForUnitTest),
		},
		{
			name:   "ERROR:Payment transaction not found",
			status: constant.StatusSuccess,
			payment: &paymentModel.Payment{
				UUID:       "85c84247-cdba-42ab-9aa6-4c263e43e5b2",
				MerchantID: "c4811ee2-9b12-40d6-8fa6-aa4ff69ca7b7",
				PaymentMethod: paymentModel.PaymentMethod{
					Type: paymentConstant.PAYMENT_METHOD_CREDIT_CARD,
				},
				Type: constant.UnifiedPaymentTypeSingle,
			},
			setupMocks: func() {
				merchantRepo.On(
					"GetSettlementConfig", mock.Anything, mock.MatchedBy(func(p merchantModel.GetSettlementConfigRequest) bool {
						return p.Reference == constant.ReferencePayment &&
							*p.Method == paymentConstant.PAYMENT_METHOD_CREDIT_CARD
					}),
				).Once().Return(&merchant.SettlementConfig{
					Type: constant.SettlementTypeInstant,
				}, nil)

				accountTrxRepo.On(
					"UpdatePaymentTransactionStatusAndMetadataByID", constant.ValueCtxMockType(), mock.Anything, mock.Anything,
				).Once().Return(constant.ErrDataNotFound)

			},
			wantErr: pkgErr.New(response.HttpErrUnprocessableContent, constant.ErrDataNotFound),
		},
		{
			name:   "ERROR:Update payment transaction status and metadata",
			status: constant.StatusSuccess,
			setupMocks: func() {
				merchantRepo.On("GetSettlementConfig", mock.Anything, mock.Anything).Once().Return(&merchant.SettlementConfig{
					Type: constant.SettlementTypeInstant,
				}, nil)
				accountTrxRepo.On(
					"UpdatePaymentTransactionStatusAndMetadataByID", constant.ValueCtxMockType(), mock.Anything, mock.Anything,
				).Once().Return(constant.ErrSomeErrorForUnitTest)
				logger.On(
					"Error", constant.ValueCtxMockType(), "Update pending ledger - failed update payment transaction status and metadata", mock.Anything,
				).Once().Return()
			},
			wantErr: pkgErr.New(response.HttpErrDatabase, constant.ErrSomeErrorForUnitTest),
		},
		{
			name:   "SUCCESS:Update failed status",
			status: constant.StatusFailed,
			setupMocks: func() {
				merchantRepo.On("GetSettlementConfig", mock.Anything, mock.Anything).Once().Return(&merchant.SettlementConfig{
					Type: constant.SettlementTypeInstant,
				}, nil)
				accountTrxRepo.On(
					"UpdatePaymentTransactionStatusAndMetadataByID", constant.ValueCtxMockType(), mock.Anything, mock.Anything,
				).Return(nil)
			},
		},
		{
			name:    "SUCCESS:Is first authorization method one dollar",
			payment: recurringPayment,
			status:  constant.StatusSuccess,
			setupMocks: func() {
				merchantRepo.On("GetSettlementConfig", mock.Anything, mock.Anything).Once().Return(&merchant.SettlementConfig{
					Type: constant.SettlementTypeInstant,
				}, nil)
			},
		},
		{
			name:   "ERROR:Failed to get find feeTrx by reference id",
			status: constant.StatusSuccess,
			setupMocks: func() {
				merchantRepo.On("GetSettlementConfig", mock.Anything, mock.Anything).Once().Return(&merchant.SettlementConfig{
					Type: constant.SettlementTypeInstant,
				}, nil)
				orchestratorService.On(
					"FindByReference", constant.ValueCtxMockType(), mock.Anything, mock.Anything,
				).Once().Return(nil, constant.ErrSomeErrorForUnitTest)
				logger.On(
					"Error", constant.ValueCtxMockType(), "Update pending ledger - failed to get fee transaction", mock.Anything,
				).Once().Return()
			},
			wantErr: pkgErr.New(response.HttpErrDatabase, constant.ErrSomeErrorForUnitTest),
		},
		{
			name:   "ERROR:Post create fee transaction",
			status: constant.StatusSuccess,
			setupMocks: func() {
				merchantRepo.On("GetSettlementConfig", mock.Anything, mock.Anything).Once().Return(&merchant.SettlementConfig{
					Type: constant.SettlementTypeInstant,
				}, nil)
				orchestratorService.On(
					"FindByReference", constant.ValueCtxMockType(), mock.Anything, mock.Anything,
				).Return(nil, nil)

				feeService.On(
					"CalculateFee", constant.ValueCtxMockType(), mock.Anything, mock.Anything,
				).Return(0.0, 0.0)

				orchestratorService.On(
					"PostAccountTransaction", constant.ValueCtxMockType(), mock.Anything,
				).Once().Return(constant.ErrSomeErrorForUnitTest)
				logger.On(
					"Error", constant.ValueCtxMockType(), "error when create account transaction for payment fee", mock.Anything,
				).Return()
			},
			wantErr: pkgErr.New(response.HttpErrInternal, constant.ErrSomeErrorForUnitTest),
		},
		{
			name:   "SUCCESS",
			status: constant.StatusSuccess,
			setupMocks: func() {
				merchantRepo.On("GetSettlementConfig", mock.Anything, mock.Anything).Once().Return(&merchant.SettlementConfig{
					Type: constant.SettlementTypeInstant,
				}, nil)
				orchestratorService.On("PostAccountTransaction", constant.ValueCtxMockType(), mock.Anything).Return(nil)
				paymentRepo.On("GetPaymentById", constant.ValueCtxMockType(), mock.Anything).Return(&paymentModel.Payment{}, nil)
				merchantRepo.On("FindMerchantByID", constant.ValueCtxMockType(), mock.Anything).Return(&merchant.Merchant{}, nil)
			},
		},
		{
			name:   "SUCCESS: Card Funded Payout with SettlementMethod",
			status: constant.StatusSuccess,
			payment: &paymentModel.Payment{
				UUID:       uuid.NewString(),
				MerchantID: uuid.NewString(),
				PaymentMethod: paymentModel.PaymentMethod{
					Type: paymentConstant.PAYMENT_METHOD_CREDIT_CARD,
				},
				Type: constant.PaymentTypeCardFundedPayout,
				CardFundedPayout: &unifiedPaymentModel.CardFundedPayout{
					SettlementMethod: constant.SettlementTypeStandard,
				},
			},
			setupMocks: func() { /* Empty Function */ },
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMocks()

			if test.payment == nil {
				test.payment = normalPayment
			}
			request := orchestrator_model.UpdatePaymentTransactionRequest{Status: test.status}

			assert.Equal(
				t, test.wantErr, service.UpdatePendingLedger(context.Background(), test.payment, request),
			)
		})
	}
}

func TestBindingPaymentMethodDetail(t *testing.T) {
	service := &PaymentService{}

	tests := []struct {
		name    string
		channel string
		payment *paymentModel.Payment
		result  any
	}{
		{
			name:    "Empty payment metadata",
			payment: &paymentModel.Payment{},
			result:  nil,
		},
		{
			name:    "QRIS:Empty snap core response",
			channel: constant.ChannelQris,
			payment: &paymentModel.Payment{
				Metadata: &map[string]any{},
			},
			result: nil,
		},
		{
			name:    "QRIS:Data found",
			channel: constant.ChannelQris,
			payment: &paymentModel.Payment{
				ReferenceID: util.ValueToPtr("1234"),
				Metadata: &map[string]any{
					"qrType":       "DYNAMIC",
					"qrMethodType": "MPM",
					"snapCore":     map[string]any{},
				},
			},
			result: orchestrator_model.MetadataPaymentMethodQRIS{
				PartnerReferenceNo: "1234",
				QrType:             "DYNAMIC",
				QrMethodType:       "MPM",
			},
		},
		{
			name:    "VA:Empty snap core response",
			channel: constant.ChannelVirtualAccount,
			payment: &paymentModel.Payment{
				Metadata: &map[string]any{},
			},
			result: nil,
		},
		{
			name:    "VA:Data found",
			channel: constant.ChannelVirtualAccount,
			payment: &paymentModel.Payment{
				Metadata: &map[string]any{},
			},
			result: nil,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.result, service.BindingPaymentMethodDetail(test.channel, test.payment))
		})
	}
}
