package card

type RecurringBillingCycle struct {
	Interval               uint8   `json:"interval"`
	IntervalUnit           string  `json:"intervalUnit"`
	Count                  uint16  `json:"count,omitempty"`
	ExpiryDate             string  `json:"expiryDate"`
	MinDaysBetweenPayments uint16  `json:"minDaysBetweenPayments"`
	MinAmountPerPayment    float64 `json:"minAmountPerPayment"`
	MaxAmountPerPayment    float64 `json:"maxAmountPerPayment"`
}
