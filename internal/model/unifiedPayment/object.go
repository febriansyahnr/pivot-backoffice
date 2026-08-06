package unifiedPaymentModel

import (
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	constantPayment "github.com/paper-indonesia/pivot-backoffice/constant/payment"
	"github.com/paper-indonesia/pivot-backoffice/pkg/types"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	snapVa "github.com/paper-indonesia/pdk/go/snap/structs/va"
)

type Amount struct {
	Value    float64 `json:"value" validate:"min=0" db:"value"`
	Currency string  `json:"currency" db:"currency"`
}

type Description struct {
	MerchantDescription       string `json:"merchantDescription"`
	StatementDescriptor       string `json:"statementDescriptor" validate:"max=20"`
	StatementDescriptorSuffix string `json:"statementDescriptorSuffix"`

	FinalStatementDescriptor string `json:"finalStatementDescriptor,omitempty"`
}

type PaymentMethodOptions struct {
	VirtualAccount *PaymentMethodOptionVirtualAccount `json:"virtualAccount,omitempty"`
	QR             *PaymentMethodOptionQR             `json:"qr,omitempty"`
	Card           *PaymentMethodOptionCard           `json:"card,omitempty"`
	Ewallet        *PaymentMethodOptionEwallet        `json:"ewallet,omitempty"`
}

// Merge merges base PaymentMethodOptions with override PaymentMethodOptions
// Override values take precedence over base values at the field level
func (p *PaymentMethodOptions) Merge(base *PaymentMethodOptions) *PaymentMethodOptions {
	if p == nil && base == nil {
		return nil
	}
	if p == nil {
		return base
	}
	if base == nil {
		return p
	}

	result := &PaymentMethodOptions{}

	// Merge VirtualAccount
	if p.VirtualAccount != nil {
		result.VirtualAccount = p.VirtualAccount.merge(base.VirtualAccount)
	} else if base.VirtualAccount != nil {
		result.VirtualAccount = base.VirtualAccount
	}

	// Merge QR
	if p.QR != nil {
		result.QR = p.QR.merge(base.QR)
	} else if base.QR != nil {
		result.QR = base.QR
	}

	// Merge Card
	if p.Card != nil {
		result.Card = p.Card.merge(base.Card)
	} else if base.Card != nil {
		result.Card = base.Card
	}

	// Merge Ewallet
	if p.Ewallet != nil {
		result.Ewallet = p.Ewallet.merge(base.Ewallet)
	} else if base.Ewallet != nil {
		result.Ewallet = base.Ewallet
	}

	return result
}

// For Monitoring Data
func (p *PaymentMethodOptions) GetPaymentMethodOptionDetail() string {
	if p.VirtualAccount != nil {
		return p.VirtualAccount.Channel
	}
	if p.Ewallet != nil {
		return p.Ewallet.Channel
	}
	return ""
}

type PaymentMethod struct {
	Type                    string                                   `json:"type" validate:"required,oneof=VIRTUAL_ACCOUNT QR CARD EWALLET"`
	CardPaymentMethodDetail *CardPaymentMethodDetail                 `json:"card,omitempty"`
	QrPaymentMethodDetail   *ChargePaymentMethodDetailQr             `json:"qr,omitempty"`
	VAPaymentMethodDetail   *ChargePaymentMethodDetailVirtualAccount `json:"virtualAccount,omitempty"`
}

// For Monitoring Data
func (p *PaymentMethod) GetPaymentMethodTypeDetail() string {
	if p.CardPaymentMethodDetail != nil {
		if p.CardPaymentMethodDetail.EncryptedCard != "" {
			return "ENCRYPTED_CARD"
		}
		if p.CardPaymentMethodDetail.Token != "" {
			return "SAVED_CARD"
		}
	}
	if p.QrPaymentMethodDetail != nil {
		return p.QrPaymentMethodDetail.Acquirer
	}

	return ""
}

type CardPaymentMethodDetail struct {
	Token         string `json:"token" validate:"omitempty"`         // Should be choice one of token or encryptedCard
	EncryptedCard string `json:"encryptedCard" validate:"omitempty"` // Should be choice one of token or encryptedCard, also contained network token detail
	CVC           string `json:"cvc" validate:"omitempty,numeric,len=3"`
}

