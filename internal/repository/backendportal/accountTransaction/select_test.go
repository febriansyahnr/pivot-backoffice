package accounttransaction_repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"testing"
	"time"

	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	orchestrator_model "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/orchestrator"
	mysqlMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestGetDetailById(t *testing.T) {
	mysql := mysqlMocks.NewIMySqlExt(t)
	logger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

	repo := New(mysql, logger)

	ptrTrxDisbursementRespMockType := mock.AnythingOfType("*orchestrator_model.TransactionDisbursementResp")
	ptrTrxHistoryDetailRespMockType := mock.AnythingOfType("*orchestrator_model.TransactionHistoryDetailResp")
	ptrTrxWithdrawalRespMockType := mock.AnythingOfType("*orchestrator_model.TransactionWithdrawalResp")
	ptrTrxPaymentRespMockType := mock.AnythingOfType("*orchestrator_model.TransactionPaymentResp")
	ptrTrxTransferRespMockType := mock.AnythingOfType("*orchestrator_model.TransactionTransferResp")

	tests := []struct {
		name      string
		mockSetup func()
		wantErr   string
		wantNil   bool
	}{
		{
			name: "ERROR:Some error in query account_transactions",
			mockSetup: func() {
				mysql.On(
					"GetContext", constant.ValueCtxMockType(), ptrTrxHistoryDetailRespMockType, constant.StringMockType(), constant.StringMockType(), constant.StringMockType(),
				).Once().Return(fmt.Errorf("Q1: %v", constant.ErrSomeErrorForUnitTest))
			},
			wantErr: "Q1: some error",
		},
		{
			name: "ERROR:Data not found",
			mockSetup: func() {
				mysql.On(
					"GetContext", constant.ValueCtxMockType(), ptrTrxHistoryDetailRespMockType, constant.StringMockType(), constant.StringMockType(), constant.StringMockType(),
				).Once().Return(sql.ErrNoRows)
			},
			wantNil: true,
		},
		{
			name: "SUCCESS:Payment transaction",
			mockSetup: func() {
				mysql.On(
					"GetContext", constant.ValueCtxMockType(), ptrTrxHistoryDetailRespMockType, constant.StringMockType(), constant.StringMockType(), constant.StringMockType(),
				).Once().Run(func(args mock.Arguments) {
					dest := args.Get(1).(*orchestrator_model.TransactionHistoryDetailResp)
					dest.Type = "PAYMENT"
					dest.ReferenceId = ""
				}).Return(nil)
			},
		},
		{
			name: "ERROR:Some error in query disbursement",
			mockSetup: func() {
				mysql.On(
					"GetContext", constant.ValueCtxMockType(), ptrTrxHistoryDetailRespMockType, constant.StringMockType(), constant.StringMockType(), constant.StringMockType(),
				).Once().Run(func(args mock.Arguments) {
					dest := args.Get(1).(*orchestrator_model.TransactionHistoryDetailResp)
					dest.Type = "DISBURSEMENT"
					dest.ReferenceId = uuid.NewString()
				}).Return(nil)

				mysql.On(
					"GetContext", constant.ValueCtxMockType(), ptrTrxDisbursementRespMockType, constant.StringMockType(), constant.StringMockType(),
				).Once().Return(fmt.Errorf("Q2: %v", constant.ErrSomeErrorForUnitTest))
			},
			wantErr: "Q2: some error",
		},
		{
			name: "SUCCESS:Disbursement transaction",
			mockSetup: func() {
				mysql.On(
					"GetContext", constant.ValueCtxMockType(), ptrTrxHistoryDetailRespMockType, constant.StringMockType(), constant.StringMockType(), constant.StringMockType(),
				).Once().Run(func(args mock.Arguments) {
					dest := args.Get(1).(*orchestrator_model.TransactionHistoryDetailResp)
					dest.Type = "DISBURSEMENT"
					dest.ReferenceId = uuid.NewString()
				}).Return(nil)

				mysql.On(
					"GetContext", constant.ValueCtxMockType(), ptrTrxDisbursementRespMockType, constant.StringMockType(), constant.StringMockType(),
				).Once().Return(nil)
			},
		},
		{
			name: "SUCCESS:Withdrawal transaction",
			mockSetup: func() {
				refId := uuid.NewString()
				mysql.On(
					"GetContext", constant.ValueCtxMockType(), ptrTrxHistoryDetailRespMockType, constant.StringMockType(), constant.StringMockType(), constant.StringMockType(),
				).Once().Run(func(args mock.Arguments) {
					dest := args.Get(1).(*orchestrator_model.TransactionHistoryDetailResp)
					dest.Type = "TEST_WITHDRAWAL"
					dest.ReferenceId = refId
				}).Return(nil)

				mysql.On(
					"GetContext", constant.ValueCtxMockType(), ptrTrxWithdrawalRespMockType, constant.StringMockType(), constant.StringMockType(),
				).Once().Return(nil)
			},
		},
		{
			name: "ERROR:Some error in query withdrawal",
			mockSetup: func() {
				refId := uuid.NewString()
				mysql.On(
					"GetContext", constant.ValueCtxMockType(), ptrTrxHistoryDetailRespMockType, constant.StringMockType(), constant.StringMockType(), constant.StringMockType(),
				).Once().Run(func(args mock.Arguments) {
					dest := args.Get(1).(*orchestrator_model.TransactionHistoryDetailResp)
					dest.Type = "TEST_WITHDRAWAL"
					dest.ReferenceId = refId
				}).Return(nil)

				mysql.On(
					"GetContext", constant.ValueCtxMockType(), ptrTrxWithdrawalRespMockType, constant.StringMockType(), constant.StringMockType(),
				).Once().Return(fmt.Errorf("Q3: %v", constant.ErrSomeErrorForUnitTest))
			},
			wantErr: "Q3: some error",
		},
		{
			name: "SUCCESS:Fee transaction with linked transaction",
			mockSetup: func() {
				mysql.On(
					"GetContext", constant.ValueCtxMockType(), ptrTrxHistoryDetailRespMockType, constant.StringMockType(), constant.StringMockType(), constant.StringMockType(),
				).Once().Run(func(args mock.Arguments) {
					dest := args.Get(1).(*orchestrator_model.TransactionHistoryDetailResp)
					dest.Type = "FEE"
					dest.ReferenceId = uuid.NewString()
					dest.LinkedTransactionId = uuid.NewString()
				}).Return(nil)
			},
		},
		{
			name: "SUCCESS:Payment transaction with reference",
			mockSetup: func() {
				refId := uuid.NewString()
				mysql.On(
					"GetContext", constant.ValueCtxMockType(), ptrTrxHistoryDetailRespMockType, constant.StringMockType(), constant.StringMockType(), constant.StringMockType(),
				).Once().Run(func(args mock.Arguments) {
					dest := args.Get(1).(*orchestrator_model.TransactionHistoryDetailResp)
					dest.Type = "PAYMENT"
					dest.ReferenceId = refId
				}).Return(nil)

				mysql.On(
					"GetContext", constant.ValueCtxMockType(), ptrTrxPaymentRespMockType, constant.StringMockType(), constant.StringMockType(),
				).Once().Return(nil)
			},
		},
		{
			name: "ERROR:Some error in query payment",
			mockSetup: func() {
				refId := uuid.NewString()
				mysql.On(
					"GetContext", constant.ValueCtxMockType(), ptrTrxHistoryDetailRespMockType, constant.StringMockType(), constant.StringMockType(), constant.StringMockType(),
				).Once().Run(func(args mock.Arguments) {
					dest := args.Get(1).(*orchestrator_model.TransactionHistoryDetailResp)
					dest.Type = "PAYMENT"
					dest.ReferenceId = refId
				}).Return(nil)

				mysql.On(
					"GetContext", constant.ValueCtxMockType(), ptrTrxPaymentRespMockType, constant.StringMockType(), constant.StringMockType(),
				).Once().Return(fmt.Errorf("Q4: %v", constant.ErrSomeErrorForUnitTest))
			},
			wantErr: "Q4: some error",
		},
		{
			name: "SUCCESS:Transfer transaction",
			mockSetup: func() {
				refId := uuid.NewString()
				mysql.On(
					"GetContext", constant.ValueCtxMockType(), ptrTrxHistoryDetailRespMockType, constant.StringMockType(), constant.StringMockType(), constant.StringMockType(),
				).Once().Run(func(args mock.Arguments) {
					dest := args.Get(1).(*orchestrator_model.TransactionHistoryDetailResp)
					dest.Type = "TRANSFER"
					dest.ReferenceId = refId
					dest.MerchantID = uuid.NewString()
				}).Return(nil)

				mysql.On(
					"GetContext", constant.ValueCtxMockType(), ptrTrxTransferRespMockType, constant.StringMockType(), constant.StringMockType(),
				).Once().Return(nil)
			},
		},
		{
			name: "ERROR:Some error in query transfer",
			mockSetup: func() {
				refId := uuid.NewString()
				mysql.On(
					"GetContext", constant.ValueCtxMockType(), ptrTrxHistoryDetailRespMockType, constant.StringMockType(), constant.StringMockType(), constant.StringMockType(),
				).Once().Run(func(args mock.Arguments) {
					dest := args.Get(1).(*orchestrator_model.TransactionHistoryDetailResp)
					dest.Type = "TRANSFER"
					dest.ReferenceId = refId
					dest.MerchantID = uuid.NewString()
				}).Return(nil)

				mysql.On(
					"GetContext", constant.ValueCtxMockType(), ptrTrxTransferRespMockType, constant.StringMockType(), constant.StringMockType(),
				).Once().Return(fmt.Errorf("Q5: %v", constant.ErrSomeErrorForUnitTest))
			},
			wantErr: "Q5: some error",
		},
		{
			name: "SUCCESS:Data not found in disbursement detail query",
			mockSetup: func() {
				mysql.On(
					"GetContext", constant.ValueCtxMockType(), ptrTrxHistoryDetailRespMockType, constant.StringMockType(), constant.StringMockType(), constant.StringMockType(),
				).Once().Run(func(args mock.Arguments) {
					dest := args.Get(1).(*orchestrator_model.TransactionHistoryDetailResp)
					dest.Type = "DISBURSEMENT"
					dest.ReferenceId = uuid.NewString()
				}).Return(nil)

				mysql.On(
					"GetContext", constant.ValueCtxMockType(), ptrTrxDisbursementRespMockType, constant.StringMockType(), constant.StringMockType(),
				).Once().Return(sql.ErrNoRows)
			},
		},
		{
			name: "SUCCESS:Data not found in withdrawal detail query",
			mockSetup: func() {
				refId := uuid.NewString()
				mysql.On(
					"GetContext", constant.ValueCtxMockType(), ptrTrxHistoryDetailRespMockType, constant.StringMockType(), constant.StringMockType(), constant.StringMockType(),
				).Once().Run(func(args mock.Arguments) {
					dest := args.Get(1).(*orchestrator_model.TransactionHistoryDetailResp)
					dest.Type = "TEST_WITHDRAWAL"
					dest.ReferenceId = refId
				}).Return(nil)

				mysql.On(
					"GetContext", constant.ValueCtxMockType(), ptrTrxWithdrawalRespMockType, constant.StringMockType(), constant.StringMockType(),
				).Once().Return(sql.ErrNoRows)
			},
		},
		{
			name: "SUCCESS:Data not found in payment detail query",
			mockSetup: func() {
				refId := uuid.NewString()
				mysql.On(
					"GetContext", constant.ValueCtxMockType(), ptrTrxHistoryDetailRespMockType, constant.StringMockType(), constant.StringMockType(), constant.StringMockType(),
				).Once().Run(func(args mock.Arguments) {
					dest := args.Get(1).(*orchestrator_model.TransactionHistoryDetailResp)
					dest.Type = "PAYMENT"
					dest.ReferenceId = refId
				}).Return(nil)

				mysql.On(
					"GetContext", constant.ValueCtxMockType(), ptrTrxPaymentRespMockType, constant.StringMockType(), constant.StringMockType(),
				).Once().Return(sql.ErrNoRows)
			},
		},
		{
			name: "SUCCESS:Data not found in transfer detail query",
			mockSetup: func() {
				refId := uuid.NewString()
				mysql.On(
					"GetContext", constant.ValueCtxMockType(), ptrTrxHistoryDetailRespMockType, constant.StringMockType(), constant.StringMockType(), constant.StringMockType(),
				).Once().Run(func(args mock.Arguments) {
					dest := args.Get(1).(*orchestrator_model.TransactionHistoryDetailResp)
					dest.Type = "TRANSFER"
					dest.ReferenceId = refId
					dest.MerchantID = uuid.NewString()
				}).Return(nil)

				mysql.On(
					"GetContext", constant.ValueCtxMockType(), ptrTrxTransferRespMockType, constant.StringMockType(), constant.StringMockType(),
				).Once().Return(sql.ErrNoRows)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.mockSetup()

			merchantId := uuid.NewString()
			transactionId := uuid.NewString()

			result, err := repo.GetDetailById(context.Background(), merchantId, transactionId)
			if test.wantErr == "" {
				require.NoError(t, err)
				if !test.wantNil {
					assert.NotNil(t, result)
				} else {
					assert.Nil(t, result)
				}
			} else {
				require.Error(t, err)
				assert.ErrorContains(t, err, test.wantErr)
			}
		})
	}
}

