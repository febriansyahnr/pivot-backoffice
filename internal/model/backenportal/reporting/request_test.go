package reportingModel

import (
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	cdcModel "github.com/paper-indonesia/pivot-backoffice/internal/model/cdc"
	orchestrator_model "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestUpsertBalanceHistoryRequestShouldExcludeEvent(t *testing.T) {
	tests := []struct {
		name       string
		request    UpsertBalanceHistoryRequest
		wantResult bool
	}{
		{
			name: "create event with non-success status should be excluded",
			request: UpsertBalanceHistoryRequest{
				Event: &cdcModel.Event[cdcModel.AccountTransaction]{
					Op: cdcModel.OpCreate,
					After: &cdcModel.AccountTransaction{
						Status: constant.StatusPending,
					},
				},
			},
			wantResult: true,
		},
		{
			name: "delete event with non-success status should be excluded",
			request: UpsertBalanceHistoryRequest{
				Event: &cdcModel.Event[cdcModel.AccountTransaction]{
					Op: cdcModel.OpDelete,
					Before: &cdcModel.AccountTransaction{
						Status: constant.StatusPending,
					},
				},
			},
			wantResult: true,
		},
		{
			name: "create event with success status should not be excluded",
			request: UpsertBalanceHistoryRequest{
				Event: &cdcModel.Event[cdcModel.AccountTransaction]{
					Op: cdcModel.OpCreate,
					After: &cdcModel.AccountTransaction{
						Status: constant.StatusSuccess,
						Debit:  decimal.NewFromInt(100), // NOSONAR
					},
				},
			},
			wantResult: false,
		},
		{
			name: "update event should not be excluded based on operation",
			request: UpsertBalanceHistoryRequest{
				Event: &cdcModel.Event[cdcModel.AccountTransaction]{
					Op: cdcModel.OpUpdate,
					Before: &cdcModel.AccountTransaction{
						Status: constant.StatusPending,
						Debit:  decimal.NewFromInt(100), // NOSONAR
					},
					After: &cdcModel.AccountTransaction{
						Status: constant.StatusSuccess,
						Debit:  decimal.NewFromInt(100), // NOSONAR
					},
				},
			},
			wantResult: false,
		},
		{
			name: "zero debit and credit should be excluded",
			request: UpsertBalanceHistoryRequest{
				Event: &cdcModel.Event[cdcModel.AccountTransaction]{
					Op: cdcModel.OpUpdate,
					Before: &cdcModel.AccountTransaction{
						Status: constant.StatusPending,
					},
					After: &cdcModel.AccountTransaction{
						Status: constant.StatusPending,
					},
				},
			},
			wantResult: true,
		},
		{
			name: "fee with VA channel and disbursement reference should be excluded",
			request: UpsertBalanceHistoryRequest{
				Event: &cdcModel.Event[cdcModel.AccountTransaction]{
					Op: cdcModel.OpCreate,
					After: &cdcModel.AccountTransaction{
						Status:    constant.StatusSuccess,
						Type:      constant.TypeFee,
						Channel:   constant.ChannelVirtualAccount,
						Reference: util.ValueToPtr(constant.ReferenceDisbursement),
						Debit:     decimal.NewFromInt(100),
					},
				},
			},
			wantResult: true,
		},
		{
			name: "card-funded payout transaction should be excluded",
			request: UpsertBalanceHistoryRequest{
				Event: &cdcModel.Event[cdcModel.AccountTransaction]{
					Op: cdcModel.OpCreate,
					After: &cdcModel.AccountTransaction{
						Status:    constant.StatusSuccess,
						Type:      constant.TypeDisbursement,
						Channel:   constant.ChannelBankTransfer,
						Reference: util.ValueToPtr(constant.ReferencePaymentFundedPayout),
						Debit:     decimal.NewFromInt(2_000_0000),
					},
				},
			},
			wantResult: true,
		},
		{
			name: "sub payment transaction should be excluded",
			request: UpsertBalanceHistoryRequest{
				Event: &cdcModel.Event[cdcModel.AccountTransaction]{
					Op: cdcModel.OpCreate,
					After: &cdcModel.AccountTransaction{
						Status:    constant.StatusSuccess,
						Type:      constant.TypeDisbursement,
						Channel:   constant.ChannelBankTransfer,
						Reference: util.ValueToPtr(string(constant.ReferenceSubPayment)),
						Credit:    decimal.NewFromInt(2_000_0000),
					},
				},
			},
			wantResult: true,
		},
		{
			name: "auto split parent payment transaction should not be excluded",
			request: UpsertBalanceHistoryRequest{
				Event: &cdcModel.Event[cdcModel.AccountTransaction]{
					Op: cdcModel.OpCreate,
					After: &cdcModel.AccountTransaction{
						Status:    constant.StatusSuccess,
						Type:      constant.TypeDisbursement,
						Channel:   constant.ChannelBankTransfer,
						Reference: util.ValueToPtr(string(constant.ReferencePayment)),
						AdditionalInfo: cdcModel.AccountTransactionAdditionalInfo{
							SubPaymentSummary: &orchestrator_model.MetadataSubPaymentSummary{
								TotalCreditAmount: decimal.NewFromFloat(200000),
							},
						},
						Debit:  decimal.NewFromInt(0),
						Credit: decimal.NewFromInt(0),
					},
				},
			},
			wantResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.request.ShouldExcludeEvent()
			assert.Equal(t, tt.wantResult, result)
		})
	}
}

