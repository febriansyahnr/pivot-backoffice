package constant

import (
	"strings"
	"time"

	paymentConstant "github.com/paper-indonesia/pivot-backoffice/constant/payment"
)

const (
	UnifiedPaymentModeRedirect = "REDIRECT"
	UnifiedPaymentModeAPI      = "API"
)

const (
	UnifiedPaymentTypeSingle             = "SINGLE"
	UnifiedPaymentTypeMultiple           = "MULTIPLE"
	UnifiedPaymentTypeSubPayment         = "SUB_PAYMENT"
	UnifiedPaymentOneDollarAuthorization = "ONE_DOLLAR_AUTHORIZATION"
)

const (
	UnifiedPaymentUseCaseCardFundedPayoutSavedCards = "CARD_FUNDED_PAYOUT_SAVED_CARDS"
)

const (
	DashboardPaymentLinkMinAmount = 10000
	DashboardPaymentLinkMaxAmount = 9999999999
)

const (
	UnifiedPaymentOneDollarAuthorizationAmount = 10_000.00
)

const (
	UnifiedPaymentSessionStatusRequirePaymentMethod = "REQUIRE_PAYMENT_METHOD"
	UnifiedPaymentSessionStatusRequireConfirmation  = "REQUIRE_CONFIRMATION"
	UnifiedPaymentSessionStatusRequireAction        = "REQUIRE_ACTION"
	UnifiedPaymentSessionStatusProcessing           = "PROCESSING"
	UnifiedPaymentSessionStatusCancelled            = "CANCELLED"
	UnifiedPaymentSessionStatusExpired              = "EXPIRED"
	UnifiedPaymentSessionStatusPaid                 = "PAID"

	// PaymentStatusRefunded is a virtual status for filtering payments that have refunds
	PaymentStatusRefunded = "REFUNDED"

	UnifiedStaticPaymentStatusActive   = "ACTIVE"
	UnifiedStaticPaymentStatusInactive = "INACTIVE"

	UnifiedPaymentMaxMetadataLength = 512
)

const (
	ChargeStatusWaitingForUserAction     = "WAITING_FOR_USER_ACTION"
	ChargeStatusWaitingForAuthentication = "WAITING_FOR_AUTHENTICATION"
	ChargeStatusWaitingForExternalFDS    = "WAITING_FOR_EXTERNAL_FDS"
	ChargeStatusProcessing               = "PROCESSING"
	ChargeStatusWaitingForCapture        = "WAITING_FOR_CAPTURE"
	ChargeStatusFailed                   = "FAILED"
	ChargeStatusExpired                  = "EXPIRED"
	ChargeStatusSuccess                  = "SUCCESS"
)

const (
	AutoSplitStatusProcessing     = "PROCESSING"
	AutoSplitStatusPartialSuccess = "PARTIAL_SUCCESS"
	AutoSplitStatusSuccess        = "SUCCESS"
	AutoSplitStatusFailed         = "FAILED"
	AutoSplitStatusCancelled      = "CANCELLED"
)

const (
	InquiryStatusPending = "PENDING"
	InquiryStatusSuccess = "SUCCESS"
	InquiryStatusFailed  = "FAILED"
	InquiryStatusExpired = "EXPIRED"
)

const (
	FailureCodeDeclinedByChannel     = "DECLINED_BY_CHANNEL"
	FailureCodeInvalidAccount        = "INVALID_ACCOUNT"
	FailureCodeAuthenticationFailed  = "AUTHENTICATION_FAILED"
	FailureCodeSuspectedFraud        = "SUSPECTED_FRAUD"
	FailureCodeBlockedByFDS          = "BLOCKED_BY_FDS" // Pivot FDS
	FailureCodeRequireReview         = "REQUIRE_REVIEW" // Pivot FDS
	FailureCodeInsufficientFund      = "INSUFFICIENT_FUND"
	FailureCodeChannelUnavailable    = "CHANNEL_UNAVAILABLE"
	FailureCodeCancelledByUser       = "CANCELLED_BY_USER"
	FailureCodeChargeExpired         = "CHARGE_EXPIRED"
	FailureCodeExceededCapturePeriod = "EXCEEDED_CAPTURE_PERIOD"
	FailureCodeUnknown               = "UNKNOWN"
)

const (
	UnifiedPaymentMethodVA      = "VIRTUAL_ACCOUNT"
	UnifiedPaymentMethodQris    = "QR"
	UnifiedPaymentMethodCard    = "CARD"
	UnifiedPaymentMethodEWallet = "EWALLET"
)