func TestGetPlatformTransactionActivities(t *testing.T) {
	db := mysqlMocks.NewIMySqlExt(t)

	repo := New(db, nil)

	rawQuery := "SELECT ???"
	merchantId := uuid.NewString()
	startDate := time.Date(2024, 8, 31, 17, 0, 0, 0, time.UTC)
	endDate := time.Date(2024, 9, 9, 16, 59, 59, 0, time.UTC)
	transaction := []orchestrator_model.TransactionActivity{
		{MerchantID: merchantId, Total: 6},
	}
	ptrSliceTrxActivitiesMockType := mock.AnythingOfType("*[]orchestrator_model.TransactionActivity")

	tests := []struct {
		name       string
		setupMock  func()
		wantErr    error
		wantResult []orchestrator_model.TransactionActivity
	}{
		{
			name: "ERROR:Build IN query",
			setupMock: func() {
				db.On(
					"In", constant.StringMockType(), mock.AnythingOfType("[]string"), constant.TimeMockType(), constant.TimeMockType(),
				).Once().Return("", nil, constant.ErrSomeErrorForUnitTest)
			},
			wantErr: fmt.Errorf("failed generate in statement: %v", constant.ErrSomeErrorForUnitTest),
		},
		{
			name: "ERROR:Some error",
			setupMock: func() {
				db.On(
					"In", constant.StringMockType(), mock.AnythingOfType("[]string"), constant.TimeMockType(), constant.TimeMockType(),
				).Return(rawQuery, []interface{}{merchantId, startDate, endDate}, nil)
				db.On("Rebind", constant.StringMockType()).Return(rawQuery)

				db.On(
					"SelectContext", constant.ValueCtxMockType(), ptrSliceTrxActivitiesMockType, rawQuery, merchantId, startDate, endDate,
				).Once().Return(constant.ErrSomeErrorForUnitTest)
			},
			wantErr: constant.ErrSomeErrorForUnitTest,
		},
		{
			name: "ERROR:Data not found",
			setupMock: func() {
				db.On(
					"SelectContext", constant.ValueCtxMockType(), ptrSliceTrxActivitiesMockType, rawQuery, merchantId, startDate, endDate,
				).Once().Return(sql.ErrNoRows)
			},
		},
		{
			name: "SUCCESS",
			setupMock: func() {
				db.On(
					"SelectContext", constant.ValueCtxMockType(), ptrSliceTrxActivitiesMockType, rawQuery, merchantId, startDate, endDate,
				).Run(func(args mock.Arguments) {
					(*args.Get(1).(*[]orchestrator_model.TransactionActivity)) = transaction
				}).Return(nil)
			},
			wantResult: transaction,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			result, err := repo.GetPlatformTransactionActivities(context.Background(), []string{merchantId}, startDate, endDate)
			assert.Equal(t, test.wantErr, err)
			assert.Equal(t, test.wantResult, result)
		})
	}
}

