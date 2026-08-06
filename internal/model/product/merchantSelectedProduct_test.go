package product

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewMerchantSelectedProduct(t *testing.T) {

	req := &AddMerchantProductRequest{
		MerchantID: "merchant-id",
		ProductID:  "product-id",
	}
	merchantSelectedProduct := NewMerchantSelectedProduct(req)
	assert.NotEmpty(t, merchantSelectedProduct.CreatedAt)
	assert.NotEmpty(t, merchantSelectedProduct.UpdatedAt)
	assert.NotEmpty(t, merchantSelectedProduct.MerchantID)
	assert.NotEmpty(t, merchantSelectedProduct.ProductID)
	assert.NotEmpty(t, merchantSelectedProduct.UUID)
	assert.True(t, merchantSelectedProduct.Active)
}
