package splitRoutingPaymentModel

import "time"

type SplitRoutingPaymentDetailResponse struct {
	PaymentID             string    `json:"paymentId"`
	ClientReferenceID     string    `json:"clientReferenceId"`
	Currency              string    `json:"currency"`
	Amount                float64   `json:"amount"`
	Remarks               string    `json:"remarks"`
	SourceMerchantID      string    `json:"sourceMerchantId"`
	DestinationMerchantID string    `json:"destinationMerchantId"`
	CreatedAt             time.Time `json:"createdAt"`
	UpdatedAt             time.Time `json:"updatedAt"`
	TransferID            string    `json:"transferId"`
}
