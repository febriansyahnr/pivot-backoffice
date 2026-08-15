package card

import (
	creditcardCoreProcessorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/creditcardCoreProcessor"
)

type EncryptedCardAuthenticationRequest struct {
	PaymentID              string
	MerchantID             string
	ClientTransactionID    string
	CardID                 string
	CVC                    string
	Amount                 float64
	Fee                    float64
	Currency               string
	Fingerprint            string // Should be choice on of fingerprint or encrypted card
	EncryptedCard          string // Should be choice on of fingerprint or encrypted card
	EncryptedEncryptionKey string
	CardHolderName         string // Only requires if fingerprint is set
	SavedFutureUse         *bool
	BillingInformation     *BillingInformation
	// Recurring Payment Feature
	RecurringID                string
	InitiateFirstAuthorization *bool
	FirstAuthorizationMethod   string
	FirstAuthorizationOrderID  *string
	RecurringBillingCycle      *RecurringBillingCycle
	// External MPI Feature
	ThreeDsMethod       string
	ExternalThreeDsInfo *ExternalThreeDsInfo
	// Others
	CardFundedPayout *creditcardCoreProcessorModel.CardFundedPayout
	CardOnFile       *CardOnFile
}

type BillingInformation struct {
	GivenName     string
	SureName      string
	Email         string
	PhoneNumber   *PhoneNumber
	Address1      string
	Address2      string
	City          string
	ProvinceState string
	Country       string
	PostalCode    string
}

type ExternalThreeDsInfo struct {
	CAVV                 string
	TransactionID        string
	ThreeDSVersion       string
	ECI                  string
	TransactionStatus    string
	AuthenticationScheme string
	ACSTransactionID     string
	ACSReference         string
	Time                 string
}

type PhoneNumber struct {
	CountryCode string
	Number      string
}

type CardOnFile struct {
	Initiator                    string
	Type                         string
	PreviousNetworkTransactionID string
}

func (request *EncryptedCardAuthenticationRequest) ToProcessorRequestModel() *creditcardCoreProcessorModel.EncryptedCardAuthenticationRequest {
	r := &creditcardCoreProcessorModel.EncryptedCardAuthenticationRequest{
		PaymentID:              request.PaymentID,
		MerchantID:             request.MerchantID,
		ClientTransactionID:    request.ClientTransactionID,
		CardID:                 request.CardID,
		CVC:                    request.CVC,
		Amount:                 request.Amount,
		Fee:                    request.Fee,
		Currency:               request.Currency,
		Fingerprint:            request.Fingerprint,
		CardHolderName:         request.CardHolderName,
		EncryptedCard:          request.EncryptedCard,
		EncryptedEncryptionKey: request.EncryptedEncryptionKey,
		SavedFutureUse:         request.SavedFutureUse,
		ThreeDsMethod:          request.ThreeDsMethod,
		CardFundedPayout:       request.CardFundedPayout,
	}

	if request.BillingInformation != nil {
		r.BillingInformation = &creditcardCoreProcessorModel.BillingInformation{
			GivenName:     request.BillingInformation.GivenName,
			SureName:      request.BillingInformation.SureName,
			Email:         request.BillingInformation.Email,
			Address1:      request.BillingInformation.Address1,
			Address2:      request.BillingInformation.Address2,
			City:          request.BillingInformation.City,
			ProvinceState: request.BillingInformation.ProvinceState,
			Country:       request.BillingInformation.Country,
			PostalCode:    request.BillingInformation.PostalCode,
		}

		if request.BillingInformation.PhoneNumber != nil {
			r.BillingInformation.PhoneNumber = &creditcardCoreProcessorModel.PhoneNumber{
				CountryCode: request.BillingInformation.PhoneNumber.CountryCode,
				Number:      request.BillingInformation.PhoneNumber.Number,
			}
		}
	}

	if request.RecurringID != "" {
		r.RecurringID = request.RecurringID
		r.InitiateFirstAuthorization = request.InitiateFirstAuthorization
		r.FirstAuthorizationMethod = request.FirstAuthorizationMethod
		r.FirstAuthorizationOrderID = request.FirstAuthorizationOrderID
		r.BillingInterval = request.RecurringBillingCycle.Interval
		r.BillingIntervalUnit = request.RecurringBillingCycle.IntervalUnit
		r.BillingCycleCount = request.RecurringBillingCycle.Count
	}

	if request.ExternalThreeDsInfo != nil {
		r.ExternalThreeDsInfo = &creditcardCoreProcessorModel.ExternalThreeDsInfo{
			TransactionID:        request.ExternalThreeDsInfo.TransactionID,
			ThreeDSVersion:       request.ExternalThreeDsInfo.ThreeDSVersion,
			ECI:                  request.ExternalThreeDsInfo.ECI,
			TransactionStatus:    request.ExternalThreeDsInfo.TransactionStatus,
			AuthenticationScheme: request.ExternalThreeDsInfo.AuthenticationScheme,
			ACSTransactionID:     request.ExternalThreeDsInfo.ACSTransactionID,
			ACSReference:         request.ExternalThreeDsInfo.ACSReference,
			CAVV:                 request.ExternalThreeDsInfo.CAVV,
			Time:                 request.ExternalThreeDsInfo.Time,
		}
	}

	if request.CardOnFile != nil {
		r.CardOnFile = &creditcardCoreProcessorModel.CardOnFile{
			Initiator:                    request.CardOnFile.Initiator,
			Type:                         request.CardOnFile.Type,
			PreviousNetworkTransactionID: request.CardOnFile.PreviousNetworkTransactionID,
		}
	}

	return r
}

