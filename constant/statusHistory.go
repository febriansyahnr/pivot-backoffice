package constant

// Labels
const (
	StatusLabelPayoutCreated = "Payout Created"
)

const (
	DisbursementStatusHistoryWaiting         = "WAITING"
	DisbursementStatusHistoryApproved        = "APPROVED"
	DisbursementStatusHistoryRejected        = "REJECTED"
	DisbursementStatusHistoryCancelled       = "CANCELLED"
	DisbursementStatusHistoryWaitingForTopUp = "WAITING_FOR_TOP_UP"
	DisbursementStatusHistoryProcessing      = "PROCESSING"
	DisbursementStatusHistorySuccess         = "SUCCESS"
	DisbursementStatusHistoryFailed          = "FAILED"
)

// Payment Status History Constants
const (
	PaymentStatusHistoryPending                = "PENDING"
	PaymentStatusHistorySuccess                = "SUCCESS"
	PaymentStatusHistoryVoid                   = "VOID"
	PaymentStatusHistoryExpired                = "EXPIRED"
	PaymentStatusHistoryCancelled              = "CANCELLED"
	PaymentStatusHistoryRequirePaymentMethod   = "REQUIRE_PAYMENT_METHOD"
	PaymentStatusHistoryRequireConfirmation    = "REQUIRE_CONFIRMATION"
	PaymentStatusHistoryRequireAction          = "REQUIRE_ACTION"
	PaymentStatusHistoryProcessing             = "PROCESSING"
	PaymentStatusHistoryPaid                   = "PAID"
	PaymentStatusHistoryFailed                 = "FAILED"
	PaymentStatusHistoryInvestigationInProcess = "INVESTIGATION_IN_PROCESS"
	PaymentStatusHistoryInvestigationSuccess   = "INVESTIGATION_SUCCESS"
	PaymentStatusHistoryInvestigationFailed    = "INVESTIGATION_FAILED"
)

// Charge Status History Constants
const (
	ChargeStatusHistoryWaitingForUserAction     = "WAITING_FOR_USER_ACTION"
	ChargeStatusHistoryWaitingForAuthentication = "WAITING_FOR_AUTHENTICATION"
	ChargeStatusHistoryWaitingForExternalFDS    = "WAITING_FOR_EXTERNAL_FDS"
	ChargeStatusHistoryProcessing               = "PROCESSING"
	ChargeStatusHistoryWaitingForCapture        = "WAITING_FOR_CAPTURE"
	ChargeStatusHistoryFailed                   = "FAILED"
	ChargeStatusHistoryExpired                  = "EXPIRED"
	ChargeStatusHistorySuccess                  = "SUCCESS"
)

// Refund Status History Constants
const (
	RefundStatusHistoryPending             = "PENDING"
	RefundStatusHistoryWaitingBankTransfer = "WAITING_BANK_TRANSFER"
	RefundStatusHistoryFailed              = "FAILED"
	RefundStatusHistorySuccess             = "SUCCESS"
	RefundStatusHistoryCancelled           = "CANCELLED"
)

// Actor Constants for Status History
const (
	StatusHistoryActorUser      = "user"
	StatusHistoryActorSystem    = "system"
	StatusHistoryActorProcessor = "processor"
	StatusHistoryActorCRM       = "crm"
)

// Disbursement Reason Types for Failed Status
const (
	DisbursementReasonTypeInvalidAccount     = "INVALID_ACCOUNT"
	DisbursementReasonTypeInactiveAccount    = "INACTIVE_ACCOUNT"
	DisbursementReasonTypeDormantAccount     = "DORMANT_ACCOUNT"
	DisbursementReasonTypeFeatureUnavailable = "FEATURE_UNAVAILABLE"
	DisbursementReasonTypeBeneficiaryLimit   = "BENEFICIARY_LIMIT"
	DisbursementReasonTypeOther              = "OTHER"
	DisbursementReasonTypeBlockedByFDS       = "BLOCKED_BY_FDS"
)

