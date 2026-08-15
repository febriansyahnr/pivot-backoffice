package customerModel

import (
	"database/sql"
	"testing"
	"time"

	cardFundedPayoutModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/backendportal/cardFundedPayout"
	unifiedPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/backendportal/unifiedPayment"
	"github.com/stretchr/testify/assert"
)

func TestToCustomerModel(t *testing.T) {
	time := time.Now()
	customerDbModel := CustomerDBModel{
		UUID:             "uuid",
		MerchantID:       "merchantId",
		Email:            sql.NullString{String: "email", Valid: true},
		PhoneCountryCode: sql.NullString{String: "+62", Valid: true},
		PhoneNumber:      "89123456789",
		Metadata:         []byte(`{"key":"value"}`),
		FirstName:        "firstName",
		LastName:         sql.NullString{String: "lastName", Valid: true},
		BusinessName:     sql.NullString{String: "businessName", Valid: true},
		City:             sql.NullString{String: "city", Valid: true},
		CreatedAt:        time,
		UpdatedAt:        time,
		DeletedAt:        sql.NullTime{Time: time, Valid: true},
		Country:          sql.NullString{String: "country", Valid: true},
		AddressLine1:     sql.NullString{String: "addressLine1", Valid: true},
		AddressLine2:     sql.NullString{String: "addressLine2", Valid: true},
		PostalCode:       sql.NullString{String: "postalCode", Valid: true},
		State:            sql.NullString{String: "state", Valid: true},
	}
	customerModel := customerDbModel.ToCustomerModel()
	assert.Equal(t, customerModel.UUID, "uuid")
	assert.Equal(t, customerModel.MerchantID, "merchantId")
	assert.Equal(t, customerModel.Email, "email")
	assert.Equal(t, customerModel.PhoneCountryCode, "+62")
	assert.Equal(t, customerModel.PhoneNumber, "089123456789")
	assert.Equal(t, customerModel.Metadata, map[string]interface{}{"key": "value"})
	assert.Equal(t, customerModel.FirstName, "firstName")
	assert.Equal(t, customerModel.LastName, "lastName")
	assert.Equal(t, customerModel.BusinessName, "businessName")
	assert.Equal(t, customerModel.City, "city")
	assert.Equal(t, customerModel.CreatedAt, time)
	assert.Equal(t, customerModel.UpdatedAt, time)
	assert.Equal(t, customerModel.DeletedAt, &time)
	assert.Equal(t, customerModel.Country, "country")
	assert.Equal(t, customerModel.AddressLine1, "addressLine1")
	assert.Equal(t, customerModel.AddressLine2, "addressLine2")
	assert.Equal(t, customerModel.PostalCode, "postalCode")
	assert.Equal(t, customerModel.State, "state")

	customerDbModel.PhoneCountryCode = sql.NullString{String: "+64", Valid: true}
	customerModel = customerDbModel.ToCustomerModel()
	assert.Equal(t, customerModel.PhoneCountryCode, "+64")
	assert.Equal(t, customerModel.PhoneNumber, "+6489123456789")

}

