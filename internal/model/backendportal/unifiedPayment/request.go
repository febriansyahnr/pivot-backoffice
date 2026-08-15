package unifiedPaymentModel

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/common"
	creditcardModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/creditcard"
	feeModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/fee"
	paymentMethodModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/paymentMethod"
	splitRoutingPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/splitRoutingPayment"
	pkgErr "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"

	"github.com/shopspring/decimal"
)

type CreateUnifiedPaymentSessionRequest struct {
	// Mandatory Request
	ClientReferenceID string `json:"clientReferenceId" validate:"required"`
	Amount            Amount `json:"amount"`

	// Optional Request
	AutoConfirm                bool                                                         `json:"autoConfirm,omitempty"`
	StatementDescriptor        string                                                       `json:"statementDescriptor,omitempty" validate:"omitempty,max=20"`
	ExpiryAt                   time.Time                                                    `json:"expiryAt,omitempty"`
	Mode                       string                                                       `json:"mode,omitempty"`
	RedirectUrl                RedirectUrl                                                  `json:"redirectUrl" validate:"required"`
	PaymentMethod              *PaymentMethod                                               `json:"paymentMethod" validate:"omitempty"`
	PaymentMethodOptions       PaymentMethodOptions                                         `json:"paymentMethodOptions,omitempty"`
	PaymentType                string                                                       `json:"paymentType,omitempty" validate:"omitempty,oneof=SINGLE MULTIPLE"`
	SplitRoutingConfigurations *[]splitRoutingPaymentModel.PaymentSplitRoutingConfiguration `json:"splitRoutingConfigurations,omitempty" validate:"omitempty,dive"`
	OrderInformation           *PaymentOrder                                                `json:"orderInformation,omitempty" validate:"omitempty"`
	CustomerID                 string                                                       `json:"customerId,omitempty" validate:"omitempty"`
	CustomerInformation        *CustomerInformation                                         `json:"customer,omitempty" validate:"omitempty"`
	Metadata                   interface{}                                                  `json:"metadata,omitempty" validate:"omitempty,dive"`
	SaveForFutureUse           *bool                                                        `json:"saveForFutureUse,omitempty" validate:"omitempty"`
	ShowSavedPayment           *bool                                                        `json:"showSavedPayment,omitempty" validate:"omitempty"`
	ExpirationMode             string                                                       `json:"expirationMode,omitempty" validate:"omitempty,oneof=LOOSE STRICT"`
	// Recurring Payment Feature
	RecurringID                string                `json:"recurringId"`
	InitiateFirstAuthorization bool                  `json:"initiateFirstAuthorization"`
	FirstAuthorizationMethod   string                `json:"-"`
	FirstAuthorizationOrderID  *string               `json:"-"`
	RecurringStatus            string                `json:"-"`
	RecurringBillingCycle      RecurringBillingCycle `json:"-"`
	// A function to remove the prepared payment lock that has already been created, but the calling main function encountered an error.
	CleanupPreparedRecurringPaymentLock func(context.Context) `json:"-" validate:"-"`
	// Bypass status page for mode redirect
	BypassStatusPage bool `json:"bypassStatusPage"`

	PaymentID              string `json:"-"`
	MerchantID             string `json:"-"`
	ParentMerchantID       string `json:"-"`
	PaymentURL             string `json:"-"`
	ShortPaymentURL        string `json:"-"`
	CreatedBy              string `json:"-"`
	CreatedFrom            string
	IsMigratingFromV1      bool                        `json:"-"`
	IsSnap                 bool                        `json:"-"`
	VirtualTerminal        *VirtualTerminal            `json:"-"`
	OneDollarAuthorization *OneDollarAuthorization     `json:"-"`
	CardFundedPayout       *CardFundedPayout           `json:"-"`
	AutoSplitPayment       *AutoSplitPayment           `json:"-"`
	FeeDetail              *feeModel.FeeMetadataObject `json:"-"`
}

