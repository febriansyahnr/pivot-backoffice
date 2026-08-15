package card

import (
	"time"

	"github.com/google/uuid"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	creditcardCoreProcessorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/creditcardCoreProcessor"
	"github.com/shopspring/decimal"
)

type CreateCardPaymentResponse struct {
	UUID                  string          `json:"uuid"`
	AcquirerTransactionID string          `json:"acquirerTransactionId,omitempty"`
	MerchantID            uuid.UUID       `json:"merchantId"`
	BankMerchantID        string          `json:"bankMerchantId,omitempty"`
	ReferenceID           string          `json:"referenceId"`
	Amount                decimal.Decimal `json:"amount"`
	Currency              string          `json:"currency"`
	Created               string          `json:"created"`
	Expired               string          `json:"expired"`
	PaymentURL            string          `json:"paymentUrl"`
	Status                string          `json:"status"`
}

type OpenAPIGetCardPaymentByIdResponse struct {
	UUID           string                        `json:"uuid"`
	ChargeID       string                        `json:"chargeId"` // means account_transactions.uuid or ledger uuid
	MerchantID     string                        `json:"merchantId"`
	BankMerchantID string                        `json:"bankMerchantId"`
	ReferenceID    string                        `json:"referenceId"`
	PaymentStatus  string                        `json:"paymentStatus"`
	Amount         decimal.Decimal               `json:"amount"`
	Currency       string                        `json:"currency"`
	PaymentURL     string                        `json:"paymentUrl"`
	CardData       SendCallbackCardDataRequest   `json:"cardData"`
	RedirectUrl    *CreditcardRedirectUrlRequest `json:"redirectUrl,omitempty"`
	Created        string                        `json:"created"`
	Updated        string                        `json:"updated"`
}

type CustomerInfo struct {
	UUID         string `json:"uuid"`
	MerchantUUID string `json:"merchantId"`
	FirstName    string `json:"firstName"`
	LastName     string `json:"lastName"`
	Email        string `json:"email"`
	Phone        string `json:"phone"`
}

type InternalGetCardPaymentByIdResponse struct {
	UUID           string              `json:"uuid"`
	ChargeID       string              `json:"chargeId"` // means account_transactions.uuid or ledger uuid
	MerchantID     string              `json:"merchantId"`
	Customer       *CustomerInfo       `json:"customer,omitempty"`
	BankMerchantID string              `json:"bankMerchantId"`
	ReferenceID    string              `json:"referenceId"`
	RecurringID    *string             `json:"recurringId,omitempty"`
	PaymentType    string              `json:"paymentType,omitempty"`
	PaymentStatus  string              `json:"paymentStatus"`
	Amount         decimal.Decimal     `json:"amount"`
	Fee            *decimal.Decimal    `json:"fee,omitempty"`
	Discount       *decimal.Decimal    `json:"discount,omitempty"`
	TotalAmount    decimal.Decimal     `json:"totalAmount"`
	Currency       string              `json:"currency"`
	PaymentURL     string              `json:"paymentUrl"`
	Metadata       *CreditcardMetadata `json:"metadata"`
	ExpirationMode string              `json:"expirationMode,omitempty"`
	ExpiredAt      time.Time           `json:"expiredAt"`
	Created        string              `json:"created"`
	Updated        string              `json:"updated"`
}

type VoidResponse struct {
	Status                string          `json:"status"`
	AcquirerTransactionID string          `json:"acquirer_transaction_id,omitempty"`
	GrandTotalAmount      decimal.Decimal `json:"grand_total_amount"`
	Currency              string          `json:"currency,omitempty"`
	CardBrand             string          `json:"card_brand"`
	CreatedAt             string          `json:"created_at"`
}

type GetTransactionListResult struct {
	TransactionDate       string              `json:"transactionDate"`
	PaymentUUID           uuid.UUID           `json:"paymentUuid"`
	ClientTransactionID   string              `json:"clientTransactionId"`
	AcquirerTransactionID []string            `json:"acquirerTransactionId"`
	PayerTransactionID    string              `json:"payerTransactionId"`
	VoidTransactionID     string              `json:"voidTransactionId"`
	AuthorizationData     *AuthorizationData  `json:"authorizationData"`
	AuthenticationData    *AuthenticationData `json:"authenticationData"`
	CardData              *CardData           `json:"cardData"`
	IssuingBank           string              `json:"issuingBank"`
	ChargeStatus          string              `json:"chargeStatus"`
	ChargeAt              string              `json:"chargeAt"`
	VoidStatus            string              `json:"voidStatus"`
	VoidAt                string              `json:"voidAt"`
	RefundDetail          *RefundDetail       `json:"refundDetail"`
	TransactionType       []string            `json:"transactionType"`
	FDS                   string              `json:"fds"`
	Amount                commonModel.Amount  `json:"amount"`
}

type RefundDetail struct {
	RefundStatus string `json:"refundStatus"`
	RefundAt     string `json:"refundAt"`
	ARN          string `json:"arn"`
}

type AuthorizationData struct {
	AuthorizationResult   string             `json:"authorizationResult"`
	OrderID               string             `json:"orderId"`
	TransactionStatus     string             `json:"transactionStatus"`
	AuthorizationID       string             `json:"authorizationId"`
	ApprovalCode          string             `json:"approvalCode"`
	BankMerchantID        string             `json:"bankMerchantId"`
	AcquirerTransactionID string             `json:"acquirerTransactionId"`
	TransactionReference  string             `json:"transactionReference"`
	CvvResult             string             `json:"cvvResult"`
	AcquirerResponseCode  string             `json:"acquirerResponseCode"`
	Stan                  string             `json:"stan"`
	AvsResult             string             `json:"avsResult"`
	ErrorMessage          string             `json:"errorMessage"`
	Amount                commonModel.Amount `json:"amount"`
}

