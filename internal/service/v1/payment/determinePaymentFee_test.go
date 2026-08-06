package paymentService_test

import (
	"context"
	"database/sql"
	"testing"

	paymentConstant "github.com/paper-indonesia/pivot-backoffice/constant/payment"
	orchestraModel "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	unifiedPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/unifiedPayment"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	feeModel "github.com/paper-indonesia/pivot-backoffice/internal/model/fee"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	. "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/payment"
	loggerMock "github.com/paper-indonesia/pivot-backoffice/mocks/pdk/logger"
	repoMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"

	"github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestDeterminePaymentFee(t *testing.T) {
	log := loggerMock.NewILogger(t)
	feeSvc := serviceMocks.NewIFeeService(t)
	paymentRepo := repoMocks.NewIPaymentRepository(t)
	orchestratorSvc := serviceMocks.NewIOrchestratorService(t)

	service := New(paymentRepo, log, nil, nil, nil, nil, nil, WithFeeService(feeSvc), WithOrchestratorService(orchestratorSvc))

	parentId := "3a90495f-2cc6-44e4-95a9-b43b58d8a03f"
	paymentOnBehalf := paymentModel.Payment{
		Metadata: &map[string]any{
			"onBehalf": map[string]any{
				"parentMerchantId": parentId,
			},
		},
	}
	paymentWithNormalCase := paymentModel.Payment{Metadata: &map[string]any{}}

	emptyCtx := context.Background()
	ctxWithParentId := context.WithValue(emptyCtx, constant.CtxParentMerchantId, parentId)

	feeDetail := &feeModel.FeeMetadataObject{}

	tests := []struct {
		name            string
		payment         paymentModel.Payment
		setupMock       func()
		wantErr         error
		wantFeeDetail   any
		wantFeeOnBehalf any
		wantCtx         context.Context
	}{
		{
			name: "ERROR:Get fee calculation details",
			payment: func() paymentModel.Payment {
				payment := paymentWithNormalCase
				payment.PaymentMethod = paymentModel.PaymentMethod{
					Type: paymentConstant.PAYMENT_METHOD_CREDIT_CARD,
				}
				payment.Type = constant.TypeVirtualTerminal
				return payment
			}(),
			setupMock: func() {
				orchestratorSvc.On(
					"FindByReference", constant.BackgroundCtxMockType(), constant.StringMockType(), constant.StringMockType(),
				).Return(&orchestraModel.AccountTransactionWithUseCase{SettlementModel: sql.NullString{Valid: true, String: constant.PaymentMethodChannelTypeFacilitator}}, nil)
				feeSvc.On(
					"GetFeeCalculationAndDetail", mock.Anything, mock.MatchedBy(func(p *feeModel.GetFeeRequest) bool {
						return p.Reference == constant.ReferencePayment &&
							p.PaymentMethod == constant.TypeVirtualTerminal
					}),
				).Once().Return(0.0, nil, assert.AnError)
				log.On(
					"Error", mock.Anything, "Failed while get fee calculation and detail", logger.Error(assert.AnError),
				).Once().Return()
			},
			wantErr:       assert.AnError,
			wantFeeDetail: nil,
			wantCtx:       emptyCtx,
		},
		{
			name: "ERROR:Update payment metadata by id",
			payment: paymentModel.Payment{
				PaymentMethod: paymentModel.PaymentMethod{
					Type: paymentConstant.PAYMENT_METHOD_CREDIT_CARD,
				},
				Type: constant.UnifiedPaymentTypeSingle,
			},
			setupMock: func() {
				feeSvc.On(
					"GetFeeCalculationAndDetail", mock.Anything, mock.MatchedBy(func(p *feeModel.GetFeeRequest) bool {
						return p.Reference == constant.ReferencePayment &&
							p.PaymentMethod == paymentConstant.PAYMENT_METHOD_CREDIT_CARD
					}),
				).Once().Return(0.0, feeDetail, nil)
				feeSvc.On("IncrementLadderCounter", mock.Anything, feeDetail.LadderCounterKey, feeDetail.LadderCounterIncrement).Once()
				paymentRepo.On(
					"UpdatePaymentMetadataById", mock.Anything, mock.Anything, mock.Anything,
				).Once().Return(assert.AnError)
			},
			wantErr:       assert.AnError,
			wantFeeDetail: feeDetail,
			wantCtx:       emptyCtx,
		},
		{
			name:    "SUCCESS",
			payment: paymentOnBehalf,
			setupMock: func() {
				feeSvc.On(
					"GetFeeCalculationAndDetail", mock.Anything, mock.Anything,
				).Return(0.0, feeDetail, nil).Once()
				feeSvc.On("IncrementLadderCounter", mock.Anything, feeDetail.LadderCounterKey, feeDetail.LadderCounterIncrement).Once()
				paymentRepo.On(
					"UpdatePaymentMetadataById", mock.Anything, mock.Anything, mock.Anything,
				).Once().Return(nil)
			},
			wantFeeDetail: feeDetail,
			wantCtx:       ctxWithParentId,
		},
		{
			name: "SUCCESS:Is first authorization method one dollar",
			payment: paymentModel.Payment{
				ReferenceID: util.ValueToPtr("123401d1-3d21-407c-9c8e-f43e8f503f5f"),
				RecurringPayment: &unifiedPaymentModel.MetadataRecurringPayment{
					InitiateFirstAuthorization: true,
					FirstAuthorizationMethod:   constant.RecurringContractAuthMethodOneDollar,
				},
				Metadata: &map[string]any{},
			},
			setupMock: func() { /* Empty Function */ },
			wantCtx:   emptyCtx,
		},
		{
			name: "SUCCESS:Is card-funded payout",
			payment: paymentModel.Payment{
				CardFundedPayout: &unifiedPaymentModel.CardFundedPayout{
					FeeConfig: feeModel.FeeMetadataObject{Amount: 2_000, FinalAmount: 2_000},
				},
				Metadata: &map[string]any{},
			},
			setupMock:     func() { /* Empty Function */ },
			wantCtx:       emptyCtx,
			wantFeeDetail: &feeModel.FeeMetadataObject{Amount: 2_000, FinalAmount: 2_000},
		},
		{
			name: "SUCCESS:Auto split sub-payment (first payment) is fee exempt",
			payment: paymentModel.Payment{
				AutoSplitPayment: &unifiedPaymentModel.AutoSplitPayment{
					TransactionType: constant.AutoSplitPaymentTypeFirstPayment,
				},
				Metadata: &map[string]any{},
			},
			setupMock:     func() { /* Empty Function */ },
			wantCtx:       emptyCtx,
			wantFeeDetail: nil,
		},
		{
			name: "SUCCESS:Auto split sub-payment (subsequent payment) is fee exempt",
			payment: paymentModel.Payment{
				AutoSplitPayment: &unifiedPaymentModel.AutoSplitPayment{
					TransactionType: constant.AutoSplitPaymentTypeSubsequentPayment,
				},
				Metadata: &map[string]any{},
			},
			setupMock:     func() { /* Empty Function */ },
			wantCtx:       emptyCtx,
			wantFeeDetail: nil,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			test.setupMock()

			ctx := context.Background()
			err := service.DeterminePaymentFee(&ctx, &test.payment)

			assert.Equal(t, test.wantErr, err)
			assert.Equal(t, test.wantFeeDetail, (*test.payment.Metadata)["feeDetail"])
			assert.Equal(t, test.wantFeeOnBehalf, (*test.payment.Metadata)["feeOnBehalf"])
			assert.Equal(t, test.wantCtx, ctx)
		})
	}
}