func NewCreateUnifiedPaymentSessionRequest() *CreateUnifiedPaymentSessionRequest {
	return &CreateUnifiedPaymentSessionRequest{
		Mode:        constant.UnifiedPaymentModeRedirect,
		PaymentType: constant.UnifiedPaymentTypeSingle,
		Amount: Amount{
			Value:    0,
			Currency: constant.CurrencyIDR,
		},
	}
}

type ConfirmUnifiedPaymentSessionRequest struct {
	PaymentType          string                `json:"paymentType,omitempty"`
	PaymentMethod        *PaymentMethod        `json:"paymentMethod,omitempty"`
	PaymentMethodOptions *PaymentMethodOptions `json:"paymentMethodOptions,omitempty"`

	PaymentSessionID         string                  `json:"-"`
	ClientReferenceID        string                  `json:"-"`
	Fee                      decimal.Decimal         `json:"-"`
	Amount                   Amount                  `json:"-" validate:"omitempty"`
	ExpiryAt                 time.Time               `json:"-"`
	ParentMerchantID         string                  `json:"-"`
	MerchantID               string                  `json:"-"`
	MerchantExternalID       string                  `json:"-"`
	DerivedMID               string                  `json:"-"`
	DerivedMerchantID        string                  `json:"-"`
	DerivedMerchantShortname string                  `json:"-"`
	PaymentMethodChannelType string                  `json:"-"`
	Mode                     string                  `json:"-"`
	StatementDescriptor      string                  `json:"-"`
	RedirectUrl              RedirectPaymentUIUrl    `json:"-"`
	VirtualTerminal          *VirtualTerminal        `json:"-"`
	CardFundedPayout         *CardFundedPayout       `json:"-"`
	OneDollarAuthorization   *OneDollarAuthorization `json:"-"`
	AutoSplitPayment         *AutoSplitPayment       `json:"-"` // This attribute is intended for internal use only.
	// Recurring Payment Feature
	RecurringID                string                `json:"-"`
	InitiateFirstAuthorization bool                  `json:"-"`
	FirstAuthorizationMethod   string                `json:"-"`
	FirstAuthorizationOrderID  *string               `json:"-"`
	RecurringBillingCycle      RecurringBillingCycle `json:"-"`
	// Payment Partner Configuration
	PaymentPartnerConfig   *paymentMethodModel.SetupPaymentMethodPartnerConfigRequest
	UnifiedPaymentMetadata *MetadataUnifiedPayment
	// Flags
	// SkipInsertAccountTransaction prevent creating new ledger
	SkipInsertAccountTransaction bool

	// BypassExternalFDS allowed to empty value
	BypassExternalFDS *bool `json:"-"`
}

type GetUnifiedPaymentSessionRequest struct {
	PaymentSessionID string `json:"-"`
	MerchantID       string `json:"-"`
}

type GetUnifiedPaymentChargeRequest struct {
	ChargeID   string `json:"-"`
	MerchantID string `json:"-"`
}

type CancelUnifiedPaymentSessionRequest struct {
	CancellationReason string `json:"cancellationReason" validate:"required,oneof=REQUESTED_BY_CUSTOMER DUPLICATED VOID_SESSION CHARGE_FAILED FRAUDULENT"`
	Source             string `json:"source,omitempty" validate:"omitempty,oneof=MERCHANT CUSTOMER"`

	PaymentSessionID string `json:"-"`
	MerchantID       string `json:"-"`
}

