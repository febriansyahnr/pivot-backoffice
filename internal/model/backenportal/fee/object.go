package feeModel

import (
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
)

type FeeMetadataObject struct {
	Type          string     `json:"type"`
	ReferenceType string     `json:"referenceType"`
	Channel       string     `json:"channel,omitempty"`
	Method        string     `json:"method"`
	DeductionType string     `json:"deductionType"`
	DeductionDay  *int16     `json:"deductionDay,omitempty"`
	TrxAmount     float64    `json:"trxAmount,omitempty"`
	AmountType    string     `json:"amountType"`
	Amount        float64    `json:"amount"`
	MaxFeeAmount  *float64   `json:"maxFeeAmount,omitempty"`
	Percentage    float64    `json:"percentage"`
	TaxType       string     `json:"taxType"`
	TaxPercentage float64    `json:"taxPercentage"`
	TaxAmount     float64    `json:"taxAmount"`
	FinalAmount   float64    `json:"finalAmount"`
	CanceledAt    *time.Time `json:"canceledAt,omitempty"`
	Notes         string     `json:"notes,omitempty"`
	TransferId    string     `json:"transferId,omitempty"`
	// Internal Process
	DeductionLastDate *time.Time `json:"-"`
	IsDefaultConfig   bool       `json:"-"`
	// For fee that have a parent transaction recorded in the account_transactions table
	LinkedTransactionId string `json:"linked_transaction_id,omitempty"`
	// Used for ladder tiering fee calculation, to identify the Redis key and increment value for the ladder counter.
	LadderCounterKey       string `json:"ladderCounterKey,omitempty"`
	LadderCounterIncrement int64  `json:"ladderCounterIncrement,omitempty"`
}

type ApplicableFee struct {
	AmountType    string
	Amount        float64
	Percentage    float64
	MaxFeeAmount  *float64
	TaxType       string
	TaxPercentage float64
}

type TrxFeeOnBehalfMetadata struct {
	// ParentMerchantId string  `json:"parentMerchantId,omitempty"` // Deprecated: Use the merchant.OnBehalfObject struct to store the parent merchant id information.
	Reference   string  `json:"-"`
	Type        string  `json:"type"`
	AmountType  string  `json:"amountType"`
	Amount      float64 `json:"amount"`
	Percentage  float64 `json:"percentage"`
	FinalAmount float64 `json:"finalAmount"`
}

func (f *TrxFeeOnBehalfMetadata) ToFeeResponse() FeeResponse {
	return FeeResponse{
		Type:          f.Reference,
		DeductionType: constant.MerchantFeeDeductionTypeDirect,
		AmountType:    f.AmountType,
		Amount:        f.Amount,
		Percentage:    f.Percentage,
		TaxType:       constant.MerchantTaxTypeNonPKP,
		TotalAmount:   f.FinalAmount,
	}
}
