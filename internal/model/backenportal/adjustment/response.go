package adjustment

import (
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
)

type ManualTopupResponse struct {
	ID string `json:"id"`
}

type BalanceAdjustmentResponse struct {
	ID string `json:"id"`
}

type MerchantBalanceAdjustmentResponse struct {
	UUID        string    `json:"uuid"`
	MerchantID  string    `json:"merchantId"`
	Type        string    `json:"type"`
	Action      string    `json:"action"`
	Currency    string    `json:"currency"`
	Amount      float64   `json:"amount"`
	ReferenceID string    `json:"referenceId"`
	Status      string    `json:"status"`
	Notes       string    `json:"notes"`
	CreatedBy   string    `json:"createdBy"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

func (a *ManualAdjustmentHistory) ToMerchantBalanceAdjustmentResponse() *MerchantBalanceAdjustmentResponse {
	return &MerchantBalanceAdjustmentResponse{
		UUID:        a.UUID,
		MerchantID:  a.MerchantID,
		Type:        a.Type,
		Action:      a.Action,
		Currency:    a.Currency,
		Amount:      a.Amount,
		ReferenceID: a.ReferenceID,
		Status:      constant.StatusSuccess,
		Notes:       a.Notes,
		CreatedBy:   a.CreatedBy,
		CreatedAt:   a.CreatedAt,
		UpdatedAt:   a.UpdatedAt,
	}
}

type HoldMerchantBalanceResponse struct {
	Amount      float64 `json:"amount"`
	MerchantID  string  `json:"merchantId"`
	AccountType string  `json:"accountType"`
	Type        string  `json:"type"`
}

type GetHoldedMerchantBalanceResponse struct {
	Amount      float64 `json:"amount"`
	MerchantID  string  `json:"merchantId"`
	AccountType string  `json:"accountType"`
}
