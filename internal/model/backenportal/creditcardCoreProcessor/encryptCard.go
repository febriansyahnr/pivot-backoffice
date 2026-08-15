package creditcardCoreProcessorModel

type EncryptCardDetailRequest struct {
	Number      string `json:"number" validate:"required,numberstring"`
	ExpiryMonth string `json:"expiry_month" validate:"required,numberstring,min=2,max=2"`
	ExpiryYear  string `json:"expiry_year" validate:"required,numberstring,min=2,max=2"`
	CVC         string `json:"cvc"`
	NameOnCard  string `json:"name_on_card" validate:"required,nospecialchars"`
}

type EncryptCardRequest struct {
	MerchantID        string                   `json:"merchant_id" `
	ClientReferenceID string                   `json:"client_reference_id" validate:"required"`
	CardRequest       EncryptCardDetailRequest `json:"card" validate:"required"`
	DeviceInformation DeviceInformation        `json:"device_informations" validate:"required"`
	Metadata          map[string]string        `json:"metadata,omitempty"`
}

type DeviceInformation struct {
	Type           string `json:"type"`
	UserAgent      string `json:"user_agent"`
	IpAddress      string `json:"ip_address"`
	AcceptLanguage string `json:"accept_language"`
	CookieToken    string `json:"cookie_token"`
	DeviceID       string `json:"device_id"`
	BrowserWidth   string `json:"browser_width"`
	BrowserHeight  string `json:"browser_height"`
	Country        string `json:"country"`
}

type EncryptedCardInformationResponse struct {
	First8Digits     string `json:"first_8_digits"`
	First6Digits     string `json:"first_6_digits"`
	Last4Digits      string `json:"last_4_digits"`
	ExpiryMonth      string `json:"expiry_month"`
	ExpiryYear       string `json:"expiry_year"`
	HasAssociatedCVC bool   `json:"has_associated_cvc"`
	Fingerprint      string `json:"fingerprint"`
}

type EncryptedCardResponse struct {
	ClientReferenceID        string                           `json:"client_reference_id"`
	EncryptedCard            string                           `json:"encrypted_card"`
	EncryptedCardInformation EncryptedCardInformationResponse `json:"encrypted_card_informations"`
	DeviceInfomation         DeviceInformation                `json:"device_information"`
	BinDetail                Bin                              `json:"card_bin_detail,omitempty"`
	CreatedAt                string                           `json:"created_at"`
	Metadata                 map[string]string                `json:"metadata,omitempty"`
}

type BaseEncryptedCardResponse struct {
	Code    string                `json:"code"`
	Message string                `json:"message,omitempty"`
	Error   interface{}           `json:"error,omitempty"`
	Data    EncryptedCardResponse `json:"data,omitempty"`
}
