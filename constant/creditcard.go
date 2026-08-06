package constant

import "time"

const (
	ACQUIRER_CC_HARSYA = "harsya"
)

const (
	// Status
	CreditCardStatusWaitingForPayment        = "WAITING_FOR_PAYMENT"
	CreditCardStatusWaitingForAuthentication = "WAITING_FOR_AUTHENTICATION"
	CreditCardStatusExpired                  = "EXPIRED"
	CreditCardStatusPAID                     = "PAID"
	CreditCardStatusFailed                   = "FAILED"
	CreditCardStatusRefunded                 = "REFUNDED"
	CreditCardStatusBlocked                  = "BLOCKED"
	CreditCardStatusVoid                     = "VOID"
	CreditCardStatusSuccess                  = "SUCCESS"
	CreditCardStatusProcessing               = "PROCESSING"

	CreditCardProcessorStatusPending  = "PENDING"
	CreditCardProcessorStatusSuccess  = "SUCCESS"
	CreditCardProcessorStatusFailed   = "FAILED"
	CreditCardProcessorStatusRefunded = "REFUNDED"
	CreditCardProcessorStatusBlocked  = "BLOCKED"
	CreditCardProcessorStatusUknown   = "UKNOWN"
	CreditCardProcessorStatusExpired  = "EXPIRED"
	CreditCardProcessorStatusVoid     = "VOID"

	// Method
	CreditCardMethodChallenge    = "CHALLENGE"
	CreditCardMethodFrictionless = "FRICTIONLESS"

	// Amount
	CreditCardMinAmount = 10000

	// Timer
	CreditCardPaymentExpired = 60 * time.Minute

	CreditCardAuthenticationSuccess  = "AUTHENTICATION_SUCCESSFUL"
	CreditCardAuthenticationRequired = "AUTHENTICATION_REQUIRED"
	CreditCardAuthenticationFailed   = "AUTHENTICATION_FAILED"
	CreditCardAuthorizationFailed    = "AUTHORIZATION_FAILED"
	CreditCardGatewayCodeAborted     = "ABORTED"
)

const (
	CreditCardMidTypeAggregator = "AGGREGATOR"
	CreditCardMidTypeDirect     = "DIRECT"

	CreditCardMidTransactionTypeDirectPay   = "DIRECT_PAY"
	CreditCardMidTransactionTypeInstallment = "INSTALLMENT"

	CreditCardMidInstallmentTypeOnUs  = "ON_US"
	CreditCardMidInstallmentTypeOffUs = "OFF_US"
)

// Credit Card Reconciliation Constants
const (
	ReconCCStatusInvalid        = "INVALID"
	ReconCCStatusValid          = "VALID"
	ReconCCCodeInvalidAmount    = TReconCode("INVALID_AMOUNT")
	ReconCCCodeInvalidReference = TReconCode("INVALID_REFERENCE")
	ReconCCCodeInvalidStatus    = TReconCode("INVALID_STATUS")
	ReconCCCodeInvalidDate      = TReconCode("INVALID_DATE")
	ReconCCCodeOk               = TReconCode("OK")
)

func ChannelTypeToMidType(input string) string {
	switch input {
	case PaymentMethodChannelTypeAggregator:
		return CreditCardMidTypeAggregator
	case PaymentMethodChannelTypeDirect:
		return CreditCardMidTypeDirect
	default:
		return input // or return "" if you want to enforce only known mappings
	}
}

func MidTypeToChannelType(input string) string {
	switch input {
	case CreditCardMidTypeAggregator:
		return PaymentMethodChannelTypeAggregator
	case CreditCardMidTypeDirect:
		return PaymentMethodChannelTypeDirect
	default:
		return input // or return "" if you want to enforce only known mappings
	}
}

const (
	CardThreeDsMethodAutomatic = "AUTOMATIC"
	CardThreeDsMethodChallenge = "CHALLENGE"
	CardThreeDsMethodNever     = "NEVER"
	CardThreeDsMethodExternal  = "EXTERNAL"
)

const (
	CardTransactionTypeAuthorization = "AUTHORIZATION"
	CardTransactionTypeCapture       = "CAPTURE"
)

const (
	CardTransactionStatusAuthorized        = "AUTHORIZED"
	CardTransactionStatusPartiallyCaptured = "PARTIALLY_CAPTURED"
)

const (
	CardOriginLocal   = "LOCAL"
	CardOriginForeign = "FOREIGN"
)
