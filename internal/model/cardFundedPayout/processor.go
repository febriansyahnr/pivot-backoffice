package cardFundedPayoutModel

type CardAuthenticationRequest struct {
	PaymentID           string                                    `json:"payment_id"`
	MerchantID          string                                    `json:"merchant_id"`
	ClientTransactionID string                                    `json:"client_transaction_id"`
	Amount              float64                                   `json:"amount"`
	Fee                 float64                                   `json:"fee"`
	Currency            string                                    `json:"currency"`
	Card                CardAuthenticationRequestCard             `json:"card"`
	CardFundedPayout    CardAuthenticationRequestCardFundedPayout `json:"card_funded_payout"`
}

type CardAuthenticationRequestCard struct {
	Fingerprint string `json:"fingerprint"`
}

type CardAuthenticationRequestCardFundedPayout struct {
	Count          int    `json:"count"`
	Sequence       int    `json:"sequence"`
	VendorID       string `json:"vendor_id"`
	VendorName     string `json:"vendor_name"`
	FirstPaymentID string `json:"first_payment_id,omitempty"`
}