func TestGetAccumulateTransactionFees(t *testing.T) {

	db := mysqlMocks.NewIMySqlExt(t)

	repo := New(db, nil)

	ptrResultDataMockType := mock.AnythingOfType("*orchestrator_model.AccumulateTransactionFees")

	var (
		reference  = constant.ReferenceDisbursement
		method     = ""
		merchantId = uuid.NewString()
		startDate  = time.Now().Add(-time.Second)
		endDate    = time.Now().Add(time.Second)
	)

	rawTrxIds := `["1", "2", "3", "4"]`
	transaction := &orchestrator_model.AccumulateTransactionFees{
		RawTransactionIds: []byte(rawTrxIds),
		TransactionIds:    []string{"1", "2", "3", "4"},
	}
	tests := []struct {
		name       string
		setupMock  func()
		wantErr    error
		wantResult *orchestrator_model.AccumulateTransactionFees
	}{
		{
			name: "ERROR:Some error", // NOSONAR
			setupMock: func() {
				db.On(
					"GetContext", constant.ValueCtxMockType(), ptrResultDataMockType, constant.StringMockType(), merchantId, startDate, endDate, reference, method,
				).Once().Return(constant.ErrSomeErrorForUnitTest)
			},
			wantErr: constant.ErrSomeErrorForUnitTest, // NOSONAR
		},
		{
			name: "SUCCESS",
			setupMock: func() {
				db.On(
					"GetContext", constant.ValueCtxMockType(), ptrResultDataMockType, constant.StringMockType(), merchantId, startDate, endDate, reference, method,
				).Run(func(args mock.Arguments) {
					(*args.Get(1).(*orchestrator_model.AccumulateTransactionFees)) = orchestrator_model.AccumulateTransactionFees{
						RawTransactionIds: []byte(rawTrxIds),
					}
				}).Return(nil)
			},
			wantResult: transaction,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			test.setupMock()

			result, err := repo.GetAccumulateTransactionFees(
				context.Background(), merchantId, reference, method, startDate, endDate,
			)
			assert.Equal(t, test.wantErr, err)
			assert.Equal(t, test.wantResult, result)
		})
	}
}

