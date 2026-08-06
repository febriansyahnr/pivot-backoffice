package config

import (
	"slices"

	paymentConstant "github.com/paper-indonesia/pivot-backoffice/constant/payment"
)

type UnifiedPaymentConfig struct {
	MasterQRDuplicationID                     string                              `mapstructure:"MASTER_QR_DUPLICATION_ID"` // it just for staging only
	ChangePaymentMethodFromVA                 bool                                `mapstructure:"CHANGE_PAYMENT_METHOD_FROM_VA"`
	ChangePaymentMethodFromQris               bool                                `mapstructure:"CHANGE_PAYMENT_METHOD_FROM_QRIS"`
	ChangePaymentMethodFromCreditCard         bool                                `mapstructure:"CHANGE_PAYMENT_METHOD_FROM_CREDIT_CARD"`
	ChangePaymentMethodFromEwallet            bool                                `mapstructure:"CHANGE_PAYMENT_METHOD_FROM_EWALLET"`
	MaxAuthorizeTransactionMinutes            int                                 `mapstructure:"MAX_AUTHORIZE_TRANSACTION_MINUTES"`
	RetryExpiringAuthorizedTransactionMinutes int                                 `mapstructure:"RETRY_EXPIRING_AUTHORIZED_TRANSACTION_MINUTES"`
	VirtualAccountConfig                      *UnifiedPaymentVirtualAccountConfig `mapstructure:"VIRTUAL_ACCOUNT"`
	QrConfig                                  *UnifiedPaymentQrConfig             `mapstructure:"QR"`
	CardConfig                                *UnifiedPaymentCardConfig           `mapstructure:"CARD"`
	EwalletConfig                             *UnifiedPaymentEwalletConfig        `mapstructure:"EWALLET"`
	ExpiringProcessedBackoffMinutes           []int                               `mapstructure:"EXPIRING_PROCESSED_PAYMENT_BACKOFF_MINUTES"`
	ExpiryConfig                              *UnifiedPaymentExpiryConfig         `mapstructure:"EXPIRY_CONFIG"`
	AutoInquiryConfig                         *PaymentAutoInquiryConfig           `mapstructure:"AUTO_INQUIRY"`
}

type UnifiedPaymentVirtualAccountConfig struct {
	MinAmount                 *float64 `mapstructure:"MIN_AMOUNT"`
	MaxAmount                 *float64 `mapstructure:"MAX_AMOUNT"`
	DefaultVaConfigMerchantId string   `mapstructure:"DEFAULT_VA_CONFIG_MERCHANT_ID"`
	DefaultVaRangeStart       string   `mapstructure:"DEFAULT_VA_RANGE_START"`
	DefaultVaRangeEnd         string   `mapstructure:"DEFAULT_VA_RANGE_END"`
	// max expiry duration unit: SECONDS, MINUTES, HOURS, DAYS
	MaxExpiryDurationUnit string `mapstructure:"MAX_EXPIRY_DURATION_UNIT"`
	MaxExpiryDuration     int    `mapstructure:"MAX_EXPIRY_DURATION"`
}

type UnifiedPaymentQrConfig struct {
	MinAmount                    *float64 `mapstructure:"MIN_AMOUNT"`
	MaxAmount                    *float64 `mapstructure:"MAX_AMOUNT"`
	MaxActiveStaticQRPerMerchant int      `mapstructure:"MAX_ACTIVE_STATIC_QR_PER_MERCHANT"`
	// max expiry duration unit: SECONDS, MINUTES, HOURS, DAYS
	MaxExpiryDurationUnit string `mapstructure:"MAX_EXPIRY_DURATION_UNIT"`
	MaxExpiryDuration     int    `mapstructure:"MAX_EXPIRY_DURATION"`
}

type UnifiedPaymentCardConfig struct {
	MinAmount             *float64 `mapstructure:"MIN_AMOUNT"`
	MaxAmount             *float64 `mapstructure:"MAX_AMOUNT"`
	AcceptedChannels      []string `mapstructure:"ACCEPTED_CHANNELS"`
	MaxExpiryDuration     int      `mapstructure:"MAX_EXPIRY_DURATION"`
	MaxExpiryDurationUnit string   `mapstructure:"MAX_EXPIRY_DURATION_UNIT"`
}

type UnifiedPaymentEwalletConfig struct {
	MinAmount             *float64 `mapstructure:"MIN_AMOUNT"`
	MaxAmount             *float64 `mapstructure:"MAX_AMOUNT"`
	MaxExpiryDuration     int      `mapstructure:"MAX_EXPIRY_DURATION"`
	MaxExpiryDurationUnit string   `mapstructure:"MAX_EXPIRY_DURATION_UNIT"`
}

type UnifiedPaymentExpiryConfig struct {
	Mode              string   `mapstructure:"MODE"`
	ExcludedMerchants []string `mapstructure:"EXCLUDED_MERCHANTS"`
}

// ShouldValidateExpiry check if merchant should validate unified payment expiry
func (u *UnifiedPaymentExpiryConfig) ShouldValidateExpiry(merchantID string) bool {
	if u.Mode == paymentConstant.UnifiedPaymentExpiryModeFull {
		return true
	}

	if u.Mode == paymentConstant.UnifiedPaymentExpiryModePartial {
		return !slices.Contains(u.ExcludedMerchants, merchantID)
	}

	return false
}