type PaymentMethodOptionVirtualAccount struct {
	Channel               string     `json:"channel" validate:"required"`
	VirtualAccountTrxType string     `json:"virtualAccountTrxType,omitempty"`
	VirtualAccountName    string     `json:"virtualAccountName,omitempty"`
	VirtualAccountNumber  string     `json:"virtualAccountNumber,omitempty" validate:"omitempty,numeric"`
	ExpiryAt              *time.Time `json:"expiryAt,omitempty"`

	// common snap payload
	BillDetails *[]snapVa.BillDetail `json:"billDetails,omitempty" validate:"-"`
}

func (p *PaymentMethodOptionVirtualAccount) merge(base *PaymentMethodOptionVirtualAccount) *PaymentMethodOptionVirtualAccount {
	if p == nil {
		return base
	}
	if base == nil {
		return p
	}

	result := &PaymentMethodOptionVirtualAccount{}

	// Override fields - request takes precedence
	if p.Channel != "" {
		result.Channel = p.Channel
	} else {
		result.Channel = base.Channel
	}

	if p.VirtualAccountTrxType != "" {
		result.VirtualAccountTrxType = p.VirtualAccountTrxType
	} else {
		result.VirtualAccountTrxType = base.VirtualAccountTrxType
	}

	if p.VirtualAccountName != "" {
		result.VirtualAccountName = p.VirtualAccountName
	} else {
		result.VirtualAccountName = base.VirtualAccountName
	}

	if p.VirtualAccountNumber != "" {
		result.VirtualAccountNumber = p.VirtualAccountNumber
	} else {
		result.VirtualAccountNumber = base.VirtualAccountNumber
	}

	if p.ExpiryAt != nil {
		result.ExpiryAt = p.ExpiryAt
	} else {
		result.ExpiryAt = base.ExpiryAt
	}

	if p.BillDetails != nil {
		result.BillDetails = p.BillDetails
	} else {
		result.BillDetails = base.BillDetails
	}

	return result
}

type PaymentMethodOptionQR struct {
	ExpiryAt *time.Time `json:"expiryAt,omitempty"`
}

func (p *PaymentMethodOptionQR) merge(base *PaymentMethodOptionQR) *PaymentMethodOptionQR {
	if p == nil {
		return base
	}
	if base == nil {
		return p
	}

	result := &PaymentMethodOptionQR{}

	if p.ExpiryAt != nil {
		result.ExpiryAt = p.ExpiryAt
	} else {
		result.ExpiryAt = base.ExpiryAt
	}

	return result
}

type PaymentMethodOptionCard struct {
	CardOnFile       *PaymentMethodOptionCardOnFileConfig     `json:"cardOnFile,omitempty"`
	CaptureMethod    string                                   `json:"captureMethod,omitempty" validate:"omitempty,oneof=automatic manual AUTOMATIC MANUAL"` // Support backward compatibility for old payload
	ThreeDsMethod    string                                   `json:"threeDsMethod,omitempty" validate:"omitempty,oneof=AUTOMATIC CHALLENGE NEVER EXTERNAL"`
	ProcessingConfig *PaymentMethodOptionCardProcessingConfig `json:"processingConfig,omitempty"`
	Installment      *PaymentMethodOptionCardInstallment      `json:"installment,omitempty"`
	ThreeDsInfo      *PaymentMethodOptionCardThreeDsInfo      `json:"threeDsInfo,omitempty"`
	ExpiryAt         *time.Time                               `json:"expiryAt,omitempty"`
	AutoSplit        *bool                                    `json:"autoSplit,omitempty" validate:"-"`
}

type PaymentMethodOptionCardOnFileConfig struct {
	Initiator                    string `json:"initiator,omitempty" validate:"omitempty,oneof=MERCHANT CUSTOMER"`
	Type                         string `json:"type,omitempty" validate:"omitempty,oneof=SCHEDULED UNSCHEDULED INSTALLMENT"`
	PreviousNetworkTransactionID string `json:"PreviousNetworkTransactionID,omitempty" validate:"required_if=initiator MERCHANT"`
}

