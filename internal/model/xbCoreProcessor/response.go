package xbCoreProcessorModel

import (
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx/types"
	"github.com/shopspring/decimal"
)

type GetFxRateResponse struct {
	Data    GetFxRateResponseData `json:"data"`
	Code    string                `json:"code"`
	Message string                `json:"message,omitempty"`
	Error   interface{}           `json:"error,omitempty"`
}

type GetFxRateResponseData struct {
	SourceCurrency      string          `json:"source_currency"`
	DestinationCurrency string          `json:"destination_currency"`
	OriginalFxRate      decimal.Decimal `json:"original_fx_Rate"`
	MarkupFxRate        decimal.Decimal `json:"markup_fx_Rate"`
	DestinationFxRate   decimal.Decimal `json:"destination_fx_rate"`
	SpreadType          string          `json:"spread_type"`
	SpreadValue         decimal.Decimal `json:"spread_value"`
	ExpiryAt            time.Time       `json:"expiry_at"`
}

type CreatePayoutSessionResponse struct {
	Data    CreatePayoutSessionResponseData `json:"data"`
	Code    string                          `json:"code"`
	Message string                          `json:"message,omitempty"`
	Error   interface{}                     `json:"error,omitempty"`
}

type CreatePayoutSessionResponseData struct {
	Uuid                  string                  `json:"uuid"`
	AcquirerTransactionId string                  `json:"acquirer_transaction_id"`
	MerchantId            string                  `json:"merchant_id"`
	SourceCurrency        string                  `json:"source_currency"`
	DestinationCurrency   string                  `json:"destination_currency"`
	DestinationAmount     decimal.Decimal         `json:"destination_amount"`
	FxRate                decimal.Decimal         `json:"fx_rate"`
	DestinationFxRate     decimal.Decimal         `json:"destination_fx_rate"`
	SpreadValue           decimal.Decimal         `json:"spread_value"`
	SpreadType            string                  `json:"spread_type"`
	TotalAmount           decimal.Decimal         `json:"total_amount"`
	CreatedAt             time.Time               `json:"created_at"`
	ExpiredAt             time.Time               `json:"expired_at"`
	Status                string                  `json:"status"`
	SenderId              string                  `json:"remitter_uuid"`
	BeneficiaryId         string                  `json:"beneficiary_uuid"`
	BeneficiaryData       BeneficiaryDataResponse `json:"beneficiary_data"`
	SenderData            SenderDataResponse      `json:"remitter_data"` // sender_data = remitter_data
	RoutingCode           string                  `json:"routing_code"`
	RoutingValue          string                  `json:"routing_value"`
}

type BeneficiaryDataResponse struct {
	Name               string `json:"name"`
	Address            string `json:"address"`
	City               string `json:"city"`
	Postcode           string `json:"postcode"`
	State              string `json:"state"`
	CountryCode        string `json:"country_code"`
	CountryName        string `json:"country_name,omitempty"`
	AccountType        string `json:"account_type"`
	AccountNumber      string `json:"account_number"`
	BankName           string `json:"bank_name"`
	BankCode           string `json:"bank_code"`
	ContactCountryCode string `json:"contact_country_code"`
	ContactNumber      string `json:"contact_number"`
	Email              string `json:"email"`
}

type SenderDataResponse struct {
	Name                 string `json:"name"`
	CountryCode          string `json:"country_code"`
	CountryName          string `json:"country_name,omitempty"`
	State                string `json:"state"`
	City                 string `json:"city"`
	Address              string `json:"address"`
	Postcode             string `json:"postcode"`
	AccountType          string `json:"account_type"`
	IdentificationType   string `json:"identification_type"`
	IdentificationNumber string `json:"identification_number"`
	BankAccountNumber    string `json:"bank_account_number"`
	Dob                  string `json:"dob"`
	ContactCountryCode   string `json:"contact_country_code"`
	ContactNumber        string `json:"contact_number"`
	SourceOfIncome       string `json:"source_of_income"`
}

type ConfirmPayoutResponse struct {
	Data    ConfirmPayoutResponseData `json:"data"`
	Code    string                    `json:"code"`
	Message string                    `json:"message,omitempty"`
	Error   interface{}               `json:"error,omitempty"`
}