type BaseProcessorRequest struct {
	PaymentMethod            *PaymentMethod
	Mode                     string
	PaymentMethodType        string
	PaymentMethodOptions     *PaymentMethodOptions
	PaymentMethodChannelType string
	Fee                      decimal.Decimal
	Amount                   Amount
	ExpiryAt                 time.Time
	PaymentPartnerConfig     *paymentMethodModel.SetupPaymentMethodPartnerConfigRequest // Payment Partner Configuration

	PaymentID          string
	ClientReferenceID  string
	ChargeID           string
	MerchantID         string
	MerchantExternalID string

	DerivedMID               string
	DerivedMerchantID        string
	DerivedMerchantShortName string
	SuccessRedirectUrl       string
	PaymentURL               string
	IsStaticPayment          bool
	IsSnap                   bool

	// Recurring Payment Feature
	RecurringID                string
	InitiateFirstAuthorization bool
	FirstAuthorizationMethod   string
	FirstAuthorizationOrderID  *string
	RecurringBillingCycle      RecurringBillingCycle

	// Card Funded Payout
	CardFundedPayout *CardFundedPayout

	// Auto Split Payment
	AutoSplitPayment *AutoSplitPayment

	// Card On File
	CardOnFile *CardOnFile
}

type CardOnFile struct {
	Initiator                    string
	Type                         string
	PreviousNetworkTransactionID string
}

type InitProcessorVARequest struct {
	*BaseProcessorRequest

	VANumber              string    `json:"vaNumber"`
	VAAccountName         string    `json:"vAAccountName"`
	Acquirer              string    `json:"acquirer"`
	VirtualAccountTrxType string    `json:"virtualAccountTrxType"`
	ExpiryAt              time.Time `json:"expiryAt"`
}

type InitProcessorQRISRequest struct {
	*BaseProcessorRequest

	ExpiryAt time.Time `json:"expiryAt"`
}

type GetListFilterRequest struct {
	UUID              string     `json:"uuid"`
	MerchantID        string     `json:"merchantId"`
	ClientReferenceID string     `json:"clientReferenceID"`
	PaymentMethodType string     `json:"paymentMethodType"`
	Status            string     `json:"status"`
	StartCreatedAt    *time.Time `json:"startCreatedAt"`
	EndCreatedAt      *time.Time `json:"endCreatedAt"`

	Sort    string `json:"sort"`
	SortBy  string `json:"sortBy"`
	Page    int    `json:"page"`
	PerPage int    `json:"perPage"`
}

type PaymentNotificationRequest struct {
	PaymentSessionID       string    `json:"paymentSessionId"`
	PaymentMethodType      string    `json:"paymentMethodType"`
	ChargeID               string    `json:"chargeId"`
	ChargeStatus           string    `json:"chargeStatus"`
	Amount                 Amount    `json:"amount"`
	CapturedAmount         *Amount   `json:"capturedAmount"`
	AuthorizedAmount       *Amount   `json:"authorizedAmount"`
	Processor              string    `json:"processor"`
	ProcessorID            string    `json:"processorId"`
	ProcessorTransactionID string    `json:"processorTransactionId"`
	TrxDatetime            time.Time `json:"trxDatetime"`

	ChargePaymentMethodDetails *ChargePaymentMethodDetails `json:"-"`
	ProcessorReferenceNumber   string                      `json:"-"`
}

type FilterChargeRequest struct {
	UUID              string    `json:"uuid"`
	PaymentSessionID  string    `json:"paymentSessionId"`
	MerchantID        string    `json:"merchantId"`
	ClientReferenceID string    `json:"clientReferenceId"`
	Status            string    `json:"status"`
	StartCreatedAt    time.Time `json:"startCreatedAt"`
	EndCreatedAt      time.Time `json:"endCreatedAt"`
	StartPaymentDate  time.Time `json:"startPaymentDate"`
	EndPaymentDate    time.Time `json:"endPaymentDate"`

	// internal
	PaymentTypes []string `json:"-"`

	Sort    string `json:"sort"`
	SortBy  string `json:"sortBy"`
	Page    int    `json:"page"`
	PerPage int    `json:"perPage"`
}

type SimulatePaymentRequest struct {
	PaymentSessionID string `json:"paymentSessionId" validate:"required"`
	ChargeID         string `json:"chargeId"`
	ChargeStatus     string `json:"chargeStatus" validate:"required,oneof=SUCCESS FAILED EXPIRED PROCESSING"`
	Amount           Amount `json:"amount,omitempty"`

	MerchantID string `json:"-"`
}
type PaymentOrder struct {
	ProductDetails      []*ProductDetail     `json:"productDetails" validate:"required,dive"`
	BillingInformation  *BillingInformation  `json:"billingInfo,omitempty" validate:"omitempty"`
	ShippingInformation *ShippingInformation `json:"shippingInfo,omitempty" validate:"omitempty"`
}

