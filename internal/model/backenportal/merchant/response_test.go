package merchant

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToSubMerchantResponse(t *testing.T) {
	// Setup the original Merchant object
	originalMerchant := Merchant{
		UUID:              "merchant-uuid",
		Name:              "Test Merchant",
		ShortName:         "TM",
		MID:               sql.NullString{String: "MID123", Valid: true},
		Description:       "A test merchant",
		Logo:              "http://logo-url.com",
		MerchantEmail:     "merchant@example.com",
		MerchantPhone:     "123456789",
		BusinessCountry:   sql.NullString{String: "ID", Valid: true},
		BusinessType:      sql.NullString{String: "Retail", Valid: true},
		BusinessStructure: sql.NullString{String: "Sole Proprietor", Valid: true},
		PICName:           sql.NullString{String: "John Doe", Valid: true},
		PICEmail:          "pic@example.com",
		PICPhone:          "987654321",
		PICJobTitle:       sql.NullString{String: "Manager", Valid: true},
		Address:           "123 Main Street",
		PostCode:          "54321",
		ParentID:          sql.NullString{String: "parent-uuid", Valid: true},
	}

	// Expected response
	expectedResponse := &SubMerchantResponse{
		UUID:              "merchant-uuid",
		Name:              "Test Merchant",
		ShortName:         "TM",
		Description:       "A test merchant",
		Logo:              "http://logo-url.com",
		MerchantEmail:     "merchant@example.com",
		MerchantPhone:     "123456789",
		BusinessCountry:   "ID",
		BusinessType:      "Retail",
		BusinessStructure: "Sole Proprietor",
		PICName:           "John Doe",
		PICEmail:          "pic@example.com",
		PICPhone:          "987654321",
		PICJobTitle:       "Manager",
		Address:           "123 Main Street",
		PostCode:          "54321",
		ParentID:          "parent-uuid",
	}

	// Call the function
	actualResponse := originalMerchant.ToSubMerchantResponse()

	// Assert the response is as expected
	assert.Equal(t, expectedResponse, actualResponse)
}
