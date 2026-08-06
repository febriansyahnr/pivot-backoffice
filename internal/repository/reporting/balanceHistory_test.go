package reportingRepository_test

import (
	"database/sql"
	"strings"
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	orchestratorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	model "github.com/paper-indonesia/pivot-backoffice/internal/model/reporting"
	. "github.com/paper-indonesia/pivot-backoffice/internal/repository/reporting"
	loggerMock "github.com/paper-indonesia/pivot-backoffice/mocks/pdk/logger"
	mySqlExtMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"

	"github.com/google/uuid"
	pdkLog "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestUpsertBalanceHistory(t *testing.T) {
	db := mySqlExtMock.NewIMySqlExt(t)
	logger := loggerMock.NewILogger(t)

	repo := New(db, logger, config.AppConfig{})

	data := model.BalanceHistory{
		TransactionID:    "TRX-123456",        //NOSONAR
		MerchantID:       "merchant-uuid-123", //NOSONAR
		ReferenceID:      "REF-001",           //NOSONAR
		Type:             constant.TypePayment,
		BalanceType:      "PAYMENT",                                     //NOSONAR
		Channel:          "VIRTUAL_ACCOUNT",                             //NOSONAR
		TransactionType:  "PAYMENT",                                     //NOSONAR
		Currency:         "IDR",                                         //NOSONAR
		Amount:           decimal.NewFromInt(100000),                    //NOSONAR
		Fee:              decimal.NewFromInt(2500),                      //NOSONAR
		Remarks:          "Test transaction",                            //NOSONAR
		Status:           "SUCCESS",                                     //NOSONAR
		SettlementStatus: "PENDING",                                     //NOSONAR
		CreatedAt:        time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC), //NOSONAR
		SourceID:         "source-uuid-123",                             //NOSONAR
		IngestedAt:       time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC), //NOSONAR
	}

	tests := []struct {
		name      string
		setupMock func()
		wantError error
	}{
		{
			name: "SUCCESS",
			setupMock: func() {
				db.On("NamedExecContext", mock.Anything, mock.Anything, data).Once().Return(true, nil)
			},
			wantError: nil,
		},
		{
			name: "ERROR:Database error",
			setupMock: func() {
				db.On("NamedExecContext", mock.Anything, mock.Anything, data).Once().Return(false, assert.AnError)
				logger.On("Error", mock.Anything, "Failed to upsert balance history data", pdkLog.Error(assert.AnError)).Once().Return()
			},
			wantError: assert.AnError,
		},
		{
			name: "ERROR:Connection error",
			setupMock: func() {
				db.On("NamedExecContext", mock.Anything, mock.Anything, data).Once().Return(false, sql.ErrConnDone)
				logger.On("Error", mock.Anything, "Failed to upsert balance history data", pdkLog.Error(sql.ErrConnDone)).Once().Return()
			},
			wantError: sql.ErrConnDone,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			err := repo.UpsertBalanceHistory(t.Context(), data)
			assert.Equal(t, test.wantError, err)

			db.AssertExpectations(t)
			logger.AssertExpectations(t)
		})
	}
}

func TestHardDeleteBalanceHistory(t *testing.T) {
	db := mySqlExtMock.NewIMySqlExt(t)
	logger := loggerMock.NewILogger(t)

	repo := New(db, logger, config.AppConfig{})

	transactionID := "TRX-123456"

	tests := []struct {
		name      string
		setupMock func()
		wantError error
	}{
		{
			name: "SUCCESS",
			setupMock: func() {
				db.On("ExecContext", mock.Anything, mock.Anything, transactionID).Once().Return(true, nil)
			},
			wantError: nil,
		},
		{
			name: "ERROR:Database error",
			setupMock: func() {
				db.On("ExecContext", mock.Anything, mock.Anything, transactionID).Once().Return(false, assert.AnError)
				logger.On("Error", mock.Anything, "Failed to hard delete balance history data", pdkLog.Error(assert.AnError)).Once().Return()
			},
			wantError: assert.AnError,
		},
		{
			name: "ERROR:Connection error",
			setupMock: func() {
				db.On("ExecContext", mock.Anything, mock.Anything, transactionID).Once().Return(false, sql.ErrConnDone)
				logger.On("Error", mock.Anything, "Failed to hard delete balance history data", pdkLog.Error(sql.ErrConnDone)).Once().Return()
			},
			wantError: sql.ErrConnDone,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			err := repo.HardDeleteBalanceHistory(t.Context(), transactionID)
			assert.Equal(t, test.wantError, err)

			db.AssertExpectations(t)
			logger.AssertExpectations(t)
		})
	}
}

