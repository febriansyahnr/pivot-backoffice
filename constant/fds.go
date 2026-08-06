package constant

const (
	FDS_PROCESSOR      = "FDS_PROCESSOR"
	PROVIDER_FRAUD_NET = "FRAUD_NET"
	PROVIDER_SOKRATECH = "SOKRATECH"
)

const (
	FDS_STATUS_PASSED        = "PASSED"
	FDS_STATUS_REVIEW        = "REVIEW"
	FDS_STATUS_REJECTED      = "REJECTED"
	FDS_STATUS_NOT_EVALUATED = "NOT_EVALUATED"
)

// Payment Method
const (
	FRAUD_NET_PAYMENT_ACH                   = "ach"
	FRAUD_NET_PAYMENT_APPLE_PAY             = "apple_pay"
	FRAUD_NET_PAYMENT_ANDROID_PAY           = "android_pay"
	FRAUD_NET_PAYMENT_BILL_ME_LATER         = "bill_me_later"
	FRAUD_NET_PAYMENT_BITCOIN               = "bitcoin"
	FRAUD_NET_PAYMENT_BPAY_UPPER            = "BPay"
	FRAUD_NET_PAYMENT_BPAY                  = "bpay"
	FRAUD_NET_PAYMENT_CASH                  = "cash"
	FRAUD_NET_PAYMENT_CHECK                 = "check"
	FRAUD_NET_PAYMENT_CREDIT_CARD           = "credit_card"
	FRAUD_NET_PAYMENT_CREDIT                = "credit"
	FRAUD_NET_PAYMENT_DIRECT_DEBIT          = "direct_debit"
	FRAUD_NET_PAYMENT_ECHECK                = "eCheck"
	FRAUD_NET_PAYMENT_EFT                   = "eft"
	FRAUD_NET_PAYMENT_EPAYMENT              = "epayment"
	FRAUD_NET_PAYMENT_GIFT_CARD             = "gift_card"
	FRAUD_NET_PAYMENT_GOOGLE_WALLET         = "google_wallet"
	FRAUD_NET_PAYMENT_INVOICE               = "invoice"
	FRAUD_NET_PAYMENT_MASTERPASS            = "masterpass"
	FRAUD_NET_PAYMENT_MONEY_ORDER           = "money_order"
	FRAUD_NET_PAYMENT_PAYPAL                = "paypal"
	FRAUD_NET_PAYMENT_REWARDS_POINTS        = "rewards_points"
	FRAUD_NET_PAYMENT_VOUCHER               = "voucher"
	FRAUD_NET_PAYMENT_THIRD_PARTY_PROCESSOR = "third_party_processor"
	FRAUD_NET_PAYMENT_OTHER                 = "other"
	FRAUD_NET_PAYMENT_ATM                   = "atm"
	FRAUD_NET_PAYMENT_DIRECT_DEPOSIT        = "direct_deposit"
	FRAUD_NET_PAYMENT_EWALLETS              = "ewallets"
	FRAUD_NET_PAYMENT_GOODS_SERVICE         = "goods_service"
	FRAUD_NET_PAYMENT_LOCAL_BANK_TRANSFERS  = "local_bank_transfers"
	FRAUD_NET_PAYMENT_PBAY                  = "pbay"
	FRAUD_NET_PAYMENT_SEPA                  = "sepa"
	FRAUD_NET_PAYMENT_WIRE_TRANSFER         = "wire_transfer"
	FRAUD_NET_PAYMENT_NPP                   = "npp"
	FRAUD_NET_PAYMENT_FASTER_PAYMENT        = "faster_payment"
	FRAUD_NET_PAYMENT_BACS                  = "bacs"
	FRAUD_NET_PAYMENT_CHAPS                 = "chaps"
	FRAUD_NET_PAYMENT_INTERNAL_TRANSFER     = "internal_transfer"
	FRAUD_NET_PAYMENT_FEE                   = "fee"
	FRAUD_NET_PAYMENT_SWIFT                 = "swift"
	FRAUD_NET_PAYMENT_DEBIT_CARD            = "debit_card"
	FRAUD_NET_PAYMENT_CARRIER_BILLING       = "carrier_billing"
	FRAUD_NET_PAYMENT_DAPI                  = "dapi"
	FRAUD_NET_PAYMENT_PAYEZ                 = "payez"
	FRAUD_NET_PAYMENT_P2P                   = "p2p"
)