func (p *PaymentMethodOptionCard) merge(base *PaymentMethodOptionCard) *PaymentMethodOptionCard {
	if p == nil {
		return base
	}
	if base == nil {
		return p
	}

	result := &PaymentMethodOptionCard{}

	// Override fields - request takes precedence
	if p.CaptureMethod != "" {
		result.CaptureMethod = p.CaptureMethod
	} else {
		result.CaptureMethod = base.CaptureMethod
	}

	if p.ThreeDsMethod != "" {
		result.ThreeDsMethod = p.ThreeDsMethod
	} else {
		result.ThreeDsMethod = base.ThreeDsMethod
	}

	if p.ProcessingConfig != nil {
		result.ProcessingConfig = p.ProcessingConfig.merge(base.ProcessingConfig)
	} else {
		result.ProcessingConfig = base.ProcessingConfig
	}

	if p.Installment != nil {
		result.Installment = p.Installment.merge(base.Installment)
	} else {
		result.Installment = base.Installment
	}

	if p.ThreeDsInfo != nil {
		result.ThreeDsInfo = p.ThreeDsInfo
	} else {
		result.ThreeDsInfo = base.ThreeDsInfo
	}

	if p.ExpiryAt != nil {
		result.ExpiryAt = p.ExpiryAt
	} else {
		result.ExpiryAt = base.ExpiryAt
	}

	result.AutoSplit = p.AutoSplit

	return result
}

type PaymentMethodOptionCardProcessingConfig struct {
	BankMerchantId string `json:"bankMerchantId,omitempty"`
	MerchantIdTag  string `json:"merchantIdTag,omitempty"`
}

func (p *PaymentMethodOptionCardProcessingConfig) merge(base *PaymentMethodOptionCardProcessingConfig) *PaymentMethodOptionCardProcessingConfig {
	if p == nil {
		return base
	}
	if base == nil {
		return p
	}

	result := &PaymentMethodOptionCardProcessingConfig{}

	if p.BankMerchantId != "" {
		result.BankMerchantId = p.BankMerchantId
	} else {
		result.BankMerchantId = base.BankMerchantId
	}

	if p.MerchantIdTag != "" {
		result.MerchantIdTag = p.MerchantIdTag
	} else {
		result.MerchantIdTag = base.MerchantIdTag
	}

	return result
}

type PaymentMethodOptionCardInstallment struct {
	Enabled        bool          `json:"enabled,omitempty"`
	AvailablePlans []interface{} `json:"availablePlans,omitempty"`
	Plan           interface{}   `json:"plan,omitempty"`
}

func (p *PaymentMethodOptionCardInstallment) merge(base *PaymentMethodOptionCardInstallment) *PaymentMethodOptionCardInstallment {
	if p == nil {
		return base
	}
	if base == nil {
		return p
	}

	result := &PaymentMethodOptionCardInstallment{}

	// For boolean, we prioritize request value even if false
	result.Enabled = p.Enabled

	if p.AvailablePlans != nil {
		result.AvailablePlans = p.AvailablePlans
	} else {
		result.AvailablePlans = base.AvailablePlans
	}

	if p.Plan != nil {
		result.Plan = p.Plan
	} else {
		result.Plan = base.Plan
	}

	return result
}

type PaymentMethodOptionCardThreeDsInfo struct {
	// Mandatory
	TransactionID        string `json:"transactionId" validate:"required,uuid"`
	ThreeDSVersion       string `json:"threeDsVersion" validate:"required,min=1,max=20"`
	ECI                  string `json:"eci" validate:"required,oneof=01 02 05 06 07"`
	TransactionStatus    string `json:"transactionStatus" validate:"required,oneof=Y N U A R"`
	AuthenticationScheme string `json:"authenticationScheme" validate:"required,oneof=VISA MASTERCARD JCB AMEX UNIONPAY"`

	// Optional
	CAVV               string `json:"cavv,omitempty" validate:"omitempty,min=28,max=40"`
	ACSTransactionID   string `json:"acsTransactionId,omitempty" validate:"omitempty,len=36"`
	ACSReference       string `json:"acsReference,omitempty" validate:"omitempty,max=32"`
	MCC                string `json:"mcc,omitempty"`            // For now, only collect the data
	BankMerchantId     string `json:"bankMerchantId,omitempty"` // For now, only collect the data
	AuthenticationTime string `json:"authenticationTime,omitempty" validate:"omitempty,datetime=2006-01-02T15:04:05Z07:00"`
}