func TestSoftDeleteBalanceHistory(t *testing.T) {
	db := mySqlExtMock.NewIMySqlExt(t)
	logger := loggerMock.NewILogger(t)

	repo := New(db, logger, config.AppConfig{})

	transactionID := "TRX-123456"
	ingestedAt := time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC)

	tests := []struct {
		name      string
		setupMock func()
		wantError error
	}{
		{
			name: "SUCCESS",
			setupMock: func() {
				db.On("ExecContext", mock.Anything, mock.Anything, mock.Anything, ingestedAt, transactionID).Once().Return(true, nil)
			},
			wantError: nil,
		},
		{
			name: "ERROR:Database error",
			setupMock: func() {
				logger.On("Error", mock.Anything, "Failed to soft delete balance history data", pdkLog.Error(assert.AnError)).Once().Return()
				db.On("ExecContext", mock.Anything, mock.Anything, mock.Anything, ingestedAt, transactionID).Once().Return(false, assert.AnError)
			},
			wantError: assert.AnError,
		},
		{
			name: "ERROR:Connection error",
			setupMock: func() {
				logger.On("Error", mock.Anything, "Failed to soft delete balance history data", pdkLog.Error(sql.ErrConnDone)).Once().Return()
				db.On("ExecContext", mock.Anything, mock.Anything, mock.Anything, ingestedAt, transactionID).Once().Return(false, sql.ErrConnDone)
			},
			wantError: sql.ErrConnDone,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			err := repo.SoftDeleteBalanceHistory(t.Context(), transactionID, ingestedAt)
			assert.Equal(t, test.wantError, err)

			db.AssertExpectations(t)
			logger.AssertExpectations(t)
		})
	}
}

func TestUpdateSettlementBalanceHistory(t *testing.T) {
	db := mySqlExtMock.NewIMySqlExt(t)
	logger := loggerMock.NewILogger(t)

	repo := New(db, logger, config.AppConfig{})

	settlementAt := time.Date(2025, 1, 16, 10, 0, 0, 0, time.UTC)
	ingestedAt := time.Date(2025, 1, 16, 10, 0, 0, 0, time.UTC)

	data := model.BalanceHistory{
		TransactionID:    "TRX-123456",
		SettlementStatus: "SETTLED",
		SettlementAt:     settlementAt,
		IngestedAt:       ingestedAt,
	}

	tests := []struct {
		name      string
		setupMock func()
		wantError error
	}{
		{
			name: "SUCCESS",
			setupMock: func() {
				db.On("NamedExecContext", mock.Anything, mock.Anything, data).Once().Return(true, nil)
			},
			wantError: nil,
		},
		{
			name: "ERROR:Database error",
			setupMock: func() {
				db.On("NamedExecContext", mock.Anything, mock.Anything, data).Once().Return(false, assert.AnError)
				logger.On("Error", mock.Anything, "Failed to update settlement balance history data", pdkLog.Error(assert.AnError)).Once().Return()
			},
			wantError: assert.AnError,
		},
		{
			name: "ERROR:No rows effected",
			setupMock: func() {
				db.On("NamedExecContext", mock.Anything, mock.Anything, data).Once().Return(false, nil)
			},
			wantError: constant.ErrNoRowsAffected,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			err := repo.UpdateSettlementBalanceHistory(t.Context(), data)
			assert.Equal(t, test.wantError, err)

			db.AssertExpectations(t)
			logger.AssertExpectations(t)
		})
	}
}