// Card Type
const (
	FRAUD_NET_CARD_TYPE_AMEX        = "amex"
	FRAUD_NET_CARD_TYPE_DISCOVER    = "discover"
	FRAUD_NET_CARD_TYPE_DINERS_CLUB = "diners_club"
	FRAUD_NET_CARD_TYPE_MC          = "mc"
	FRAUD_NET_CARD_TYPE_OTHER       = "other"
	FRAUD_NET_CARD_TYPE_VISA        = "visa"
)

// Direction
const (
	FRAUD_NET_DIRECTION_IN  = "in"
	FRAUD_NET_DIRECTION_OUT = "out"
)

// Transction Type
const (
	FRAUD_NET_TRANSACTION_AUTH = "authentication"
)

// IP Type
const (
	FRAUD_NET_IPV4 = "v4"
)

// Transaction Status
const (
	FRAUD_NET_TRX_STATUS_NEW       = "new"
	FRAUD_NET_TRX_STATUS_HOLD      = "hold"
	FRAUD_NET_TRX_STATUS_QUEUED    = "queued"
	FRAUD_NET_TRX_STATUS_APPROVED  = "approved"
	FRAUD_NET_TRX_STATUS_CANCELLED = "cancelled"
	FRAUD_NET_TRX_STATUS_FULFILLED = "fulfilled"
	FRAUD_NET_TRX_STATUS_RETURNED  = "returned"
)

// Payment Status
const (
	FRAUD_NET_PAYMENT_STATUS_AUTH             = "auth"
	FRAUD_NET_PAYMENT_STATUS_PAID             = "paid"
	FRAUD_NET_PAYMENT_STATUS_PARTIAL_PAID     = "partially paid"
	FRAUD_NET_PAYMENT_STATUS_INVOICED         = "invoiced"
	FRAUD_NET_PAYMENT_STATUS_REFUNDED         = "refunded"
	FRAUD_NET_PAYMENT_STATUS_PARTIAL_REFUNDED = "partially refunded"
	FRAUD_NET_PAYMENT_STATUS_DEFAULT          = "default"
	FRAUD_NET_PAYMENT_STATUS_PARTIAL_DEFAULT  = "partially default"
	FRAUD_NET_PAYMENT_STATUS_DECLINED         = "declined"
	FRAUD_NET_PAYMENT_STATUS_CHARGEBACK       = "chargeback"
	FRAUD_NET_PAYMENT_STATUS_VOID             = "void"
)

// Card Status
const (
	FRAUD_NET_CARD_STATUS_DECLINE   = "decline"
	FRAUD_NET_CARD_STATUS_EXPIRED   = "expired"
	FRAUD_NET_CARD_STATUS_INACTIVE  = "inactive"
	FRAUD_NET_CARD_STATUS_STOLEN    = "stolen"
	FRAUD_NET_CARD_STATUS_SUSPENDED = "suspended"
)

// AdditionalInfo Keys
const (
	FdsRiskAssesment = "fdsRiskAssessment"
)

const (
	WindowUnitSecond = "SECOND"
	WindowUnitMinute = "MINUTE"
	WindowUnitHour   = "HOUR"
	WindowUnitDay    = "DAY"
)

const (
	FDSVelocityMerchantUploadPoPKeyFmt = "backend-portal:fds:velocity:merchants:%s:payments" // $1 is merchant id
)

const (
	WorkflowFDSResultApprove = "APPROVE"
	WorkflowFDSResultReject  = "REJECT"
	WorkflowFDSResultReview  = "REVIEW"
	WorkflowFDSResultError   = "ERROR"
)

const (
	FDSWorkflowNamePaymentRules = "PaymentRules"
	FDSWorkflowNamePayoutRules  = "PayoutRules"
)
