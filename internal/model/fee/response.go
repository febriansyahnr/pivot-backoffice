package feeModel

type FeeResponse struct {
	Type          string   `json:"type"`
	Method        string   `json:"method"`
	DeductionType string   `json:"deductionType"`
	DeductionDay  *int16   `json:"deductionDay,omitempty"`
	AmountType    string   `json:"amountType"`
	Amount        float64  `json:"amount"`
	MaxFeeAmount  *float64 `json:"maxFeeAmount,omitempty"`
	Percentage    float64  `json:"percentage"`
	TaxType       string   `json:"taxType"`
	TaxPercentage float64  `json:"taxPercentage"`
	TaxAmount     float64  `json:"taxAmount"`
	TotalAmount   float64  `json:"totalAmount"`
}

func (feeDetail *FeeMetadataObject) WithTotalAmountResponse(totalAmount float64) *FeeResponse {
	return &FeeResponse{
		Type:          feeDetail.Type,
		Method:        feeDetail.Method,
		DeductionType: feeDetail.DeductionType,
		DeductionDay:  feeDetail.DeductionDay,
		AmountType:    feeDetail.AmountType,
		Amount:        feeDetail.Amount,
		MaxFeeAmount:  feeDetail.MaxFeeAmount,
		Percentage:    feeDetail.Percentage,
		TaxType:       feeDetail.TaxType,
		TaxPercentage: feeDetail.TaxPercentage,
		TaxAmount:     feeDetail.TaxAmount,
		TotalAmount:   totalAmount,
	}
}

func (f *FeeMetadataObject) ToFeeResponse() FeeResponse {
	return FeeResponse{
		Type:          f.Type,
		Method:        f.Method,
		DeductionType: f.DeductionType,
		DeductionDay:  f.DeductionDay,
		AmountType:    f.AmountType,
		Amount:        f.Amount,
		MaxFeeAmount:  f.MaxFeeAmount,
		Percentage:    f.Percentage,
		TaxType:       f.TaxType,
		TaxPercentage: f.TaxPercentage,
		TaxAmount:     f.TaxAmount,
		TotalAmount:   f.FinalAmount,
	}
}