func TestPrepareAdvancedBalanceHistoryData(t *testing.T) {
	db := mySqlExtMock.NewIMySqlExt(t)
	logger := loggerMock.NewILogger(t)

	repo := New(db, logger, config.AppConfig{})

	var (
		sourceID          = "source-uuid-123"
		clientReferenceID = "client-reference-id-001"
		sourceCreatedBy   = "username"
		feeAmount         = decimal.NewFromInt(500)
		merchantID        = "merchant-uuid-123"
		createdAt         = time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC)
	)

	tests := []struct {
		name       string
		data       *model.BalanceHistory
		setupMock  func()
		wantError  error
		wantResult func(data *model.BalanceHistory)
	}{
		{
			name: "SUCCESS:Type Disbursement",
			data: &model.BalanceHistory{
				Type:     constant.TypeDisbursement,
				SourceID: sourceID,
			},
			setupMock: func() {
				db.On("GetContext", mock.Anything, mock.Anything, mock.MatchedBy(func(query string) bool {
					return strings.Contains(query, "disbursements d") &&
						strings.Contains(query, "LEFT JOIN merchants m ON m.uuid = d.created_by") &&
						strings.Contains(query, "LEFT JOIN users u ON u.uuid = d.created_by") &&
						strings.Contains(query, "LEFT JOIN merchants mt ON mt.uuid = d.merchant_id")
				}), sourceID).Once().Run(func(args mock.Arguments) {
					dest := args.Get(1).(*model.BalanceHistory)
					dest.ReferenceID = clientReferenceID
					dest.SourceCreatedBy = sourceCreatedBy
				}).Return(nil)
			},
			wantResult: func(data *model.BalanceHistory) {
				assert.Equal(t, clientReferenceID, data.ReferenceID)
				assert.Equal(t, sourceCreatedBy, data.SourceCreatedBy)
			},
		},
		{
			name: "SUCCESS:Type Payment",
			data: &model.BalanceHistory{
				Type:     constant.TypePayment,
				SourceID: sourceID,
				Fee:      feeAmount,
			},
			setupMock: func() {
				db.On("GetContext", mock.Anything, mock.Anything, mock.MatchedBy(func(query string) bool {
					return strings.Contains(query, "payments p") &&
						strings.Contains(query, "LEFT JOIN merchants m ON m.uuid = p.created_by") &&
						strings.Contains(query, "LEFT JOIN users u ON u.uuid = p.created_by")
				}), feeAmount, feeAmount, sourceID).Once().Run(func(args mock.Arguments) {
					dest := args.Get(1).(*model.BalanceHistory)
					dest.ReferenceID = clientReferenceID
					dest.SourceCreatedBy = sourceCreatedBy
				}).Return(nil)
			},
			wantResult: func(data *model.BalanceHistory) {
				assert.Equal(t, clientReferenceID, data.ReferenceID)
				assert.Equal(t, sourceCreatedBy, data.SourceCreatedBy)
			},
		},
		{
			name: "SUCCESS:Type Withdrawal",
			data: &model.BalanceHistory{
				Type:     constant.TypeWithdrawal,
				SourceID: sourceID,
			},
			setupMock: func() {
				db.On("GetContext", mock.Anything, mock.Anything, mock.MatchedBy(func(query string) bool {
					return strings.Contains(query, "withdrawals w") &&
						strings.Contains(query, "LEFT JOIN merchants m ON m.uuid = w.created_by") &&
						strings.Contains(query, "LEFT JOIN users u ON u.uuid = w.created_by")
				}), sourceID).Once().Run(func(args mock.Arguments) {
					dest := args.Get(1).(*model.BalanceHistory)
					dest.ReferenceID = clientReferenceID
					dest.SourceCreatedBy = sourceCreatedBy
				}).Return(nil)
			},
			wantResult: func(data *model.BalanceHistory) {
				assert.Equal(t, clientReferenceID, data.ReferenceID)
				assert.Equal(t, sourceCreatedBy, data.SourceCreatedBy)
			},
		},
		{
			name: "SUCCESS:Type Transfer",
			data: &model.BalanceHistory{
				Type:       constant.TypeTransfer,
				SourceID:   sourceID,
				MerchantID: merchantID,
				CreatedAt:  createdAt,
			},
			setupMock: func() {
				db.On("GetContext", mock.Anything, mock.Anything, mock.MatchedBy(func(query string) bool {
					return strings.Contains(query, "transfers t") &&
						strings.Contains(query, "LEFT JOIN account_transactions at ON at.reference_id = t.uuid")
				}), merchantID, createdAt, createdAt, sourceID).Once().Run(func(args mock.Arguments) {
					dest := args.Get(1).(*model.BalanceHistory)
					dest.ReferenceID = clientReferenceID
					dest.SourceCreatedBy = sourceCreatedBy
				}).Return(nil)
			},
			wantResult: func(data *model.BalanceHistory) {
				assert.Equal(t, clientReferenceID, data.ReferenceID)
				assert.Equal(t, sourceCreatedBy, data.SourceCreatedBy)
			},
		},
		{
			name: "SUCCESS:Disbursement Fee transaction type",
			data: &model.BalanceHistory{
				Type:            constant.TypeFee,
				TransactionType: constant.TypeDisbursement + "_FEE",
				SourceID:        sourceID,
			},
			setupMock: func() {
				db.On("GetContext", mock.Anything, mock.Anything, mock.MatchedBy(func(query string) bool {
					return strings.Contains(query, "disbursements")
				}), sourceID).Once().Run(func(args mock.Arguments) {
					dest := args.Get(1).(*model.BalanceHistory)
					dest.ReferenceID = clientReferenceID
					dest.SourceCreatedBy = sourceCreatedBy
				}).Return(nil)
			},
			wantResult: func(data *model.BalanceHistory) {
				assert.Equal(t, clientReferenceID, data.ReferenceID)
				assert.Equal(t, sourceCreatedBy, data.SourceCreatedBy)
			},
		},
		{
			name: "SUCCESS:Type GeneralTopUp",
			data: &model.BalanceHistory{
				Type:       constant.TypeGeneralTopUp,
				Channel:    constant.ChannelManualTransfer,
				MerchantID: merchantID,
				SourceID:   sourceID,
				CreatedAt:  createdAt,
			},
			setupMock: func() {
				db.On("GetContext", mock.Anything, mock.Anything, mock.MatchedBy(func(query string) bool {
					return strings.Contains(query, "account_transactions")
				}), sourceID, merchantID, constant.TypeGeneralTopUp, createdAt, createdAt).Once().Return(nil)
			},
			wantResult: func(*model.BalanceHistory) { /* Empty Function */ },
		},
		{
			name: "SUCCESS:Type Refund",
			data: &model.BalanceHistory{
				MerchantID: merchantID,
				Type:       constant.TypeRefund,
				SourceID:   sourceID,
				CreatedAt:  createdAt,
			},
			setupMock: func() {
				db.On("GetContext", mock.Anything, mock.Anything, mock.MatchedBy(func(query string) bool {
					return strings.Contains(query, "refunds r") &&
						strings.Contains(query, "JOIN merchants m ON m.uuid = r.merchant_id") &&
						strings.Contains(query, "JOIN payments p ON p.uuid = r.payment_id")
				}), merchantID, createdAt, createdAt, sourceID).Once().Return(nil)
			},
			wantResult: func(*model.BalanceHistory) { /* Empty Function */ },
		},
		{
			name: "SUCCESS:Type Virtual Terminal",
			data: &model.BalanceHistory{
				Type:     constant.TypeVirtualTerminal,
				SourceID: sourceID,
				Fee:      feeAmount,
			},
			setupMock: func() {
				db.On("GetContext", mock.Anything, mock.Anything, mock.MatchedBy(func(query string) bool {
					return strings.Contains(query, "payments p") &&
						strings.Contains(query, "LEFT JOIN merchants m ON m.uuid = p.created_by") &&
						strings.Contains(query, "LEFT JOIN users u ON u.uuid = p.created_by")
				}), feeAmount, feeAmount, sourceID).Once().Run(func(args mock.Arguments) {
					dest := args.Get(1).(*model.BalanceHistory)
					dest.ReferenceID = clientReferenceID
					dest.SourceCreatedBy = sourceCreatedBy
				}).Return(nil)
			},
			wantResult: func(data *model.BalanceHistory) {
				assert.Equal(t, constant.TypeVirtualTerminal, data.Type)
			},
		},
		{
			name: "SUCCESS:Type Merchant Payment",
			data: &model.BalanceHistory{
				Type:       constant.TypeMerchantPayment,
				MerchantID: merchantID,
				SourceID:   sourceID,
				CreatedAt:  createdAt,
			},
			setupMock: func() {
				db.On("GetContext", mock.Anything, mock.Anything, mock.MatchedBy(func(query string) bool {
					return query == "SELECT name AS source_created_by FROM merchants WHERE uuid = ?;"
				}), merchantID).Once().Return(nil)
			},
			wantResult: func(*model.BalanceHistory) { /* Empty Function */ },
		},
		{
			name: "SUCCESS:Type Manual Adjustment",
			data: &model.BalanceHistory{
				Type:       constant.TypeManualAdjust,
				MerchantID: merchantID,
				SourceID:   sourceID,
				CreatedAt:  createdAt,
			},
			setupMock: func() {
				db.On("GetContext", mock.Anything, mock.Anything, mock.MatchedBy(func(query string) bool {
					return strings.Contains(query, "manual_adjustment_histories")
				}), sourceID).Once().Return(nil)
			},
			wantResult: func(*model.BalanceHistory) { /* Empty Function */ },
		},
		{
			name: "SUCCESS:Type Manual Adjustment",
			data: &model.BalanceHistory{
				Type:            constant.TypeReversal,
				TransactionType: constant.TypeDisbursement,
				MerchantID:      merchantID,
				SourceID:        sourceID,
				CreatedAt:       createdAt,
			},
			setupMock: func() {
				db.On("GetContext", mock.Anything, mock.Anything, mock.MatchedBy(func(query string) bool {
					return strings.Contains(query, "disbursements")
				}), sourceID).Once().Return(nil)
			},
			wantResult: func(*model.BalanceHistory) { /* Empty Function */ },
		},
		{
			name: "SUCCESS:Unknown type returns nil without error",
			data: &model.BalanceHistory{
				Type:            "UNKNOWN_TYPE",
				TransactionType: "UNKNOWN",
				SourceID:        sourceID,
			},
			setupMock:  func() { /* Empty Function */ },
			wantResult: func(*model.BalanceHistory) { /* Empty Function */ },
		},
		{
			name: "ERROR:Database error on GetContext",
			data: &model.BalanceHistory{
				Type:     constant.TypeDisbursement,
				SourceID: sourceID,
			},
			setupMock: func() {
				db.On("GetContext", mock.Anything, mock.Anything, mock.Anything, sourceID).Once().Return(assert.AnError)
				logger.On("Warn", mock.Anything, "Failed to run the balance history data preparation query", mock.Anything, mock.Anything).Once().Return()
			},
			wantError:  assert.AnError,
			wantResult: func(*model.BalanceHistory) { /* Empty Function */ },
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			err := repo.PrepareAdvancedBalanceHistoryData(t.Context(), test.data)

			test.wantResult(test.data)
			assert.Equal(t, test.wantError, err)

			db.AssertExpectations(t)
			logger.AssertExpectations(t)
		})
	}
}

