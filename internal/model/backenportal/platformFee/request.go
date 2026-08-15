package platformFee

type PlatformFeeRequest struct {
	MerchantID  string
	Amount      float64
	ReferenceID string
	Usecase     string
}

type PlatformReversalFeeRequest struct {
	MerchantID          string
	ReferenceID         string
	ReversalReferenceID string
	Remarks             string
}
