package config

type DisbursementConfig struct {
	MinAmount                             float64                        `mapstructure:"MIN_AMOUNT"`
	MaxAmount                             float64                        `mapstructure:"MAX_AMOUNT"`
	OverbookingBankMaxAmount              float64                        `mapstructure:"OVERBOOKING_BANK_MAX_AMOUNT"`
	OverbookingBankMaxAmountForCustomRule float64                        `mapstructure:"OVERBOOKING_BANK_MAX_AMOUNT_FOR_CUSTOM_RULE"`
	DailyLimitMerchant                    float64                        `mapstructure:"DAILY_LIMIT_MERCHANT"`
	DailyLimitMerchantPlatform            float64                        `mapstructure:"DAILY_LIMIT_MERCHANT_PLATFORM"`
	CutOffTimeWindow                      DisbursementCutOffTimeWindow   `mapstructure:"CUT_OFF_TIME_WINDOW"`
	RetryInquiryConfig                    DisbursementRetryInquiryConfig `mapstructure:"RETRY_INQUIRY_CONFIG"`
	BeneficiaryLimit                      DisbursementBeneficiaryLimit   `mapstructure:"BENEFICIARY_LIMIT"`
	SchedulePayoutAlertInMinute           int                            `mapstructure:"SCHEDULE_PAYOUT_ALERT_IN_MINUTE"`
}

type DisbursementBeneficiaryLimit struct {
	Amount   float64 `mapstructure:"AMOUNT"`
	Velocity int64   `mapstructure:"VELOCITY"`
}

type DisbursementCutOffTimeWindow struct {
	Enabled                       bool   `mapstructure:"ENABLED"`
	SameDay                       bool   `mapstructure:"SAME_DAY"`
	GMT                           int    `mapstructure:"GMT"`
	StartTime                     string `mapstructure:"START_TIME"`
	EndTime                       string `mapstructure:"END_TIME"`
	TimeLagForSendingReportSecond int    `mapstructure:"TIME_LAG_FOR_SENDING_REPORT_SECOND"`
	BannerShowBeforeMinute        int    `mapstructure:"BANNER_SHOW_BEFORE_MINUTE"`
	BannerStatus                  string `mapstructure:"BANNER_STATUS"`
	TransactionInfo               string `mapstructure:"TRANSACTION_INFO"`
}

type DisbursementRetryInquiryConfig struct {
	DelayTimeMinute     int `mapstructure:"DELAY_TIME_MINUTE"`
	RetryIntervalMinute int `mapstructure:"RETRY_INTERVAL_MINUTE"`
}