// mockArgsMatcher creates a flexible matcher for variadic arguments
func mockArgsMatcher(numArgs int) []any {
	args := make([]any, numArgs)
	for i := range args {
		args[i] = mock.Anything
	}
	return args
}

func TestListBalanceHistory(t *testing.T) {
	db := mySqlExtMock.NewIMySqlExt(t)
	logger := loggerMock.NewILogger(t)

	appConfig := config.AppConfig{
		UseOverFetchPagination: false,
		InitialPageWindow:      0,
	}
	repo := New(db, logger, appConfig)

	var (
		merchantID    = uuid.New()
		startDate     = time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
		endDate       = time.Date(2025, 1, 31, 23, 59, 59, 0, time.UTC)
		transactionID = uuid.New()
		updatedAt     = time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC)
	)

	baseFilter := &orchestratorModel.TransactionHistoryFilterRequest{
		MerchantID:          merchantID.String(),
		StartSettlementDate: startDate,
		EndSettlementDate:   endDate,
		FilteredSortQuery:   "status_updated_at DESC",
	}

	tests := []struct {
		name       string
		filter     *orchestratorModel.TransactionHistoryFilterRequest
		page       int64
		perPage    int64
		numArgs    int // Number of expected variadic args (base filter = 3: merchantID, startDate, endDate)
		setupMock  func(numArgs int)
		wantError  error
		wantResult func(t *testing.T, result *commonModel.PaginationResponse)
	}{
		{
			name:    "SUCCESS:List with basic filter",
			filter:  baseFilter,
			page:    1,
			perPage: 10,
			numArgs: 3, // merchantID, startDate, endDate
			setupMock: func(numArgs int) {
				// Mock SelectContext for fetching data
				args := append([]any{mock.Anything, mock.Anything, mock.MatchedBy(func(query string) bool {
					return strings.Contains(query, "SELECT") &&
						strings.Contains(query, "FROM") &&
						strings.Contains(query, "report_balance_histories")
				})}, mockArgsMatcher(numArgs)...)
				db.On("SelectContext", args...).Once().Run(func(args mock.Arguments) {
					dest := args.Get(1).(*[]orchestratorModel.AccountTransactionWithUseCase)
					*dest = []orchestratorModel.AccountTransactionWithUseCase{
						{
							UUID:                   transactionID,
							ReferenceID:            "REF-001",
							Type:                   "DISBURSEMENT",
							Channel:                "VIRTUAL_ACCOUNT",
							Status:                 "SUCCESS",
							UpdatedAt:              updatedAt,
							Debit:                  100000,
							Credit:                 0,
							Fee:                    2500,
							BeneficiaryAccountName: "John Doe",
							BalanceType:            "Payout Balance",
							MerchantReferenceID:    "-",
							SettlementStatus:       sql.NullString{String: "SUCCESS", Valid: true},
							SettlementAt:           sql.NullTime{Time: updatedAt, Valid: true},
							SettlementModel:        sql.NullString{String: "INSTANT", Valid: true},
						},
					}
				}).Return(nil)

				// Mock GetContext for count query
				countArgs := append([]any{mock.Anything, mock.Anything, mock.MatchedBy(func(query string) bool {
					return strings.Contains(query, "COUNT")
				})}, mockArgsMatcher(numArgs)...)
				db.On("GetContext", countArgs...).Once().Run(func(args mock.Arguments) {
					dest := args.Get(1).(*int64)
					*dest = 1
				}).Return(nil)
			},
			wantError: nil,
			wantResult: func(t *testing.T, result *commonModel.PaginationResponse) {
				assert.NotNil(t, result)
				assert.Equal(t, int64(1), result.Meta.Page)
				assert.Equal(t, int64(10), result.Meta.PerPage)
				assert.Equal(t, int64(1), result.Meta.TotalItems)

				data, ok := result.Data.([]any)
				require.True(t, ok)
				require.Len(t, data, 1)
				originData := data[0].(orchestratorModel.TransactionHistoryResponse)
				assert.Equal(t, transactionID.String(), originData.ID)
				assert.Equal(t, "DISBURSEMENT", originData.TrxType)
				assert.Equal(t, "SUCCESS", originData.Status)
			},
		},
		{
			name: "SUCCESS:List with balance types filter",
			filter: &orchestratorModel.TransactionHistoryFilterRequest{
				MerchantID:          merchantID.String(),
				StartSettlementDate: startDate,
				EndSettlementDate:   endDate,
				FilteredSortQuery:   "status_updated_at DESC",
				BalanceTypes:        []string{"DISBURSEMENT", "PAYMENT"},
			},
			page:    1,
			perPage: 10,
			numArgs: 5, // merchantID, startDate, endDate + 2 balance types
			setupMock: func(numArgs int) {
				args := append([]any{mock.Anything, mock.Anything, mock.Anything}, mockArgsMatcher(numArgs)...)
				db.On("SelectContext", args...).Once().Run(func(args mock.Arguments) {
					dest := args.Get(1).(*[]orchestratorModel.AccountTransactionWithUseCase)
					*dest = []orchestratorModel.AccountTransactionWithUseCase{}
				}).Return(nil)

				countArgs := append([]any{mock.Anything, mock.Anything, mock.Anything}, mockArgsMatcher(numArgs)...)
				db.On("GetContext", countArgs...).Once().Run(func(args mock.Arguments) {
					dest := args.Get(1).(*int64)
					*dest = 0
				}).Return(nil)
			},
			wantError: nil,
			wantResult: func(t *testing.T, result *commonModel.PaginationResponse) {
				assert.NotNil(t, result)
				assert.Equal(t, int64(0), result.Meta.TotalItems)
			},
		},
		{
			name: "SUCCESS:List with transaction types filter",
			filter: &orchestratorModel.TransactionHistoryFilterRequest{
				MerchantID:          merchantID.String(),
				StartSettlementDate: startDate,
				EndSettlementDate:   endDate,
				FilteredSortQuery:   "status_updated_at DESC",
				TrxTypes:            []string{"DISBURSEMENT", "BULK_DISBURSEMENT"},
			},
			page:    1,
			perPage: 10,
			numArgs: 7, // merchantID, startDate, endDate + (2 args per trx type) = 3 + 4 = 7
			setupMock: func(numArgs int) {
				args := append([]any{mock.Anything, mock.Anything, mock.Anything}, mockArgsMatcher(numArgs)...)
				db.On("SelectContext", args...).Once().Run(func(args mock.Arguments) {
					dest := args.Get(1).(*[]orchestratorModel.AccountTransactionWithUseCase)
					*dest = []orchestratorModel.AccountTransactionWithUseCase{}
				}).Return(nil)

				countArgs := append([]any{mock.Anything, mock.Anything, mock.Anything}, mockArgsMatcher(numArgs)...)
				db.On("GetContext", countArgs...).Once().Run(func(args mock.Arguments) {
					dest := args.Get(1).(*int64)
					*dest = 0
				}).Return(nil)
			},
			wantError: nil,
			wantResult: func(t *testing.T, result *commonModel.PaginationResponse) {
				assert.NotNil(t, result)
			},
		},
		{
			name: "SUCCESS:List with status filter",
			filter: &orchestratorModel.TransactionHistoryFilterRequest{
				MerchantID:          merchantID.String(),
				StartSettlementDate: startDate,
				EndSettlementDate:   endDate,
				FilteredSortQuery:   "status_updated_at DESC",
				Status:              "SUCCESS",
			},
			page:    1,
			perPage: 10,
			numArgs: 4, // merchantID, startDate, endDate + status
			setupMock: func(numArgs int) {
				args := append([]any{mock.Anything, mock.Anything, mock.Anything}, mockArgsMatcher(numArgs)...)
				db.On("SelectContext", args...).Once().Run(func(args mock.Arguments) {
					dest := args.Get(1).(*[]orchestratorModel.AccountTransactionWithUseCase)
					*dest = []orchestratorModel.AccountTransactionWithUseCase{}
				}).Return(nil)

				countArgs := append([]any{mock.Anything, mock.Anything, mock.Anything}, mockArgsMatcher(numArgs)...)
				db.On("GetContext", countArgs...).Once().Run(func(args mock.Arguments) {
					dest := args.Get(1).(*int64)
					*dest = 0
				}).Return(nil)
			},
			wantError: nil,
			wantResult: func(t *testing.T, result *commonModel.PaginationResponse) {
				assert.NotNil(t, result)
			},
		},
		{
			name: "SUCCESS:List with transaction id filter",
			filter: &orchestratorModel.TransactionHistoryFilterRequest{
				MerchantID:          merchantID.String(),
				StartSettlementDate: startDate,
				EndSettlementDate:   endDate,
				FilteredSortQuery:   "status_updated_at DESC",
				TransactionId:       "TRX-123456",
			},
			page:    1,
			perPage: 10,
			numArgs: 5, // merchantID, startDate, endDate + 2 (transactionId twice for OR condition)
			setupMock: func(numArgs int) {
				args := append([]any{mock.Anything, mock.Anything, mock.Anything}, mockArgsMatcher(numArgs)...)
				db.On("SelectContext", args...).Once().Run(func(args mock.Arguments) {
					dest := args.Get(1).(*[]orchestratorModel.AccountTransactionWithUseCase)
					*dest = []orchestratorModel.AccountTransactionWithUseCase{}
				}).Return(nil)

				countArgs := append([]any{mock.Anything, mock.Anything, mock.Anything}, mockArgsMatcher(numArgs)...)
				db.On("GetContext", countArgs...).Once().Run(func(args mock.Arguments) {
					dest := args.Get(1).(*int64)
					*dest = 0
				}).Return(nil)
			},
			wantError: nil,
			wantResult: func(t *testing.T, result *commonModel.PaginationResponse) {
				assert.NotNil(t, result)
			},
		},
		{
			name:    "SUCCESS:List with pagination page 2",
			filter:  baseFilter,
			page:    2,
			perPage: 5,
			numArgs: 3,
			setupMock: func(numArgs int) {
				args := append([]any{mock.Anything, mock.Anything, mock.Anything}, mockArgsMatcher(numArgs)...)
				db.On("SelectContext", args...).Once().Run(func(args mock.Arguments) {
					dest := args.Get(1).(*[]orchestratorModel.AccountTransactionWithUseCase)
					*dest = []orchestratorModel.AccountTransactionWithUseCase{
						{
							UUID:        transactionID,
							ReferenceID: "REF-006",
							Type:        "PAYMENT",
							Channel:     "VIRTUAL_ACCOUNT",
							Status:      "SUCCESS",
							UpdatedAt:   updatedAt,
							Credit:      50000,
							Fee:         1000,
						},
					}
				}).Return(nil)

				countArgs := append([]any{mock.Anything, mock.Anything, mock.Anything}, mockArgsMatcher(numArgs)...)
				db.On("GetContext", countArgs...).Once().Run(func(args mock.Arguments) {
					dest := args.Get(1).(*int64)
					*dest = 6
				}).Return(nil)
			},
			wantError: nil,
			wantResult: func(t *testing.T, result *commonModel.PaginationResponse) {
				assert.NotNil(t, result)
				assert.Equal(t, int64(2), result.Meta.Page)
				assert.Equal(t, int64(5), result.Meta.PerPage)
				assert.Equal(t, int64(6), result.Meta.TotalItems)
				assert.Equal(t, int64(2), result.Meta.TotalPages)
			},
		},
		{
			name:    "SUCCESS:Empty result",
			filter:  baseFilter,
			page:    1,
			perPage: 10,
			numArgs: 3,
			setupMock: func(numArgs int) {
				args := append([]any{mock.Anything, mock.Anything, mock.Anything}, mockArgsMatcher(numArgs)...)
				db.On("SelectContext", args...).Once().Run(func(args mock.Arguments) {
					dest := args.Get(1).(*[]orchestratorModel.AccountTransactionWithUseCase)
					*dest = []orchestratorModel.AccountTransactionWithUseCase{}
				}).Return(nil)

				countArgs := append([]any{mock.Anything, mock.Anything, mock.Anything}, mockArgsMatcher(numArgs)...)
				db.On("GetContext", countArgs...).Once().Run(func(args mock.Arguments) {
					dest := args.Get(1).(*int64)
					*dest = 0
				}).Return(nil)
			},
			wantError: nil,
			wantResult: func(t *testing.T, result *commonModel.PaginationResponse) {
				assert.NotNil(t, result)
				assert.Equal(t, int64(0), result.Meta.TotalItems)

				data, ok := result.Data.([]any)
				assert.True(t, ok)
				assert.Len(t, data, 0)
			},
		},
		{
			name:    "ERROR:SelectContext database error",
			filter:  baseFilter,
			page:    1,
			perPage: 10,
			numArgs: 3,
			setupMock: func(numArgs int) {
				// Both SelectContext and GetContext run in parallel via errgroup
				// So we need to mock both, even though only SelectContext error is returned
				logger.On("Error", mock.Anything, "error when get paginated list", pdkLog.Error(assert.AnError)).Once().Return()

				args := append([]any{mock.Anything, mock.Anything, mock.Anything}, mockArgsMatcher(numArgs)...)
				db.On("SelectContext", args...).Once().Return(assert.AnError)

				// GetContext is also called in parallel - its error is ignored
				countArgs := append([]any{mock.Anything, mock.Anything, mock.Anything}, mockArgsMatcher(numArgs)...)
				db.On("GetContext", countArgs...).Once().Return(nil)
			},
			wantError:  assert.AnError,
			wantResult: func(t *testing.T, result *commonModel.PaginationResponse) {},
		},
		{
			name:    "ERROR:GetContext count error (should not return error, just set count to 0)",
			filter:  baseFilter,
			page:    1,
			perPage: 10,
			numArgs: 3,
			setupMock: func(numArgs int) {
				args := append([]any{mock.Anything, mock.Anything, mock.Anything}, mockArgsMatcher(numArgs)...)
				db.On("SelectContext", args...).Once().Run(func(args mock.Arguments) {
					dest := args.Get(1).(*[]orchestratorModel.AccountTransactionWithUseCase)
					*dest = []orchestratorModel.AccountTransactionWithUseCase{}
				}).Return(nil)

				countArgs := append([]any{mock.Anything, mock.Anything, mock.Anything}, mockArgsMatcher(numArgs)...)
				db.On("GetContext", countArgs...).Once().Return(assert.AnError)
			},
			wantError: nil,
			wantResult: func(t *testing.T, result *commonModel.PaginationResponse) {
				assert.NotNil(t, result)
				assert.Equal(t, int64(0), result.Meta.TotalItems)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock(test.numArgs)

			result, err := repo.ListBalanceHistory(t.Context(), test.filter, test.page, test.perPage)

			assert.Equal(t, test.wantError, err)
			test.wantResult(t, result)

			db.AssertExpectations(t)
			logger.AssertExpectations(t)
		})
	}
}