type PaymentMethodOptionEwallet struct {
	Channel  string     `json:"channel" validate:"required,oneof=SHOPEEPAY DANA"`
	ExpiryAt *time.Time `json:"expiryAt,omitempty"`
}

func (p *PaymentMethodOptionEwallet) merge(base *PaymentMethodOptionEwallet) *PaymentMethodOptionEwallet {
	if p == nil {
		return base
	}
	if base == nil {
		return p
	}

	result := &PaymentMethodOptionEwallet{}

	if p.Channel != "" {
		result.Channel = p.Channel
	} else {
		result.Channel = base.Channel
	}

	if p.ExpiryAt != nil {
		result.ExpiryAt = p.ExpiryAt
	} else {
		result.ExpiryAt = base.ExpiryAt
	}

	return result
}

type RedirectUrl struct {
	SuccessReturnUrl    string `json:"successReturnUrl"`
	FailureReturnUrl    string `json:"failureReturnUrl"`
	ExpirationReturnUrl string `json:"expirationReturnUrl"`
}

// RedirectPaymentUIUrl Only for CC
type RedirectPaymentUIUrl struct {
	SuccessUrl    string `json:"successUrl"`
	FailedUrl     string `json:"failedUrl"`
	ProcessingUrl string `json:"processingUrl,omitempty"`
}

type ChargePaymentData struct {
	CustomerReference       string `json:"customerReference"`
	PaymentPartnerReference string `json:"paymentPartnerReference"`
	PartnerReferenceNo      string `json:"partnerReferenceNo"`
}

type ChargePaymentMethodDetails struct {
	VirtualAccount *ChargePaymentMethodDetailVirtualAccount `json:"virtualAccount,omitempty" db:"virtual_account"`
	Card           *ChargePaymentMethodDetailCard           `json:"card,omitempty" db:"card"`
	Qr             *ChargePaymentMethodDetailQr             `json:"qr,omitempty" db:"qr"`
	Ewallet        *ChargePaymentMethodDetailEwallet        `json:"ewallet,omitempty" db:"ewallet"`

	ProcessorReference            string    `json:"-"`
	ProcessorReferenceID          string    `json:"processorReferenceId,omitempty"`
	ProcessorTransactionID        string    `json:"-"`
	ProcessorTransactionTimestamp time.Time `json:"-"`
	ProcessorExpiredAt            time.Time `json:"-"`
	ProcessorReferenceNo          string    `json:"-"`
	StatementDescriptor           string    `json:"-"`
}

type ChargePaymentMethodDetailVirtualAccount struct {
	Channel               string    `json:"channel"`
	VirtualAccountNumber  string    `json:"virtualAccountNumber"`
	VirtualAccountName    string    `json:"virtualAccountName"`
	VirtualAccountTrxType string    `json:"virtualAccountTrxType,omitempty"`
	ExpiryAt              time.Time `json:"expiryAt"`
	BankReferenceNo       string    `json:"bankReferenceNo,omitempty"`
}

type ChargePaymentMethodDetailQr struct {
	Acquirer                 string    `json:"acquirer"`
	QrContent                string    `json:"qrContent"`
	QrUrl                    string    `json:"qrUrl"`
	QrType                   string    `json:"qrType,omitempty"`
	RetrievalReferenceNumber string    `json:"retrievalReferenceNumber"`
	IssuerName               string    `json:"issuerName"`
	ExpiryAt                 time.Time `json:"expiryAt"`
	MerchantName             string    `json:"merchantName,omitempty"`
	StoreID                  string    `json:"storeId,omitempty"`
}

