package merchantTopUp

import "time"

type MerchantTopUpResponse struct {
	UUID            string  `json:"uuid" example:"58f61006-7847-4548-9024-507a6e05746b"`
	MerchantID      string  `json:"merchantId" example:"aef998d3-38d8-42d2-a864-4c16af9c50cf"`
	PaymentMethodID string  `json:"paymentMethodId" example:"a60ce359-d2ec-47b8-9b95-28a32541c7da"`
	ReferenceNumber string  `json:"referenceNumber" example:"reference_number"`
	CreatedAt       string  `json:"createdAt" example:"2021-01-01T00:00:00Z"`
	UpdatedAt       string  `json:"updatedAt" example:"2021-01-01T00:00:00Z"`
	Instructions    *string `json:"instructions,omitempty"`
}

// ToResponse is a function to convert DisbursementTopUp to DisbursementTopUpResponse
func (d *MerchantTopUp) ToResponse() *MerchantTopUpResponse {
	return &MerchantTopUpResponse{
		UUID:            d.ID,
		MerchantID:      d.MerchantID,
		PaymentMethodID: d.PaymentMethodID,
		ReferenceNumber: d.ReferenceNumber,
		Instructions:    d.Instructions,
		CreatedAt:       d.CreatedAt.Format(time.RFC3339),
		UpdatedAt:       d.UpdatedAt.Format(time.RFC3339),
	}
}
