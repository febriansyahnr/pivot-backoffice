package customerModel

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestToGeneralCustomerResponse(t *testing.T) {
	customer := &Customer{
		UUID:         "12345",
		MerchantID:   "67890",
		Email:        "test@example.com",
		PhoneNumber:  "123-456-7890",
		FirstName:    "John",
		LastName:     "Doe",
		BusinessName: "Doe Inc.",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		City:         "Test City",
		Country:      "Test Country",
		AddressLine1: "123 Test St",
		AddressLine2: "Apt 4",
		PostalCode:   "12345",
		State:        "Test State",
		Metadata:     map[string]interface{}{"key": "value"},
	}

	response := customer.ToGeneralCustomerResponse()
	assert.Equal(t, customer.UUID, response.UUID)
	assert.Equal(t, customer.MerchantID, response.MerchantID)
	assert.Equal(t, customer.Email, response.Email)
	assert.Equal(t, customer.PhoneNumber, response.PhoneNumber)
	assert.Equal(t, customer.FirstName, response.FirstName)
	assert.Equal(t, customer.LastName, response.LastName)
	assert.Equal(t, customer.BusinessName, response.BusinessName)
	assert.Equal(t, customer.CreatedAt, response.CreatedAt)
	assert.Equal(t, customer.UpdatedAt, response.UpdatedAt)
	assert.Equal(t, customer.City, response.City)
	assert.Equal(t, customer.Country, response.Country)
	assert.Equal(t, customer.AddressLine1, response.AddressLine1)
	assert.Equal(t, customer.AddressLine2, response.AddressLine2)
	assert.Equal(t, customer.PostalCode, response.PostalCode)
	assert.Equal(t, customer.State, response.State)
	assert.Equal(t, customer.Metadata, response.Metadata)
}