type ChargePaymentMethodDetailCard struct {
	First6               string                                             `json:"first6"`
	First8               string                                             `json:"first8"`
	Last4                string                                             `json:"last4"`
	ExpMonth             types.String                                       `json:"expMonth"`
	ExpYear              types.String                                       `json:"expYear"`
	Fingerprint          string                                             `json:"fingerprint,omitempty"`
	CardHolderName       string                                             `json:"cardHolderName,omitempty"`
	CardName             string                                             `json:"cardName,omitempty"` // CardName is TAG for OTA Vendor
	BinInformations      ChargePaymentMethodDetailBinInformation            `json:"binInformations"`
	AuthenticationResult *ChargePaymentMethodDetailCardAuthenticationResult `json:"authenticationResult"`
	AuthorizationResult  *ChargePaymentMethodDetailCardAuthorizationResult  `json:"authorizationResult"`
	ACSURL               string                                             `json:"acsUrl,omitempty"`
	ResponseCode         *ChargePaymentMethodDetailCardResponseCode         `json:"responseCode,omitempty"`
	BankMerchantID       string                                             `json:"bankMerchantId,omitempty"`
	SaveForFutureUse     *bool                                              `json:"saveForFutureUse,omitempty"`
	MIDInfo              *MIDInfo                                           `json:"midInfo,omitempty"` // MID info from processor
	ApprovalCode         string                                             `json:"approvalCode,omitempty"`
	IsNetworkToken       bool                                               `json:"isNetworkToken,omitempty"`
	MerchantCategoryCode string                                             `json:"merchantCategoryCode,omitempty"`
	Description          string                                             `json:"description,omitempty"`    // Transaction description
	SettlementDate       string                                             `json:"settlementDate,omitempty"` // Transaction settlement date
	Device               *ChargePaymentMethodDetailCardDevice               `json:"device,omitempty"`
	Error                *ChargePaymentMethodDetailCardError                `json:"error,omitempty"`
	// Customer token, used exclusively for internal purposes.
	Token string `json:"-"`
}

type ChargePaymentMethodDetailEwallet struct {
	AppRedirectURL     string `json:"appRedirectUrl,omitempty"`
	WebRedirectURL     string `json:"webRedirectUrl,omitempty"`
	ReferenceNo        string `json:"referenceNo,omitempty"`
	PartnerReferenceNo string `json:"partnerReferenceNo,omitempty"`
	Channel            string `json:"channel"`
}

type ChargePaymentMethodDetailBinInformation struct {
	Type        string `json:"type"`
	IssuingBank string `json:"issuingBank"`
	Brand       string `json:"brand"`
	Country     string `json:"country"`
}

func (b *ChargePaymentMethodDetailBinInformation) IsForeign() bool {
	if b.Country != "ID" {
		return true
	}
	return false
}

type ChargePaymentMethodDetailCardResponseCode struct {
	AcquirerCode          string `json:"acquirerCode"`
	AcquirerMessage       string `json:"acquirerMessage"`
	GatewayCode           string `json:"gatewayCode"`
	GatewayRecommendation string `json:"gatewayRecommendation"`
}

type ChargePaymentMethodDetailCardError struct {
	Cause       string `json:"cause"`
	Explanation string `json:"explanation"`
}

type ChargePaymentMethodDetailCardDevice struct {
	Browser          string `json:"browser"`
	IPAddress        string `json:"ipAddress"`
	MobilePhoneModel string `json:"mobilePhoneModel"`
}

type ChargePaymentMethodDetailCardAuthenticationResult struct {
	ThreeDsVersion       string     `json:"threeDsVersion"`
	ThreeDsResult        string     `json:"threeDsResult"`
	ThreeDsMethod        string     `json:"threeDsMethod"`
	EciCode              string     `json:"eciCode"`
	TransactionID        string     `json:"transactionId,omitempty"`
	TransactionStatus    string     `json:"transactionStatus,omitempty"`
	AuthenticationScheme string     `json:"authenticationScheme,omitempty"`
	AcsTransactionID     string     `json:"acsTransactionId,omitempty"`
	AcsReference         string     `json:"acsReference,omitempty"`
	AuthenticationTime   *time.Time `json:"authenticationTime,omitempty"`
	// Auto split payment
	CallbackTransactionID string `json:"callbackTransactionId,omitempty"`
}

