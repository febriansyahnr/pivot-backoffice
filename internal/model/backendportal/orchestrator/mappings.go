package orchestrator_model

import (
	"strings"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// FormatChannelName converts channel names to their display format
func FormatChannelName(channel string) string {
	if channel == "" {
		return ""
	}
	if channel == constant.ChannelQris {
		return "QRIS"
	}
	if channel == constant.ChannelXB {
		return "XB"
	}
	if channel == constant.ChannelPPOB {
		return "PPOB"
	}
	words := strings.Split(strings.ToLower(channel), "_")
	titleCase := cases.Title(language.English)
	for i, word := range words {
		words[i] = titleCase.String(word)
	}
	return strings.Join(words, " ")
}

func TransactionTypeForUser(typ, channel string) string {
	if typ == constant.TypeTopUp {
		return "VA Top Up"

	} else if strings.ToUpper(typ) == constant.TypeDisbursement {
		return "Single Payout"

	} else if strings.ToUpper(typ) == constant.TypeBulkDisbursement {
		return "Bulk Payout"

	} else if typ == constant.TypeManualAdjust && channel == constant.ChannelManualTransfer {
		return "Manual Top Up"

	} else if typ == constant.TypeManualAdjust && channel == constant.ChannelBalanceAdjustment {
		return "Balance Adjustment"

	} else if typ == constant.TypeXB {
		return "International Payout"

	} else if typ == constant.TypeXB+"_"+constant.TypeFee {
		return "Cross Border Fee"

	} else if typ == constant.TypeDisbursement+"_"+constant.TypeFee {
		return "Payout Fee"

	} else if typ == constant.TypeRefund {
		return "Payment Refund"

	} else if typ == constant.TypePayment {
		if channel == constant.ChannelVirtualAccount {
			return "VA Payment"
		}
		if channel == constant.ChannelQris {
			return "QRIS Payment"
		}
		if channel == constant.ChannelCreditCard || channel == constant.ChannelCard {
			return "Cards Payment"
		}
	}

	// Type FEE will be handled here
	return util.ToTitle(typ)
}