func TestUpsertBalanceHistoryRequestshouldExcludeFee(t *testing.T) {
	tests := []struct {
		name       string
		payload    *cdcModel.AccountTransaction
		wantResult bool
	}{
		{
			name: "non-fee type should not be excluded",
			payload: &cdcModel.AccountTransaction{
				Type: constant.TypePayment,
			},
			wantResult: false,
		},
		{
			name: "fee with VA channel and disbursement reference should be excluded",
			payload: &cdcModel.AccountTransaction{
				Type:      constant.TypeFee,
				Channel:   constant.ChannelVirtualAccount,
				Reference: util.ValueToPtr(constant.ReferenceDisbursement),
			},
			wantResult: true,
		},
		{
			name: "fee with DISBURSEMENT additional info type should be excluded",
			payload: &cdcModel.AccountTransaction{
				Type: constant.TypeFee,
				AdditionalInfo: cdcModel.AccountTransactionAdditionalInfo{
					Type: constant.ReferenceDisbursement,
				},
			},
			wantResult: true,
		},
		{
			name: "fee with DISBURSEMENT_VA additional info type should be excluded",
			payload: &cdcModel.AccountTransaction{
				Type: constant.TypeFee,
				AdditionalInfo: cdcModel.AccountTransactionAdditionalInfo{
					Type: constant.ReferenceDisbursementVA,
				},
			},
			wantResult: true,
		},
		{
			name: "fee with ON-BEHALF notes should not be excluded",
			payload: &cdcModel.AccountTransaction{
				Type: constant.TypeFee,
				AdditionalInfo: cdcModel.AccountTransactionAdditionalInfo{
					Type:  constant.ReferenceDisbursement,
					Notes: "ON-BEHALF",
				},
			},
			wantResult: false,
		},
		{
			name: "fee with INQUIRY_ACCOUNT additional info type should not be excluded",
			payload: &cdcModel.AccountTransaction{
				Type: constant.TypeFee,
				AdditionalInfo: cdcModel.AccountTransactionAdditionalInfo{
					Type: constant.ReferenceAccountInquiry,
				},
			},
			wantResult: false,
		},
		{
			name: "fee with PAYMENT additional info type should be excluded",
			payload: &cdcModel.AccountTransaction{
				Type:    constant.TypeFee,
				Channel: "QRIS",
				AdditionalInfo: cdcModel.AccountTransactionAdditionalInfo{
					Type: constant.ReferencePayment,
				},
			},
			wantResult: true,
		},
		{
			name: "fee with XB additional info type should be excluded",
			payload: &cdcModel.AccountTransaction{
				Type: constant.TypeFee,
				AdditionalInfo: cdcModel.AccountTransactionAdditionalInfo{
					Type: constant.ReferenceXB,
				},
			},
			wantResult: true,
		},
		{
			name: "fee with PLATFORM_TRANSACTION additional info type should be excluded",
			payload: &cdcModel.AccountTransaction{
				Type: constant.TypeFee,
				AdditionalInfo: cdcModel.AccountTransactionAdditionalInfo{
					Type: constant.ReferencePlatformTransaction,
				},
			},
			wantResult: true,
		},
		{
			name: "fee with other additional info type should not be excluded",
			payload: &cdcModel.AccountTransaction{
				Type:    constant.TypeFee,
				Channel: "OTHER",
				AdditionalInfo: cdcModel.AccountTransactionAdditionalInfo{
					Type: "UNKNOWN_TYPE",
				},
			},
			wantResult: false,
		},
		{
			name: "fee with empty additional info type should not be excluded",
			payload: &cdcModel.AccountTransaction{
				Type:    constant.TypeFee,
				Channel: "OTHER",
				AdditionalInfo: cdcModel.AccountTransactionAdditionalInfo{
					Type: "",
				},
			},
			wantResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &UpsertBalanceHistoryRequest{}
			result := req.shouldExcludeFee(tt.payload)
			assert.Equal(t, tt.wantResult, result)
		})
	}
}

