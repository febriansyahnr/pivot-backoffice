package config

type CreditCardReferences struct {
	DefaultFee                CreditCardDefaultFee         `mapstructure:"DEFAULT_FEE"`
	DefaultVirtualTerminalFee PaymentFeeDefault            `mapstructure:"DEFAULT_VIRTUAL_TERMINAL_FEE"`
	DefaultSplitPaymentFee    PaymentFeeDefault            `mapstructure:"DEFAULT_SPLIT_PAYMENT_FEE"`
	DefaultCardFundedPayout   map[string]PaymentFeeDefault `mapstructure:"DEFAULT_CARD_FUNDED_PAYOUT_FEE"`
	CardBrands                []string                     `mapstructure:"CARD_BRANDS"`
}

type CreditCardDefaultFee struct {
	OtherChannel  CreditCardFeeConfig            `mapstructure:"OTHER_CHANNEL"`
	CustomChannel map[string]CreditCardFeeConfig `mapstructure:"CUSTOM_CHANNEL"`
}

type CreditCardFeeConfig struct {
	Amount     float64 `mapstructure:"AMOUNT"`
	Percentage float64 `mapstructure:"PERCENTAGE"`
}

type CreditcardConfig struct {
	WebviewURL   string `mapstructure:"WEBVIEW_URL"`
	ProcessorURL string `mapstructure:"PROCESSOR_URL"`
}

var creditCardReferences CreditCardReferences

func GetCreditCardReferences() CreditCardReferences {
	return creditCardReferences
}