type AuthenticationData struct {
	AuthenticationResult string `json:"authenticationResult"`
	AuthenticationID     string `json:"authenticationId"`
	PaRes                string `json:"paRes"`
	VeRes                string `json:"veRes"`
	XID                  string `json:"xid"`
	CAVV                 string `json:"cavv"`
	EciCode              string `json:"eciCode"`
	ThreeDsVer           string `json:"threeDsVer"`
	ChallengeCode        string `json:"challengeCode"`
	MID                  string `json:"mid"`
}

type CardData struct {
	First8Digit    string `json:"first8Digit,omitempty"`
	Last4Digit     string `json:"last4Digit,omitempty"`
	CardType       string `json:"cardType,omitempty"`
	CardBrand      string `json:"cardBrand,omitempty"`
	CardIssuing    string `json:"cardIssuing,omitempty"`
	CountryCode    string `json:"countryCode,omitempty"`
	IssuingCountry string `json:"issuingCountry,omitempty"`
	Fingerprint    string `json:"fingerprint,omitempty"`
}

type BlockCardResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

func ToCreditcardCoreGetTransactionListResponse(response *creditcardCoreProcessorModel.GetTransactionDataList) *commonModel.PaginationResponse {
	results := []*GetTransactionListResult{}

	for _, r := range response.Result {

		var (
			authorizationData  *AuthorizationData
			authenticationData *AuthenticationData
			cardData           *CardData
		)

		if r.AuthorizationData != nil {
			authorizationData = &AuthorizationData{
				AuthorizationResult:   r.AuthorizationData.AuthorizationResult,
				OrderID:               r.AuthorizationData.OrderID,
				TransactionStatus:     r.AuthorizationData.TransactionStatus,
				AuthorizationID:       r.AuthorizationData.AuthorizationID,
				ApprovalCode:          r.AuthorizationData.ApprovalCode,
				BankMerchantID:        r.AuthorizationData.BankMerchantID,
				AcquirerTransactionID: r.AuthorizationData.AcquirerTransactionID,
				TransactionReference:  r.AuthorizationData.TransactionReference,
				CvvResult:             r.AuthorizationData.CvvResult,
				AcquirerResponseCode:  r.AuthorizationData.AcquirerResponseCode,
				Stan:                  r.AuthorizationData.Stan,
				AvsResult:             r.AuthorizationData.AvsResult,
				ErrorMessage:          r.AuthorizationData.ErrorMessage,
				Amount:                r.AuthorizationData.Amount,
			}
		}

		if r.AuthenticationData != nil {
			mid := ""
			// Get MID from AuthorizationData if available
			if r.AuthorizationData != nil {
				mid = r.AuthorizationData.BankMerchantID
			}
			authenticationData = &AuthenticationData{
				AuthenticationResult: r.AuthenticationData.AuthenticationResult,
				AuthenticationID:     r.AuthenticationData.AuthenticationID,
				PaRes:                r.AuthenticationData.PaRes,
				VeRes:                r.AuthenticationData.VeRes,
				XID:                  r.AuthenticationData.XID,
				CAVV:                 r.AuthenticationData.CAVV,
				EciCode:              r.AuthenticationData.EciCode,
				ThreeDsVer:           r.AuthenticationData.ThreeDsVer,
				ChallengeCode:        r.AuthenticationData.ChallengeCode,
				MID:                  mid,
			}
		}

		if r.CardData != nil {
			cardData = &CardData{
				First8Digit:    r.CardData.First8Digit,
				Last4Digit:     r.CardData.Last4Digit,
				CardType:       r.CardData.CardType,
				CardBrand:      r.CardData.CardBrand,
				CardIssuing:    r.CardData.CardIssuing,
				CountryCode:    r.CardData.CountryCode,
				IssuingCountry: r.CardData.IssuingCountry,
				Fingerprint:    r.CardData.Fingerprint,
			}
		}

		// Map RefundDetail from creditcard-core-processor
		var refundDetail *RefundDetail
		if r.RefundDetail != nil {
			refundDetail = &RefundDetail{
				RefundStatus: r.RefundDetail.RefundStatus,
				RefundAt:     r.RefundDetail.RefundAt,
				ARN:          r.RefundDetail.ARN,
			}
		}

		results = append(results, &GetTransactionListResult{
			TransactionDate:       r.TransactionDate,
			PaymentUUID:           r.PaymentUUID,
			ClientTransactionID:   r.ClientTransactionID,
			AcquirerTransactionID: r.AcquirerTransactionID,
			PayerTransactionID:    r.PayerTransactionID,
			VoidTransactionID:     r.VoidTransactionID,
			AuthorizationData:     authorizationData,
			AuthenticationData:    authenticationData,
			CardData:              cardData,
			IssuingBank:           r.IssuingBank,
			ChargeStatus:          r.ChargeStatus,
			ChargeAt:              r.ChargeAt,
			VoidStatus:            r.VoidStatus,
			VoidAt:                r.VoidAt,
			RefundDetail:          refundDetail,
			TransactionType:       r.TransactionType,
			FDS:                   r.FDS,
			Amount:                r.Amount,
		})
	}

	return &commonModel.PaginationResponse{
		Meta: commonModel.Meta{
			Page:       int64(response.Pagination.PageNumber),
			PerPage:    int64(response.Pagination.PageLimit),
			TotalItems: int64(response.Pagination.TotalRecord),
			TotalPages: int64(response.Pagination.TotalPage),
		},
		Data: results,
	}
}
