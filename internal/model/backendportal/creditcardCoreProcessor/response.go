package creditcardCoreProcessorModel

import (
	"github.com/google/uuid"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/backendportal/common"
	"github.com/shopspring/decimal"
)

type ResponseType interface {
	AuthenticationResponse
}

type Response[T ResponseType] struct {
	Code    string `json:"code"`
	Message string `json:"message,omitempty"`
	Error   any    `json:"error,omitempty"`
	Data    *T     `json:"data,omitempty"`
}

type VoidResponseData struct {
	Status                string          `json:"status"`
	AcquirerTransactionID string          `json:"acquirer_transaction_id,omitempty"`
	GrandTotalAmount      decimal.Decimal `json:"grand_total_amount"`
	Currency              string          `json:"currency,omitempty"`
	CardBrand             string          `json:"card_brand"`
	CreatedAt             string          `json:"created_at"`
}

type VoidResponse struct {
	Data    VoidResponseData `json:"data"`
	Code    string           `json:"code"`
	Message string           `json:"message,omitempty"`
	Error   interface{}      `json:"error,omitempty"`
}

type GetTransactionList struct {
	Data    *GetTransactionDataList `json:"data"`
	Code    string                  `json:"code"`
	Message string                  `json:"message,omitempty"`
	Error   interface{}             `json:"error,omitempty"`
}

type GenericApiResponse struct {
	Data    interface{} `json:"data"`
	Code    string      `json:"code"`
	Message string      `json:"message,omitempty"`
	Error   interface{} `json:"error,omitempty"`
}

type GetTransactionDataList struct {
	Result     []*GetTransactionListResult `json:"results"`
	Pagination PaginationResponse          `json:"pagination"`
}

type PaginationResponse struct {
	PageLimit   int `json:"page_limit"`
	PageNumber  int `json:"page_number"`
	TotalRecord int `json:"total_record"`
	TotalPage   int `json:"total_page"`
}

type GetTransactionListResult struct {
	TransactionDate       string              `json:"transaction_date"`
	PaymentUUID           uuid.UUID           `json:"payment_uuid"`
	ClientTransactionID   string              `json:"client_transaction_id"`
	AcquirerTransactionID []string            `json:"acquirer_transaction_id"`
	PayerTransactionID    string              `json:"payer_transaction_id"`
	VoidTransactionID     string              `json:"void_transaction_id"`
	AuthorizationData     *AuthorizationData  `json:"authorization_data"`
	AuthenticationData    *AuthenticationData `json:"authentication_data"`
	CardData              *CardData           `json:"card_data"`
	IssuingBank           string              `json:"issuing_bank"`
	ChargeStatus          string              `json:"charge_status"`
	ChargeAt              string              `json:"charge_at"`
	VoidStatus            string              `json:"void_status"`
	VoidAt                string              `json:"void_at"`
	RefundDetail          *RefundDetail       `json:"refundDetail"`
	TransactionType       []string            `json:"transaction_type"`
	FDS                   string              `json:"fds"`
	Amount                commonModel.Amount  `json:"amount"`
}

type RefundDetail struct {
	RefundStatus string `json:"refundStatus"`
	RefundAt     string `json:"refundAt"`
	ARN          string `json:"arn"`
}

type AuthorizationData struct {
	AuthorizationResult   string             `json:"authorization_result"`
	OrderID               string             `json:"order_id"`
	TransactionStatus     string             `json:"transaction_status"`
	AuthorizationID       string             `json:"authorization_id"`
	ApprovalCode          string             `json:"approval_code"`
	BankMerchantID        string             `json:"bank_merchant_id"`
	AcquirerTransactionID string             `json:"acquirer_transaction_id"`
	TransactionReference  string             `json:"transaction_reference"`
	CvvResult             string             `json:"cvv_result"`
	AcquirerResponseCode  string             `json:"acquirer_response_code"`
	Stan                  string             `json:"stan"`
	AvsResult             string             `json:"avs_result"`
	Amount                commonModel.Amount `json:"amount"`
	ErrorMessage          string             `json:"error_message"`
}

