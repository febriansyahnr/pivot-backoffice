package settlementModel

import "time"

type ProcessSettlementRequest struct {
	MerchantID           string `json:"merchantId"`
	Type                 string `json:"type,omitempty"`
	TransactionID        string `json:"transactionId"`
	FeeTransactionID     string `json:"feeTransactionId,omitempty"`
	ByPassSettlementHold bool   `json:"byPassSettlementHold,omitempty"`
}

type ProcessHoldReleaseSettlementRequest struct {
	ReferenceID    string
	Action         string
	LastActionTime time.Time
}
