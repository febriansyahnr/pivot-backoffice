package accounttransaction_repository

import (
	"context"
	"database/sql"
	"testing"
	"time"

	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	orchestratorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/orchestrator"
	"github.com/paper-indonesia/pivot-backoffice/internal/repository"
	mysqlMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestGetList(t *testing.T) {
	list := &[]*orchestratorModel.AccountTransactionWithUseCase{
		{
			UUID: uuid.New(),
		},
		{
			UUID: uuid.New(),
		},
	}

	testCase := []struct {
		name           string
		ctx            context.Context
		mockSetup      func(mysqlMock *mysqlMocks.IMySqlExt)
		filter         *orchestratorModel.TransactionHistoryFilterRequest
		wantErr        bool
		withAppConfig  bool
		initialPageWin int64
	}{
		{
			name: "SUCCESS:Get List without any filter and for balance history open api",
			ctx:  context.WithValue(context.Background(), constant.CtxFeatureName, constant.FeatureBalanceHistoryOpenApi),
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					constant.ValueCtxMockType(), mock.Anything, constant.StringMockType(),
				).Return(nil).Run(func(args mock.Arguments) {
					dataPtr := args.Get(1).(*[]*orchestratorModel.AccountTransactionWithUseCase)
					*dataPtr = *list
				})

				mysqlMock.On(
					"GetContext",
					constant.ValueCtxMockType(), mock.AnythingOfType(constant.MockTypeInt64Reference), constant.StringMockType(),
				).Return(nil)
			},
			filter:  &orchestratorModel.TransactionHistoryFilterRequest{},
			wantErr: false,
		},
		{
			name: "SUCCESS:Get List without any filter",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					constant.ValueCtxMockType(), mock.Anything, constant.StringMockType(),
				).Return(nil).Run(func(args mock.Arguments) {
					dataPtr := args.Get(1).(*[]*orchestratorModel.AccountTransactionWithUseCase)
					*dataPtr = *list
				})

				mysqlMock.On(
					"GetContext",
					constant.ValueCtxMockType(), mock.AnythingOfType(constant.MockTypeInt64Reference), constant.StringMockType(),
				).Return(nil)
			},
			filter:  &orchestratorModel.TransactionHistoryFilterRequest{},
			wantErr: false,
		},
		{
			name: "SUCCESS:Get List without any filter and total items is zero",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					constant.ValueCtxMockType(), mock.Anything, constant.StringMockType(),
				).Return(nil)

				mysqlMock.On(
					"GetContext",
					constant.ValueCtxMockType(), mock.AnythingOfType(constant.MockTypeInt64Reference), constant.StringMockType(),
				).Return(constant.ErrNoRowsData)
			},
			filter:  &orchestratorModel.TransactionHistoryFilterRequest{},
			wantErr: false,
		},
		{
			name: "SUCCESS:Get List with filter [Bulk Disbursement]",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					constant.ValueCtxMockType(), mock.Anything, constant.StringMockType(), constant.StringMockType(), constant.StringMockType(), constant.StringMockType(),
				).Return(nil)

				mysqlMock.On(
					"GetContext",
					constant.ValueCtxMockType(), mock.AnythingOfType(constant.MockTypeInt64Reference), constant.StringMockType(), constant.StringMockType(), constant.StringMockType(), constant.StringMockType()).Return(constant.ErrNoRowsData)
			},
			filter: &orchestratorModel.TransactionHistoryFilterRequest{
				MerchantID: uuid.NewString(),
				StartDate:  util.TimeNow,
				EndDate:    util.TimeNow,
				TrxID:      uuid.NewString(),
				TrxTypes:   []string{"BULK_DISBURSEMENT"},
				Status:     constant.StatusPending,
			},
			wantErr: false,
		},
		{
			name: "SUCCESS:Get List with filter [Payout Withdrawal]",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					constant.ValueCtxMockType(), mock.Anything, constant.StringMockType(), constant.StringMockType(), constant.StringMockType(),
					constant.StringMockType(), constant.StringMockType(),
				).Return(nil)

				mysqlMock.On(
					"GetContext",
					constant.ValueCtxMockType(), mock.AnythingOfType(constant.MockTypeInt64Reference), constant.StringMockType(), constant.StringMockType(), constant.StringMockType(),
					constant.StringMockType(), constant.StringMockType()).Return(constant.ErrNoRowsData)
			},
			filter: &orchestratorModel.TransactionHistoryFilterRequest{
				MerchantID: uuid.NewString(),
				StartDate:  util.TimeNow,
				EndDate:    util.TimeNow,
				TrxID:      uuid.NewString(),
				TrxTypes:   []string{"DISBURSEMENT_WITHDRAWAL"},
				Status:     constant.StatusSuccess,
			},
			wantErr: false,
		},
		{
			name: "SUCCESS:Get List with filter [Disbursement]",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					constant.ValueCtxMockType(), mock.Anything, constant.StringMockType(), constant.StringMockType(), constant.StringMockType(), constant.StringMockType(),
				).Return(nil)

				mysqlMock.On(
					"GetContext",
					constant.ValueCtxMockType(), mock.AnythingOfType(constant.MockTypeInt64Reference), constant.StringMockType(), constant.StringMockType(), constant.StringMockType(), constant.StringMockType(),
				).Return(constant.ErrNoRowsData)
			},
			filter: &orchestratorModel.TransactionHistoryFilterRequest{
				MerchantID: uuid.NewString(),
				StartDate:  util.TimeNow,
				EndDate:    util.TimeNow,
				TrxID:      uuid.NewString(),
				TrxTypes:   []string{"DISBURSEMENT"},
				Status:     constant.StatusPending,
			},
			wantErr: false,
		},
		{
			name: "SUCCESS:Get List with filter [Transaction ID]",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					constant.ValueCtxMockType(), mock.Anything, constant.StringMockType(), constant.StringMockType(), constant.StringMockType(), constant.StringMockType(),
					constant.StringMockType(), constant.StringMockType(), constant.StringMockType(), constant.StringMockType(),
					constant.StringMockType(), constant.StringMockType(), // Extra params for TransactionId
				).Return(nil)

				mysqlMock.On(
					"GetContext",
					constant.ValueCtxMockType(), mock.AnythingOfType(constant.MockTypeInt64Reference), constant.StringMockType(), constant.StringMockType(),
					constant.StringMockType(), constant.StringMockType(), constant.StringMockType(), constant.StringMockType(),
					constant.StringMockType(), constant.StringMockType(), constant.StringMockType(), constant.StringMockType(), // Extra params for TransactionId
				).Return(constant.ErrNoRowsData)
			},
			filter: &orchestratorModel.TransactionHistoryFilterRequest{
				MerchantID:    uuid.NewString(),
				StartDate:     util.TimeNow,
				EndDate:       util.TimeNow,
				TrxID:         uuid.NewString(),
				TrxTypes:      []string{"BULK_DISBURSEMENT"},
				Status:        constant.StatusPending,
				TransactionId: uuid.NewString(),
			},
			wantErr: false,
		},
		{
			name: "SUCCESS:Get List with filter [Manual Top Up]",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					constant.ValueCtxMockType(),
					mock.Anything,
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(nil)

				mysqlMock.On(
					"GetContext",
					constant.ValueCtxMockType(),
					mock.AnythingOfType(constant.MockTypeInt64Reference),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(constant.ErrNoRowsData)
			},
			filter: &orchestratorModel.TransactionHistoryFilterRequest{
				MerchantID: uuid.NewString(),
				StartDate:  util.TimeNow,
				EndDate:    util.TimeNow,
				TrxID:      uuid.NewString(),
				TrxTypes:   []string{"MANUAL_TOP_UP"},
				Status:     constant.StatusPending,
			},
			wantErr: false,
		},
		{
			name: "SUCCESS:Get List with filter [Balance Adjustment]",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					constant.ValueCtxMockType(),
					mock.Anything,
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(nil)

				mysqlMock.On(
					"GetContext",
					constant.ValueCtxMockType(),
					mock.AnythingOfType(constant.MockTypeInt64Reference),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(constant.ErrNoRowsData)
			},
			filter: &orchestratorModel.TransactionHistoryFilterRequest{
				MerchantID: uuid.NewString(),
				StartDate:  util.TimeNow,
				EndDate:    util.TimeNow,
				TrxID:      uuid.NewString(),
				TrxTypes:   []string{"BALANCE_ADJUSTMENT"},
				Status:     constant.StatusPending,
			},
			wantErr: false,
		},
		{
			name: "SUCCESS:Get List with filter [Balance Type]",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					constant.ValueCtxMockType(),
					mock.Anything,
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(nil)

				mysqlMock.On(
					"GetContext",
					constant.ValueCtxMockType(),
					mock.AnythingOfType(constant.MockTypeInt64Reference),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(constant.ErrNoRowsData)
			},
			filter: &orchestratorModel.TransactionHistoryFilterRequest{
				MerchantID:   uuid.NewString(),
				StartDate:    util.TimeNow,
				EndDate:      util.TimeNow,
				TrxID:        uuid.NewString(),
				BalanceTypes: []string{"Payout Balance"},
				Status:       constant.StatusPending,
			},
			wantErr: false,
		},
		{
			name: "SUCCESS:Get List with filter [Merchant Reference ID]",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					constant.ValueCtxMockType(),
					mock.Anything,
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(nil)

				mysqlMock.On(
					"GetContext",
					constant.ValueCtxMockType(),
					mock.AnythingOfType(constant.MockTypeInt64Reference),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(constant.ErrNoRowsData)
			},
			filter: &orchestratorModel.TransactionHistoryFilterRequest{
				MerchantID:          uuid.NewString(),
				StartDate:           util.TimeNow,
				EndDate:             util.TimeNow,
				BalanceTypes:        []string{"Payout Balance"},
				MerchantReferenceID: "Ref123",
				Status:              constant.StatusPending,
			},
			wantErr: false,
		},
		{
			name: "SUCCESS:Get List with settlement date filter",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					constant.ValueCtxMockType(),
					mock.Anything,
					constant.StringMockType(),
					constant.StringMockType(),
					constant.TimeMockType(),
					constant.TimeMockType(),
					constant.TimeMockType(),
					constant.TimeMockType(),
					constant.TimeMockType(),
					constant.TimeMockType(),
					constant.TimeMockType(),
					constant.TimeMockType(),
				).Return(nil)

				mysqlMock.On(
					"GetContext",
					constant.ValueCtxMockType(),
					mock.AnythingOfType(constant.MockTypeInt64Reference),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.TimeMockType(),
					constant.TimeMockType(),
					constant.TimeMockType(),
					constant.TimeMockType(),
					constant.TimeMockType(),
					constant.TimeMockType(),
					constant.TimeMockType(),
					constant.TimeMockType(),
				).Return(constant.ErrNoRowsData)
			},
			filter: &orchestratorModel.TransactionHistoryFilterRequest{
				MerchantID:          uuid.NewString(),
				StartSettlementDate: time.Now().AddDate(0, -1, 0), // 1 month ago
				EndSettlementDate:   time.Now(),                   // now
				Status:              constant.StatusPending,
			},
			wantErr: false,
		},
		{
			name: "SUCCESS:Get List with filter [VA_PAYMENT]",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					constant.ValueCtxMockType(),
					mock.Anything,
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(nil)

				mysqlMock.On(
					"GetContext",
					constant.ValueCtxMockType(),
					mock.AnythingOfType(constant.MockTypeInt64Reference),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(constant.ErrNoRowsData)
			},
			filter: &orchestratorModel.TransactionHistoryFilterRequest{
				MerchantID: uuid.NewString(),
				StartDate:  util.TimeNow,
				EndDate:    util.TimeNow,
				TrxID:      uuid.NewString(),
				TrxTypes:   []string{"VA_PAYMENT"},
				Status:     constant.StatusPending,
			},
			wantErr: false,
		},
		{
			name: "SUCCESS:Get List with filter [QRIS_PAYMENT]",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					constant.ValueCtxMockType(),
					mock.Anything,
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(nil)

				mysqlMock.On(
					"GetContext",
					constant.ValueCtxMockType(),
					mock.AnythingOfType(constant.MockTypeInt64Reference),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(constant.ErrNoRowsData)
			},
			filter: &orchestratorModel.TransactionHistoryFilterRequest{
				MerchantID: uuid.NewString(),
				StartDate:  util.TimeNow,
				EndDate:    util.TimeNow,
				TrxID:      uuid.NewString(),
				TrxTypes:   []string{"QRIS_PAYMENT"},
				Status:     constant.StatusPending,
			},
			wantErr: false,
		},
		{
			name: "SUCCESS:Get List with filter [CARD_PAYMENT]",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					constant.ValueCtxMockType(),
					mock.Anything,
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(nil)

				mysqlMock.On(
					"GetContext",
					constant.ValueCtxMockType(),
					mock.AnythingOfType(constant.MockTypeInt64Reference),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(constant.ErrNoRowsData)
			},
			filter: &orchestratorModel.TransactionHistoryFilterRequest{
				MerchantID: uuid.NewString(),
				StartDate:  util.TimeNow,
				EndDate:    util.TimeNow,
				TrxID:      uuid.NewString(),
				TrxTypes:   []string{"CARD_PAYMENT"},
				Status:     constant.StatusPending,
			},
			wantErr: false,
		},
		{
			name: "SUCCESS:Get List with filter [Fee Type]",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					constant.ValueCtxMockType(),
					mock.Anything,
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(nil)

				mysqlMock.On(
					"GetContext",
					constant.ValueCtxMockType(),
					mock.AnythingOfType(constant.MockTypeInt64Reference),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(constant.ErrNoRowsData)
			},
			filter: &orchestratorModel.TransactionHistoryFilterRequest{
				MerchantID: uuid.NewString(),
				StartDate:  util.TimeNow,
				EndDate:    util.TimeNow,
				TrxID:      uuid.NewString(),
				TrxTypes:   []string{"PAYMENT_FEE"},
				Status:     constant.StatusPending,
			},
			wantErr: false,
		},
		{
			name: "SUCCESS:Get List with filter [Other Transaction Type]",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					constant.ValueCtxMockType(),
					mock.Anything,
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(nil)

				mysqlMock.On(
					"GetContext",
					constant.ValueCtxMockType(),
					mock.AnythingOfType(constant.MockTypeInt64Reference),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(constant.ErrNoRowsData)
			},
			filter: &orchestratorModel.TransactionHistoryFilterRequest{
				MerchantID: uuid.NewString(),
				StartDate:  util.TimeNow,
				EndDate:    util.TimeNow,
				TrxID:      uuid.NewString(),
				TrxTypes:   []string{"CUSTOM_TYPE"}, // This will hit the default else branch
				Status:     constant.StatusPending,
			},
			wantErr: false,
		},
		{
			name: "SUCCESS:Get List with CreatedAt filter",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					constant.ValueCtxMockType(),
					mock.Anything,
					constant.StringMockType(),
					constant.StringMockType(),
					constant.TimeMockType(),
					constant.TimeMockType(),
					constant.StringMockType(),
				).Return(nil)

				mysqlMock.On(
					"GetContext",
					constant.ValueCtxMockType(),
					mock.AnythingOfType(constant.MockTypeInt64Reference),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.TimeMockType(),
					constant.TimeMockType(),
					constant.StringMockType(),
				).Return(constant.ErrNoRowsData)
			},
			filter: &orchestratorModel.TransactionHistoryFilterRequest{
				MerchantID:         uuid.NewString(),
				CreatedAtStartDate: time.Now().AddDate(0, -1, 0), // 1 month ago
				CreatedAtEndDate:   time.Now(),                   // now
				TrxID:              uuid.NewString(),
				Status:             constant.StatusPending,
			},
			wantErr: false,
		},
		{
			name: "SUCCESS:Get List with filter [INTERNATIONAL_PAYOUT]",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					constant.ValueCtxMockType(),
					mock.Anything,
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(nil)

				mysqlMock.On(
					"GetContext",
					constant.ValueCtxMockType(),
					mock.AnythingOfType(constant.MockTypeInt64Reference),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(constant.ErrNoRowsData)
			},
			filter: &orchestratorModel.TransactionHistoryFilterRequest{
				MerchantID: uuid.NewString(),
				StartDate:  util.TimeNow,
				EndDate:    util.TimeNow,
				TrxID:      uuid.NewString(),
				TrxTypes:   []string{"INTERNATIONAL_PAYOUT"},
				Status:     constant.StatusPending,
			},
			wantErr: false,
		},
		{
			name: "SUCCESS:Get List with filter [VA_TOP_UP]",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					constant.ValueCtxMockType(),
					mock.Anything,
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(nil)

				mysqlMock.On(
					"GetContext",
					constant.ValueCtxMockType(),
					mock.AnythingOfType(constant.MockTypeInt64Reference),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(constant.ErrNoRowsData)
			},
			filter: &orchestratorModel.TransactionHistoryFilterRequest{
				MerchantID: uuid.NewString(),
				StartDate:  util.TimeNow,
				EndDate:    util.TimeNow,
				TrxID:      uuid.NewString(),
				TrxTypes:   []string{"VA_TOP_UP"},
				Status:     constant.StatusPending,
			},
			wantErr: false,
		},
		{
			name: "SUCCESS:Get List with filter [CUSTOMER_TOP_UP]",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					constant.ValueCtxMockType(),
					mock.Anything,
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(nil)

				mysqlMock.On(
					"GetContext",
					constant.ValueCtxMockType(),
					mock.AnythingOfType(constant.MockTypeInt64Reference),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(constant.ErrNoRowsData)
			},
			filter: &orchestratorModel.TransactionHistoryFilterRequest{
				MerchantID: uuid.NewString(),
				StartDate:  util.TimeNow,
				EndDate:    util.TimeNow,
				TrxID:      uuid.NewString(),
				TrxTypes:   []string{"CUSTOMER_TOP_UP"},
				Status:     constant.StatusPending,
			},
			wantErr: false,
		},
		{
			name: "SUCCESS:Get List with SettlementModel [Aggregator]",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					constant.ValueCtxMockType(),
					mock.Anything,
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(nil)

				mysqlMock.On(
					"GetContext",
					constant.ValueCtxMockType(),
					mock.AnythingOfType(constant.MockTypeInt64Reference),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(constant.ErrNoRowsData)
			},
			filter: &orchestratorModel.TransactionHistoryFilterRequest{
				MerchantID:      uuid.NewString(),
				StartDate:       util.TimeNow,
				EndDate:         util.TimeNow,
				TrxID:           uuid.NewString(),
				SettlementModel: constant.PaymentMethodChannelTypeAggregator,
				Status:          constant.StatusPending,
			},
			wantErr: false,
		},
		{
			name: "SUCCESS:Get List with SettlementModel [Facilitator]",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					constant.ValueCtxMockType(),
					mock.Anything,
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(nil)

				mysqlMock.On(
					"GetContext",
					constant.ValueCtxMockType(),
					mock.AnythingOfType(constant.MockTypeInt64Reference),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(constant.ErrNoRowsData)
			},
			filter: &orchestratorModel.TransactionHistoryFilterRequest{
				MerchantID:      uuid.NewString(),
				StartDate:       util.TimeNow,
				EndDate:         util.TimeNow,
				TrxID:           uuid.NewString(),
				SettlementModel: constant.PaymentMethodChannelTypeFacilitator,
				Status:          constant.StatusPending,
			},
			wantErr: false,
		},
		{
			name: "SUCCESS:Get List with AppConfig set",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					constant.ValueCtxMockType(),
					mock.Anything,
					constant.StringMockType(),
				).Return(nil)

				mysqlMock.On(
					"GetContext",
					constant.ValueCtxMockType(),
					mock.AnythingOfType(constant.MockTypeInt64Reference),
					constant.StringMockType(),
				).Return(constant.ErrNoRowsData)
			},
			filter:         &orchestratorModel.TransactionHistoryFilterRequest{},
			wantErr:        false,
			withAppConfig:  true,
			initialPageWin: 100,
		},
		{
			name: "FAILED:Get List on error get table",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					constant.ValueCtxMockType(),
					mock.Anything,
					constant.StringMockType(),
				).Return(constant.ErrSomeErrorForUnitTest)

				mysqlMock.On(
					"GetContext",
					constant.ValueCtxMockType(),
					mock.AnythingOfType(constant.MockTypeInt64Reference),
					constant.StringMockType(),
				).Return(constant.ErrSomeErrorForUnitTest)

			},
			filter:  &orchestratorModel.TransactionHistoryFilterRequest{},
			wantErr: true,
		},
	}

	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			mockMysql := mysqlMocks.NewIMySqlExt(t)
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			tc.mockSetup(mockMysql)

			var repo repository.IAccountTransactionRepository
			if tc.withAppConfig {
				appConfig := &config.AppConfig{
					InitialPageWindow: tc.initialPageWin,
				}
				repo = New(mockMysql, mockLogger, WithAppConfig(appConfig))
			} else {
				repo = New(mockMysql, mockLogger)
			}

			if tc.ctx == nil {
				tc.ctx = context.Background()
			}
			_, err := repo.GetList(tc.ctx, tc.filter, 0, 20)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockMysql.AssertExpectations(t)

		})
	}
}

