package reportingRepository

import (
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	orchestratorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"

	"github.com/stretchr/testify/assert"
)

func TestBuildConditionForBalanceHistory(t *testing.T) {
	merchantID := "merchant-uuid-123"
	startDate := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2025, 1, 31, 23, 59, 59, 0, time.UTC)

	tests := []struct {
		name           string
		filter         *orchestratorModel.TransactionHistoryFilterRequest
		wantConditions []string
		wantArgs       []any
	}{
		{
			name: "SUCCESS:Only required fields",
			filter: &orchestratorModel.TransactionHistoryFilterRequest{
				MerchantID:          merchantID,
				StartSettlementDate: startDate,
				EndSettlementDate:   endDate,
			},
			wantConditions: []string{
				"merchant_id = ? AND (status_updated_at BETWEEN ? AND ?)",
			},
			wantArgs: []any{merchantID, startDate, endDate},
		},
		{
			name: "SUCCESS:With balance types",
			filter: &orchestratorModel.TransactionHistoryFilterRequest{
				MerchantID:          merchantID,
				StartSettlementDate: startDate,
				EndSettlementDate:   endDate,
				BalanceTypes:        []string{"WALLET", "HOLDING"},
			},
			wantConditions: []string{
				"merchant_id = ? AND (status_updated_at BETWEEN ? AND ?)",
				"(balance_type = ? OR balance_type = ?)",
			},
			wantArgs: []any{merchantID, startDate, endDate, "WALLET", "HOLDING"},
		},
		{
			name: "SUCCESS:With single balance type",
			filter: &orchestratorModel.TransactionHistoryFilterRequest{
				MerchantID:          merchantID,
				StartSettlementDate: startDate,
				EndSettlementDate:   endDate,
				BalanceTypes:        []string{"WALLET"},
			},
			wantConditions: []string{
				"merchant_id = ? AND (status_updated_at BETWEEN ? AND ?)",
				"(balance_type = ?)",
			},
			wantArgs: []any{merchantID, startDate, endDate, "WALLET"},
		},
		{
			name: "SUCCESS:With status",
			filter: &orchestratorModel.TransactionHistoryFilterRequest{
				MerchantID:          merchantID,
				StartSettlementDate: startDate,
				EndSettlementDate:   endDate,
				Status:              "SETTLED",
			},
			wantConditions: []string{
				"merchant_id = ? AND (status_updated_at BETWEEN ? AND ?)",
				"settlement_status = ?",
			},
			wantArgs: []any{merchantID, startDate, endDate, "SETTLED"},
		},
		{
			name: "SUCCESS:With transaction ID",
			filter: &orchestratorModel.TransactionHistoryFilterRequest{
				MerchantID:          merchantID,
				StartSettlementDate: startDate,
				EndSettlementDate:   endDate,
				TransactionId:       "TRX-123456",
			},
			wantConditions: []string{
				"merchant_id = ? AND (status_updated_at BETWEEN ? AND ?)",
				"(transaction_id = ? OR reference_id = ?)",
			},
			wantArgs: []any{merchantID, startDate, endDate, "TRX-123456", "TRX-123456"},
		},
		{
			name: "SUCCESS:With single trx type",
			filter: &orchestratorModel.TransactionHistoryFilterRequest{
				MerchantID:          merchantID,
				StartSettlementDate: startDate,
				EndSettlementDate:   endDate,
				TrxTypes:            []string{"DISBURSEMENT"},
			},
			wantConditions: []string{
				"merchant_id = ? AND (status_updated_at BETWEEN ? AND ?)",
				"((transaction_type = ? AND channel != ?))",
			},
			wantArgs: []any{merchantID, startDate, endDate, constant.TypeDisbursement, constant.ChannelXB},
		},
		{
			name: "SUCCESS:With multiple trx types",
			filter: &orchestratorModel.TransactionHistoryFilterRequest{
				MerchantID:          merchantID,
				StartSettlementDate: startDate,
				EndSettlementDate:   endDate,
				TrxTypes:            []string{"DISBURSEMENT", "VA_PAYMENT"},
			},
			wantConditions: []string{
				"merchant_id = ? AND (status_updated_at BETWEEN ? AND ?)",
				"((transaction_type = ? AND channel != ?) OR (transaction_type = ? AND channel = ?))",
			},
			wantArgs: []any{merchantID, startDate, endDate, constant.TypeDisbursement, constant.ChannelXB, constant.TypePayment, constant.ChannelVirtualAccount},
		},
		{
			name: "SUCCESS:With all filters",
			filter: &orchestratorModel.TransactionHistoryFilterRequest{
				MerchantID:          merchantID,
				StartSettlementDate: startDate,
				EndSettlementDate:   endDate,
				BalanceTypes:        []string{"WALLET"},
				TrxTypes:            []string{"DISBURSEMENT"},
				Status:              "SETTLED",
				TransactionId:       "TRX-123456",
			},
			wantConditions: []string{
				"merchant_id = ? AND (status_updated_at BETWEEN ? AND ?)",
				"(balance_type = ?)",
				"((transaction_type = ? AND channel != ?))",
				"settlement_status = ?",
				"(transaction_id = ? OR reference_id = ?)",
			},
			wantArgs: []any{merchantID, startDate, endDate, "WALLET", constant.TypeDisbursement, constant.ChannelXB, "SETTLED", "TRX-123456", "TRX-123456"},
		},
		{
			name: "SUCCESS:With unknown trx type falls back to default",
			filter: &orchestratorModel.TransactionHistoryFilterRequest{
				MerchantID:          merchantID,
				StartSettlementDate: startDate,
				EndSettlementDate:   endDate,
				TrxTypes:            []string{"UNKNOWN_TYPE"},
			},
			wantConditions: []string{
				"merchant_id = ? AND (status_updated_at BETWEEN ? AND ?)",
				"(transaction_type = ?)",
			},
			wantArgs: []any{merchantID, startDate, endDate, "UNKNOWN_TYPE"},
		},
		{
			name: "SUCCESS:Empty filter optional fields",
			filter: &orchestratorModel.TransactionHistoryFilterRequest{
				MerchantID:          merchantID,
				StartSettlementDate: startDate,
				EndSettlementDate:   endDate,
				BalanceTypes:        []string{},
				TrxTypes:            []string{},
				Status:              "",
				TransactionId:       "",
			},
			wantConditions: []string{
				"merchant_id = ? AND (status_updated_at BETWEEN ? AND ?)",
			},
			wantArgs: []any{merchantID, startDate, endDate},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conditions, args := buildConditionForBalanceHistory(tt.filter)

			assert.Equal(t, tt.wantConditions, conditions, "conditions mismatch")
			assert.Equal(t, tt.wantArgs, args, "args mismatch")
		})
	}
}

