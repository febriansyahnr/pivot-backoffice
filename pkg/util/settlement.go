package util

import "github.com/paper-indonesia/pivot-backoffice/constant"

func IsValidSettlementTime(settlementType string) bool {
	return IsPatternMatch(constant.SettlementTimeTransactionBasedPattern, settlementType) || IsSettlementTimeDayBased(settlementType)
}

func IsSettlementTimeDayBased(settlementType string) bool {
	return IsPatternMatch(constant.SettlementTimeDayBasedPattern, settlementType)
}