// Disbursement Reason Types for Processing Status
const (
	DisbursementReasonTypeDelayed    = "DELAYED"
	DisbursementReasonTypeCutOffTime = "CUT_OFF_TIME"
)

// Disbursement Status Labels and Descriptions (without reason type)
var DisbursementStatusHistoryLabelsAndDescriptions = map[string]map[string]string{
	DisbursementStatusHistoryWaiting: {
		"label":       "Payout Created",
		"description": "Waiting for approval.",
	},
	DisbursementStatusHistoryApproved: {
		"label":       "Payout Created",
		"description": "Transaction approved.",
	},
	DisbursementStatusHistoryRejected: {
		"label":       "Payout Created",
		"description": "Transaction rejected.",
	},
	DisbursementStatusHistoryWaitingForTopUp: {
		"label":          "Waiting for Top Up",
		"description":    "Transaction could not be processed because your balance is insufficient.",
		"recommendation": "Top up your balance and retry the transaction via Need Action – Waiting for Top Up.",
	},
	DisbursementStatusHistoryCancelled: {
		"label":          "Cancelled",
		"description":    "Transaction was cancelled.",
		"recommendation": "Create a new transaction if needed.",
	},
	DisbursementStatusHistoryProcessing: {
		"label":       "Processing",
		"description": "Transaction is being processed by our bank partner.",
	},
	DisbursementStatusHistorySuccess: {
		"label":       "Success",
		"description": "Transaction has been successfully completed.",
	},
	DisbursementStatusHistoryFailed: {
		"label":       "Failed",
		"description": "Transaction failed. Please check the transaction details for more information.",
	},
}

// Disbursement Failed Status Descriptions by Reason Type
var DisbursementFailedReasonDescriptions = map[string]map[string]string{
	DisbursementReasonTypeInvalidAccount: {
		"description":    "Transaction failed because the beneficiary account number is invalid.",
		"recommendation": "Ensure the beneficiary account number is correct and active before retrying.",
	},
	DisbursementReasonTypeInactiveAccount: {
		"description":    "Transaction failed because the beneficiary account is inactive.",
		"recommendation": "Ensure the beneficiary account is correct and active before retrying.",
	},
	DisbursementReasonTypeDormantAccount: {
		"description":    "Transaction failed because the beneficiary account is dormant.",
		"recommendation": "Ensure the beneficiary account is correct and active before retrying.",
	},
	DisbursementReasonTypeFeatureUnavailable: {
		"description":    "Transaction failed because this feature is currently not allowed or unavailable.",
		"recommendation": "Contact our Customer Support for assistance.",
	},
	DisbursementReasonTypeBeneficiaryLimit: {
		"description":    "Transaction declined because the beneficiary has reached their limit.",
		"recommendation": "Contact our Customer Support for assistance.",
	},
	DisbursementReasonTypeOther: {
		"description":    "Transaction could not be created due to an unexpected issue.",
		"recommendation": "Please create a new transaction.",
	},
	DisbursementReasonTypeBlockedByFDS: {
		"description":    "Transaction blocked due to risk detected by the FDS",
		"recommendation": "Contact our Customer Support for assistance.",
	},
}