type ConfirmPayoutResponseData struct {
	Uuid                  string                  `json:"uuid"`
	AcquirerTransactionId string                  `json:"acquirer_transaction_id"`
	PartnerTransactionId  string                  `json:"partner_transaction_id"`
	MerchantId            string                  `json:"merchant_id"`
	SourceCurrency        string                  `json:"source_currency"`
	DestinationCurrency   string                  `json:"destination_currency"`
	DestinationAmount     decimal.Decimal         `json:"destination_amount"`
	FxRate                decimal.Decimal         `json:"fx_rate"`
	DestinationFxRate     decimal.Decimal         `json:"destination_fx_rate"`
	SpreadValue           decimal.Decimal         `json:"spread_value"`
	SpreadType            string                  `json:"spread_type"`
	TotalAmount           decimal.Decimal         `json:"total_amount"`
	StatementNarrative    string                  `json:"statement_narrative"`
	CreatedAt             time.Time               `json:"created_at"`
	UpdatedAt             time.Time               `json:"updated_at"`
	Status                string                  `json:"status"`
	SenderId              string                  `json:"remitter_uuid"`
	BeneficiaryId         string                  `json:"beneficiary_uuid"`
	BeneficiaryData       BeneficiaryDataResponse `json:"beneficiary_data"`
	SenderData            SenderDataResponse      `json:"remitter_data"` // sender_data = remitter_data
}

type UploadUnderlyingDocumentResponse struct {
	Data    UploadUnderlyingDocumentResponseData `json:"data"`
	Code    string                               `json:"code"`
	Message string                               `json:"message,omitempty"`
	Error   interface{}                          `json:"error,omitempty"`
}

type UploadUnderlyingDocumentResponseData struct {
	DocumentReference string `json:"document_reference"`
}

type SubmitRfiDetailsResponse struct {
	Data    SubmitRfiDetailsResponseData `json:"data"`
	Code    string                       `json:"code"`
	Message string                       `json:"message,omitempty"`
	Error   interface{}                  `json:"error,omitempty"`
}

type SubmitRfiDetailsResponseData struct {
	GetRfiDetailsResponseData
}

type CommonResponse struct {
	Code         string            `json:"code"`
	Message      string            `json:"message,omitempty"`
	Error        any               `json:"error,omitempty"`
	ErrorType    string            `json:"error_type,omitempty"`
	ErrorDetails []*ApiErrorDetail `json:"error_details,omitempty"`
}

type CreateBeneficiaryResponse struct {
	CommonResponse
	Data CreateBeneficiaryData `json:"data"`
}