func TestListBalanceHistoryWithOverFetch(t *testing.T) {
	db := mySqlExtMock.NewIMySqlExt(t)
	logger := loggerMock.NewILogger(t)

	appConfig := config.AppConfig{
		UseOverFetchPagination: true,
		InitialPageWindow:      3,
	}
	repo := New(db, logger, appConfig)

	var (
		merchantID = uuid.New()
		startDate  = time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
		endDate    = time.Date(2025, 1, 31, 23, 59, 59, 0, time.UTC)
	)

	filter := &orchestratorModel.TransactionHistoryFilterRequest{
		MerchantID:          merchantID.String(),
		StartSettlementDate: startDate,
		EndSettlementDate:   endDate,
		FilteredSortQuery:   "status_updated_at DESC",
	}

	tests := []struct {
		name       string
		page       int64
		perPage    int64
		setupMock  func()
		wantError  error
		wantResult func(t *testing.T, result *commonModel.PaginationResponse)
	}{
		{
			name:    "SUCCESS:Over-fetch pagination with trim",
			page:    1,
			perPage: 10,
			setupMock: func() {
				args := append([]any{mock.Anything, mock.Anything, mock.Anything}, mockArgsMatcher(3)...)
				db.On("SelectContext", args...).Once().Run(func(args mock.Arguments) {
					dest := args.Get(1).(*[]orchestratorModel.AccountTransactionWithUseCase)
					// Return 15 items (more than perPage)
					*dest = make([]orchestratorModel.AccountTransactionWithUseCase, 15)
					for i := range *dest {
						(*dest)[i] = orchestratorModel.AccountTransactionWithUseCase{
							UUID:      uuid.New(),
							Type:      "PAYMENT",
							UpdatedAt: time.Now(),
						}
					}
				}).Return(nil)
			},
			wantError: nil,
			wantResult: func(t *testing.T, result *commonModel.PaginationResponse) {
				assert.NotNil(t, result)
				assert.Equal(t, int64(1), result.Meta.Page)
				assert.Equal(t, int64(10), result.Meta.PerPage)

				data, ok := result.Data.([]any)
				assert.True(t, ok)
				// Should be trimmed to perPage (10)
				assert.Len(t, data, 10)
			},
		},
		{
			name:    "SUCCESS:Over-fetch pagination without trim",
			page:    1,
			perPage: 10,
			setupMock: func() {
				args := append([]any{mock.Anything, mock.Anything, mock.Anything}, mockArgsMatcher(3)...)
				db.On("SelectContext", args...).Once().Run(func(args mock.Arguments) {
					dest := args.Get(1).(*[]orchestratorModel.AccountTransactionWithUseCase)
					// Return fewer items than perPage
					*dest = make([]orchestratorModel.AccountTransactionWithUseCase, 5)
					for i := range *dest {
						(*dest)[i] = orchestratorModel.AccountTransactionWithUseCase{
							UUID:      uuid.New(),
							Type:      "PAYMENT",
							UpdatedAt: time.Now(),
						}
					}
				}).Return(nil)
			},
			wantError: nil,
			wantResult: func(t *testing.T, result *commonModel.PaginationResponse) {
				assert.NotNil(t, result)

				data, ok := result.Data.([]any)
				assert.True(t, ok)
				assert.Len(t, data, 5)
			},
		},
		{
			name:    "ERROR:Over-fetch database error",
			page:    1,
			perPage: 10,
			setupMock: func() {
				logger.On("Error", mock.Anything, "error when get paginated list with over-fetch", pdkLog.Error(assert.AnError)).Once().Return()
				args := append([]any{mock.Anything, mock.Anything, mock.Anything}, mockArgsMatcher(3)...)
				db.On("SelectContext", args...).Once().Return(assert.AnError)
			},
			wantError:  assert.AnError,
			wantResult: func(t *testing.T, result *commonModel.PaginationResponse) {},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			result, err := repo.ListBalanceHistory(t.Context(), filter, test.page, test.perPage)

			assert.Equal(t, test.wantError, err)
			test.wantResult(t, result)

			db.AssertExpectations(t)
			logger.AssertExpectations(t)
		})
	}
}

