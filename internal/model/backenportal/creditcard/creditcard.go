package card

import (
	"encoding/json"

	feeModel "github.com/paper-indonesia/pivot-backoffice/internal/model/fee"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	paymentMethodModel "github.com/paper-indonesia/pivot-backoffice/internal/model/paymentMethod"
	pb "github.com/paper-indonesia/pivot-backoffice/internal/model/proto/messages/callback"
	splitRoutingPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/splitRoutingPayment"
)

type CreditcardMetadata struct {
	AuthenticationMethod string                                        `json:"authenticationMethod"`
	BankMerchantID       string                                        `json:"bankMerchantId,omitempty"`
	ProcessorStatus      string                                        `json:"processorStatus,omitempty"`
	CardData             *CardDataRequest                              `json:"cardData,omitempty"`
	AuthorizationData    *PaymentNotificationAuthorizationDataRequest  `json:"authorizationData,omitempty"`
	AuthenticationData   *PaymentNotificationAuthenticationDataRequest `json:"authenticationData,omitempty"`
	FeeDetail            *feeModel.FeeMetadataObject                   `json:"feeDetail,omitempty"`
	RedirectUrl          CreditcardRedirectUrlRequest                  `json:"redirectUrl,omitempty"`
	FeeOnBehalf          *feeModel.TrxFeeOnBehalfMetadata              `json:"feeOnBehalf,omitempty"`
	OnBehalf             *merchantModel.OnBehalfObject                 `json:"onBehalf,omitempty"`

	// Using interface{} to support both V2 (successReturnUrl/failureReturnUrl) and legacy (successUrl/failedUrl) formats
	ClientRedirectUrl        interface{}                                                       `json:"clientRedirectUrl"`
	IsUnifiedPayment         bool                                                              `json:"isUnifiedPayment"`
	StatementDescriptor      string                                                            `json:"statementDescriptor"`
	CardConfig               *CreditCardConfig                                                 `json:"cardConfig,omitempty"`
	ThreeDsMethod            string                                                            `json:"threeDsMethod,omitempty"`
	CaptureMethod            string                                                            `json:"captureMethod,omitempty"`
	AutoSplitPayment         *AutoSplitPayment                                                 `json:"autoSplitPayment,omitempty"`
	CardPartnerConfigs       *paymentMethodModel.SetupPaymentMethodPartnerConfigForCardRequest `json:"cardPartnerConfigs,omitempty"`
	ParentCardPartnerConfigs *paymentMethodModel.SetupPaymentMethodPartnerConfigForCardRequest `json:"parentCardPartnerConfigs,omitempty"`

	SplitRoutingConfigurations *[]splitRoutingPaymentModel.PaymentSplitRoutingConfiguration `json:"splitRoutingConfigurations,omitempty" validate:"omitempty,dive"`
	ProcessingConfig           *CreditCardProcessingConfig                                  `json:"processingConfig,omitempty"`
	RecurringPayment           *RecurringPayment                                            `json:"recurringPayment,omitempty"`

	BypassExternalFdsEvaluation bool `json:"bypassExternalFdsEvaluation,omitempty"`
}

type CreditCardConfig struct {
	SavedFutureUse   bool `json:"savedFutureUse"`
	ShowSavedPayment bool `json:"showSavedPayment"`
}

type CreditCardProcessingConfig struct {
	BankMerchantId string `json:"bankMerchantId,omitempty"`
	MerchantIdTag  string `json:"merchantIdTag,omitempty"`
}

type CardDataRequest struct {
	First8Digit    string `json:"first8Digit"`
	Last4Digit     string `json:"last4Digit"`
	CardType       string `json:"cardType"`
	CardBrand      string `json:"cardBrand"`
	CardIssuing    string `json:"cardIssuing"`
	CountryCode    string `json:"countryCode"`
	IssuingCountry string `json:"issuingCountry"`
	Fingerprint    string `json:"fingerprint"`
	ExpiryMonth    string `json:"expiryMonth"`
	ExpiryYear     string `json:"expiryYear"`
	CardHolderName string `json:"cardHolderName"`
	SavedFutureUse bool   `json:"savedFutureUse"`
	CardName       string `json:"cardName"`
}

type RecurringPayment struct {
	InitiateFirstAuthorization bool                  `json:"initiateFirstAuthorization"`
	FirstAuthorizationMethod   string                `json:"firstAuthorizationMethod,omitempty"`
	FirstAuthorizationOrderID  *string               `json:"firstAuthorizationOrderID,omitempty"`
	BillingCycle               RecurringBillingCycle `json:"billingCycle"`
}

type AutoSplitPayment struct {
	TransactionType string  `json:"transactionType"`
	Processor       string  `json:"processor,omitempty"`
	ProcessorLimit  float64 `json:"processorLimit,omitempty"`
	CITMerchantID   string  `json:"citMerchantId,omitempty"`
	MITMerchantID   string  `json:"mitMerchantId,omitempty"`
}

func (c *CardDataRequest) ToSendCallbackCardDataRequest() SendCallbackCardDataRequest {
	if c == nil {
		return SendCallbackCardDataRequest{}
	}
	return SendCallbackCardDataRequest{
		CardType:    c.CardType,
		CardBrand:   c.CardBrand,
		CardIssuing: c.CardIssuing,
		CountryCode: c.CountryCode,
		Fingerprint: c.Fingerprint,
	}
}

func (c *CardDataRequest) ToPaymentCreditCardDataRequest() *pb.PaymentCreditCardData {
	if c == nil {
		return &pb.PaymentCreditCardData{}
	}
	return &pb.PaymentCreditCardData{
		CardType:    c.CardType,
		CardBrand:   c.CardBrand,
		CardIssuing: c.CardIssuing,
		CountryCode: c.CountryCode,
		Fingerprint: c.Fingerprint,
	}
}

type SendCallbackCardDataRequest struct {
	CardType    string `json:"cardType"`
	CardBrand   string `json:"cardBrand"`
	CardIssuing string `json:"cardIssuing"`
	CountryCode string `json:"countryCode"`
	Fingerprint string `json:"fingerprint"`
}

func UpdateCreditcardMetaData(
	metadata *map[string]any,
	cardData *CardDataRequest,
	AuthorizationData *PaymentNotificationAuthorizationDataRequest,
	AuthenticationData *PaymentNotificationAuthenticationDataRequest,
	processorStatus string,
) error {
	jsonData, err := json.Marshal(metadata)
	if err != nil {
		return err
	}

	var creditCardMetadata CreditcardMetadata
	err = json.Unmarshal(jsonData, &creditCardMetadata)
	if err != nil {
		return err
	}

	if processorStatus != "" {
		creditCardMetadata.ProcessorStatus = processorStatus
	}

	if cardData != nil {
		creditCardMetadata.CardData = cardData
	}

	if AuthenticationData != nil {
		creditCardMetadata.AuthenticationData = AuthenticationData
	}

	if AuthorizationData != nil {
		creditCardMetadata.AuthorizationData = AuthorizationData
	}

	creditCardMetadataByte, _ := json.Marshal(creditCardMetadata)

	_ = json.Unmarshal(creditCardMetadataByte, &metadata)

	return nil
}