type AuthenticationData struct {
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

type CardData struct {
	First8Digit    string `json:"first_8_digit,omitempty"`
	Last4Digit     string `json:"last_4_digit,omitempty"`
	CardType       string `json:"card_type,omitempty"`
	CardBrand      string `json:"card_brand,omitempty"`
	CardIssuing    string `json:"card_issuing,omitempty"`
	CountryCode    string `json:"country_code,omitempty"`
	IssuingCountry string `json:"issuing_country,omitempty"`
	Fingerprint    string `json:"fingerprint,omitempty"`
}

type RefundResponseData struct {
	Status                string          `json:"status"`
	AcquirerTransactionID string          `json:"acquirer_transaction_id,omitempty"`
	GrandTotalAmount      decimal.Decimal `json:"grand_total_amount"`
	Currency              string          `json:"currency,omitempty"`
	CardBrand             string          `json:"card_brand"`
	CreatedAt             string          `json:"created_at"`
}

type RefundResponse struct {
	Data    RefundResponseData `json:"data"`
	Code    string             `json:"code"`
	Message string             `json:"message,omitempty"`
	Error   interface{}        `json:"error,omitempty"`
}

type MIDResponseData struct {
	Uuid               uuid.UUID `json:"uuid"`
	Mid                string    `json:"mid"`
	Name               string    `json:"name"`
	Description        string    `json:"description"`
	Type               string    `json:"type"`
	TransactionType    string    `json:"transaction_type"`
	InstallmentType    string    `json:"installment_type"`
	InstallmentTenor   int       `json:"installment_tenor"`
	Processor          string    `json:"processor"`
	PrincipalAvailable []string  `json:"principal_available"`
	IsActive           bool      `json:"is_active"`
	IsDefault          bool      `json:"is_default"`
	BaseURL            string    `json:"base_url"`
	Acquirer           string    `json:"acquirer"`
	CreatedAt          string    `json:"created_at"`
	UpdatedAt          string    `json:"updated_at"`
}

type MIDResponse struct {
	Data    MIDResponseData `json:"data"`
	Code    string          `json:"code"`
	Message string          `json:"message,omitempty"`
	Error   interface{}     `json:"error,omitempty"`
}

type CreateMIDResponseData struct {
	Uuid    uuid.UUID `json:"uuid"`
	Created bool      `json:"created"`
}

type CreateMIDResponse struct {
	Data    CreateMIDResponseData `json:"data"`
	Code    string                `json:"code"`
	Message string                `json:"message,omitempty"`
	Error   interface{}           `json:"error,omitempty"`
}

type UpdateMIDResponseData struct {
	Uuid    uuid.UUID `json:"uuid"`
	Updated bool      `json:"updated"`
}

type UpdateMIDResponse struct {
	Data    UpdateMIDResponseData `json:"data"`
	Code    string                `json:"code"`
	Message string                `json:"message,omitempty"`
	Error   interface{}           `json:"error,omitempty"`
}

type CreateMIDMapResponseData struct {
	Uuid    uuid.UUID `json:"uuid"`
	Created bool      `json:"created"`
}

type CreateMIDMapResponse struct {
	Data    CreateMIDMapResponseData `json:"data"`
	Code    string                   `json:"code"`
	Message string                   `json:"message,omitempty"`
	Error   interface{}              `json:"error,omitempty"`
}

type ValidateMIDInstallmentBinsResponse struct {
	Code    string      `json:"code"`
	Message string      `json:"message,omitempty"`
	Error   interface{} `json:"error,omitempty"`
}

type MIDListResponse struct {
	Data    MIDListResponseData `json:"data"`
	Code    string              `json:"code"`
	Message string              `json:"message,omitempty"`
	Error   interface{}         `json:"error,omitempty"`
}

type MIDListResponseData struct {
	Results    []MIDResponseData  `json:"results"`
	Pagination PaginationResponse `json:"pagination"`
}

type MIDMapListResponse struct {
	Data    MIDMapListResponseData `json:"data"`
	Code    string                 `json:"code"`
	Message string                 `json:"message,omitempty"`
	Error   interface{}            `json:"error,omitempty"`
}

type MIDMapListResponseData struct {
	Results    []MIDMapResponseData `json:"results"`
	Pagination PaginationResponse   `json:"pagination"`
}

type MIDMapResponse struct {
	Data    MIDMapResponseData `json:"data"`
	Code    string             `json:"code"`
	Message string             `json:"message,omitempty"`
	Error   interface{}        `json:"error,omitempty"`
}

type MIDMapResponseData struct {
	Uuid         uuid.UUID       `json:"uuid"`
	MerchantId   string          `json:"merchant_id"`
	MerchantName string          `json:"merchant_name"`
	IsActive     bool            `json:"is_active"`
	Priority     int             `json:"priority"`
	MidDetail    MIDResponseData `json:"mid_detail"`
	CreatedAt    string          `json:"created_at"`
}

type UpdateMIDMapResponse struct {
	Data    UpdateMIDMapResponseData `json:"data"`
	Code    string                   `json:"code"`
	Message string                   `json:"message,omitempty"`
	Error   interface{}              `json:"error,omitempty"`
}

type UpdateMIDMapResponseData struct {
	Uuid    uuid.UUID `json:"uuid"`
	Updated bool      `json:"updated"`
}

type CaptureResponseData struct {
	ID                     string  `json:"id"`
	Status                 string  `json:"status"`
	AcquirerTransactionID  string  `json:"acquirer_transaction_id"`
	ReleaseRemainingAmount bool    `json:"release_remaining_amount"`
	Currency               string  `json:"currency"`
	Amount                 float64 `json:"amount"`
	CreatedAt              string  `json:"created_at"`
}

type CaptureResponse struct {
	Data    CaptureResponseData `json:"data"`
	Code    string              `json:"code"`
	Message string              `json:"message,omitempty"`
	Error   interface{}         `json:"error,omitempty"`
}

type GenericApiResponseGen[T comparable] struct {
	Data    T      `json:"data"`
	Code    string `json:"code"`
	Message string `json:"message,omitempty"`
	Error   any    `json:"error,omitempty"`
}

type GetBinDetailResponse struct {
	UUID          uuid.UUID `json:"uuid"`
	BinNumber     string    `json:"bin_number"`
	CardType      string    `json:"card_type"`
	CardBrand     string    `json:"card_brand"`
	ConsumerType  string    `json:"consumer_type"`
	CardLevel     string    `json:"card_level"`
	IssuerName    string    `json:"issuer_name"`
	IssuerCountry string    `json:"issuer_country"`
	Currency      string    `json:"currency"`
	Status        string    `json:"status"`
	IsBlocked     bool      `json:"is_blocked"`
	CreatedAt     string    `json:"created_at"`
	UpdatedAt     string    `json:"updated_at"`
}

type AuthenticationResponse struct {
	Status                string                         `json:"status"`
	Message               string                         `json:"message,omitempty"`
	SessionID             string                         `json:"session_id,omitempty"`
	AcquirerTransactionID string                         `json:"acquirer_transaction_id,omitempty"`
	Amount                decimal.Decimal                `json:"amount"`
	Currency              string                         `json:"currency,omitempty"`
	AuthenticationURL     *AuthenticationResponseThreeds `json:"authentication_url,omitempty"`
	AuthenticationData    *AuthenticationData            `json:"authentication_data,omitempty"`
}

type AuthenticationResponseThreeds struct {
	ActionUrl   string `json:"action_url,omitempty"`
	CReq        string `json:"creq,omitempty"`
	Method      string `json:"method,omitempty"`
	URL         string `json:"url,omitempty"`
	AccessToken string `json:"access_token,omitempty"`
	HTML        string `json:"html,omitempty"`
	Version     string `json:"version,omitempty"`
	CreatedAt   string `json:"created_at,omitempty"`
}
