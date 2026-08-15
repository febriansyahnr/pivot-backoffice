package card

import (
	"time"

	"github.com/google/uuid"
	creditcardCoreProcessorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/creditcardCoreProcessor"
	splitRoutingPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/splitRoutingPayment"
	"github.com/shopspring/decimal"
)

type CreateCardPaymentRequest struct {
	PaymentUUID          uuid.UUID                    `json:"uuid"`
	ReferenceID          string                       `json:"referenceId" validate:"required"`
	BankMerchantID       string                       `json:"bankMerchantId"`
	Amount               decimal.Decimal              `json:"amount" validate:"required"`
	Currency             string                       `json:"currency" validate:"required"`
	AuthenticationMethod string                       `json:"authenticationMethod" validate:"required"`
	MerchantID           uuid.UUID                    `json:"-"`
	RedirectUrl          CreditcardRedirectUrlRequest `json:"redirectUrl"`

	UnifiedPaymentRedirectUrl UnifiedPaymentRedirectUrl `json:"-"`
	IsUnifiedPayment          bool                      `json:"-"`
	CreatedBy                 string                    `json:"-"`

	SplitRoutingConfigurations *[]splitRoutingPaymentModel.PaymentSplitRoutingConfiguration `json:"splitRoutingConfigurations,omitempty" validate:"omitempty,dive"`
}

type CardPaymentNotificationRequest struct {
	Event string                         `json:"event" validate:"required"`
	Data  PaymentNotificationDataRequest `json:"data"`
}

type PaymentNotificationDataRequest struct {
	PaymentUUID           uuid.UUID                                     `json:"uuid" validate:"required"`
	TransactionID         uuid.UUID                                     `json:"transactionId"`
	AuthenticationMethod  string                                        `json:"authenticationMethod" validate:"required"`
	AcquirerTransactionID string                                        `json:"acquirer_transaction_id" validate:"required"`
	MerchantID            uuid.UUID                                     `json:"merchantId" validate:"required"`
	MerchantCategoryCode  string                                        `json:"merchantCategoryCode" validate:"-"`
	BankMerchantID        string                                        `json:"bankMerchantId"`
	ReferenceID           string                                        `json:"referenceId" validate:"required"`
	Amount                decimal.Decimal                               `json:"amount" validate:"required"`
	Currency              string                                        `json:"currency" validate:"required"`
	PaymentURL            string                                        `json:"paymentUrl"`
	CardData              *CardDataRequest                              `json:"cardData"`
	AuthorizationData     *PaymentNotificationAuthorizationDataRequest  `json:"authorizationData"`
	AuthenticationData    *PaymentNotificationAuthenticationDataRequest `json:"authenticationData"`
	ResponseCode          *PaymentNotificationResponseCode              `json:"responseCode,omitempty"`
	PaymentStatus         string                                        `json:"paymentStatus" validate:"required"`
	Description           string                                        `json:"description,omitempty" validate:"-"` // Transaction description
	SettlementDate        string                                        `json:"settlementDate,omitempty"`           // Transaction settlement date
	Updated               time.Time                                     `json:"updated" validate:"required"`
	RedirectUrl           CreditcardRedirectUrlRequest                  `json:"redirectUrl"`
	MIDInfo               *MIDInfo                                      `json:"midInfo,omitempty"`
	Type                  string                                        `json:"type,omitempty"`
	IsNetworkToken        bool                                          `json:"isNetworkToken,omitempty"`
	Device                *PaymentNotificationDevice                    `json:"device,omitempty" validate:"-"`
	Error                 *PaymentNotificationError                     `json:"error,omitempty" validate:"-"`
}

type SendCallbackPaymentNotificationDataRequest struct {
	Created        string                      `json:"created" validate:"required"`
	PaymentUUID    string                      `json:"uuid" validate:"required"`
	MerchantID     string                      `json:"merchantId" validate:"required"`
	BankMerchantID string                      `json:"bankMerchantId"`
	ReferenceID    string                      `json:"referenceId" validate:"required"`
	Amount         decimal.Decimal             `json:"amount" validate:"required"`
	Currency       string                      `json:"currency" validate:"required"`
	PaymentURL     string                      `json:"paymentUrl"`
	CardData       SendCallbackCardDataRequest `json:"cardData"`
	PaymentStatus  string                      `json:"paymentStatus" validate:"required"`
	Updated        string                      `json:"updated" validate:"required"`
}