func TestCalculatingMerchantTPVToDetermineFeeTierLevel(t *testing.T) {
	db := mysqlMocks.NewIMySqlExt(t)

	repo := New(db, nil)

	tpvSummariesMockType := mock.AnythingOfType("*[]orchestrator_model.CalculatingMerchantTPVSummary")

	tests := []struct {
		name       string
		setupMock  func()
		wantErr    error
		wantResult map[string]orchestrator_model.CalculatingMerchantTPVSummary
	}{
		{
			name: "ERROR:Some error",
			setupMock: func() {
				db.On(
					"SelectContext", constant.ValueCtxMockType(), tpvSummariesMockType, constant.StringMockType(), constant.StringMockType(), constant.TimeMockType(), constant.TimeMockType(),
				).Once().Return(constant.ErrSomeErrorForUnitTest)
			},
			wantErr: constant.ErrSomeErrorForUnitTest,
		},
		{
			name: "ERROR:Data not found",
			setupMock: func() {
				db.On(
					"SelectContext", constant.ValueCtxMockType(), tpvSummariesMockType, constant.StringMockType(), constant.StringMockType(), constant.TimeMockType(), constant.TimeMockType(),
				).Once().Return(sql.ErrNoRows)
			},
		},
		{
			name: "SUCCESS",
			setupMock: func() {
				db.On(
					"SelectContext", constant.ValueCtxMockType(), tpvSummariesMockType, constant.StringMockType(), constant.StringMockType(), constant.TimeMockType(), constant.TimeMockType(),
				).Run(func(args mock.Arguments) {
					(*args.Get(1).(*[]orchestrator_model.CalculatingMerchantTPVSummary)) = []orchestrator_model.CalculatingMerchantTPVSummary{
						{
							Type:      constant.TypePayment,
							Channel:   constant.ChannelQris,
							Frequency: 1, Volume: 1_000_000,
						},
						{
							Type:       constant.TypeFee,
							Additional: util.ValueToPtr(constant.ReferenceAccountInquiry),
							Frequency:  10, Volume: 1_000,
						},
					}
				}).Return(nil)
			},
			wantResult: map[string]orchestrator_model.CalculatingMerchantTPVSummary{
				"PAYMENT_QRIS": {
					Type:      constant.TypePayment,
					Channel:   constant.ChannelQris,
					Frequency: 1, Volume: 1_000_000,
				},
				"ACCOUNT_INQUIRY": {
					Type:       constant.TypeFee,
					Additional: util.ValueToPtr(constant.ReferenceAccountInquiry),
					Frequency:  10, Volume: 1_000,
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			result, err := repo.CalculatingMerchantTPVToDetermineFeeTierLevel(context.Background(), "12345", time.Now().Add(-1*time.Minute), time.Now())
			assert.Equal(t, test.wantErr, err)
			assert.Equal(t, test.wantResult, result)
		})
	}
}

func TestCalculatingMerchantTPVForLadderTiering(t *testing.T) {
	db := mysqlMocks.NewIMySqlExt(t)

	repo := New(db, nil)

	tpvSummariesMockType := mock.AnythingOfType("*[]orchestrator_model.CalculatingMerchantTPVSummary")

	tests := []struct {
		name       string
		setupMock  func()
		wantErr    error
		wantResult map[string]orchestrator_model.CalculatingMerchantTPVSummary
	}{
		{
			name: "ERROR:Some error",
			setupMock: func() {
				db.On(
					"SelectContext", constant.ValueCtxMockType(), tpvSummariesMockType, constant.StringMockType(), constant.StringMockType(), constant.TimeMockType(), constant.TimeMockType(),
				).Once().Return(constant.ErrSomeErrorForUnitTest)
			},
			wantErr: constant.ErrSomeErrorForUnitTest,
		},
		{
			name: "ERROR:Data not found",
			setupMock: func() {
				db.On(
					"SelectContext", constant.ValueCtxMockType(), tpvSummariesMockType, constant.StringMockType(), constant.StringMockType(), constant.TimeMockType(), constant.TimeMockType(),
				).Once().Return(sql.ErrNoRows)
			},
		},
		{
			name: "SUCCESS - includes pending settlement transactions",
			setupMock: func() {
				db.On(
					"SelectContext", constant.ValueCtxMockType(), tpvSummariesMockType, constant.StringMockType(), constant.StringMockType(), constant.TimeMockType(), constant.TimeMockType(),
				).Run(func(args mock.Arguments) {
					(*args.Get(1).(*[]orchestrator_model.CalculatingMerchantTPVSummary)) = []orchestrator_model.CalculatingMerchantTPVSummary{
						{
							Type:      constant.TypePayment,
							Channel:   constant.ChannelVirtualAccount,
							Frequency: 11, Volume: 1_980_000,
						},
						{
							Type:      constant.TypePayment,
							Channel:   constant.ChannelEwallet,
							Frequency: 1, Volume: 10_000,
						},
						{
							Type:      constant.TypeDisbursement,
							Channel:   constant.ChannelBankTransfer,
							Frequency: 10, Volume: 100_000,
						},
					}
				}).Return(nil)
			},
			wantResult: map[string]orchestrator_model.CalculatingMerchantTPVSummary{
				"PAYMENT_VIRTUAL_ACCOUNT": {
					Type:      constant.TypePayment,
					Channel:   constant.ChannelVirtualAccount,
					Frequency: 11, Volume: 1_980_000,
				},
				"PAYMENT_EWALLET": {
					Type:      constant.TypePayment,
					Channel:   constant.ChannelEwallet,
					Frequency: 1, Volume: 10_000,
				},
				"DISBURSEMENT_BANK_TRANSFER": {
					Type:      constant.TypeDisbursement,
					Channel:   constant.ChannelBankTransfer,
					Frequency: 10, Volume: 100_000,
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			result, err := repo.CalculatingMerchantTPVForLadderTiering(context.Background(), "12345", time.Now().Add(-1*time.Minute), time.Now())
			assert.Equal(t, test.wantErr, err)
			assert.Equal(t, test.wantResult, result)
		})
	}
}

func TestCalculatingTPVForPlatformActivitiesToDetermineFeeTierLevel(t *testing.T) {
	db := mysqlMocks.NewIMySqlExt(t)

	repo := New(db, nil)

	tpvSummaryMockType := mock.AnythingOfType("*orchestrator_model.CalculatingMerchantTPVSummary")

	tests := []struct {
		name       string
		setupMock  func()
		wantErr    error
		wantResult map[string]orchestrator_model.CalculatingMerchantTPVSummary
	}{
		{
			name: "ERROR:In statement",
			setupMock: func() {
				db.On(
					"In", constant.StringMockType(), constant.ArrayStringMockType(), constant.TimeMockType(), constant.TimeMockType(),
				).Once().Return("", nil, constant.ErrSomeErrorForUnitTest)
			},
			wantErr: fmt.Errorf("in statement: %w", constant.ErrSomeErrorForUnitTest),
		},
		{
			name: "ERROR:Get context",
			setupMock: func() {
				db.On(
					"In", constant.StringMockType(), constant.ArrayStringMockType(), constant.TimeMockType(), constant.TimeMockType(),
				).Return("SELECT ?", []interface{}{"123", time.Now(), time.Now().Add(time.Minute)}, nil)

				db.On(
					"GetContext", constant.ValueCtxMockType(), tpvSummaryMockType, constant.StringMockType(), constant.StringMockType(), constant.TimeMockType(), constant.TimeMockType(),
				).Once().Return(constant.ErrSomeErrorForUnitTest)
			},
			wantErr: constant.ErrSomeErrorForUnitTest,
		},
		{
			name: "SUCCESS:Data not found",
			setupMock: func() {
				db.On(
					"GetContext", constant.ValueCtxMockType(), tpvSummaryMockType, constant.StringMockType(), constant.StringMockType(), constant.TimeMockType(), constant.TimeMockType(),
				).Once().Return(sql.ErrNoRows)
			},
			wantResult: map[string]orchestrator_model.CalculatingMerchantTPVSummary{constant.ReferencePlatformActivity: {}},
		},
		{
			name: "SUCCESS:Data found",
			setupMock: func() {
				db.On(
					"GetContext", constant.ValueCtxMockType(), tpvSummaryMockType, constant.StringMockType(), constant.StringMockType(), constant.TimeMockType(), constant.TimeMockType(),
				).Run(func(args mock.Arguments) {
					(*args.Get(1).(*orchestrator_model.CalculatingMerchantTPVSummary)) = orchestrator_model.CalculatingMerchantTPVSummary{
						Type:      constant.ReferencePlatformActivity,
						Frequency: 11,
						Volume:    1_200_000,
					}
				}).Return(nil)
			},
			wantResult: map[string]orchestrator_model.CalculatingMerchantTPVSummary{
				constant.ReferencePlatformActivity: {
					Type:      constant.ReferencePlatformActivity,
					Frequency: 11,
					Volume:    1_200_000,
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			result, err := repo.CalculatingTPVForPlatformActivitiesToDetermineFeeTierLevel(context.Background(), []string{"12345"}, time.Now().Add(-1*time.Minute), time.Now())
			assert.Equal(t, test.wantErr, err)
			assert.Equal(t, test.wantResult, result)
		})
	}
}

func TestFindLastMerchantTransactionDate(t *testing.T) {
	timeNow := time.Now()

	testCase := []struct {
		name      string
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		wantErr   bool
	}{
		{
			name: "SUCCESS",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					constant.ValueCtxMockType(),
					constant.TimeReferenceMockType(), constant.StringMockType(), constant.StringMockType(),
				).Run(func(args mock.Arguments) {
					existedPtr := args.Get(1).(*time.Time)
					*existedPtr = timeNow
				}).Return(nil)
			},
		},
		{
			name: "ERROR: Mysql error",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					constant.ValueCtxMockType(),
					constant.TimeReferenceMockType(), constant.StringMockType(), constant.StringMockType(),
				).Return(constant.ErrSomeErrorForUnitTest)

			},
			wantErr: true,
		},
		{
			name: "ERROR: Data not found",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					constant.ValueCtxMockType(),
					constant.TimeReferenceMockType(), constant.StringMockType(), constant.StringMockType(),
				).Return(sql.ErrNoRows)
			},
		},
	}

	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			mockMysql := mysqlMocks.NewIMySqlExt(t)

			ctx := context.Background()

			tc.mockSetup(mockMysql)

			repo := New(mockMysql, mockLogger)

			if _, err := repo.FindLastMerchantTransactionDate(ctx, uuid.NewString()); tc.wantErr {
				assert.Error(t, err)

			} else {
				assert.NoError(t, err)
			}

			mockMysql.AssertExpectations(t)

		})
	}
}