func TestUpsertBalanceHistoryRequestToCreateBalanceHistory(t *testing.T) {
	now := time.Now()
	createdAt := now.Add(-time.Hour)
	settlementAt := now.Add(time.Minute)

	tests := []struct {
		name     string
		request  UpsertBalanceHistoryRequest
		validate func(t *testing.T, result BalanceHistory)
	}{
		{
			name: "basic mapping with debit amount",
			request: UpsertBalanceHistoryRequest{
				Event: &cdcModel.Event[cdcModel.AccountTransaction]{
					Op:   cdcModel.OpUpdate,
					TsMs: now.UnixMilli(),
					After: &cdcModel.AccountTransaction{
						UUID:                "tx-123",                            // NOSONAR
						MerchantID:          "merchant-456",                      // NOSONAR
						ReferenceID:         "ref-789",                           // NOSONAR
						MerchantReferenceID: util.ValueToPtr("merchant-ref-001"), // NOSONAR
						Currency:            "IDR",                               // NOSONAR
						Credit:              decimal.NewFromInt(1000),            // NOSONAR
						Type:                constant.TypePayment,
						Channel:             "VIRTUAL_ACCOUNT", // NOSONAR
						Status:              constant.StatusSuccess,
						CreatedAt:           now,
						UpdatedAt:           now,
						Reference:           util.ValueToPtr(constant.TypePayment),
					},
				},
			},
			validate: func(t *testing.T, result BalanceHistory) {
				assert.Equal(t, "tx-123", result.TransactionID)
				assert.Equal(t, "merchant-456", result.MerchantID)
				assert.Equal(t, "merchant-ref-001", result.ReferenceID)
				assert.Equal(t, constant.TypePayment, result.Type)
				assert.Equal(t, constant.TypePayment, result.TransactionType)
				assert.Equal(t, "VIRTUAL_ACCOUNT", result.Channel)
				assert.Equal(t, "IDR", result.Currency)
				assert.Equal(t, decimal.NewFromInt(1000), result.Amount)
				assert.Equal(t, constant.StatusSuccess, result.Status)
				assert.Equal(t, constant.StatusSuccess, result.SettlementStatus)
				assert.Equal(t, result.SettlementAt, now)
				assert.Equal(t, "ref-789", result.SourceID)
			},
		},
		{
			name: "basic mapping with credit amount",
			request: UpsertBalanceHistoryRequest{
				Event: &cdcModel.Event[cdcModel.AccountTransaction]{
					Op:   cdcModel.OpUpdate,
					TsMs: now.UnixMilli(),
					After: &cdcModel.AccountTransaction{
						UUID:       "tx-123",
						MerchantID: "merchant-456",
						Currency:   "IDR",
						Debit:      decimal.Zero,
						Credit:     decimal.NewFromInt(2000),
						Type:       constant.TypePayment,
						Channel:    "VIRTUAL_ACCOUNT",
						Status:     constant.StatusSuccess,
						Remarks:    "Test remarks",
						CreatedAt:  now,
						UpdatedAt:  now,
						Reference:  util.ValueToPtr(constant.TypeWallet),
					},
				},
			},
			validate: func(t *testing.T, result BalanceHistory) {
				assert.Equal(t, decimal.NewFromInt(2000), result.Amount)
			},
		},
		{
			name: "withdrawal type should append to balance type",
			request: UpsertBalanceHistoryRequest{
				Event: &cdcModel.Event[cdcModel.AccountTransaction]{
					Op:   cdcModel.OpUpdate,
					TsMs: now.UnixMilli(),
					After: &cdcModel.AccountTransaction{
						UUID:       "tx-123",                 // NOSONAR
						MerchantID: "merchant-456",           // NOSONAR
						Debit:      decimal.NewFromInt(1000), // NOSONAR
						Type:       constant.TypeWithdrawal,
						Status:     constant.StatusSuccess,
						CreatedAt:  now, UpdatedAt: now,
						Reference: util.ValueToPtr(constant.TypeDisbursement),
					},
				},
			},
			validate: func(t *testing.T, result BalanceHistory) {
				assert.Equal(t, constant.TypeDisbursement+"_"+constant.TypeWithdrawal, result.TransactionType)
			},
		},
		{
			name: "topup type should set type to merchant topup and source fields",
			request: UpsertBalanceHistoryRequest{
				Event: &cdcModel.Event[cdcModel.AccountTransaction]{
					Op:   cdcModel.OpUpdate,
					TsMs: now.UnixMilli(),
					After: &cdcModel.AccountTransaction{
						UUID:       "tx-123",                 // NOSONAR
						MerchantID: "merchant-456",           // NOSONAR
						Credit:     decimal.NewFromInt(1000), // NOSONAR
						Type:       constant.TypeTopUp,
						Status:     constant.StatusSuccess,
						CreatedAt:  createdAt, UpdatedAt: now,
						Reference: util.ValueToPtr(constant.TypeDisbursement),
					},
				},
			},
			validate: func(t *testing.T, result BalanceHistory) {
				assert.Equal(t, constant.TypeMerchantTopUp, result.Type)
				assert.Equal(t, *result.SourceCreatedAt, createdAt)
				assert.Equal(t, "System", result.SourceCreatedBy)
			},
		},
		{
			name: "merchant topup type should set source fields",
			request: UpsertBalanceHistoryRequest{
				Event: &cdcModel.Event[cdcModel.AccountTransaction]{
					Op:   cdcModel.OpUpdate,
					TsMs: now.UnixMilli(),
					After: &cdcModel.AccountTransaction{
						UUID:       "tx-123",                 // NOSONAR
						MerchantID: "merchant-456",           // NOSONAR
						Credit:     decimal.NewFromInt(1000), // NOSONAR
						Type:       constant.TypeMerchantTopUp,
						Status:     constant.StatusSuccess,
						CreatedAt:  createdAt, UpdatedAt: now,
						Reference: util.ValueToPtr(constant.TypeDisbursement),
					},
				},
			},
			validate: func(t *testing.T, result BalanceHistory) {
				assert.Equal(t, constant.TypeMerchantTopUp, result.Type)
				assert.Equal(t, *result.SourceCreatedAt, createdAt)
				assert.Equal(t, "System", result.SourceCreatedBy)
			},
		},
		{
			name: "fee type with non-wallet reference should set transaction type and source fields",
			request: UpsertBalanceHistoryRequest{
				Event: &cdcModel.Event[cdcModel.AccountTransaction]{
					Op:   cdcModel.OpUpdate,
					TsMs: now.UnixMilli(),
					After: &cdcModel.AccountTransaction{
						UUID:       "tx-123",                // NOSONAR
						MerchantID: "merchant-456",          // NOSONAR
						Debit:      decimal.NewFromInt(100), // NOSONAR
						Type:       constant.TypeFee,
						Status:     constant.StatusSuccess,
						CreatedAt:  createdAt, UpdatedAt: now,
						Reference: util.ValueToPtr(constant.TypeDisbursement),
						AdditionalInfo: cdcModel.AccountTransactionAdditionalInfo{
							Type:  constant.ReferenceDisbursement,
							Notes: "ON-BEHALF", // NOSONAR
						},
					},
				},
			},
			validate: func(t *testing.T, result BalanceHistory) {
				assert.Equal(t, constant.ReferenceDisbursement+"_FEE", result.TransactionType)
				assert.Equal(t, *result.SourceCreatedAt, createdAt)
				assert.Equal(t, "System", result.SourceCreatedBy)
			},
		},
		{
			name: "fee type with wallet reference should not modify transaction type",
			request: UpsertBalanceHistoryRequest{
				Event: &cdcModel.Event[cdcModel.AccountTransaction]{
					Op:   cdcModel.OpUpdate,
					TsMs: now.UnixMilli(),
					After: &cdcModel.AccountTransaction{
						UUID:       "tx-123",
						MerchantID: "merchant-456",
						Debit:      decimal.NewFromInt(100),
						Type:       constant.TypeFee,
						Status:     constant.StatusSuccess,
						CreatedAt:  createdAt, UpdatedAt: now,
						Reference: util.ValueToPtr(constant.ReferenceWallet),
						AdditionalInfo: cdcModel.AccountTransactionAdditionalInfo{
							Type: constant.ReferenceWallet,
						},
					},
				},
			},
			validate: func(t *testing.T, result BalanceHistory) {
				assert.Equal(t, constant.TypeFee, result.TransactionType)
			},
		},
		{
			name: "empty channel should use additional info reference type",
			request: UpsertBalanceHistoryRequest{
				Event: &cdcModel.Event[cdcModel.AccountTransaction]{
					Op:   cdcModel.OpUpdate,
					TsMs: now.UnixMilli(),
					After: &cdcModel.AccountTransaction{
						UUID:       "tx-123",                // NOSONAR
						MerchantID: "merchant-456",          // NOSONAR
						Debit:      decimal.NewFromInt(100), // NOSONAR
						Type:       constant.TypeFee,
						Status:     constant.StatusSuccess,
						CreatedAt:  now,
						UpdatedAt:  now,
						Reference:  util.ValueToPtr(constant.TypeWallet),
						AdditionalInfo: cdcModel.AccountTransactionAdditionalInfo{
							ReferenceType: "EWALLET", // NOSONAR
						},
					},
				},
			},
			validate: func(t *testing.T, result BalanceHistory) {
				assert.Equal(t, "EWALLET", result.Channel) // NOSONAR
			},
		},
		{
			name: "fee detail should set fee amount",
			request: UpsertBalanceHistoryRequest{
				Event: &cdcModel.Event[cdcModel.AccountTransaction]{
					Op:   cdcModel.OpUpdate,
					TsMs: now.UnixMilli(),
					After: &cdcModel.AccountTransaction{
						UUID:       "tx-123",                 // NOSONAR
						MerchantID: "merchant-456",           // NOSONAR
						Credit:     decimal.NewFromInt(1000), // NOSONAR
						Type:       constant.TypePayment,
						Status:     constant.StatusSuccess,
						CreatedAt:  now, UpdatedAt: now,
						Reference: util.ValueToPtr(constant.TypePayment),
						AdditionalInfo: cdcModel.AccountTransactionAdditionalInfo{
							FeeDetail: &cdcModel.FeeDetail{
								FinalAmount: 150.50, // NOSONAR
							},
						},
					},
				},
			},
			validate: func(t *testing.T, result BalanceHistory) {
				assert.Equal(t, decimal.NewFromFloat(150.50), result.Fee)
			},
		},
		{
			name: "nil settlement status should default to success",
			request: UpsertBalanceHistoryRequest{
				Event: &cdcModel.Event[cdcModel.AccountTransaction]{
					Op:   cdcModel.OpUpdate,
					TsMs: now.UnixMilli(),
					After: &cdcModel.AccountTransaction{
						UUID:       "tx-123",                 // NOSONAR
						MerchantID: "merchant-456",           // NOSONAR
						Credit:     decimal.NewFromInt(1000), // NOSONAR
						Type:       constant.TypePayment,
						Channel:    "CARD", // NOSONAR
						Status:     constant.StatusSuccess,
						CreatedAt:  createdAt, UpdatedAt: now,
						Reference: util.ValueToPtr(constant.TypePayment),
					},
				},
			},
			validate: func(t *testing.T, result BalanceHistory) {
				assert.Equal(t, constant.StatusSuccess, result.SettlementStatus)
				assert.Equal(t, now, result.SettlementAt)
			},
		},
		{
			name: "settlement status from payload should be used",
			request: UpsertBalanceHistoryRequest{
				Event: &cdcModel.Event[cdcModel.AccountTransaction]{
					Op:   cdcModel.OpUpdate,
					TsMs: now.UnixMilli(),
					After: &cdcModel.AccountTransaction{
						UUID:             "tx-123",                 // NOSONAR
						MerchantID:       "merchant-456",           // NOSONAR
						Credit:           decimal.NewFromInt(1000), // NOSONAR
						Type:             constant.TypePayment,
						Status:           constant.StatusSuccess,
						SettlementStatus: util.ValueToPtr("PENDING"),
						CreatedAt:        createdAt, UpdatedAt: now,
						Reference: util.ValueToPtr(constant.TypePayment),
					},
				},
			},
			validate: func(t *testing.T, result BalanceHistory) {
				assert.Equal(t, "PENDING", result.SettlementStatus)
				assert.Equal(t, now, result.SettlementAt)
			},
		},
		{
			name: "settlement at from payload should be used when not zero",
			request: UpsertBalanceHistoryRequest{
				Event: &cdcModel.Event[cdcModel.AccountTransaction]{
					Op:   cdcModel.OpUpdate,
					TsMs: now.UnixMilli(),
					After: &cdcModel.AccountTransaction{
						UUID:         "tx-123",                 // NOSONAR
						MerchantID:   "merchant-456",           // NOSONAR
						Credit:       decimal.NewFromInt(1000), // NOSONAR
						Type:         constant.TypePayment,
						Status:       constant.StatusSuccess,
						SettlementAt: &settlementAt,
						CreatedAt:    createdAt,
						UpdatedAt:    now,
						Reference:    util.ValueToPtr(constant.TypePayment),
					},
				},
			},
			validate: func(t *testing.T, result BalanceHistory) {
				assert.Equal(t, settlementAt, result.SettlementAt)
				assert.Equal(t, constant.StatusSuccess, result.SettlementStatus)
			},
		},
		{
			name: "settlement at from estimate when payload is zero",
			request: UpsertBalanceHistoryRequest{
				Event: &cdcModel.Event[cdcModel.AccountTransaction]{
					Op:   cdcModel.OpUpdate,
					TsMs: now.UnixMilli(),
					After: &cdcModel.AccountTransaction{
						UUID:         "tx-123",                 // NOSONAR
						MerchantID:   "merchant-456",           // NOSONAR
						Credit:       decimal.NewFromInt(1000), // NOSONAR
						Type:         constant.TypePayment,
						Status:       constant.StatusSuccess,
						SettlementAt: &time.Time{},
						CreatedAt:    createdAt, UpdatedAt: now,
						Reference: util.ValueToPtr(constant.TypePayment),
						AdditionalInfo: cdcModel.AccountTransactionAdditionalInfo{
							SettlementDetail: &cdcModel.SettlementDetail{
								EstimateSettlementAt: &settlementAt,
							},
						},
					},
				},
			},
			validate: func(t *testing.T, result BalanceHistory) {
				assert.Equal(t, settlementAt, result.SettlementAt)
				assert.Equal(t, constant.StatusSuccess, result.SettlementStatus)
			},
		},
		{
			name: "settlement at from T+ days when estimate is nil",
			request: UpsertBalanceHistoryRequest{
				Event: &cdcModel.Event[cdcModel.AccountTransaction]{
					Op:   cdcModel.OpUpdate,
					TsMs: now.UnixMilli(),
					After: &cdcModel.AccountTransaction{
						UUID:         "tx-123",                 // NOSONAR
						MerchantID:   "merchant-456",           // NOSONAR
						Credit:       decimal.NewFromInt(1000), // NOSONAR
						Type:         constant.TypePayment,
						Status:       constant.StatusSuccess,
						SettlementAt: &time.Time{},
						CreatedAt:    createdAt, UpdatedAt: now,
						Reference: util.ValueToPtr(constant.TypePayment),
						AdditionalInfo: cdcModel.AccountTransactionAdditionalInfo{
							SettlementDetail: &cdcModel.SettlementDetail{Type: "T+2"},
						},
					},
				},
			},
			validate: func(t *testing.T, result BalanceHistory) {
				assert.Equal(t, now.AddDate(0, 0, 2), result.SettlementAt)
			},
		},
		{
			name: "settlement at defaults to updated at when no other info",
			request: UpsertBalanceHistoryRequest{
				Event: &cdcModel.Event[cdcModel.AccountTransaction]{
					Op:   cdcModel.OpUpdate,
					TsMs: now.UnixMilli(),
					After: &cdcModel.AccountTransaction{
						UUID:         "tx-123",                 // NOSONAR
						MerchantID:   "merchant-456",           // NOSONAR
						Credit:       decimal.NewFromInt(1000), // NOSONAR
						Type:         constant.TypePayment,
						Status:       constant.StatusSuccess,
						SettlementAt: &time.Time{},
						CreatedAt:    createdAt, UpdatedAt: now,
						Reference: util.ValueToPtr(constant.TypePayment),
					},
				},
			},
			validate: func(t *testing.T, result BalanceHistory) {
				assert.Equal(t, now, result.SettlementAt)
			},
		},
		{
			name: "auto split parent payment should update the amount from metadata",
			request: UpsertBalanceHistoryRequest{
				Event: &cdcModel.Event[cdcModel.AccountTransaction]{
					Op:   cdcModel.OpUpdate,
					TsMs: now.UnixMilli(),
					After: &cdcModel.AccountTransaction{
						UUID:         "tx-123",              // NOSONAR
						MerchantID:   "merchant-456",        // NOSONAR
						Credit:       decimal.NewFromInt(0), // NOSONAR
						Type:         constant.TypePayment,
						Status:       constant.StatusSuccess,
						SettlementAt: &time.Time{},
						CreatedAt:    createdAt, UpdatedAt: now,
						Reference: util.ValueToPtr(constant.TypePayment),
						AdditionalInfo: cdcModel.AccountTransactionAdditionalInfo{
							SubPaymentSummary: &orchestrator_model.MetadataSubPaymentSummary{
								TotalCreditAmount: decimal.NewFromInt(11000000),
							},
						},
					},
				},
			},
			validate: func(t *testing.T, result BalanceHistory) {
				assert.Equal(t, decimal.NewFromInt(11000000), result.Amount) // NOSONAR
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.request.ToCreateBalanceHistory()
			tt.validate(t, result)
		})
	}
}