func TestToCustomerDBModel(t *testing.T) {
	time := time.Now()
	customer := Customer{
		UUID:         "uuid",
		MerchantID:   "merchantId",
		Email:        "email",
		PhoneNumber:  "phoneNumber",
		Metadata:     map[string]interface{}{"key": "value"},
		FirstName:    "firstName",
		LastName:     "lastName",
		BusinessName: "businessName",
		City:         "city",
		CreatedAt:    time,
		UpdatedAt:    time,
		DeletedAt:    &time,
		Country:      "country",
		AddressLine1: "addressLine1",
		AddressLine2: "addressLine2",
		PostalCode:   "postalCode",
		State:        "state",
	}
	customerDbModel := customer.ToCustomerDBModel()
	assert.Equal(t, customerDbModel.UUID, "uuid")
	assert.Equal(t, customerDbModel.MerchantID, "merchantId")
	assert.Equal(t, customerDbModel.Email.String, "email")
	assert.Equal(t, customerDbModel.PhoneNumber, "phoneNumber")
	assert.Equal(t, customerDbModel.Metadata, []byte(`{"key":"value"}`))
	assert.Equal(t, customerDbModel.FirstName, "firstName")
	assert.Equal(t, customerDbModel.LastName.String, "lastName")
	assert.Equal(t, customerDbModel.BusinessName.String, "businessName")
	assert.Equal(t, customerDbModel.City.String, "city")
	assert.Equal(t, customerDbModel.CreatedAt, time)
	assert.Equal(t, customerDbModel.UpdatedAt, time)
	assert.Equal(t, customerDbModel.DeletedAt.Time, time)
	assert.Equal(t, customerDbModel.Country.String, "country")

	customer2 := Customer{
		UUID:         "uuid",
		MerchantID:   "merchantId",
		Email:        "email",
		PhoneNumber:  "phoneNumber",
		FirstName:    "firstName",
		LastName:     "lastName",
		BusinessName: "businessName",
		City:         "city",
		CreatedAt:    time,
		UpdatedAt:    time,
		DeletedAt:    &time,
		Country:      "country",
		AddressLine1: "addressLine1",
		AddressLine2: "addressLine2",
		PostalCode:   "postalCode",
		State:        "state",
	}
	customerDbModel2 := customer2.ToCustomerDBModel()
	assert.Equal(t, customerDbModel2.UUID, "uuid")
	assert.Equal(t, customerDbModel2.MerchantID, "merchantId")
	assert.Equal(t, customerDbModel2.Email.String, "email")
	assert.Equal(t, customerDbModel2.PhoneNumber, "phoneNumber")
	assert.Equal(t, customerDbModel2.Metadata, []byte(`{}`))
	assert.Equal(t, customerDbModel2.FirstName, "firstName")
	assert.Equal(t, customerDbModel2.LastName.String, "lastName")
	assert.Equal(t, customerDbModel2.BusinessName.String, "businessName")
	assert.Equal(t, customerDbModel2.City.String, "city")
	assert.Equal(t, customerDbModel2.CreatedAt, time)
	assert.Equal(t, customerDbModel2.UpdatedAt, time)
	assert.Equal(t, customerDbModel2.DeletedAt.Time, time)
	assert.Equal(t, customerDbModel2.Country.String, "country")
}

func TestFullNameToFirstNameAndLastName(t *testing.T) {
	firstName, lastName := FullNameToFirstNameAndLastName("John Stephen Doe")
	assert.Equal(t, firstName, "John Stephen")
	assert.Equal(t, lastName, "Doe")

	firstName, lastName = FullNameToFirstNameAndLastName("John Doe")
	assert.Equal(t, firstName, "John")
	assert.Equal(t, lastName, "Doe")

	firstName, lastName = FullNameToFirstNameAndLastName("John")
	assert.Equal(t, firstName, "John")
	assert.Equal(t, lastName, "")
}

func TestFirstNameAndLastNameToFullName(t *testing.T) {
	fullName := FirstNameAndLastNameToFullName("John", "Doe")
	assert.Equal(t, fullName, "John Doe")

	fullName = FirstNameAndLastNameToFullName("John", "")
	assert.Equal(t, fullName, "John")
}

func TestCreateCustomer(t *testing.T) {
	createCustomerRequest := CreateCustomerRequest{
		MerchantID:   "merchantId",
		Email:        "email",
		PhoneNumber:  "phoneNumber",
		FirstName:    "firstName",
		LastName:     "lastName",
		BusinessName: "businessName",
		City:         "city",
		Country:      "country",
		AddressLine1: "addressLine1",
		AddressLine2: "addressLine2",
		PostalCode:   "postalCode",
		State:        "state",
		Metadata:     map[string]interface{}{"key": "value"},
	}
	customer := CreateCustomer(&createCustomerRequest)
	assert.Equal(t, customer.MerchantID, "merchantId")
	assert.Equal(t, customer.Email, "email")
	assert.Equal(t, customer.PhoneNumber, "phoneNumber")
	assert.Equal(t, customer.FirstName, "firstName")
	assert.Equal(t, customer.LastName, "lastName")
	assert.Equal(t, customer.BusinessName, "businessName")
	assert.Equal(t, customer.City, "city")
	assert.Equal(t, customer.Country, "country")
	assert.Equal(t, customer.AddressLine1, "addressLine1")
	assert.Equal(t, customer.AddressLine2, "addressLine2")
	assert.Equal(t, customer.PostalCode, "postalCode")
	assert.Equal(t, customer.State, "state")
	assert.Equal(t, customer.Metadata, map[string]interface{}{"key": "value"})

}

