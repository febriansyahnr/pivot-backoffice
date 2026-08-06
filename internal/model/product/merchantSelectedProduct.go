package product

import (
	"time"

	"github.com/google/uuid"
)

type MerchantSelectedProduct struct {
	UUID       string    `json:"uuid" db:"uuid"`
	MerchantID string    `json:"merchantId" db:"merchant_id"`
	ProductID  string    `json:"productId" db:"product_id"`
	Active     bool      `json:"active" db:"active"`
	CreatedAt  time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt  time.Time `json:"updatedAt" db:"updated_at"`
}

type AddMerchantProductRequest struct {
	MerchantID string `json:"merchantId" validate:"required,uuid"`
	ProductID  string `json:"productId" validate:"required,uuid"`
}

func NewMerchantSelectedProduct(req *AddMerchantProductRequest) *MerchantSelectedProduct {
	return &MerchantSelectedProduct{
		UUID:       uuid.NewString(),
		MerchantID: req.MerchantID,
		ProductID:  req.ProductID,
		Active:     true,
		CreatedAt:  time.Now().UTC(),
		UpdatedAt:  time.Now().UTC(),
	}
}

type UpdateMerchantSelectedProductAvailabilityRequest struct {
	MerchantID string `json:"merchantId" validate:"required,uuid"`
	ProductID  string `json:"productId" validate:"required,uuid"`
	Active     bool   `json:"active"`
}

type MerchantWithProductName struct {
	ProductID   string    `json:"productId" db:"product_id"`
	ProductName string    `json:"productName" db:"name"`
	Active      bool      `json:"active" db:"active"`
	CreatedAt   time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt   time.Time `json:"updatedAt" db:"updated_at"`
}

type ValidateMerchantProductAvailability struct {
	MerchantID  string
	ProductName string
}