type CreditcardRedirectUrlRequest struct {
	SuccessUrl    string `json:"successUrl" validate:"required"`
	FailedUrl     string `json:"failedUrl" validate:"required"`
	ProcessingUrl string `json:"processingUrl,omitempty" validate:"-"`
}

type UnifiedPaymentRedirectUrl struct {
	SuccessUrl string `json:"successUrl"`
	FailedUrl  string `json:"failedUrl"`
}

type PaymentNotificationAuthorizationDataRequest struct {
	AuthorizationResult      string `json:"authorizationResult"`
	OrderID                  string `json:"orderId"`
	TransactionStaus         string `json:"transactionStatus"`
	AuthorizationID          string `json:"authorizationId"`
	ApprovalCode             string `json:"approvalCode"`
	BankMerchantID           string `json:"bankMerchantId"`
	AcquirerTransactionID    string `json:"acquirerTransactionId"`
	TransactionReference     string `json:"transactionReference"`
	RetrievalReferenceNumber string `json:"retrievalReferenceNumber"`
	CvvResult                string `json:"cvvResult"`
	AcquirerResponseCode     string `json:"acquirerResponseCode"`
	Stan                     string `json:"stan"`
	AvsResult                string `json:"avsResult"`
	MerchantAdviceCode       string `json:"merchantAdviceCode"`
}

type PaymentNotificationResponseCode struct {
	AcquirerCode          string `json:"acquirerCode"`
	AcquirerMessage       string `json:"acquirerMessage"`
	GatewayCode           string `json:"gatewayCode"`
	GatewayRecommendation string `json:"gatewayRecommendation"`
}

type PaymentNotificationAuthenticationDataRequest struct {
	AuthenticationResult string     `json:"authenticationResult"`
	AuthenticationID     string     `json:"authenticationId"`
	PaRes                string     `json:"paRes"`
	VeRes                string     `json:"veRes"`
	XID                  string     `json:"xid"`
	CAVV                 string     `json:"cavv"`
	EciCode              string     `json:"eciCode"`
	ThreeDsVer           string     `json:"3dsVer"`
	ChallengeCode        string     `json:"challengeCode"`
	TransactionID        string     `json:"transactionId,omitempty"`
	TransactionStatus    string     `json:"transactionStatus,omitempty"`
	AuthenticationScheme string     `json:"authenticationScheme,omitempty"`
	AcsTransactionID     string     `json:"acsTransactionId,omitempty"`
	AcsReference         string     `json:"acsReference,omitempty"`
	AuthenticationTime   *time.Time `json:"authenticationTime,omitempty"`
	// Auto split payment
	CallbackTransactionID string `json:"callbackTransactionId,omitempty"`
}

type PaymentNotificationError struct {
	Cause       string `json:"cause"`
	Explanation string `json:"explanation"`
}

type PaymentNotificationDevice struct {
	Browser          string `json:"browser"`
	IPAddress        string `json:"ipAddress"`
	MobilePhoneModel string `json:"mobilePhoneModel"`
}

type VoidRequest struct {
	MerchantID  uuid.UUID `json:"merchantId" validate:"required"`
	ReferenceID string    `json:"referenceId" validate:"required"`
}

type GetTransactionListRequest struct {
	Page                int    `query:"page,omitempty" validate:"min=0"`
	PerPage             int    `query:"perPage,omitempty" validate:"min=0"`
	DateFrom            string `query:"dateFrom"`
	DateTo              string `query:"dateTo"`
	TrxType             string `query:"type"`
	ChargeStatus        string `query:"chargeStatus"`
	VoidStatus          string `query:"voidStatus"`
	ClientTransactionID string `query:"clientTransactionId"`
	IssuingBank         string `query:"issuingBank"`
	CardFingerprint     string `query:"cardFingerprint"`
	PaymentUUID         string `query:"paymentUuid"`
	MerchantID          string `query:"merchantId"`
	ChargeFrom          string `query:"chargeFrom"`
	ChargeTo            string `query:"chargeTo"`
	RefundFrom          string `query:"refundFrom"`
	RefundTo            string `query:"refundTo"`
}

func (r *GetTransactionListRequest) ToCreditcardCoreGetTransactionListRequest() *creditcardCoreProcessorModel.GetTransactionListRequest {
	return &creditcardCoreProcessorModel.GetTransactionListRequest{
		Limit:               r.PerPage,
		Page:                r.Page,
		DateFrom:            r.DateFrom,
		DateTo:              r.DateTo,
		TrxType:             r.TrxType,
		ChargeStatus:        r.ChargeStatus,
		VoidStatus:          r.VoidStatus,
		ClientTransactionID: r.ClientTransactionID,
		IssuingBank:         r.IssuingBank,
		CardFingerprint:     r.CardFingerprint,
		PaymentUUID:         r.PaymentUUID,
		MerchantID:          r.MerchantID,
		ChargeFrom:          r.ChargeFrom,
		ChargeTo:            r.ChargeTo,
		RefundFrom:          r.RefundFrom,
		RefundTo:            r.RefundTo,
	}
}