type ProductDetail struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Type        string  `json:"type" validate:"required,oneof=PHYSICAL SERVICE DIGITAL"`
	Category    string  `json:"category" example:"FOOD, FASHION, ELECTRONICS"`
	SubCategory string  `json:"subCategory"`
	Quantity    float64 `json:"quantity" validate:"required"`
	Price       Amount  `json:"price" validate:"required"`
}

type BillingInformation struct {
	GivenName string `json:"givenName" validate:"required"`
	// will accept both SureName and Surname
	// for backward compatibility
	Surname       string                     `json:"surname"`
	SureName      string                     `json:"sureName"`
	Email         string                     `json:"email" validate:"required,email"`
	PhoneNumber   *UnifiedPaymentPhoneNumber `json:"phoneNumber" validate:"omitempty"`
	Address1      string                     `json:"addressLine1"`
	Address2      string                     `json:"addressLine2"`
	City          string                     `json:"city"`
	ProvinceState string                     `json:"provinceState"`
	Country       string                     `json:"country"`
	PostalCode    string                     `json:"postalCode"`
}

type ShippingInformation struct {
	GivenName string `json:"givenName" validate:"required"`
	// will accept both SureName and Surname
	// for backward compatibility
	Surname        string                    `json:"surname"`
	SureName       string                    `json:"sureName"`
	Email          string                    `json:"email" validate:"required,email"`
	PhoneNumber    UnifiedPaymentPhoneNumber `json:"phoneNumber" validate:"omitempty"`
	Address1       string                    `json:"addressLine1" validate:"required"`
	Address2       string                    `json:"addressLine2"`
	City           string                    `json:"city" validate:"required"`
	ProvinceState  string                    `json:"provinceState" validate:"required"`
	Country        string                    `json:"country" validate:"required"`
	PostalCode     string                    `json:"postalCode"`
	ShippingMethod string                    `json:"method" validate:"required,oneof=REGULAR NEXTDAY SAMEDAY INSTANT"`
	ShippingFee    Amount                    `json:"shippingFee"`
}

type CustomerInformation struct {
	CustomerID string `json:"-"`
	GivenName  string `json:"givenName" `
	// will accept both SureName and Surname
	// for backward compatibility
	Surname              *string                         `json:"surname,omitempty"`
	SureName             string                          `json:"sureName"`
	Email                string                          `json:"email" validate:"required,email"`
	PhoneNumber          *UnifiedPaymentPhoneNumber      `json:"phoneNumber" validate:"omitempty"`
	RefundPreference     *UnifiedPaymentRefundPreference `json:"refundPreference,omitempty" validate:"omitempty"`
	StoredPaymentMethods []*CustomerPaymentMethod        `json:"storedPaymentMethods,omitempty"`
}

type UnifiedPaymentPhoneNumber struct {
	CountryCode string `json:"countryCode" validate:"required" example:"+62"`
	Number      string `json:"number" validate:"required,number" example:"81234567890"`
}

type UnifiedPaymentRefundPreference struct {
	Method              string                     `json:"method" validate:"required,oneof=AUTO TRANSFER_ONLY"`
	TransferDestination *RefundTransferDestination `json:"transferDestination" validate:"required_if=Method TRANSFER_ONLY"`
}

type RefundTransferDestination struct {
	ChannelCode        string                    `json:"channelCode" validate:"required" example:"BRI, BCA, MANDIRI, GOPAY, OVO, LINKAJA"`
	ChannelInformation *RefundChannelInformation `json:"channelInformation" validate:"required"`
}

type RefundChannelInformation struct {
	AccountNumber string `json:"accountNumber" validate:"required"`
	AccountName   string `json:"accountName" validate:"required"`
}