func TestUpdate(t *testing.T) {
	customer := Customer{
		MerchantID:       "merchantId",
		Email:            "email",
		PhoneCountryCode: "phoneCountryCode",
		PhoneNumber:      "phoneNumber",
		FirstName:        "firstName",
		LastName:         "lastName",
		BusinessName:     "businessName",
		City:             "city",
		Country:          "country",
		AddressLine1:     "addressLine1",
		AddressLine2:     "addressLine2",
		PostalCode:       "postalCode",
		State:            "state",
		Metadata:         map[string]interface{}{"key": "value"},
		UUID:             "uuid",
	}
	updateCustomerRequest := UpdateCustomerRequest{
		Email:            NewString("newEmail"),
		PhoneCountryCode: NewString("newPhoneCountryCode"),
		PhoneNumber:      NewString("newPhoneNumber"),
		FirstName:        NewString("newFirstName"),
		LastName:         NewString("newLastName"),
		BusinessName:     NewString("newBusinessName"),
		City:             NewString("newCity"),
		Country:          NewString("newCountry"),
		AddressLine1:     NewString("newAddressLine1"),
		AddressLine2:     NewString("newAddressLine2"),
		PostalCode:       NewString("newPostalCode"),
		State:            NewString("newState"),
	}
	customer.Update(&updateCustomerRequest)
	assert.Equal(t, customer.Email, "newEmail")
	assert.Equal(t, customer.PhoneNumber, "newPhoneNumber")
	assert.Equal(t, customer.FirstName, "newFirstName")
	assert.Equal(t, customer.LastName, "newLastName")
	assert.Equal(t, customer.BusinessName, "newBusinessName")
	assert.Equal(t, customer.City, "newCity")
	assert.Equal(t, customer.Country, "newCountry")
	assert.Equal(t, customer.AddressLine1, "newAddressLine1")
	assert.Equal(t, customer.AddressLine2, "newAddressLine2")
	assert.Equal(t, customer.PostalCode, "newPostalCode")
	assert.Equal(t, customer.State, "newState")
	assert.Equal(t, customer.Metadata, map[string]interface{}{"key": "value"})
	assert.Equal(t, customer.UUID, "uuid")
}

func TestUpdateMetadata(t *testing.T) {
	customer := Customer{
		MerchantID: "merchantId",
		Email:      "email",
		Metadata:   map[string]interface{}{"key": "value"},
		UUID:       "uuid",
	}

	refundPref := &unifiedPaymentModel.UnifiedPaymentRefundPreference{
		Method: "AUTO",
		TransferDestination: &unifiedPaymentModel.RefundTransferDestination{
			ChannelCode: "BRI",
			ChannelInformation: &unifiedPaymentModel.RefundChannelInformation{
				AccountNumber: "1234567890",
				AccountName:   "John Doe",
			},
		},
	}

	// Create a new metadata map to update with
	newMetadata := map[string]interface{}{
		"key":              "updated",
		"newKey":           "newValue",
		"version":          2,
		"refundPreference": refundPref,
	}

	updateCustomerRequest := UpdateCustomerRequest{
		Metadata: newMetadata,
	}

	customer.Update(&updateCustomerRequest)

	// Verify that the metadata was updated
	assert.Equal(t, newMetadata, customer.Metadata)
	assert.Equal(t, "updated", customer.Metadata["key"])
	assert.Equal(t, "newValue", customer.Metadata["newKey"])
	assert.Equal(t, 2, customer.Metadata["version"])

	// Verify other fields remain unchanged
	assert.Equal(t, "merchantId", customer.MerchantID)
	assert.Equal(t, "email", customer.Email)
	assert.Equal(t, "uuid", customer.UUID)

	// when the refundPreference metadata is nil in request, then should not update it
	newMetadata = map[string]interface{}{
		"key":     "updated2",
		"newKey":  "newValue2",
		"version": 3,
	}

	updateCustomerRequest = UpdateCustomerRequest{
		Metadata: newMetadata,
	}
	customer.Update(&updateCustomerRequest)

	// Verify that the metadata was updated
	assert.Equal(t, map[string]interface{}{
		"key":              "updated2",
		"newKey":           "newValue2",
		"version":          3,
		"refundPreference": refundPref,
	}, customer.Metadata)
}