type CreateMIDRequest struct {
	Mid                string   `json:"mid" validate:"required"`
	Name               string   `json:"name" validate:"required"`
	Description        string   `json:"description"`
	Type               string   `json:"type"`
	TransactionType    string   `json:"transactionType"`
	InstallmentType    string   `json:"installmentType"`
	InstallmentTenor   int      `json:"installmentTenor"`
	Processor          string   `json:"processor" validate:"required"`
	PrincipalAvailable []string `json:"principalAvailable"  validate:"required"`
	IsActive           bool     `json:"isActive"`
	IsDefault          bool     `json:"isDefault"`
	BaseURL            string   `json:"baseUrl" validate:"required"`
	Password           string   `json:"password" validate:"required"`
	Acquirer           string   `json:"acquirer"`
}

func (r *CreateMIDRequest) ToCreditCardCoreRequest() *creditcardCoreProcessorModel.CreateMIDRequest {
	return &creditcardCoreProcessorModel.CreateMIDRequest{
		Mid:                r.Mid,
		Name:               r.Name,
		Description:        r.Description,
		Type:               r.Type,
		TransactionType:    r.TransactionType,
		InstallmentType:    r.InstallmentType,
		InstallmentTenor:   r.InstallmentTenor,
		Processor:          r.Processor,
		PrincipalAvailable: r.PrincipalAvailable,
		IsActive:           r.IsActive,
		IsDefault:          r.IsDefault,
		BaseURL:            r.BaseURL,
		Password:           r.Password,
		Acquirer:           r.Acquirer,
	}
}

type UpdateMIDRequest struct {
	Mid                string   `json:"mid,omitempty"`
	Name               string   `json:"name,omitempty"`
	Description        string   `json:"description,omitempty"`
	Type               string   `json:"type,omitempty"`
	TransactionType    string   `json:"transactionType,omitempty"`
	InstallmentType    string   `json:"installmentType,omitempty"`
	InstallmentTenor   int      `json:"installmentTenor,omitempty"`
	Processor          string   `json:"processor,omitempty"`
	PrincipalAvailable []string `json:"principalAvailable,omitempty"`
	IsActive           bool     `json:"isActive"`
	IsDefault          bool     `json:"isDefault"`
	BaseURL            string   `json:"base_Ul,omitempty"`
	Password           string   `json:"password,omitempty"`
	Acquirer           string   `json:"acquirer,omitempty"`

	UUID string `json:"-"`
}

type InquiryTransactionRequest struct {
	MerchantID           string
	ClientReferenceID    string
	ProcessorReferenceID string
}

type BlockCardRequest struct {
	CardUUID    string    `json:"cardUuid" validate:"required"`
	IsBlocked   bool      `json:"isBlocked"`
	BlockedTo   time.Time `json:"blockedTo,omitempty"`
	BlockReason string    `json:"blockReason,omitempty"`
}

type MIDInfo struct {
	MID  string `json:"mid"`
	Type string `json:"type"`
}

func (r *UpdateMIDRequest) ToCreditCardCoreRequest() *creditcardCoreProcessorModel.UpdateMIDRequest {
	return &creditcardCoreProcessorModel.UpdateMIDRequest{
		Mid:                r.Mid,
		Name:               r.Name,
		Description:        r.Description,
		Type:               r.Type,
		TransactionType:    r.TransactionType,
		InstallmentType:    r.InstallmentType,
		InstallmentTenor:   r.InstallmentTenor,
		Processor:          r.Processor,
		PrincipalAvailable: r.PrincipalAvailable,
		IsActive:           r.IsActive,
		IsDefault:          r.IsDefault,
		BaseURL:            r.BaseURL,
		Password:           r.Password,
		Acquirer:           r.Acquirer,
		UUID:               r.UUID,
	}
}

type GetMIDListRequest struct {
	Page            int
	PerPage         int
	Mid             string
	Acquirer        string
	Name            string
	Type            string
	TransactionType string
	InstallmentType string
	IsDefault       *bool
	IsActive        *bool
}

type ValidateMIDInstallmentBinsRequest struct {
	MidID string
	Bins  []string
}
