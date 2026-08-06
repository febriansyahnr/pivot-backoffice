package constant

const (
	SettlementTransaction  = "SETTLEMENT_TRANSACTION"
	SettlementFeeOnly      = "SETTLEMENT_FEE_ONLY"
	SettlementTypeInstant  = "INSTANT"
	SettlementTypeStandard = "STANDARD"
)

const (
	SettlementDayTypeAnyday  = "ANYDAY"
	SettlementTypeTimePlus01 = "T+1"
)

const (
	SettlementTimeTransactionBasedPrefix  = "T+"
	SettlementTimeDayBasedPrefix          = "D+"
	SettlementTimeTransactionBasedPattern = `^T\+[1-9]\d*$`
	SettlementTimeDayBasedPattern         = `^D\+[1-9]\d*$`
)

const (
	SettlementMethodInstant  = "INSTANT"
	SettlementMethodStandard = "STANDARD"
)
