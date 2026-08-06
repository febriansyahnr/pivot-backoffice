package constant

// Channel
const (
	ChannelBalance           = "BALANCE"
	ChannelVirtualAccount    = "VIRTUAL_ACCOUNT"
	ChannelQris              = "QRIS"
	ChannelEwallet           = "EWALLET"
	ChannelCreditCard        = "CREDIT_CARD"
	ChannelBankTransfer      = "BANK_TRANSFER"
	ChannelSlackNotifier     = "SLACK_NOTIFIER"
	ChannelManualTransfer    = "MANUAL_TRANSFER"
	ChannelBalanceAdjustment = "BALANCE_ADJUSTMENT"
	ChannelManualAction      = "MANUAL_ACTION"
	ChannelPPOB              = "PPOB"
	ChannelXB                = "XB"
	ChannelBalanceTransfer   = "BALANCE_TRANSFER"
	ChannelMerchantPayment   = "MERCHANT_PAYMENT"
	ChannelCard              = "CARD"
	ChannelInvestigation     = "INVESTIGATION"
	ChannelQR                = "QR"
)

// Type
const (
	TypePayment              = "PAYMENT"
	TypeDisbursement         = "DISBURSEMENT"
	TypeBulkDisbursement     = "BULK_DISBURSEMENT"
	TypeManualAdjust         = "MANUAL_ADJUSTMENT"
	TypeFee                  = "FEE"
	TypeAccountInquiryFee    = "ACCOUNT_INQUIRY_FEE"
	TypeWallet               = "WALLET"
	TypeWalletTopUp          = "TOP_UP_WALLET"
	TypeWalletTransfer       = "TRANSFER"
	TypeWalletWithdrawal     = "WITHDRAWAL"
	TypeWalletBillPayment    = "BILL_PAYMENT"
	TypeXB                   = "XB"
	TypeTransfer             = "TRANSFER"
	TypeWithdrawal           = "WITHDRAWAL"
	TypeGeneralTopUp         = "TOP_UP"
	TypeInvalidUsecase       = "INVALID_USECASE_THAT_RETURNS_EMPTY_STRING"
	TypeMerchantTopUp        = "MERCHANT_TOP_UP"
	TypeMerchantPayment      = "MERCHANT_PAYMENT"
	TypeVoid                 = "VOID"
	TypeTopUp                = "DISBURSEMENT_TOP_UP"
	TypeRefund               = "REFUND"
	TypeFeeRefund            = "FEE_REFUND"
	TypeFeeReversal          = "FEE_REVERSAL"
	TypeFinalFailedDeduction = "FINAL_FAILED_DEDUCTION"
	TypeVirtualTerminal      = "VIRTUAL_TERMINAL"
	TypeCashback             = "CASHBACK"
	TypeCardFundedPayout     = "CARD_FUNDED_PAYOUT"
	TypePaymentFundedPayout  = "PAYMENT_FUNDED_PAYOUT"
	TypeSplitPayment         = "SPLIT_PAYMENT"
)

// For internal use only
const (
	TypeReversal        = "REVERSAL"
	TypeDisbursementFee = "DISBURSEMENT_FEE"
)

const (
	ReferencePayment             = "PAYMENT"
	ReferenceCharge              = "CHARGE"
	ReferenceDisbursement        = "DISBURSEMENT"
	ReferenceDisbursementVA      = "DISBURSEMENT_VA"
	ReferencePlatform            = "PLATFORM"
	ReferenceAccountInquiry      = "ACCOUNT_INQUIRY"
	ReferenceWallet              = "WALLET"
	ReferencePlatformActivity    = "PLATFORM_ACTIVITY"
	ReferencePlatformTransfer    = "PLATFORM_TRANSFER"
	ReferencePlatformTransaction = "PLATFORM_TRANSACTION"
	ReferenceWithdrawal          = "WITHDRAWAL"
	ReferenceTopUp               = "TOP_UP"
	ReferenceRefund              = "REFUND"
	ReferenceXB                  = "XB"
	ReferenceVirtualTerminal     = AccountNameVirtualTerminal
	ReferencePaymentFundedPayout = "PAYMENT_FUNDED_PAYOUT"
	ReferenceSubPayment          = "SUB_PAYMENT"
)

// Transaction Status
const (
	StatusSuccess = "SUCCESS"
	StatusFailed  = "FAILED"
	StatusPending = "PENDING"
)

const (
	SettlementStatusCancelled = "CANCELLED"
	SettlementStatusPending   = "PENDING"
)

const (
	SettlementModelAggregator  = "AGGREGATOR"
	SettlementModelFacilitator = "FACILITATOR" // No longer used; replaced by DIRECT.
	SettlementModelDirect      = "DIRECT"
)

const (
	EmptyUUID = "00000000-0000-0000-0000-000000000000"
)

const (
	ReasonTypeOtherReason                    = "OTHER"
	ReasonTypeInsufficientEscrowFund         = "INSUFFICIENT_ESCROW_FUND"
	ReasonTypeBeneficiaryAccountReason       = "INVALID_ACCOUNT"
	ReasonTypeBlockedByHarsya                = "BLOCKED_BY_HARSYA"
	ReasonTypeReversal                       = "REVERSAL"
	ReasonTypeBlockedByBankWhitelisted       = "BLOCKED_BY_BANK"
	ReasonTypePayoutCutOffTime               = "CUT_OFF_TIME"
	ReasonTypeDeclinedBeneficiaryRestriction = "DECLINED_BENEFICIARY_RESTRICTION"
	ReasonTypeBankNetworkError               = "BANK_NETWORK_ERROR"
	ReasonTypeCancelMerchantPayment          = "CANCEL_MERCHANT_PAYMENT"
	ReasonTypePayoutDelayed                  = "DELAYED"
	ReasonTypeBlockedByFDS                   = "BLOCKED_BY_FDS"
	ReasonTypeWaitingManualAction            = "WAITING_MANUAL_ACTION"

	ReasonDescInvalidBeneficiaryAccount  = "Invalid account"
	ReasonDescBlockedByHarsya            = "Merchant %s operation is blocked by Harsya"
	ReasonDescTransactionVoidByProcessor = "Transaction void by processor"
	ReasonDescBlockedByBankWhitelisted   = "Transaction blocked by bank"
	ReasonDescCancelMerchantPayment      = "Merchant payment cancelled"
	ReasonDescPayoutDelayed              = "Transaction taking longer than expected, requires further investigation"
	ReasonDescPayoutCutOffTime           = "Transaction held during bank cutoff window"
	ReasonDescBlockedByFDS               = "Transaction blocked due to risk detected by the FDS"
	ReasonDescWaitingManualAction        = "Transaction requires manual intervention by operations team"
)

const (
	CurrencyIDR = "IDR"
)

const (
	TransferTypeP2P    = "P2P"
	TransferTypePayIn  = "PAY_IN"
	TransferTypePayOut = "PAY_OUT"
	TransferTypeCharge = "CHARGE"
	TransferTypeCancel = "CANCEL"
	TransferTypeRefund = "REFUND"
)

const (
	ReconStatusSuccess = "TRUE"
	ReconStatusReview  = "REVIEW"
)

const (
	EarliestUpdatedAtCacheKey = "backend-portal:account-transaction:earliest-updated-at:%s:%s" // merchantID:requestHash
)
