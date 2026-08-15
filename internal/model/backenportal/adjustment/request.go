package adjustment

import (
	"mime/multipart"
)

type ManualTopupRequest struct {
	MerchantID   string                `validate:"required,uuid"`
	BankRefID    string                `validate:"required"`
	BankName     string                `validate:"required"`
	BankAccount  string                `validate:"required"`
	Currency     string                `validate:"required,oneof=IDR"`
	Amount       float64               `validate:"required,min=1"`
	CreatedBy    string                `validate:"required"`
	Notes        string                `validate:"-"`
	File         *multipart.FileHeader `validate:"-"`
	SendCallback bool                  `validate:"-"`
}

type BalanceAdjustmentRequest struct {
	MerchantID string  `json:"merchantId" validate:"required,uuid"`
	Currency   string  `json:"currency" validate:"required,oneof=IDR"`
	Amount     float64 `json:"amount" validate:"required"`
	CreatedBy  string  `json:"createdBy" validate:"required"`
	Notes      string  `json:"notes" validate:"-"`

	AdjustmentID string `json:"-"`
}

type MerchantBalanceAdjustmentRequest struct {
	MerchantId  string  `json:"merchantId" validate:"required,uuid"`
	ReferenceId string  `json:"referenceId", validate:"required"`
	BalanceType string  `json:"balanceType", validate:"required,oneof=PAYOUT_BALANCE PAYMENT_BALANCE"`
	Currency    string  `json:"currency" validate:"required,oneof=IDR"`
	Debit       float64 `json:"debit" validate:"required_if=Credit 0"`
	Credit      float64 `json:"credit" validate:"required_if=Debit 0"`
	CreatedBy   string  `json:"createdBy" validate:"required"`
	Remarks     string  `json:"remarks" validate:"required"`
}

type HoldMerchantBalanceRequest struct {
	MerchantId  string  `json:"merchantId" validate:"required,uuid"`
	AccountType string  `json:"accountType" validate:"required,oneof=PAYMENT VIRTUAL_TERMINAL WALLET"`
	Type        string  `json:"type" validate:"required,oneof=HOLD RELEASE"`
	Amount      float64 `json:"amount" validate:"required,min=1"`
}

type GetHoldedMerchantBalanceRequest struct {
	MerchantId  string `validate:"required,uuid"`
	AccountType string `validate:"required,oneof=PAYMENT VIRTUAL_TERMINAL WALLET"`
}
