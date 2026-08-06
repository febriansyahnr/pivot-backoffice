package constant

const (
	RefundMethodAuto         = "AUTO"
	RefundMethodTransferOnly = "TRANSFER_ONLY"

	RefundStatusPending             = "PENDING"
	RefundStatusWaitingBankTransfer = "WAITING_BANK_TRANSFER"
	RefundStatusFailed              = "FAILED"
	RefundStatusSuccess             = "SUCCESS"
	RefundStatusCancelled           = "CANCELLED"

	RefundReasonRequestedByCustomer = "REQUESTED_BY_CUSTOMER"
	RefundReasonSuspectFraudulent   = "SUSPECT_FRAUDULENT"
	RefundReasonDuplicate           = "DUPLICATE"
	RefundReasonCancellation        = "CANCELLATION"
	RefundReasonOthers              = "OTHERS"

	RefundDestinationTypeChannel = "CHANNEL"
	RefundDestinationTypeAccount = "ACCOUNT"

	// Refund type for payment history summary
	RefundTypeNone    = "NONE"
	RefundTypePartial = "PARTIAL"
	RefundTypeFull    = "FULL"
)
