package cardFundedPayoutModel

import "time"

type CardFundedPayment struct {
	ID              string    `json:"id" db:"uuid"`
	ChargeID        string    `json:"chargeId" db:"charge_id"`
	MerchantID      string    `json:"merchantId" db:"merchant_id"`
	ReferenceID     string    `json:"referenceId" db:"reference_id"`
	Currency        string    `json:"currency" db:"currency"`
	Fee             float64   `json:"fee" db:"fee"`
	Amount          float64   `json:"amount" db:"amount"`
	Count           int       `json:"count" db:"count"`
	Sequence        int       `json:"sequence" db:"sequence"`
	FirstPaymentID  string    `json:"firstPaymentId" db:"first_payment_id"`
	CardFingerprint string    `json:"cardFingerprint" db:"card_fingerprint"`
	CreatedAt       time.Time `json:"createdAt" db:"created_at"`
}

func (c *CardFundedPayment) ToCardAuthenticationRequest(vendorID, vendorName string) CardAuthenticationRequest {
	return CardAuthenticationRequest{
		PaymentID:           c.ID,
		MerchantID:          c.MerchantID,
		ClientTransactionID: c.ReferenceID,
		Amount:              c.Amount,
		Fee:                 c.Fee,
		Currency:            c.Currency,
		Card: CardAuthenticationRequestCard{
			Fingerprint: c.CardFingerprint,
		},
		CardFundedPayout: CardAuthenticationRequestCardFundedPayout{
			Count:          c.Count,
			Sequence:       c.Sequence,
			VendorID:       vendorID,
			VendorName:     vendorName,
			FirstPaymentID: c.FirstPaymentID,
		},
	}
}

type CardFundedPayoutFundingSummary struct {
	PayoutID               string  `json:"payoutId" db:"payout_id"`
	MerchantID             string  `json:"merchantId" db:"merchant_id"`
	TotalPayment           float64 `json:"totalPayment" db:"total_payment"`
	TotalWaiting           float64 `json:"totalWaiting" db:"total_waiting"`
	TotalFailed            float64 `json:"totalFailed" db:"total_failed"`
	TotalPendingSettlement float64 `json:"totalPendingSettlement" db:"total_pending_settlement"`
	TotalSuccessSettlement float64 `json:"totalSuccessSettlement" db:"total_success_settlement"`
	TotalFee               float64 `json:"totalFee" db:"total_fee"`
}
