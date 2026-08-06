package walletTopUp

type TopUpCallbackRequest struct {
	VaNumber        string  `json:"va_number,omitempty"`
	Amount          float64 `json:"amount,omitempty"`
	Currency        string  `json:"currency,omitempty"`
	UserReferenceId string  `json:"user_reference_id,omitempty"`
	Acquirer        string  `json:"acquirer,omitempty"`
}
