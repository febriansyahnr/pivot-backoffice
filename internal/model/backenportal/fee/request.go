package feeModel

type GetFeeRequest struct {
	MerchantID       string
	Reference        string
	ReferenceType    string
	PaymentMethod    string
	Channel          string
	ReferenceAmount  float64
	SettlementModel  string
	SettlementMethod string // INSTANT, STANDARD
	IsVirtualAccount bool
}

type GetPayoutTrxFeeRequest struct {
	MerchantId       string
	MerchantType     string
	ParentMerchantId string
	ChannelCode      string
	BankCode         string
	TrxAmount        float64
}

type GetTrxFeeOnBehalfRequest struct {
	MerchantId        string
	SubMerchantId     string
	Reference         string
	ReferenceType     string
	PaymentMethod     string
	TransactionAmount float64
}

type CalculateWhitelabelMerchantFeeRequest struct {
	MerchantID          string  `json:"-"`
	RequesterMerchantID string  `json:"merchantId"`
	ReferenceType       string  `json:"referenceType"`
	Amount              float64 `json:"amount"`
}

type GetTransactionFeeRequest struct {
	MerchantID    string  `json:"-"`
	ReferenceType string  `json:"referenceType"`
	Amount        float64 `json:"amount"`
}
