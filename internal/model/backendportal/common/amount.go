package commonModel

import (
	"encoding/json"

	pb "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/backendportal/proto/common"
	"github.com/shopspring/decimal"
)

type Amount struct {
	Currency string `json:"currency" validate:"required" example:"IDR"`
	Value    string `json:"value" validate:"required" example:"1000000.00"`
}

func (d *Amount) ProtoAmount() *pb.Amount {
	if d == nil {
		return nil
	}
	return &pb.Amount{
		Currency: d.Currency,
		Value:    d.Value,
	}
}

func (d *Amount) ToJson() ([]byte, error) {
	return json.Marshal(d)
}

func (d *Amount) ToDecimal() decimal.Decimal {
	amount, _ := decimal.NewFromString(d.Value)
	return amount
}

type Amount2 struct {
	Value    float64 `json:"value"`
	Currency string  `json:"currency"`
}

type AmountRequest struct {
	Currency string  `json:"currency" validate:"required,oneof=IDR"`
	Value    float64 `json:"value" validate:"required,min=1"`
}