func TestGetListTransactionHistories(t *testing.T) {
	db := mysqlMocks.NewIMySqlExt(t)

	repo := New(db, nil)

	tests := []struct {
		name       string
		setupMock  func()
		wantErr    string
		wantResult []orchestratorModel.TransactionHistory
	}{
		{
			name: "ERROR:Data not found",
			setupMock: func() {
				db.On(
					"SelectContext", constant.ValueCtxMockType(), mock.AnythingOfType("*[]orchestrator_model.TransactionHistory"),
					constant.StringMockType(), constant.StringMockType(),
				).Once().Return(sql.ErrNoRows)
			},
		},
		{
			name: "ERROR:Some error",
			setupMock: func() {
				db.On(
					"SelectContext", constant.ValueCtxMockType(), mock.AnythingOfType("*[]orchestrator_model.TransactionHistory"),
					constant.StringMockType(), constant.StringMockType(),
				).Once().Return(constant.ErrSomeErrorForUnitTest)
			},
			wantErr: constant.ErrSomeErrorForUnitTest.Error(),
		},
		{
			name: "SUCCESS",
			setupMock: func() {
				db.On(
					"SelectContext", constant.ValueCtxMockType(), mock.AnythingOfType("*[]orchestrator_model.TransactionHistory"),
					constant.StringMockType(), constant.StringMockType(),
				).Run(func(args mock.Arguments) {
					*args.Get(1).(*[]orchestratorModel.TransactionHistory) = []orchestratorModel.TransactionHistory{
						{Id: "trx-id", Type: constant.TypeDisbursement},
					}
				}).Return(nil)
			},
			wantResult: []orchestratorModel.TransactionHistory{{Id: "trx-id", Type: constant.TypeDisbursement}},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			filter := &orchestratorModel.TransactionHistoryFilterRequest{
				MerchantID: uuid.NewString(),
				StartDate:  time.Date(2024, 7, 1, 0, 0, 0, 0, time.UTC),
				EndDate:    time.Date(2024, 7, 31, 23, 59, 59, 0, time.UTC),
			}
			if resp, err := repo.GetListTransactionHistories(context.Background(), filter); test.wantErr == "" {
				require.NoError(t, err)
				assert.Equal(t, test.wantResult, resp)

			} else {
				require.Error(t, err)
				assert.ErrorContains(t, err, test.wantErr)
			}
		})
	}
}