// Payment Status Labels and Descriptions
var PaymentStatusHistoryLabelsAndDescriptions = map[string]map[string]string{
	PaymentStatusHistoryPending: {
		"label":       "Payment Created",
		"description": "Payment session created and waiting for payment.",
	},
	PaymentStatusHistoryRequirePaymentMethod: {
		"label":       "Requires Payment Method",
		"description": "Payment session requires payment method selection.",
	},
	PaymentStatusHistoryRequireConfirmation: {
		"label":       "Requires Confirmation",
		"description": "Payment session requires confirmation from customer.",
	},
	PaymentStatusHistoryRequireAction: {
		"label":       "Requires Action",
		"description": "Payment session requires additional action from customer.",
	},
	PaymentStatusHistoryProcessing: {
		"label":       "Processing",
		"description": "Payment is being processed.",
	},
	PaymentStatusHistorySuccess: {
		"label":       "Success",
		"description": "Payment completed successfully.",
	},
	PaymentStatusHistoryPaid: {
		"label":       "Paid",
		"description": "Payment has been paid successfully.",
	},
	PaymentStatusHistoryVoid: {
		"label":       "Void",
		"description": "Payment has been voided.",
	},
	PaymentStatusHistoryExpired: {
		"label":       "Expired",
		"description": "Payment session has expired.",
	},
	PaymentStatusHistoryCancelled: {
		"label":       "Cancelled",
		"description": "Payment session has been cancelled.",
	},
	PaymentStatusHistoryInvestigationInProcess: {
		"label":       "Investigation In Process",
		"description": "Payment is under investigation by Ops team.",
	},
	PaymentStatusHistoryInvestigationSuccess: {
		"label":       "Investigation Completed - Success",
		"description": "Investigation completed successfully. Payment verified and resolved.",
	},
	PaymentStatusHistoryInvestigationFailed: {
		"label":       "Investigation Completed - Failed",
		"description": "Investigation completed with failure. Issue identified and documented.",
	},
}

// Charge Status Labels and Descriptions
var ChargeStatusHistoryLabelsAndDescriptions = map[string]map[string]string{
	ChargeStatusHistoryWaitingForUserAction: {
		"label":       "Waiting for User Action",
		"description": "Payment is waiting for customer action.",
	},
	ChargeStatusHistoryWaitingForAuthentication: {
		"label":       "Waiting for Authentication",
		"description": "Payment is waiting for authentication.",
	},
	ChargeStatusHistoryWaitingForExternalFDS: {
		"label":       "Waiting for External FDS",
		"description": "Payment is waiting for external fraud detection system.",
	},
	ChargeStatusHistoryProcessing: {
		"label":       "Processing",
		"description": "Payment is being processed.",
	},
	ChargeStatusHistoryWaitingForCapture: {
		"label":       "Waiting for Capture",
		"description": "Payment is waiting for capture.",
	},
	ChargeStatusHistorySuccess: {
		"label":       "Success",
		"description": "Charge completed successfully.",
	},
	ChargeStatusHistoryFailed: {
		"label":       "Failed",
		"description": "Charge has failed.",
	},
	ChargeStatusHistoryExpired: {
		"label":       "Expired",
		"description": "Charge has expired.",
	},
}

// Refund Status Labels and Descriptions
var RefundStatusHistoryLabelsAndDescriptions = map[string]map[string]string{
	RefundStatusHistoryPending: {
		"label":       "Refund Created",
		"description": "Refund has been created and is pending processing.",
	},
	RefundStatusHistoryWaitingBankTransfer: {
		"label":       "Waiting for Bank Transfer",
		"description": "Refund is waiting for bank transfer processing.",
	},
	RefundStatusHistorySuccess: {
		"label":       "Success",
		"description": "Refund completed successfully.",
	},
	RefundStatusHistoryFailed: {
		"label":       "Failed",
		"description": "Refund has failed.",
	},
	RefundStatusHistoryCancelled: {
		"label":       "Cancelled",
		"description": "Refund has been cancelled.",
	},
}

// Disbursement Processing Status Descriptions by Reason Type
var DisbursementProcessingReasonDescriptions = map[string]map[string]string{
	DisbursementReasonTypeDelayed: {
		"description":    "Transaction taking longer than expected, requires further investigation.",
		"recommendation": "Our team has escalated this to our bank partner. Please allow up to 24 hours for the final status and avoid retrying the transaction. You may also check the beneficiary's statement for updates.",
	},
	DisbursementReasonTypeCutOffTime: {
		"description":    "Transaction is held during bank partner maintenance window.",
		"recommendation": "The transaction will be automatically processed after the maintenance window ends. No action is required.",
	},
	DisbursementReasonTypeOther: {
		"description":    "Transaction is being processed by our bank partner.",
		"recommendation": "",
	},
}
