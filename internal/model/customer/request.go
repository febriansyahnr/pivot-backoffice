package customerModel

type CreateCustomerRequest struct {
	MerchantID       string `json:"merchantId"`
	Email            string `json:"email"`
	PhoneCountryCode string `json:"phoneCountryCode"`
	PhoneNumber      string `json:"phoneNumber" validate:"required,numeric"`
	FirstName        string `json:"firstName"`
	LastName         string `json:"lastName"`
	BusinessName     string `json:"businessName"`

	City         string `json:"city,omitempty"`
	Country      string `json:"country,omitempty"`
	AddressLine1 string `json:"addressLine1,omitempty"`
	AddressLine2 string `json:"addressLine2,omitempty"`
	PostalCode   string `json:"postalCode,omitempty"`
	State        string `json:"state,omitempty"`

	IsBlocked   bool   `json:"isBlocked,omitempty"`
	BlockReason string `json:"blockReason,omitempty" validate:"required_if=IsBlocked true"`

	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

type CreateUnifiedPaymentCustomerRequest struct {
	MerchantID       string                 `json:"merchantId" validate:"required"`
	FirstName        string                 `json:"firstName" validate:"required"`
	LastName         string                 `json:"lastName" validate:"required"`
	Email            string                 `json:"email" validate:"required,email"`
	PhoneNumber      string                 `json:"phoneNumber"`
	PhoneCountryCode string                 `json:"phoneCountryCode"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
}

type UpdateCustomerRequest struct {
	UUID             string  `json:"uuid,omitempty"`
	MerchantID       string  `json:"merchantId"`
	Email            *string `json:"email"`
	PhoneCountryCode *string `json:"phoneCountryCode"`
	PhoneNumber      *string `json:"phoneNumber"`
	FirstName        *string `json:"firstName"`
	LastName         *string `json:"lastName"`
	BusinessName     *string `json:"businessName"`

	City         *string `json:"city,omitempty"`
	Country      *string `json:"country,omitempty"`
	AddressLine1 *string `json:"addressLine1,omitempty"`
	AddressLine2 *string `json:"addressLine2,omitempty"`
	PostalCode   *string `json:"postalCode,omitempty"`
	State        *string `json:"state,omitempty"`

	IsBlocked   *bool   `json:"isBlocked,omitempty"`
	BlockReason *string `json:"blockReason,omitempty"`

	Metadata map[string]interface{} `json:"metadata,omitempty"` // map[string]interface{} already handle nil on nil conditional
}

type GetMerchantCustomerRequest struct {
	MerchantID       string
	CustomerID       string
	Email            string
	PhoneCountryCode string
	PhoneNumber      string
}
