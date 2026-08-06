package splitRoutingPaymentModel

type PaymentSplitRoutingConfiguration struct {
	MerchantId       string  `json:"merchantId" validate:"required,uuid"`
	Type             string  `json:"type" validate:"required,oneof=PERCENTAGE FIXED"`
	Currency         string  `json:"currency" validate:"required"`
	PercentageAmount float64 `json:"percentageAmount" validate:"required_if=Type PERCENTAGE"`
	FixedAmount      float64 `json:"fixedAmount" validate:"required_if=Type FIXED"`
	Remarks          string  `json:"remarks" validate:"required"`

	TransferID string `json:"transferId,omitempty"`
}
