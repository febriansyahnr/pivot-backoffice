package orchestrator_model

import (
	"time"

	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/backendportal/common"
	feeModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/backendportal/fee"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/backendportal/merchant"
	"github.com/shopspring/decimal"
)

type UpdatePaymentTransactionRequest struct {
	ProcessorReferenceName string
	ProcessorReferenceId   string
	ProcessorTransactionId string
	LedgerId               string
	UpdatedAt              time.Time
	TrxDatetime            *time.Time
	Status                 string
	ChargeStatus           string
	SettlementStatus       *string
	SettlementAt           *time.Time
	SettlementModel        *string
	Channel                string
	Amount                 commonModel.Amount
	TransactionTimestamp   time.Time
	FailureCode            string

	MethodDetail any
}

type PaymentMethodDetail interface {
	MetadataPaymentMethodVA | MetadataPaymentMethodQRIS | MetadataPaymentMethodCC | any
}

type MetadataPayment[T PaymentMethodDetail] struct {
	SettlementDetail       *MetadataPaymentSettlementDetail `json:"settlementDetail,omitempty"`
	ReconReferenceNo       string                           `json:"reconReferenceNo"`
	ProcessorTransactionId string                           `json:"processorTransactionId"`
	ExpiredAt              time.Time                        `json:"expiredAt"`
	MethodDetail           T                                `json:"methodDetail"`
	FeeDetail              *feeModel.FeeMetadataObject      `json:"feeDetail,omitempty"`
	FeeOnBehalf            *feeModel.TrxFeeOnBehalfMetadata `json:"feeOnBehalf,omitempty"`
	ReconDetail            *MetadataReconDetail             `json:"reconDetail,omitempty"`

	// For Unified Payment
	ChargeStatus        string `json:"chargeStatus,omitempty"`
	StatementDescriptor string `json:"statementDescriptor,omitempty"`
	FailureCode         string `json:"failureCode,omitempty"`

	// For FDS flag
	BypassExternalFDS *bool `json:"bypassExternalFDS,omitempty"`

	// For Sub Payment Summary
	SubPaymentSummary *MetadataSubPaymentSummary `json:"subPaymentSummary,omitempty"`
}

type MetadataPaymentSettlementDetail merchant.SettlementConfig

type MetadataPaymentMethodVA struct {
	AccountName    string         `json:"accountName"`
	Acquirer       string         `json:"acquirer"`
	Status         string         `json:"status,omitempty"`
	CreatedAt      time.Time      `json:"createdAt"`
	IsClosedAmount bool           `json:"isClosedAmount"`
	IsSingleUse    bool           `json:"isSingleUse"`
	AdditionalInfo map[string]any `json:"additionalInfo"`
}

type MetadataPaymentMethodQRIS struct {
	QrType             string         `json:"qrType"`
	QrMethodType       string         `json:"qrMethodType"`
	PartnerReferenceNo string         `json:"partnerReferenceNo"`
	StoreID            string         `json:"storeID"`
	MerchantID         string         `json:"merchantID"`
	MerchantName       string         `json:"merchantName,omitempty"`
	QrUrl              string         `json:"qrUrl"`
	QrContent          string         `json:"qrContent"`
	AdditionalInfo     map[string]any `json:"additionalInfo"`
}

type MetadataPaymentMethodCC struct {
	AuthenticationMethod string                                 `json:"authenticationMethod"`
	BankMerchantID       string                                 `json:"bankMerchantId,omitempty"`
	ProcessorStatus      string                                 `json:"processorStatus,omitempty"`
	CardData             *MetadataPaymentMethodCCCard           `json:"cardData,omitempty"`
	AuthorizationData    *MetadataPaymentMethodCCAuthorization  `json:"authorizationData,omitempty"`
	AuthenticationData   *MetadataPaymentMethodCCAuthentication `json:"authenticationData,omitempty"`
	AdditionalInfo       map[string]any                         `json:"additionalInfo"`
}

type MetadataPaymentMethodEwallet struct {
	ResponseCode       string `json:"responseCode"`
	ResponseMessage    string `json:"responseMessage"`
	PartnerReferenceNo string `json:"partnerReferenceNo"`
	ReferenceNo        string `json:"referenceNo"`
	WebRedirectURL     string `json:"webRedirectUrl"`
	AppRedirectURL     string `json:"appRedirectUrl"`
}

type MetadataPaymentMethodCCCard struct {
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
}

type MetadataPaymentMethodCCAuthorization struct {
	AuthorizationResult   string `json:"authorizationResult"`
	OrderID               string `json:"orderId"`
	TransactionStaus      string `json:"transactionStatus"`
	AuthorizationID       string `json:"authorizationId"`
	ApprovalCode          string `json:"approvalCode"`
	BankMerchantID        string `json:"bankMerchantId"`
	AcquirerTransactionID string `json:"acquirerTransactionId"`
	TransactionReference  string `json:"transactionReference"`
	CvvResult             string `json:"cvvResult"`
	AcquirerResponseCode  string `json:"acquirerResponseCode"`
	Stan                  string `json:"stan"`
	AvsResult             string `json:"avsResult"`
}

type MetadataPaymentMethodCCAuthentication struct {
	AuthenticationResult string `json:"authenticationResult"`
	AuthenticationID     string `json:"authenticationId"`
	PaRes                string `json:"paRes"`
	VeRes                string `json:"veRes"`
	XID                  string `json:"xid"`
	CAVV                 string `json:"cavv"`
	EciCode              string `json:"eciCode"`
	ThreeDsVer           string `json:"3dsVer"`
	ChallengeCode        string `json:"challengeCode"`
}

type MetadataRefund struct {
	PaymentSessionID string `json:"paymentSessionId"`
	PaymentChargeID  string `json:"paymentChargeId"`
}

type MetadataRefundOfPaymentFee struct {
	PaymentSessionID   string                      `json:"paymentSessionId"`
	PaymentChargeID    string                      `json:"paymentChargeId"`
	PaymentFeeLedgerID string                      `json:"paymentFeeLedgerId"`
	FeeDetail          *feeModel.FeeMetadataObject `json:"feeDetail,omitempty"`
}

type MetadataReconDetail struct {
	Status   string `json:"status"`
	DateTime string `json:"datetime"`
}

type MetadataSubPaymentSummary struct {
	TotalCreditAmount decimal.Decimal `json:"totalCreditAmount"`
	TotalFeeAmount    decimal.Decimal `json:"totalFeeAmount"`
}