const (
	UnifiedPaymentEWalletDanaAcquirer      = "DANA"
	UnifiedPaymentEWalletShopeePayAcquirer = "SHOPEEPAY"

	EWalletDanaMaxExpiryTime      = 30 * time.Minute
	EWalletShopeePayMaxExpiryTime = 5 * 24 * time.Hour // 5 days
)

const (
	StoredPaymentMethodStatusActive   = "ACTIVE"
	StoredPaymentMethodStatusInactive = "INACTIVE"
)

const (
	UnifiedPaymentExpirationModeLoose  = "LOOSE"
	UnifiedPaymentExpirationModeStrict = "STRICT"
)

const (
	UnifiedPaymentCardCaptureMethodAutomatic = "AUTOMATIC"
	UnifiedPaymentCardCaptureMethodManual    = "MANUAL"
)

const (
	LockKeyAutoSplitPaymentCompleteKey   = "backend-portal:locks:auto-split-payment:complete:%s"     // $1: parent payment id
	LockKeyUnifiedPaymentNotificationKey = "backend-portal:locks:unified-payment:notification:%s:%s" // $1: payment id, $2: status
)

const (
	LockKeyUnifiedPaymentNotificationExpiry = time.Minute * 1
)

const (
	AutoSplitPaymentStatusSuccess        = "SUCCESS"
	AutoSplitPaymentStatusProcessing     = "PROCESSING"
	AutoSplitPaymentStatusFailed         = "FAILED"
	AutoSplitPaymentStatusCancelled      = "CANCELLED"
	AutoSplitPaymentStatusPartialSuccess = "PARTIAL_SUCCESS"
)

func MapChargeStatusToLedgerStatus(status string) (ledgerStatus string) {
	switch status {
	case ChargeStatusWaitingForUserAction, ChargeStatusWaitingForAuthentication, ChargeStatusProcessing, ChargeStatusWaitingForCapture:
		return StatusPending
	case ChargeStatusSuccess:
		return StatusSuccess
	case ChargeStatusFailed, ChargeStatusExpired:
		return StatusFailed
	}

	return StatusPending
}

func MapUnifiedPaymentMethod(unifiedPaymentMethod string) string {
	switch unifiedPaymentMethod {
	case UnifiedPaymentMethodVA:
		return paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT
	case UnifiedPaymentMethodQris:
		return paymentConstant.PAYMENT_METHOD_QRIS
	case UnifiedPaymentMethodCard:
		return paymentConstant.PAYMENT_METHOD_CREDIT_CARD
	case UnifiedPaymentMethodEWallet:
		return paymentConstant.PAYMENT_METHOD_EWALLET
	}

	return ""
}

func MapToUnifiedPaymentMethod(paymentMethod string) string {
	switch paymentMethod {
	case paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT:
		return UnifiedPaymentMethodVA
	case paymentConstant.PAYMENT_METHOD_QRIS:
		return UnifiedPaymentMethodQris
	case paymentConstant.PAYMENT_METHOD_CREDIT_CARD:
		return UnifiedPaymentMethodCard
	}

	return ""
}

func MapProcessorToChargeStatus(processorStatus, processorTransactionType string) string {
	processorStatus = strings.ToUpper(processorStatus)

	switch processorStatus {
	case paymentConstant.QrisStatusSuccess, paymentConstant.VirtualAccountStatusPaid:
		return ChargeStatusSuccess
	case paymentConstant.QrisStatusFailed, UnifiedPaymentSessionStatusCancelled, paymentConstant.PAYMENT_STATUS_VOID:
		return ChargeStatusFailed
	case paymentConstant.UnifiedPaymentStatusProcessing:
		if processorTransactionType == CardTransactionTypeAuthorization {
			return ChargeStatusWaitingForCapture
		}
		return ChargeStatusProcessing
	case ChargeStatusExpired:
		return ChargeStatusExpired
	case CreditCardStatusWaitingForAuthentication:
		return ChargeStatusWaitingForAuthentication
	}

	return processorStatus
}

func MapUnifiedPaymentStatusToV1(v2Status string) string {
	switch v2Status {
	case UnifiedPaymentSessionStatusRequirePaymentMethod,
		UnifiedPaymentSessionStatusRequireConfirmation,
		UnifiedPaymentSessionStatusRequireAction,
		UnifiedPaymentSessionStatusProcessing:
		return paymentConstant.UnifiedPaymentStatusWaitingForPayment
	case UnifiedPaymentSessionStatusExpired:
		return paymentConstant.UnifiedPaymentStatusExpired
	case UnifiedPaymentSessionStatusCancelled:
		return paymentConstant.UnifiedPaymentStatusFailed
	case UnifiedPaymentSessionStatusPaid:
		return paymentConstant.UnifiedPaymentStatusSuccess
	}

	return v2Status
}