func TestGetLastTransactionByAccountName(t *testing.T) {
	db := mysqlMocks.NewIMySqlExt(t)

	repo := New(db, nil)

	result := time.Date(2024, 12, 3, 10, 0, 0, 0, time.UTC)

	tests := []struct {
		name       string
		setupMock  func()
		wantErr    error
		wantResult *time.Time
	}{
		{
			name: "ERROR:Some error",
			setupMock: func() {
				db.On(
					"GetContext", constant.ValueCtxMockType(), mock.Anything, constant.StringMockType(), constant.StringMockType(), constant.StringMockType(),
				).Once().Return(constant.ErrSomeErrorForUnitTest)
			},
			wantErr: constant.ErrSomeErrorForUnitTest, wantResult: &time.Time{},
		},
		{
			name: "SUCCESS",
			setupMock: func() {
				db.On(
					"GetContext", constant.ValueCtxMockType(), mock.Anything, constant.StringMockType(), constant.StringMockType(), constant.StringMockType(),
				).Run(func(args mock.Arguments) {
					*args.Get(1).(*time.Time) = result
				}).Return(nil)
			},
			wantResult: &result,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			result, err := repo.GetLastTransactionByAccountName(context.Background(), "123456", "PAYMENT")
			assert.Equal(t, test.wantErr, err)
			assert.Equal(t, test.wantResult, result)
		})
	}
}