func TestExportBalanceHistory(t *testing.T) {
	db := mySqlExtMock.NewIMySqlExt(t)

	repo := New(db, nil, config.AppConfig{})

	var (
		merchantID = "5ccc58c5-7ac5-4c8f-a43a-b222df8f31aa"
		startDate  = time.Date(2026, 3, 30, 0, 0, 0, 0, time.UTC)
		endDate    = time.Date(2026, 3, 30, 23, 59, 59, 0, time.UTC)
	)
	tests := []struct {
		name       string
		setupMock  func()
		wantError  error
		wantResult []orchestratorModel.TransactionHistory
	}{
		{
			name: "ERROR:Transaction not found", // NOSONAR
			setupMock: func() {
				db.On("SelectContext", mock.Anything, mock.Anything, mock.Anything, merchantID, startDate, endDate).Once().Return(sql.ErrNoRows)
			},
		},
		{
			name: "ERROR:Some error", // NOSONAR
			setupMock: func() {
				db.On("SelectContext", mock.Anything, mock.Anything, mock.Anything, merchantID, startDate, endDate).Once().Return(assert.AnError)
			},
			wantError: assert.AnError,
		},
		{
			name: "SUCCESS", // NOSONAR
			setupMock: func() {
				db.On(
					"SelectContext", mock.Anything, mock.Anything, mock.Anything, merchantID, startDate, endDate,
				).Once().Run(func(args mock.Arguments) {
					*args.Get(1).(*[]orchestratorModel.AccountTransactionWithUseCase) = []orchestratorModel.AccountTransactionWithUseCase{{Credit: 1}}
				}).Return(nil)
			},
			wantResult: []orchestratorModel.TransactionHistory{{Id: "00000000-0000-0000-0000-000000000000", Amount: 1}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			result, err := repo.ExportBalanceHistory(t.Context(), &orchestratorModel.TransactionHistoryFilterRequest{
				MerchantID:          merchantID,
				StartSettlementDate: startDate,
				EndSettlementDate:   endDate,
			})
			assert.Equal(t, tt.wantError, err)
			assert.Equal(t, tt.wantResult, result)
		})
	}
}
