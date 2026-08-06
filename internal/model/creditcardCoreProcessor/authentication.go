package creditcardCoreProcessorModel

type EncryptedCardAuthenticationRequest struct {
	PaymentID           string              `json:"payment_id"`
	MerchantID          string              `json:"merchant_id"`
	ClientTransactionID string              `json:"client_transaction_id"`
	Fingerprint         string              `json:"fingerprint,omitempty"`
	CardHolderName      string              `json:"card_holder_name,omitempty"`
	EncryptedCard       string              `json:"encrypted_card,omitempty"`
	CardID              string              `json:"card_id,omitempty"`
	CVC                 string              `json:"cvc"`
	Amount              float64             `json:"amount"`
	Fee                 float64             `json:"fee"`
	Currency            string              `json:"currency"`
	SavedFutureUse      *bool               `json:"saved_future_use,omitempty"`
	BillingInformation  *BillingInformation `json:"billing_information,omitempty"`
	// Recurring Payment
	RecurringID                string  `json:"recurring_id,omitempty"`
	InitiateFirstAuthorization *bool   `json:"initiate_first_authorization,omitempty"`
	FirstAuthorizationMethod   string  `json:"first_authorization_method,omitempty"`
	FirstAuthorizationOrderID  *string `json:"first_authorization_order_id,omitempty"`
	BillingInterval            uint8   `json:"billing_interval,omitempty"`
	BillingIntervalUnit        string  `json:"billing_interval_unit,omitempty"`
	BillingCycleCount          uint16  `json:"billing_cycle_count,omitempty"`
	// External MPI
	ThreeDsMethod       string               `json:"three_ds_method,omitempty"`
	ExternalThreeDsInfo *ExternalThreeDsInfo `json:"external_three_ds_info,omitempty"`
	// Others
	CardFundedPayout       *CardFundedPayout `json:"card_funded_payout,omitempty"`
	CardOnFile             *CardOnFile       `json:"card_on_file,omitempty"`
	EncryptedEncryptionKey string            `json:"-"`
}

type BaseEncryptedCardAuthenticationResponse struct {
	Code    string                              `json:"code"`
	Message string                              `json:"message,omitempty"`
	Error   interface{}                         `json:"error,omitempty"`
	Data    EncryptedCardAuthenticationResponse `json:"data,omitempty"`
}

type EncryptedCardAuthenticationResponse struct {
	AcquirerTransactionID    string                           `json:"acquirer_transaction_id"`
	Amount                   string                           `json:"amount"`
	Currency                 string                           `json:"currency"`
	Message                  string                           `json:"message"`
	SessionID                string                           `json:"session_id"`
	Status                   string                           `json:"status"`
	AuthenticationURL        AuthenticationURLDetail          `json:"authentication_url"`
	EncryptedCardInformation EncryptedCardInformationDetail   `json:"encrypted_card_informations"`
	AuthenticationData       *EncryptedCardAuthenticationData `json:"authentication_data,omitempty"`
}

type AuthenticationURLDetail struct {
	ActionURL    string `json:"action_url"`
	CreatedAt    string `json:"created_at"`
	ThreeDSToken string `json:"creq"`
	HTML         string `json:"html"`
	Method       string `json:"method"`
	URL          string `json:"url"`
	Version      string `json:"version"`
}

type EncryptedCardInformationDetail struct {
	First8Digits string    `json:"first_8_digits"`
	First6Digits string    `json:"first_6_digits"`
	Last4Digits  string    `json:"last_4_digits"`
	ExpiryMonth  string    `json:"expiry_month"`
	ExpiryYear   string    `json:"expiry_year"`
	Fingerprint  string    `json:"fingerprint"`
	BinDetail    BinDetail `json:"bin_detail"`
}

type EncryptedCardAuthenticationData struct {
	AuthenticationResult string `json:"authentication_result"`
	AuthenticationID     string `json:"authentication_id"`
	PaRes                string `json:"pa_res"`
	VeRes                string `json:"ve_res"`
	XID                  string `json:"xid"`
	CAVV                 string `json:"cavv"`
	EciCode              string `json:"eci_code"`
	ThreeDsVer           string `json:"three_ds_ver"`
	ChallengeCode        string `json:"challenge_code"`
}

type BinDetail struct {
	CardType      string `json:"card_type"`
	IssuerName    string `json:"issuer_name"`
	CardBrand     string `json:"card_brand"`
	IssuerCountry string `json:"issuer_country"`
}

type BillingInformation struct {
	GivenName     string       `json:"given_name" validate:"required"`
	SureName      string       `json:"sure_name"`
	Email         string       `json:"email" validate:"required,email"`
	PhoneNumber   *PhoneNumber `json:"phone_number" validate:"omitempty"`
	Address1      string       `json:"address_line1"`
	Address2      string       `json:"address_line2"`
	City          string       `json:"city"`
	ProvinceState string       `json:"province_state"`
	Country       string       `json:"country"`
	PostalCode    string       `json:"postal_code"`
}

type ExternalThreeDsInfo struct {
	CAVV                 string `json:"cavv,omitempty"`
	TransactionID        string `json:"transaction_id" validate:"required"`
	ThreeDSVersion       string `json:"three_ds_version" validate:"required"`
	ECI                  string `json:"eci" validate:"required"`
	TransactionStatus    string `json:"transaction_status" validate:"required"`
	AuthenticationScheme string `json:"authentication_scheme" validate:"required"`
	ACSTransactionID     string `json:"acs_transaction_id,omitempty" validate:"omitempty"`
	ACSReference         string `json:"acs_reference,omitempty" validate:"omitempty"`
	Time                 string `json:"time,omitempty" validate:"omitempty"`
}

type PhoneNumber struct {
	CountryCode string `json:"country_code" validate:"required" example:"+62"`
	Number      string `json:"number" validate:"required,number" example:"81234567890"`
}

type CardFundedPayout struct {
	Sequence       int    `json:"sequence"`
	Count          int    `json:"count"`
	VendorID       string `json:"vendor_id"`
	VendorName     string `json:"vendor_name"`
	FirstPaymentID string `json:"first_payment_id,omitempty"`
}

type CardOnFile struct {
	Initiator                    string `json:"initiator"`
	Type                         string `json:"type"`
	PreviousNetworkTransactionID string `json:"previous_network_transaction_id"`
}
