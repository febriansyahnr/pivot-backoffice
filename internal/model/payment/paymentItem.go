package paymentModel

import (
	"encoding/json"
	"github.com/shopspring/decimal"
	"time"
)

type PaymentItem struct {
	UUID        string          `json:"uuid"`
	PaymentID   string          `json:"paymentId"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Qty         int             `json:"qty"`
	Currency    string          `json:"currency"`
	Amount      decimal.Decimal `json:"amount"`
	TotalAmount decimal.Decimal `json:"totalAmount"`
	Metadata    *map[string]any `json:"metadata"`
	CreatedAt   time.Time       `json:"createdAt"`
	UpdatedAt   time.Time       `json:"updatedAt"`
	DeletedAt   *time.Time      `json:"deletedAt"`
}

type PaymentItemDTO struct {
	UUID        string          `json:"uuid" db:"uuid"`
	PaymentID   string          `json:"paymentId" db:"payment_id"`
	Name        string          `json:"name" db:"name"`
	Description string          `json:"description" db:"description"`
	Qty         int             `json:"qty" db:"qty"`
	Currency    string          `json:"currency" db:"currency"`
	Amount      decimal.Decimal `json:"amount" db:"amount"`
	TotalAmount decimal.Decimal `json:"totalAmount" db:"total_amount"`
	Metadata    *string         `json:"metadata" db:"metadata"`
	CreatedAt   time.Time       `json:"createdAt" db:"created_at"`
	UpdatedAt   time.Time       `json:"updatedAt" db:"updated_at"`
	DeletedAt   *time.Time      `json:"deletedAt" db:"deleted_at"`
}

func (a *PaymentItem) PaymentItemFromDTO(dto *PaymentItemDTO) {
	a.UUID = dto.UUID
	a.PaymentID = dto.PaymentID
	a.Name = dto.Name
	a.Description = dto.Description
	a.Qty = dto.Qty
	a.Currency = dto.Currency
	a.Amount = dto.Amount
	a.TotalAmount = dto.TotalAmount
	a.CreatedAt = dto.CreatedAt
	a.UpdatedAt = dto.UpdatedAt
	a.DeletedAt = dto.DeletedAt

	if dto.Metadata != nil {
		var metadata map[string]interface{}
		json.Unmarshal([]byte(*dto.Metadata), &metadata)
		a.Metadata = &metadata
	}
}

func (a *PaymentItemDTO) ToPaymentItem() *PaymentItem {
	return &PaymentItem{
		UUID:        a.UUID,
		PaymentID:   a.PaymentID,
		Name:        a.Name,
		Description: a.Description,
		Qty:         a.Qty,
		Currency:    a.Currency,
		Amount:      a.Amount,
		TotalAmount: a.TotalAmount,
		CreatedAt:   a.CreatedAt,
		UpdatedAt:   a.UpdatedAt,
		DeletedAt:   a.DeletedAt,
	}
}

func (a *PaymentItem) ToPaymentResponseItem() *PaymentResponseItem {
	return &PaymentResponseItem{
		ItemID:      a.UUID,
		Name:        a.Name,
		Description: a.Description,
		Amount: Amount{
			Value:    a.Amount,
			Currency: a.Currency,
		},
		Qty: a.Qty,
	}
}