type CustomerPaymentMethod struct {
	Token          string    `json:"token" validate:"required"`
	PaymentMethod  string    `json:"paymentMethod" validate:"required"`
	PaymentChannel string    `json:"paymentChannel" validate:"required"`
	Status         string    `json:"status" validate:"required"`
	CreatedAt      time.Time `json:"createdAt" validate:"required"`

	Card *CustomerPaymentMethodCard `json:"card,omitempty"`
}

type CustomerPaymentMethodCard struct {
	Fingerprint         string `json:"fingerprint,omitempty" validate:"required"`
	Network             string `json:"network"`
	First6              string `json:"first6"`
	First8              string `json:"first8"`
	Last4               string `json:"last4"`
	ExpMonth            string `json:"expMonth"`
	ExpYear             string `json:"expYear"`
	CardHolderFirstName string `json:"cardHolderFirstName"`
	CardHolderLastName  string `json:"cardHolderLastName"`
	CardHolderEmail     string `json:"cardHolderEmail,omitempty"`
	CardHolderPhone     string `json:"cardHolderPhone,omitempty"`
	CardName            string `json:"cardName,omitempty"`
	IssuingBank         string `json:"issuingBank,omitempty"`
	CardOrigin          string `json:"cardOrigin,omitempty"`
}

func (r *CreateUnifiedPaymentSessionRequest) IsStaticPayment() bool {
	return r.PaymentType == constant.UnifiedPaymentTypeMultiple
}

func (r *ConfirmUnifiedPaymentSessionRequest) IsStaticPayment() bool {
	return r.PaymentType == constant.UnifiedPaymentTypeMultiple
}

func (r *ConfirmUnifiedPaymentSessionRequest) IsAutoSplitPaymentAuth() bool {
	return r.AutoSplitPayment != nil &&
		r.AutoSplitPayment.TransactionType == constant.AutoSplitPaymentTypeAuthentication
}

func (r *ConfirmUnifiedPaymentSessionRequest) IsAutoSplitSubPayments() bool {
	return r.AutoSplitPayment != nil &&
		(r.AutoSplitPayment.TransactionType == constant.AutoSplitPaymentTypeFirstPayment || r.AutoSplitPayment.TransactionType == constant.AutoSplitPaymentTypeSubsequentPayment)
}

func (r *CreateUnifiedPaymentSessionRequest) SetDefaultExpiryAtIfNotSet() {
	if r.PaymentType == constant.UnifiedPaymentTypeMultiple {
		return
	}

	if !r.ExpiryAt.IsZero() {
		return
	}

	r.ExpiryAt = time.Now().Add(15 * time.Minute)
}

func (r *CreateUnifiedPaymentSessionRequest) RecurringPaymentType() string {
	if r.RecurringID == "" {
		return ""
	}
	if r.InitiateFirstAuthorization {
		return constant.RecurringPaymentTypeFirstAuthorization
	}
	return constant.RecurringPaymentTypeSubsequentPayment
}

func (r *CreateUnifiedPaymentSessionRequest) IsAutoSplitCardPayment() bool {
	return r.PaymentMethodOptions.Card != nil &&
		r.PaymentMethodOptions.Card.AutoSplit != nil &&
		*r.PaymentMethodOptions.Card.AutoSplit
}