func TestGetBalanceHistoryTransactionTypeFilter(t *testing.T) {
	tests := []struct {
		name          string
		typ           string
		wantCondition string
		wantArgs      []any
	}{
		{
			name:          "DISBURSEMENT",
			typ:           "DISBURSEMENT",
			wantCondition: "(transaction_type = ? AND channel != ?)",
			wantArgs:      []any{constant.TypeDisbursement, constant.ChannelXB},
		},
		{
			name:          "BULK_DISBURSEMENT",
			typ:           "BULK_DISBURSEMENT",
			wantCondition: "(transaction_type = ? AND channel != ?)",
			wantArgs:      []any{constant.TypeBulkDisbursement, constant.ChannelXB},
		},
		{
			name:          "INTERNATIONAL_PAYOUT",
			typ:           "INTERNATIONAL_PAYOUT",
			wantCondition: "(transaction_type = ? AND channel = ?)",
			wantArgs:      []any{constant.TypeDisbursement, constant.ChannelXB},
		},
		{
			name:          "MANUAL_TOP_UP",
			typ:           "MANUAL_TOP_UP",
			wantCondition: "(transaction_type = ? AND channel = ?)",
			wantArgs:      []any{constant.TypeManualAdjust, constant.ChannelManualTransfer},
		},
		{
			name:          "BALANCE_ADJUSTMENT",
			typ:           "BALANCE_ADJUSTMENT",
			wantCondition: "(transaction_type = ? AND channel = ?)",
			wantArgs:      []any{constant.TypeManualAdjust, constant.ChannelBalanceAdjustment},
		},
		{
			name:          "VA_TOP_UP",
			typ:           "VA_TOP_UP",
			wantCondition: "(transaction_type = ? AND channel = ?)",
			wantArgs:      []any{constant.TypeTopUp, constant.ChannelVirtualAccount},
		},
		{
			name:          "CUSTOMER_TOP_UP",
			typ:           "CUSTOMER_TOP_UP",
			wantCondition: "(transaction_type = ? AND channel = ? AND balance_type = ?)",
			wantArgs:      []any{constant.TypeGeneralTopUp, constant.ChannelManualTransfer, constant.AccountNameWallet},
		},
		{
			name:          "VA_PAYMENT",
			typ:           "VA_PAYMENT",
			wantCondition: "(transaction_type = ? AND channel = ?)",
			wantArgs:      []any{constant.TypePayment, constant.ChannelVirtualAccount},
		},
		{
			name:          "QRIS_PAYMENT",
			typ:           "QRIS_PAYMENT",
			wantCondition: "(transaction_type = ? AND channel IN (?, ?))",
			wantArgs:      []any{constant.TypePayment, constant.ChannelQris, constant.ChannelQR},
		},
		{
			name:          "CARD_PAYMENT",
			typ:           "CARD_PAYMENT",
			wantCondition: "(transaction_type = ? AND channel = ?)",
			wantArgs:      []any{constant.TypePayment, constant.ChannelCard},
		},
		{
			name:          "UNKNOWN_TYPE - fallback to default",
			typ:           "UNKNOWN_TYPE",
			wantCondition: "transaction_type = ?",
			wantArgs:      []any{"UNKNOWN_TYPE"},
		},
		{
			name:          "Empty string - fallback to default",
			typ:           "",
			wantCondition: "transaction_type = ?",
			wantArgs:      []any{""},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			condition, args := getBalanceHistoryTransactionTypeFilter(tt.typ)

			assert.Equal(t, tt.wantCondition, condition)
			assert.Equal(t, tt.wantArgs, args)
		})
	}
}
