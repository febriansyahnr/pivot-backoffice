package reportingRepository

import (
	"strings"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	orchestratorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
)

func buildConditionForBalanceHistory(filter *orchestratorModel.TransactionHistoryFilterRequest) (conditions []string, args []any) {
	// Required condition
	conditions = []string{"merchant_id = ? AND (status_updated_at BETWEEN ? AND ?)"}
	args = []any{filter.MerchantID, filter.StartSettlementDate, filter.EndSettlementDate}

	// Multi-select for balance types
	if len(filter.BalanceTypes) > 0 {
		var orConditions []string
		for _, balanceType := range filter.BalanceTypes {
			args = append(args, balanceType)
			orConditions = append(orConditions, "balance_type = ?")
		}
		conditions = append(conditions, "("+strings.Join(orConditions, " OR ")+")")
	}

	// Multi-select for transaction types
	if len(filter.TrxTypes) > 0 {
		var orConditions []string
		for _, typ := range filter.TrxTypes {
			conds, values := getBalanceHistoryTransactionTypeFilter(typ)

			args = append(args, values...)
			orConditions = append(orConditions, conds)
		}
		conditions = append(conditions, "("+strings.Join(orConditions, " OR ")+")")
	}

	// Single-select for transaction status
	if filter.Status != "" {
		args = append(args, filter.Status)
		conditions = append(conditions, "settlement_status = ?")
	}

	// Single-select for transaction id
	if filter.TransactionId != "" {
		args = append(args, filter.TransactionId, filter.TransactionId)
		conditions = append(conditions, "(transaction_id = ? OR reference_id = ?)")
	}

	return conditions, args
}

func getBalanceHistoryTransactionTypeFilter(typ string) (condition string, args []any) {
	switch typ {
	case "DISBURSEMENT":
		return "(transaction_type = ? AND channel != ?)",
			[]any{constant.TypeDisbursement, constant.ChannelXB}

	case "BULK_DISBURSEMENT":
		return "(transaction_type = ? AND channel != ?)",
			[]any{constant.TypeBulkDisbursement, constant.ChannelXB}

	case "INTERNATIONAL_PAYOUT":
		return "(transaction_type = ? AND channel = ?)",
			[]any{constant.TypeDisbursement, constant.ChannelXB}

	case "MANUAL_TOP_UP":
		return "(transaction_type = ? AND channel = ?)",
			[]any{constant.TypeManualAdjust, constant.ChannelManualTransfer}

	case "BALANCE_ADJUSTMENT":
		return "(transaction_type = ? AND channel = ?)",
			[]any{constant.TypeManualAdjust, constant.ChannelBalanceAdjustment}

	case "VA_TOP_UP":
		return "(transaction_type = ? AND channel = ?)",
			[]any{constant.TypeTopUp, constant.ChannelVirtualAccount}

	case "CUSTOMER_TOP_UP":
		return "(transaction_type = ? AND channel = ? AND balance_type = ?)",
			[]any{constant.TypeGeneralTopUp, constant.ChannelManualTransfer, constant.AccountNameWallet}

	case "VA_PAYMENT":
		return "(transaction_type = ? AND channel = ?)",
			[]any{constant.TypePayment, constant.ChannelVirtualAccount}

	case "QRIS_PAYMENT":
		return "(transaction_type = ? AND channel IN (?, ?))",
			[]any{constant.TypePayment, constant.ChannelQris, constant.ChannelQR}

	case "CARD_PAYMENT":
		return "(transaction_type = ? AND channel = ?)",
			[]any{constant.TypePayment, constant.ChannelCard}

	case "WALLET_PAYMENT":
		return "(transaction_type = ? AND channel = ?)",
			[]any{constant.TypePayment, constant.ChannelEwallet}

	}
	return "transaction_type = ?", []any{typ}
}