func (r *CreateUnifiedPaymentSessionRequest) PrepareAutoSplitCardPayment(splitConfig *paymentMethodModel.SplitCardPaymentConfig, partnerConfig *paymentMethodModel.SetupPaymentMethodPartnerConfigRequest, processorLimitDefault float64) error {
	processorLimit := processorLimitDefault
	if processor, ok := splitConfig.Processors[splitConfig.ActiveProcessor]; ok {
		processorLimit = processor.Limit
	}
	r.AutoSplitPayment = &AutoSplitPayment{
		TransactionType: constant.AutoSplitPaymentTypeAuthentication,
		Processor:       splitConfig.ActiveProcessor,
		ProcessorLimit:  processorLimit,
	}
	if r.PaymentMethodOptions.Card.ProcessingConfig == nil {
		r.PaymentMethodOptions.Card.ProcessingConfig = &PaymentMethodOptionCardProcessingConfig{}
	}
	r.PaymentMethodOptions.Card.ThreeDsMethod = constant.CardThreeDsMethodChallenge

	for _, cnf := range partnerConfig.Card.Items {
		if cnf.PartnerProcessor != splitConfig.ActiveProcessor {
			continue
		}
		switch cnf.SplitPaymentType {
		case constant.CardTransactionTypeCIT:
			r.AutoSplitPayment.CITMerchantID = cnf.AcquirerMerchantID
			r.PaymentMethodOptions.Card.ProcessingConfig.BankMerchantId = cnf.AcquirerMerchantID

		case constant.CardTransactionTypeMIT:
			r.AutoSplitPayment.MITMerchantID = cnf.AcquirerMerchantID
		}
	}
	invalidSplitConfig := r.AutoSplitPayment.CITMerchantID == "" ||
		r.AutoSplitPayment.MITMerchantID == "" ||
		r.AutoSplitPayment.ProcessorLimit == 0
	if invalidSplitConfig {
		return pkgErr.New(response.HttpErrRequest, errors.New("invalid configuration for split card payment"))
	}
	return nil
}

func (r *CreateUnifiedPaymentSessionRequest) HasCardOnFile() bool {
	return r.PaymentMethodOptions.Card != nil && r.PaymentMethodOptions.Card.CardOnFile != nil
}

func (r *ConfirmUnifiedPaymentSessionRequest) HasCardOnFile() bool {
	return r.PaymentMethodOptions != nil && r.PaymentMethodOptions.Card != nil && r.PaymentMethodOptions.Card.CardOnFile != nil
}

func (r *FilterChargeRequest) HashFilter(timezone string) string {
	endDate := r.EndCreatedAt
	if time.Now().UTC().Before(r.EndCreatedAt) {
		endDate = time.Now().UTC()
	}

	buf := bytes.NewBufferString(
		r.MerchantID + "|" + r.StartCreatedAt.Format(time.DateTime) + "|" + endDate.Format(time.DateTime),
	)

	if r.Status != "" {
		_, _ = buf.WriteString("|" + r.Status)
	}
	if r.UUID != "" {
		_, _ = buf.WriteString("|" + r.UUID)
	}
	if r.ClientReferenceID != "" {
		_, _ = buf.WriteString("|" + r.UUID)
	}
	if r.PaymentSessionID != "" {
		_, _ = buf.WriteString("|" + r.UUID)
	}
	_, _ = buf.WriteString("|" + r.Sort + "|" + timezone)

	hash := sha256.Sum256(buf.Bytes())
	return hex.EncodeToString(hash[:])
}

func (r CreateUnifiedPaymentSessionRequest) GetEwalletChannel() string {
	if r.PaymentMethodOptions.Ewallet == nil {
		return ""
	}

	return r.PaymentMethodOptions.Ewallet.Channel
}

func (r CreateUnifiedPaymentSessionRequest) IsAPIMode() bool {
	return r.Mode == constant.UnifiedPaymentModeAPI
}

// RemoveCardTokenizationRequest is a structure to request removal of card tokenization from a customer
type RemoveCardTokenizationRequest struct {
	MerchantID string `validate:"required,uuid"`
	PaymentID  string `validate:"required,uuid"`
	CustomerID string `validate:"required,uuid"`
	TokenID    string `validate:"required,uuid"`
}

type CaptureRequest struct {
	ReleaseRemainingAmount bool    `json:"releaseRemainingAmount"`
	Amount                 *Amount `json:"amount" validate:"omitempty,required_if=ReleaseRemainingAmount FALSE"`

	PaymentID  string `json:"-"`
	ChargeID   string `json:"-"`
	MerchantID string `json:"-"`
}

type ProcessCaptureRequest struct {
	ID string `json:"id"`
}