type ChargePaymentMethodDetailCardAuthorizationResult struct {
	AuthorizationID          string `json:"-"` // Do not expose the auth ID
	NetworkTransactionID     string `json:"networkTransactionId"`
	AcquirerReferenceNumber  string `json:"acquirerReferenceNumber"`
	RetrievalReferenceNumber string `json:"retrievalReferenceNumber"`
	Stan                     string `json:"stan"`
	AvsResult                string `json:"avsResult"`
	CvvResult                string `json:"cvvResult"`
	AuthorizedAmount         Amount `json:"authorizedAmount"`
	IssuerAuthorizationCode  string `json:"issuerAuthorizationCode"`
	MerchantAdviceCode       string `json:"merchantAdviceCode,omitempty"`
}

type MIDInfo struct {
	Acquirer string `json:"acquirer,omitempty"` // should use merchant payment method config as SOT
	MID      string `json:"mid"`
	Type     string `json:"type"`
}

type SummaryTransaction struct {
	SumPaidAmount   float64 `json:"sumPaidAmount"`
	CountPaidAmount int     `json:"countPaidAmount"`
}

func (qr *ChargePaymentMethodDetailQr) Scan(value any) error {
	return util.ScanJSON(value, qr)
}

func (va *ChargePaymentMethodDetailVirtualAccount) Scan(value any) error {
	return util.ScanJSON(value, va)
}

func (cc *ChargePaymentMethodDetailCard) Scan(value any) error {
	return util.ScanJSON(value, cc)
}

func (cc *ChargePaymentMethodDetailEwallet) Scan(value any) error {
	return util.ScanJSON(value, cc)
}

func (a *Amount) Scan(value any) error {
	return util.ScanJSON(value, a)
}

func (c *ChargePaymentMethodDetails) GetAuthorizationCode() string {
	if c == nil {
		return ""
	}

	if c.Card != nil && c.Card.AuthorizationResult != nil {
		return c.Card.AuthorizationResult.IssuerAuthorizationCode
	}

	return ""
}

func (c ChargePaymentMethodDetails) GetCardMID() string {
	if c.Card == nil || c.Card.MIDInfo == nil {
		return ""
	}

	return c.Card.MIDInfo.MID
}

func (c *ChargePaymentMethodDetails) GetNaturalPaymentFailureMessage(paymentMethod string, failureCode string) string {
	if c == nil {
		return ""
	}

	authorizationCode := c.GetAuthorizationCode()
	switch failureCode {
	case constant.FailureCodeDeclinedByChannel:
		switch paymentMethod {
		case constantPayment.PAYMENT_METHOD_CREDIT_CARD:
			// Check for specific authorization codes that warrant different messages
			switch authorizationCode {
			case "54", "101":
				return "Your card has expired. Please use another card or payment method."
			case "N7":
				return "Card verification failed. Please check your card details and try again."
			default:
				return "Payment was declined by issuer. Please contact your card issuer or use another payment method."
			}
		case constantPayment.PAYMENT_METHOD_EWALLET:
			return "Payment was declined by issuer. Please contact your e-wallet issuer or use another payment method."
		}

	case constant.FailureCodeInvalidAccount:
		switch paymentMethod {
		case constantPayment.PAYMENT_METHOD_CREDIT_CARD:
			return "Your card is invalid. Please use another card or payment method."
		case constantPayment.PAYMENT_METHOD_EWALLET:
			return "Your account is not valid. Please use another payment method."
		}

	case constant.FailureCodeAuthenticationFailed:
		return "Payment could not be verified. Please contact your card issuer or use another card."

	case constant.FailureCodeSuspectedFraud:
		return "Payment could not be completed. Please contact your card issuer."

	case constant.FailureCodeBlockedByFDS:
		return "Payment could not be completed at this time. Please use another card or payment method."

	case constant.FailureCodeRequireReview:
		return "Payment in process. Please wait a moment."

	case constant.FailureCodeInsufficientFund:
		return "Insufficient funds. Please add funds or use another payment method."

	case constant.FailureCodeChannelUnavailable:
		return "Payment service is temporarily unavailable. Please try again later or use another payment method."

	case constant.FailureCodeCancelledByUser:
		return "Payment was cancelled."

	case constant.FailureCodeChargeExpired:
		return "Payment session expired. Please try again."

	case constant.FailureCodeExceededCapturePeriod:
		return "Payment could not be completed. Please try again."
	}

	return ""
}