type ApiErrorDetail struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type CreateBeneficiaryData struct {
	UUID                 uuid.UUID  `json:"uuid"`
	MerchantID           string     `json:"merchant_id"`
	Name                 string     `json:"name"`
	AccountType          string     `json:"account_type"`
	Address              string     `json:"address"`
	City                 string     `json:"city"`
	Postcode             string     `json:"postcode"`
	State                string     `json:"state"`
	CountryCode          string     `json:"country_code"`
	IdentificationType   string     `json:"identification_type"`
	IdentificationNumber string     `json:"identification_number"`
	AccountNumber        string     `json:"account_number"`
	BankName             string     `json:"bank_name"`
	BankCode             string     `json:"bank_code"`
	ContactCountryCode   string     `json:"contact_country_code"`
	ContactNumber        string     `json:"contact_number"`
	Email                string     `json:"email"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            *time.Time `json:"updated_at,omitempty"`
	DeactivatedAt        *time.Time `json:"deactivated_at,omitempty"`
}

type GetPayoutResponse struct {
	Data GetPayoutResponseData `json:"data"`
	CommonResponse
}

type GetRfiDetailsResponse struct {
	Data []*GetRfiDetailsResponseData `json:"data"`
	CommonResponse
}

type GetPayoutResponseData struct {
	UUID                  uuid.UUID               `json:"uuid"`
	AcquirerTransactionID string                  `json:"acquirer_transaction_id"`
	PartnerRequestID      string                  `json:"partner_request_id"`
	PartnerTransactionID  string                  `json:"partner_transaction_id"`
	MerchantID            uuid.UUID               `json:"merchant_id"`
	SourceCurrency        string                  `json:"source_currency"`
	DestinationCurrency   string                  `json:"destination_currency"`
	DestinationAmount     decimal.Decimal         `json:"destination_amount"`
	FxRate                decimal.Decimal         `json:"fx_rate"`
	DestinationFXRate     decimal.Decimal         `json:"destination_fx_rate"`
	SpreadValue           decimal.Decimal         `json:"spread_value"`
	SpreadType            string                  `json:"spread_type"`
	TotalAmount           decimal.Decimal         `json:"total_amount"`
	StatementNarrative    string                  `json:"statement_narrative"`
	BeneficiaryData       BeneficiaryDataResponse `json:"beneficiary_data"`
	SenderData            SenderDataResponse      `json:"remitter_data"` // sender_data = remitter_data
	Status                string                  `json:"status"`
	StatusDescription     string                  `json:"status_description"`
	RfiDetails            []*RfiDetails           `json:"rfi_details,omitempty"`
	CreatedAt             time.Time               `json:"created_at"`
	UpdatedAt             time.Time               `json:"updated_at"`
	RoutingCode           string                  `json:"routing_code"`
	RoutingValue          string                  `json:"routing_value"`
}

type RfiDetails struct {
	UUID                    uuid.UUID      `json:"uuid" db:"uuid"`                                                       // autogenerated
	PayoutID                uuid.UUID      `json:"payout_id" db:"payout_id"`                                             // uuid in payout_transactions -> should get partner_transaction_id -> asscociated with payment_id in nium
	PartnerDocumentID       string         `json:"partner_document_id,omitempty" db:"partner_document_id"`               // -> rfi_id
	PartnerDocumentEntityID string         `json:"partner_document_entity_id,omitempty" db:"partner_document_entity_id"` // -> rfi_entity_id
	Actor                   string         `json:"actor,omitempty" db:"actor"`                                           // BENEFICIARY, REMITTER
	Entity                  string         `json:"entity,omitempty" db:"entity"`                                         // rfi_entity
	DocumentType            string         `json:"document_type,omitempty" db:"document_type"`                           // rfi_type
	DocumentURL             string         `json:"document_url,omitempty" db:"document_url"`                             // public url
	Filename                string         `json:"filename" db:"filename"`                                               // filename
	Location                types.JSONText `json:"location,omitempty" db:"location"`                                     // to generate signed url
	Value                   string         `json:"value,omitempty" db:"value"`                                           // rfi_value
	Comment                 string         `json:"comment,omitempty" db:"comment"`                                       // comment
	Status                  string         `json:"status,omitempty" db:"status"`                                         // pending, received
	RequestedAt             *time.Time     `json:"requested_at,omitempty" db:"requested_at"`                             // rfi_requested_datetime
	CreatedAt               time.Time      `json:"created_at,omitempty" db:"created_at"`
	UpdatedAt               time.Time      `json:"updated_at,omitempty" db:"updated_at"`
	DeletedAt               *time.Time     `json:"deleted_at,omitempty" db:"deleted_at"`
}

type GetRfiDetailsResponseData struct {
	UUID                    uuid.UUID      `json:"uuid" db:"uuid"`                                                       // autogenerated
	PayoutID                uuid.UUID      `json:"payout_id" db:"payout_id"`                                             // uuid in payout_transactions -> should get partner_transaction_id -> asscociated with payment_id in nium
	PartnerDocumentID       string         `json:"partner_document_id,omitempty" db:"partner_document_id"`               // -> rfi_id
	PartnerDocumentEntityID string         `json:"partner_document_entity_id,omitempty" db:"partner_document_entity_id"` // -> rfi_entity_id
	Actor                   string         `json:"actor,omitempty" db:"actor"`                                           // BENEFICIARY, REMITTER
	Entity                  string         `json:"entity,omitempty" db:"entity"`                                         // rfi_entity
	DocumentType            string         `json:"document_type,omitempty" db:"document_type"`                           // rfi_type
	DocumentURL             string         `json:"document_url,omitempty" db:"document_url"`                             // public url
	Filename                string         `json:"filename" db:"filename"`                                               // filename
	Location                types.JSONText `json:"location,omitempty" db:"location"`                                     // to generate signed url
	Value                   string         `json:"value,omitempty" db:"value"`                                           // rfi_value
	Comment                 string         `json:"comment,omitempty" db:"comment"`                                       // comment
	Status                  string         `json:"status,omitempty" db:"status"`                                         // pending, received
	RequestedAt             *time.Time     `json:"requested_at,omitempty" db:"requested_at"`                             // rfi_requested_datetime
	CreatedAt               time.Time      `json:"created_at,omitempty" db:"created_at"`
	UpdatedAt               time.Time      `json:"updated_at,omitempty" db:"updated_at"`
	DeletedAt               *time.Time     `json:"deleted_at,omitempty" db:"deleted_at"`
}

type CreateSenderResponse struct {
	Data CreateSenderData `json:"data"`
	CommonResponse
}

type CreateSenderData struct {
	UUID                 uuid.UUID  `json:"uuid"`
	Name                 string     `json:"name"`
	AccountType          string     `json:"account_type"`
	Address              string     `json:"address"`
	City                 string     `json:"city"`
	Postcode             string     `json:"postcode"`
	State                string     `json:"state"`
	CountryCode          string     `json:"country_code"`
	IdentificationType   string     `json:"identification_type"`
	IdentificationNumber string     `json:"identification_number"`
	SourceOfFunds        string     `json:"source_of_funds"`
	BankAccountNumber    string     `json:"bank_account_number"`
	Nationality          string     `json:"nationality"`
	Dob                  string     `json:"dob"`
	ContactCountryCode   string     `json:"contact_country_code"`
	ContactNumber        string     `json:"contact_number"`
	SourceOfIncome       string     `json:"source_of_income"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            *time.Time `json:"updated_at,omitempty"`
	DeactivatedAt        *time.Time `json:"deactivated_at,omitempty"`
}