type GeneratePaymentTokenRequest struct {
	PaymentID string `json:"paymentId" validate:"required"`
	ExpiryAt  string `json:"expiryAt" validate:"required"`
}

func (b *BillingInformation) GetSurname() string {
	if b.Surname != "" {
		return b.Surname
	}
	return b.SureName
}

func (s *ShippingInformation) GetSurname() string {
	if s.Surname != "" {
		return s.Surname
	}
	return s.SureName
}

func (c *CustomerInformation) GetSurname() string {
	if c.Surname != nil {
		return *c.Surname
	}
	return c.SureName
}

type GetBinDetailRequest struct {
	MerchantId string
	BinNumber  string
}

type RecurringBillingCycle struct {
	Interval               uint8   `json:"interval"`
	IntervalUnit           string  `json:"intervalUnit" examples:"DAY or MONTH or YEAR"`
	Count                  uint16  `json:"count,omitempty"`
	ExpiryDate             string  `json:"expiryDate" example:"2026-01-31"`
	MinDaysBetweenPayments uint16  `json:"minDaysBetweenPayments"`
	MinAmountPerPayment    float64 `json:"minAmountPerPayment"`
	MaxAmountPerPayment    float64 `json:"maxAmountPerPayment"`
}

type VirtualTerminal struct {
	BatchID            string   `json:"batchId"`
	BookingID          string   `json:"bookingId"`
	TravelAgentCode    string   `json:"travelAgentCode"`
	TravelAgentName    string   `json:"travelAgentName"`
	Remarks            string   `json:"remarks"`
	AcquirerMerchantID string   `json:"acquirerMerchantId"`
	AllowedBinNumbers  []string `json:"allowedBinNumbers"`
	AllowedCardTypes   []string `json:"allowedCardTypes"`
	AllowedPrincipal   []string `json:"allowedPrincipal"`
}

type OneDollarAuthorization struct {
	UseCase string `json:"useCase"`
}

type CardFundedPayout struct {
	FirstPaymentID   string                     `json:"-"`
	SettlementMethod string                     `json:"-"`
	Sequence         int                        `json:"-"`
	Count            int                        `json:"-"`
	FeeAmount        float64                    `json:"-"`
	FeeConfig        feeModel.FeeMetadataObject `json:"-"`
	VendorID         string                     `json:"-"`
	VendorName       string                     `json:"-"`
	CardID           string                     `json:"-"`
	CardToken        string                     `json:"-"`
}

type AutoSplitPayment struct {
	TransactionType string  `json:"transactionType"`
	Processor       string  `json:"processor,omitempty"`
	ProcessorLimit  float64 `json:"processorLimit,omitempty"`
	CITMerchantID   string  `json:"citMerchantId,omitempty"`
	MITMerchantID   string  `json:"mitMerchantId,omitempty"`

	// for sequence payment
	Sequence         int    `json:"sequence,omitempty"`
	FirstPaymentID   string `json:"firstPaymentID,omitempty"`
	OrderReferenceID string `json:"orderReferenceId,omitempty"` // store parent payment id (authenticated from mpgs)

	// Optional when its done
	Summary *AutoSplitPaymentSummary `json:"summary,omitempty"`
}

func (a AutoSplitPayment) ToCardAutoSplitPayment() *creditcardModel.AutoSplitPayment {
	return &creditcardModel.AutoSplitPayment{
		TransactionType: a.TransactionType,
		Processor:       a.Processor,
		ProcessorLimit:  a.ProcessorLimit,
		CITMerchantID:   a.CITMerchantID,
		MITMerchantID:   a.MITMerchantID,
	}
}

func (p PaymentNotificationRequest) GetCardFingerprintID() string {
	if p.ChargePaymentMethodDetails == nil || p.ChargePaymentMethodDetails.Card == nil {
		return ""
	}

	return p.ChargePaymentMethodDetails.Card.Fingerprint
}