func TestGetTransactionByReferenceIdAndProcessorId(t *testing.T) {
	db := mysqlMocks.NewIMySqlExt(t)

	repo := New(db, nil)

	result := orchestrator_model.AccountTransaction{UUID: uuid.New()}

	tests := []struct {
		name       string
		setupMocks func()
		wantErr    error
		wantResult *orchestrator_model.AccountTransaction
	}{
		{
			name: "SUCCESS:Data not found",
			setupMocks: func() {
				db.On(
					"GetContext", constant.ValueCtxMockType(), mock.Anything, constant.StringMockType(), constant.StringMockType(), constant.StringMockType(),
				).Once().Return(sql.ErrNoRows)
			},
		},
		{
			name: "ERROR:Some error",
			setupMocks: func() {
				db.On(
					"GetContext", constant.ValueCtxMockType(), mock.Anything, constant.StringMockType(), constant.StringMockType(), constant.StringMockType(),
				).Once().Return(constant.ErrSomeErrorForUnitTest)
			},
			wantErr: constant.ErrSomeErrorForUnitTest,
		},
		{
			name: "SUCCESS",
			setupMocks: func() {
				db.On(
					"GetContext", constant.ValueCtxMockType(), mock.Anything, constant.StringMockType(), constant.StringMockType(), constant.StringMockType(),
				).Run(func(args mock.Arguments) {
					*args.Get(1).(*orchestrator_model.AccountTransaction) = result
				}).Return(nil)
			},
			wantResult: &result,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMocks()

			result, err := repo.GetTransactionByReferenceIdAndProcessorId(context.Background(), "1", "2")
			assert.Equal(t, test.wantErr, err)
			assert.Equal(t, test.wantResult, result)
		})
	}
}