type AuthenticationResponse struct {
	AcquirerTransactionID string
	Amount                string
	Currency              string
	Message               string
	SessionID             string
	Status                string
	AuthenticationURL     AuthenticationURLDetail
	AuthenticationData    *EncryptedCardAuthenticationData
}

type AuthenticationURLDetail struct {
	ActionURL    string
	CreatedAt    string
	ThreeDSToken string
	HTML         string
	Method       string
	URL          string
	Version      string
}

type EncryptedCardAuthenticationData struct {
	AuthenticationResult string
	AuthenticationID     string
	PaRes                string
	VeRes                string
	XID                  string
	CAVV                 string
	EciCode              string
	ThreeDsVer           string
	ChallengeCode        string
}

func GetAuthenticationResponseFromProcessor(response *creditcardCoreProcessorModel.EncryptedCardAuthenticationResponse) *AuthenticationResponse {
	resp := &AuthenticationResponse{
		AcquirerTransactionID: response.AcquirerTransactionID,
		Amount:                response.Amount,
		Currency:              response.Currency,
		Message:               response.Message,
		SessionID:             response.SessionID,
		Status:                response.Status,
		AuthenticationURL: AuthenticationURLDetail{
			ActionURL:    response.AuthenticationURL.ActionURL,
			CreatedAt:    response.AuthenticationURL.CreatedAt,
			ThreeDSToken: response.AuthenticationURL.ThreeDSToken,
			HTML:         response.AuthenticationURL.HTML,
			Method:       response.AuthenticationURL.Method,
			URL:          response.AuthenticationURL.URL,
			Version:      response.AuthenticationURL.Version,
		},
	}

	if response.AuthenticationData != nil {
		resp.AuthenticationData = &EncryptedCardAuthenticationData{
			AuthenticationResult: response.AuthenticationData.AuthenticationResult,
			AuthenticationID:     response.AuthenticationData.AuthenticationID,
			PaRes:                response.AuthenticationData.PaRes,
			VeRes:                response.AuthenticationData.VeRes,
			XID:                  response.AuthenticationData.XID,
			CAVV:                 response.AuthenticationData.CAVV,
			EciCode:              response.AuthenticationData.EciCode,
			ThreeDsVer:           response.AuthenticationData.ThreeDsVer,
			ChallengeCode:        response.AuthenticationData.ChallengeCode,
		}
	}

	return resp
}

type EncryptedCardAuthenticationResponse struct {
	CardID                 string
	CardInfo               EncryptedCardInformationResponse
	Bin                    Bin
	AuthenticationResponse AuthenticationResponse
}

func ToEncryptedCardAuthenticationResponse(request *EncryptedCardAuthenticationRequest, authResponse *creditcardCoreProcessorModel.EncryptedCardAuthenticationResponse) *EncryptedCardAuthenticationResponse {
	return &EncryptedCardAuthenticationResponse{
		CardID: request.CardID,
		CardInfo: EncryptedCardInformationResponse{
			First8Digits:     authResponse.EncryptedCardInformation.First8Digits,
			First6Digits:     authResponse.EncryptedCardInformation.First6Digits,
			Last4Digits:      authResponse.EncryptedCardInformation.Last4Digits,
			ExpiryMonth:      authResponse.EncryptedCardInformation.ExpiryMonth,
			ExpiryYear:       authResponse.EncryptedCardInformation.ExpiryYear,
			HasAssociatedCVC: true,
			Fingerprint:      authResponse.EncryptedCardInformation.Fingerprint,
		},
		Bin: Bin{
			CardType:      authResponse.EncryptedCardInformation.BinDetail.CardType,
			CardBrand:     authResponse.EncryptedCardInformation.BinDetail.CardBrand,
			IssuerName:    authResponse.EncryptedCardInformation.BinDetail.IssuerName,
			IssuerCountry: authResponse.EncryptedCardInformation.BinDetail.IssuerCountry,
		},
		AuthenticationResponse: *GetAuthenticationResponseFromProcessor(authResponse),
	}
}