func (b BaseProcessorRequest) ShouldAuthenticateEncryptedCard() bool {
	if b.AutoSplitPayment != nil && b.AutoSplitPayment.TransactionType != constant.AutoSplitPaymentTypeAuthentication {
		return false
	}

	if b.CardFundedPayout != nil && b.CardFundedPayout.Sequence == 1 {
		return true
	}

	return b.Mode == constant.UnifiedPaymentModeAPI
}

func (m PaymentNotificationRequest) GetCardThreeDSCallbackID() string {
	if m.ChargePaymentMethodDetails == nil || m.ChargePaymentMethodDetails.Card == nil {
		return ""
	}

	if m.ChargePaymentMethodDetails.Card.AuthenticationResult != nil {
		return m.ChargePaymentMethodDetails.Card.AuthenticationResult.CallbackTransactionID
	}

	return ""
}

type CardAuthenticationRequest struct {
	PaymentID           string                         `json:"payment_id"`
	MerchantID          string                         `json:"merchant_id"`
	ClientTransactionID string                         `json:"client_transaction_id"`
	Amount              float64                        `json:"amount"`
	Currency            string                         `json:"currency"`
	Card                *CardAuthenticationRequestCard `json:"card,omitempty"`
	ThreeDSCallbackID   string                         `json:"threeds_callback_id,omitempty"`
	AutoSplitPayment    *CardAuthAutoSplitPayment      `json:"auto_split_payment,omitempty"`
}

type CardAuthenticationRequestCard struct {
	Fingerprint string `json:"fingerprint"`
}

type CardAuthAutoSplitPayment struct {
	TransactionType  string `json:"transaction_type"`
	Sequence         int    `json:"sequence"`
	FirstPaymentID   string `json:"first_payment_id"`
	OrderReferenceID string `json:"order_reference_id"`
}

type AutoSplitPaymentSummary struct {
	Status                      string             `json:"status"` // PROCESSING | PARTIAL_SUCCESS | SUCCESS | FAILED | CANCELLED
	NumberOfCharges             int                `json:"numberOfCharges"`
	NumberOfSuccessfulCharges   int                `json:"numberOfSuccessfulCharges"` // PAID
	NumberOfInProcessCharges    int                `json:"numberOfInProcessCharges"`  // PROCESSING
	NumberOfFailedCharges       int                `json:"numberOfFailedCharges"`     // CANCELLED | EXPIRED
	TotalSuccessfulChargeAmount commonModel.Amount `json:"totalSuccessfulChargeAmount"`
	TotalInProgressChargeAmount commonModel.Amount `json:"totalInProgressChargeAmount"`
	TotalFailedChargeAmount     commonModel.Amount `json:"totalFailedChargeAmount"`
	ChargeDetails               []ChargeResponse   `json:"chargeDetails"`
}

func (m AutoSplitPaymentSummary) ToAutoSplitDetail() *AutoSplitDetails {
	for _, c := range m.ChargeDetails {
		c.RemoveUnusedResponse()
	}

	return &AutoSplitDetails{
		Status:                    m.Status,
		NumberOfCharges:           m.NumberOfCharges,
		NumberOfSuccessfulCharges: m.NumberOfSuccessfulCharges,
		NumberOfInProcessCharges:  m.NumberOfInProcessCharges,
		NumberOfFailedCharges:     m.NumberOfFailedCharges,
		TotalSuccessfulChargeAmount: &Amount{
			Value:    m.TotalSuccessfulChargeAmount.ToDecimal().InexactFloat64(),
			Currency: m.TotalSuccessfulChargeAmount.Currency,
		},
		TotalFailedChargeAmount: &Amount{
			Value:    m.TotalFailedChargeAmount.ToDecimal().InexactFloat64(),
			Currency: m.TotalFailedChargeAmount.Currency,
		},
		TotalInProcessChargeAmount: &Amount{
			Value:    m.TotalInProgressChargeAmount.ToDecimal().InexactFloat64(),
			Currency: m.TotalInProgressChargeAmount.Currency,
		},
		ChargesDetails: m.ChargeDetails,
	}
}