func TestGetReferenceIdByTransactionIdAndType(t *testing.T) {
	db := mysqlMocks.NewIMySqlExt(t)
	logger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

	repo := New(db, logger)

	transactionId := uuid.NewString()
	transactionType := "PAYMENT"

	tests := []struct {
		name       string
		mockSetup  func()
		wantResult string
		wantErr    bool
	}{
		{
			name: "SUCCESS: Get reference ID",
			mockSetup: func() {
				db.On(
					"GetContext",
					mock.Anything,
					mock.AnythingOfType("*string"),
					mock.Anything,
					transactionId,
					transactionType,
				).Run(func(args mock.Arguments) {
					*args.Get(1).(*string) = "ref-123456"
				}).Return(nil)
			},
			wantResult: "ref-123456",
			wantErr:    false,
		},
		{
			name: "SUCCESS: No data found",
			mockSetup: func() {
				db.On(
					"GetContext",
					mock.Anything,
					mock.AnythingOfType("*string"),
					mock.Anything,
					transactionId,
					transactionType,
				).Return(sql.ErrNoRows)
			},
			wantResult: "",
			wantErr:    false,
		},
		{
			name: "ERROR: Database error",
			mockSetup: func() {
				db.On(
					"GetContext",
					mock.Anything,
					mock.AnythingOfType("*string"),
					mock.Anything,
					transactionId,
					transactionType,
				).Return(errors.New("database error"))
			},
			wantResult: "",
			wantErr:    true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Clear previous mocks and set up new ones
			db.ExpectedCalls = nil

			test.mockSetup()

			result, err := repo.GetReferenceIdByTransactionIdAndType(context.Background(), transactionId, transactionType)

			if test.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, test.wantResult, result)
			db.AssertExpectations(t)
		})
	}
}

func TestGetTransactionTransferDetailWithMatchingMerchantID(t *testing.T) {
	mysql := mysqlMocks.NewIMySqlExt(t)
	logger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

	// Create concrete repository instead of interface
	repo := &AccountTransactionRepository{
		db:     mysql,
		logger: logger,
	}

	// Fixed merchant ID for comparison
	merchantId := "matching-merchant-id"
	transferId := "transfer-id-1234"

	// Mock setup for getTransactionTransferDetail
	mysql.On(
		"GetContext", constant.ValueCtxMockType(), mock.AnythingOfType("*orchestrator_model.TransactionTransferResp"), constant.StringMockType(), transferId,
	).Run(func(args mock.Arguments) {
		dest := args.Get(1).(*orchestrator_model.TransactionTransferResp)
		dest.MerchantID = merchantId // Set the same merchant ID to trigger the condition
	}).Return(nil)

	// Call the function directly
	result, err := repo.getTransactionTransferDetail(context.Background(), transferId, merchantId)

	// Assertions
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, constant.TransferTypeOUT, result.Type)
	require.Equal(t, merchantId, result.MerchantID)
}
