package card

import creditcardCoreProcessorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/creditcardCoreProcessor"

type EncryptCardRequest struct {
	MerchantID        string                   `json:"merchantId"`
	ClientReferenceID string                   `json:"clientReferenceId" validate:"required"`
	CardRequest       EncryptCardDetailRequest `json:"card" validate:"required"`
	DeviceInformation DeviceInformation        `json:"deviceInformations" validate:"required"`
	Metadata          map[string]string        `json:"metadata,omitempty"`
}

func (r *EncryptCardRequest) ToProcessorRequestModel() *creditcardCoreProcessorModel.EncryptCardRequest {
	return &creditcardCoreProcessorModel.EncryptCardRequest{
		MerchantID:        r.MerchantID,
		ClientReferenceID: r.ClientReferenceID,
		CardRequest: creditcardCoreProcessorModel.EncryptCardDetailRequest{
			Number:      r.CardRequest.Number,
			ExpiryMonth: r.CardRequest.ExpiryMonth,
			ExpiryYear:  r.CardRequest.ExpiryYear,
			CVC:         r.CardRequest.CVC,
			NameOnCard:  r.CardRequest.NameOnCard,
		},
		DeviceInformation: creditcardCoreProcessorModel.DeviceInformation{
			Type:           r.DeviceInformation.Type,
			UserAgent:      r.DeviceInformation.UserAgent,
			IpAddress:      r.DeviceInformation.IpAddress,
			AcceptLanguage: r.DeviceInformation.AcceptLanguage,
			CookieToken:    r.DeviceInformation.CookieToken,
			DeviceID:       r.DeviceInformation.DeviceID,
			BrowserWidth:   r.DeviceInformation.BrowserWidth,
			BrowserHeight:  r.DeviceInformation.BrowserHeight,
			Country:        r.DeviceInformation.Country,
		},
		Metadata: r.Metadata,
	}
}

type EncryptCardDetailRequest struct {
	Number      string `json:"number" validate:"required,numberstring"`
	ExpiryMonth string `json:"expiryMonth" validate:"required,numberstring,min=2,max=2"`
	ExpiryYear  string `json:"expiryYear" validate:"required,numberstring,min=2,max=2"`
	CVC         string `json:"cvc"`
	NameOnCard  string `json:"nameOnCard" validate:"required,nospecialchars"`
}

type DeviceInformation struct {
	Type           string `json:"type"`
	UserAgent      string `json:"userAgent"`
	IpAddress      string `json:"ipAddress"`
	AcceptLanguage string `json:"acceptLanguage"`
	CookieToken    string `json:"cookieToken"`
	DeviceID       string `json:"deviceId"`
	BrowserWidth   string `json:"browserWidth"`
	BrowserHeight  string `json:"browserHeight"`
	Country        string `json:"country"`
}

type EncryptedCardInformationResponse struct {
	First8Digits     string `json:"first8"`
	First6Digits     string `json:"first6"`
	Last4Digits      string `json:"last4"`
	ExpiryMonth      string `json:"expiryMonth"`
	ExpiryYear       string `json:"expiryYear"`
	HasAssociatedCVC bool   `json:"hasAssociatedCvc"`
	Fingerprint      string `json:"fingerprint"`
}

type EncryptedCardResponse struct {
	ClientReferenceID        string                           `json:"client_reference_id"`
	EncryptedCard            string                           `json:"encryptedCard"`
	EncryptedCardInformation EncryptedCardInformationResponse `json:"encryptedCardInformations"`
	DeviceInfomation         DeviceInformation                `json:"deviceInformations"`
	BinDetail                Bin                              `json:"cardBinDetail,omitempty"`
	CreatedAt                string                           `json:"createdAt"`
	Metadata                 map[string]string                `json:"metadata,omitempty"`
}

func ToCardResponseModel(card *creditcardCoreProcessorModel.EncryptedCardResponse) *EncryptedCardResponse {
	if card == nil {
		return nil
	}
	return &EncryptedCardResponse{
		ClientReferenceID: card.ClientReferenceID,
		EncryptedCard:     card.EncryptedCard,
		EncryptedCardInformation: EncryptedCardInformationResponse{
			First8Digits:     card.EncryptedCardInformation.First8Digits,
			First6Digits:     card.EncryptedCardInformation.First6Digits,
			Last4Digits:      card.EncryptedCardInformation.Last4Digits,
			ExpiryMonth:      card.EncryptedCardInformation.ExpiryMonth,
			ExpiryYear:       card.EncryptedCardInformation.ExpiryYear,
			HasAssociatedCVC: card.EncryptedCardInformation.HasAssociatedCVC,
			Fingerprint:      card.EncryptedCardInformation.Fingerprint,
		},
		DeviceInfomation: DeviceInformation{
			Type:           card.DeviceInfomation.Type,
			UserAgent:      card.DeviceInfomation.UserAgent,
			IpAddress:      card.DeviceInfomation.IpAddress,
			AcceptLanguage: card.DeviceInfomation.AcceptLanguage,
			CookieToken:    card.DeviceInfomation.CookieToken,
			DeviceID:       card.DeviceInfomation.DeviceID,
			BrowserWidth:   card.DeviceInfomation.BrowserWidth,
			BrowserHeight:  card.DeviceInfomation.BrowserHeight,
			Country:        card.DeviceInfomation.Country,
		},
		BinDetail: Bin(card.BinDetail),
		CreatedAt: card.CreatedAt,
		Metadata:  card.Metadata,
	}
}
