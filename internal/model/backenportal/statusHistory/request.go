package statusHistoryModel

// RecordDisbursementStatusHistoryRequest represents the parameters for recording status history
type RecordDisbursementStatusHistoryRequest struct {
	DisbursementID string
	Status         string
	Actor          string
	ReasonType     string
}

// RecordPaymentStatusHistoryRequest represents the parameters for recording payment status history
type RecordPaymentStatusHistoryRequest struct {
	PaymentID  string
	Status     string
	Actor      string
	ReasonType string
}

// RecordChargeStatusHistoryRequest represents the parameters for recording charge status history
type RecordChargeStatusHistoryRequest struct {
	ChargeID   string
	Status     string
	Actor      string
	ReasonType string
}

// RecordRefundStatusHistoryRequest represents the parameters for recording refund status history
type RecordRefundStatusHistoryRequest struct {
	RefundID   string
	Status     string
	Actor      string
	ReasonType string
}