func TestToCardFundedPayoutSavedCardList(t *testing.T) {
	testCases := []struct {
		desc     string
		input    CustomerDBModel
		expected *cardFundedPayoutModel.GetSavedCardResponse
	}{
		{
			desc: "SUCCESS: convert with valid metadata",
			input: CustomerDBModel{
				UUID: "customer-123",
				Metadata: []byte(`{
					"useCase": "CARD_FUNDED_PAYOUT_SAVED_CARDS",
					"paymentMethods": [
						{
							"paymentChannel": "CHANNEL",
							"card": {
								"cardName": "VISA",
								"issuingBank": "BANK",
								"last4": "1234",
								"expMonth": "12",
								"expYear": "2025"
							}
						}
					]
				}`),
			},
			expected: &cardFundedPayoutModel.GetSavedCardResponse{
				ID:             "customer-123",
				CardName:       "VISA",
				IssuingBank:    "BANK",
				Last4:          "1234",
				ExpiryMonth:    "12",
				ExpiryYear:     "2025",
				PaymentChannel: "CHANNEL",
			},
		},
		{
			desc: "ERROR: return nil when paymentMethods is missing",
			input: CustomerDBModel{
				UUID:     "customer-123",
				Metadata: []byte(`{"useCase": "CARD_FUNDED_PAYOUT_SAVED_CARDS"}`),
			},
			expected: nil,
		},
		{
			desc: "ERROR: return nil when paymentMethods is empty",
			input: CustomerDBModel{
				UUID:     "customer-123",
				Metadata: []byte(`{"useCase": "CARD_FUNDED_PAYOUT_SAVED_CARDS", "paymentMethods": []}`),
			},
			expected: nil,
		},
		{
			desc: "ERROR: return nil when card is missing",
			input: CustomerDBModel{
				UUID:     "customer-123",
				Metadata: []byte(`{"useCase": "CARD_FUNDED_PAYOUT_SAVED_CARDS", "paymentMethods": [{"paymentChannel": "CHANNEL"}]}`),
			},
			expected: nil,
		},
		{
			desc: "ERROR: return nil when metadata is empty",
			input: CustomerDBModel{
				UUID:     "customer-123",
				Metadata: []byte(``),
			},
			expected: nil,
		},
		{
			desc: "ERROR: return nil when metadata is invalid JSON",
			input: CustomerDBModel{
				UUID:     "customer-123",
				Metadata: []byte(`{invalid json}`),
			},
			expected: nil,
		},
		{
			desc: "SUCCESS: convert with partial card data",
			input: CustomerDBModel{
				UUID: "customer-456",
				Metadata: []byte(`{
					"paymentMethods": [
						{
							"paymentChannel": "MASTERCARD",
							"card": {
								"cardName": "MASTERCARD",
								"last4": "5678"
							}
						}
					]
				}`),
			},
			expected: &cardFundedPayoutModel.GetSavedCardResponse{
				ID:             "customer-456",
				CardName:       "MASTERCARD",
				IssuingBank:    "",
				Last4:          "5678",
				ExpiryMonth:    "",
				ExpiryYear:     "",
				PaymentChannel: "MASTERCARD",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			result := tc.input.ToCardFundedPayoutSavedCardList()
			if tc.expected == nil {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
				assert.Equal(t, tc.expected.ID, result.ID)
				assert.Equal(t, tc.expected.CardName, result.CardName)
				assert.Equal(t, tc.expected.IssuingBank, result.IssuingBank)
				assert.Equal(t, tc.expected.Last4, result.Last4)
				assert.Equal(t, tc.expected.ExpiryMonth, result.ExpiryMonth)
				assert.Equal(t, tc.expected.ExpiryYear, result.ExpiryYear)
				assert.Equal(t, tc.expected.PaymentChannel, result.PaymentChannel)
			}
		})
	}
}
