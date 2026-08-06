package constant

import "time"

const (
	PaymentTypeSingle           = UnifiedPaymentTypeSingle
	PaymentTypeMultiple         = UnifiedPaymentTypeMultiple
	PaymentTypeVirtualTerminal  = "VIRTUAL_TERMINAL"
	PaymentTypeCardFundedPayout = "CARD_FUNDED_PAYOUT"
	PaymentTypeOneDollarAuth    = "ONE_DOLLAR_AUTHORIZATION"
)

const (
	PaymentSettlementMethodInstant  = "INSTANT"
	PaymentSettlementMethodStandard = "STANDARD"
)

const (
	PaymentSimulatorKey = "paymentSimulator"
	IsUnifiedPaymentKey = "isUnifiedPayment"
)

const (
	PaymentTokenCacheKey            = "backend-portal:payment-token:%s"
	TemporaryPaymentRecordCacheKey  = "backend-portal:payment:temporary:%s:%s" // merchant_id, payment_id
	TemporaryPaymentRecordTTL       = time.Minute * 5
	PaymentEncryptionKeyCacheKey    = "backend-portal:payment:encryption-key"
	PaymentEncryptionKeyTTL         = 24 * time.Hour
	VCCTerminalSubmitChargeCacheKey = "backend-portal:payment:vcc-terminal:charges:%s" // payment_id
	VCCTerminalSubmitChargeTTL      = 24 * time.Hour
	PaymentNotificationLockCacheKey = "backend-portal:payment:notification:%s" // payment_id
	PaymentNotificationLockTTL      = 1 * time.Minute
)

const (
	UnifiedPaymentCallbackEventPattern       = `^PAYMENT\..+$`
	UnifiedPaymentChargeCallbackEventPattern = `^CHARGE\..+$`
)

const (
	SplitRoutingPaymentConfigKey = "splitRoutingConfigurations"

	SplitRoutingPaymentTypeFixed      = "FIXED"
	SplitRoutingPaymentTypePercentage = "PERCENTAGE"
)

// constant for payment order information
const (
	ProductDetailTypePhysical = "PHYSICAL"
	ProductDetailTypeService  = "SERVICE"
	ProductDetailTypeDigital  = "DIGITAL"

	ShippingMethodRegular = "REGULAR"
	ShippingMethodNextDay = "NEXTDAY"
	ShippingMethodSameDay = "SAMEDAY"
	ShippingMethodInstant = "INSTANT"
)

// constant for payment UI banner messages default
const (
	PaymentUIAuthCaptureBannerDefaultMessage = "Your card limit is securely reserved and will only be charged once your payment is successful"
)

const (
	GroupPaymentTypePayment          = "PAYMENT"
	GroupPaymentTypeRecurringPayment = "RECURRING_PAYMENT"
	GroupPaymentTypeVirtualTerminal  = "VIRTUAL_TERMINAL"
	GroupPaymentTypeCardFundedPayout = "CARD_FUNDED_PAYOUT"
	GroupPaymentTypeOneDollarAuth    = "ONE_DOLLAR_AUTHORIZATION"
	GroupPaymentTypeSplitPayment     = "SPLIT_PAYMENT"
)