type PaginationData struct {
	Results    interface{} `json:"results"`
	Pagination Pagination  `json:"pagination"`
}

type Pagination struct {
	Page       int  `json:"page" example:"1"`
	PerPage    int  `json:"per_page" example:"10"`
	TotalItems int  `json:"total_items" example:"100"`
	TotalPages int  `json:"total_pages" example:"10"`
	FetchAll   bool `json:"fetch_all" example:"false"`
}

type PaginationResponse struct {
	Data PaginationData `json:"data"`
	CommonResponse
}

type GetConfigSpreadDetailResponse struct {
	Data GetConfigSpreadDetailData `json:"data"`
	CommonResponse
}

type GetConfigSpreadDetailData struct {
	UUID                uuid.UUID       `json:"uuid"`
	MerchantID          uuid.UUID       `json:"merchant_id"`
	SourceCurrency      string          `json:"source_currency"`
	DestinationCurrency string          `json:"destination_currency"`
	SpreadType          string          `json:"spread_type"`
	SpreadValue         decimal.Decimal `json:"spread_value"`
	CreatedAt           time.Time       `json:"created_at"`
	UpdatedAt           time.Time       `json:"updated_at"`
}

type CreateConfigSpreadResponse struct {
	Data CreateConfigSpreadData `json:"data"`
	CommonResponse
}

type CreateConfigSpreadData struct {
	UUID    uuid.UUID `json:"uuid"`
	Created bool      `json:"created"`
}

type UpdateConfigSpreadResponse struct {
	Data UpdateConfigSpreadData `json:"data"`
	CommonResponse
}

type UpdateConfigSpreadData struct {
	UUID    uuid.UUID `json:"uuid"`
	Updated bool      `json:"updated"`
}

type ReConfirmPayoutResponse struct {
	Data    CreatePayoutSessionResponseData `json:"data"`
	Code    string                          `json:"code"`
	Message string                          `json:"message,omitempty"`
	Error   interface{}                     `json:"error,omitempty"`
}
