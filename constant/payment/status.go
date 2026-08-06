package constant

const (
	PAYMENT_STATUS_PENDING = "PENDING"
	PAYMENT_STATUS_SUCCESS = "SUCCESS"
	PAYMENT_STATUS_VOID    = "VOID"
	PaymentStatusExpired   = "EXPIRED"
	PaymentStatusFailed    = "FAILED"

	VirtualAccountStatusPaid    = "PAID"
	VirtualAccountStatusExpired = "EXPIRED"

	QrisStatusSuccess = "SUCCESS"
	QrisStatusFailed  = "FAILED"
	QrisStatusPending = "PENDING"
	QrisStatusExpired = "EXPIRED"

	EwalletStatusSuccess = "SUCCESS"
	EwalletStatusPending = "PENDING"
	EwalletStatusFailed  = "FAILED"

	UnifiedPaymentStatusWaitingForPayment = "WAITING_FOR_PAYMENT"
	UnifiedPaymentStatusSuccess           = "SUCCESS"
	UnifiedPaymentStatusFailed            = "FAILED"
	UnifiedPaymentStatusExpired           = "EXPIRED"
	UnifiedPaymentStatusProcessing        = "PROCESSING"

	InvestigationStatusInProcess = "INVESTIGATION_IN_PROCESS"
	InvestigationStatusSuccess   = "INVESTIGATION_SUCCESS"
	InvestigationStatusFailed    = "INVESTIGATION_FAILED"
)
