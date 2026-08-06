package config

type PaymentSettlementConfig struct {
	CreditCard       map[string]SettlementConfigDetail `mapstructure:"CREDIT_CARD"`
	VirtualTerminal  SettlementConfigDetail            `mapstructure:"VIRTUAL_TERMINAL"`
	CardFundedPayout SettlementConfigDetail            `mapstructure:"CARD_FUNDED_PAYOUT"`
	VirtualAccount   map[string]SettlementConfigDetail `mapstructure:"VIRTUAL_ACCOUNT"`
	Qris             map[string]SettlementConfigDetail `mapstructure:"QRIS"`
	Ewallet          map[string]SettlementConfigDetail `mapstructure:"EWALLET"`
}

type SettlementConfigDetail struct {
	Type           string                       `mapstructure:"TYPE"`
	DayType        string                       `mapstructure:"DAY_TYPE"`
	CutOff         SettlementCutoffConfigDetail `mapstructure:"CUT_OFF"`
	SettlementTime string                       `mapstructure:"SETTLEMENT_TIME"`
}

type SettlementCutoffConfigDetail struct {
	Window   SettlementCutoffWindowConfigDetail   `mapstructure:"WINDOW"`
	Deferral SettlementCutoffDeferralConfigDetail `mapstructure:"DEFERRAL"`
}

type SettlementCutoffWindowConfigDetail struct {
	EndTime   string `mapstructure:"END_TIME"`
	StartTime string `mapstructure:"START_TIME"`
}

type SettlementCutoffDeferralConfigDetail struct {
	OffsetDays    int    `mapstructure:"OFFSET_DAYS"`
	ExecutionTime string `mapstructure:"EXECUTION_TIME"`
}
