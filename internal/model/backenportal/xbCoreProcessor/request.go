package xbCoreProcessorModel

import (
	"mime/multipart"

	"github.com/google/uuid"

	"github.com/shopspring/decimal"
)

type GetFxRateRequest struct {
	MerchantId          string `json:"-"`
	SourceCurrency      string `json:"sourceCurrency"`
	DestinationCurrency string `json:"destinationCurrency"`
	RequestType         string `json:"requestType"`
}

type CreatePayoutSessionRequest struct {
	MerchantId          string          `json:"-"`
	ReferenceId         string          `json:"reference_id" validate:"required"`
	SourceCurrency      string          `json:"source_currency" validate:"required"`
	DestinationCurrency string          `json:"destination_currency" validate:"required"`
	DestinationAmount   decimal.Decimal `json:"destination_amount" validate:"min=0"`
	RemitterId          string          `json:"remitter_id" validate:"required"`
	BeneficiaryId       string          `json:"beneficiary_id" validate:"required"`
	StatementNarrative  string          `json:"statement_narrative" validate:"required"`
	PurposeCode         string          `json:"purpose_code" validate:"required"`
	RoutingValue        string          `json:"routing_value" validate:"required"`
	CNAPS               string          `json:"cnaps"`
}

type ConfirmPayoutRequest struct {
	MerchantId            string `json:"-"`
	XbPayoutId            string `json:"-"`
	AcquirerTransactionId string `json:"acquirer_transaction_id" validate:"required"`
}

type UploadUnderlyingDocumentRequest struct {
	MerchantId string                `json:"-"`
	XbPayoutId string                `json:"-"`
	Document   *multipart.FileHeader `validate:"-"`
}

type SubmitRfiDetailsRequest struct {
	MerchantId string                `json:"-"`
	PayoutId   string                `json:"payout_id"`
	DocumentId string                `json:"document_id"`
	Comment    string                `json:"comment"`
	Value      string                `json:"value"`
	Document   *multipart.FileHeader `validate:"-"`
}

type CreateBeneficiaryRequest struct {
	MerchantId           string `json:"merchant_id"`
	Name                 string `json:"name"`
	AccountType          string `json:"account_type"`
	Address              string `json:"address"`
	City                 string `json:"city"`
	Postcode             string `json:"postcode"`
	State                string `json:"state"`
	CountryCode          string `json:"country_code"`
	IdentificationType   string `json:"identification_type"`
	IdentificationNumber string `json:"identification_number"`
	AccountNumber        string `json:"account_number"`
	BankName             string `json:"bank_name"`
	BankCode             string `json:"bank_code"`
	ContactCountryCode   string `json:"contact_country_code"`
	ContactNumber        string `json:"contact_number"`
	Email                string `json:"email"`
	PayoutMethod         string `json:"payout_method"`
}

type UpdateBeneficiaryRequest struct {
	MerchantId           string `json:"merchant_id"`
	BeneficiaryId        string `json:"_"`
	Name                 string `json:"name"`
	AccountType          string `json:"account_type"`
	Address              string `json:"address"`
	City                 string `json:"city"`
	Postcode             string `json:"postcode"`
	State                string `json:"state"`
	CountryCode          string `json:"country_code"`
	IdentificationType   string `json:"identification_type"`
	IdentificationNumber string `json:"identification_number"`
	AccountNumber        string `json:"account_number"`
	BankName             string `json:"bank_name"`
	BankCode             string `json:"bank_code"`
	ContactCountryCode   string `json:"contact_country_code"`
	ContactNumber        string `json:"contact_number"`
	Email                string `json:"email"`
}

type GetListBeneficiaryRequest struct {
	MerchantId      string `json:"-"`
	Page            int    `json:"page"`
	PerPage         int    `json:"per_page"`
	FetchAll        bool   `json:"fetch_all"`
	ShowDeactivated bool   `json:"show_deactivated"`
	Name            string `json:"name"`
	CountryCode     string `json:"country_code"`
	AccountNumber   string `json:"account_number"`
	AccountType     string `json:"account_type"`
}

type GetBeneficiaryByIdRequest struct {
	MerchantId    string `json:"-"`
	BeneficiaryId string `json:"beneficiary_id"`
}

type GetPayoutRequest struct {
	MerchantId string `json:"-"`
	Id         string `json:"id"`
}

type GetRfiDetailsRequest struct {
	MerchantId string `json:"-"`
	Id         string `json:"id"`
}

type CreateSenderRequest struct {
	MerchantId           string `json:"-"`
	Name                 string `json:"name"`
	AccountType          string `json:"account_type"`
	Address              string `json:"address"`
	City                 string `json:"city"`
	Postcode             string `json:"postcode"`
	State                string `json:"state"`
	CountryCode          string `json:"country_code"`
	IdentificationType   string `json:"identification_type"`
	IdentificationNumber string `json:"identification_number"`
	BankAccountNumber    string `json:"bank_account_number"`
	Dob                  string `json:"dob"`
	ContactCountryCode   string `json:"contact_country_code"`
	ContactNumber        string `json:"contact_number"`
	SourceOfIncome       string `json:"source_of_income"`
}

type UpdateSenderRequest struct {
	MerchantId           string `json:"merchant_id"`
	SenderId             string `json:"-"`
	Name                 string `json:"name"`
	AccountType          string `json:"account_type"`
	Address              string `json:"address"`
	City                 string `json:"city"`
	Postcode             string `json:"postcode"`
	State                string `json:"state"`
	CountryCode          string `json:"country_code"`
	IdentificationType   string `json:"identification_type"`
	IdentificationNumber string `json:"identification_number"`
	BankAccountNumber    string `json:"bank_account_number"`
	Dob                  string `json:"dob"`
	ContactCountryCode   string `json:"contact_country_code"`
	ContactNumber        string `json:"contact_number"`
	SourceOfIncome       string `json:"source_of_income"`
}

type GetListSenderRequest struct {
	MerchantId      string `json:"-"`
	Page            int    `json:"page"`
	PerPage         int    `json:"per_page"`
	FetchAll        bool   `json:"fetch_all"`
	ShowDeactivated bool   `json:"show_deactivated"`
	Name            string `json:"name"`
	AccountType     string `json:"account_type"`
}

type GetSenderByIdRequest struct {
	MerchantId string `json:"-"`
	SenderId   string `json:"sender_id"`
}

type GetListMasterCountryRequest struct {
	ActiveOnly   bool   `json:"active_only"`
	CountryCode  string `json:"country_code"`
	CurrencyCode string `json:"currency_code"`
	FetchAll     bool   `json:"fetch_all"`
	Page         int    `json:"page"`
	PerPage      int    `json:"per_page"`
}

type GetListMasterCurrencyRequest struct {
	Code     string `json:"code"`
	FetchAll bool   `json:"fetch_all"`
	Page     int    `json:"page"`
	PerPage  int    `json:"per_page"`
}

type GetListMasterStateRequest struct {
	CountryCode string `json:"country_code"`
	Name        string `json:"name"`
	FetchAll    bool   `json:"fetch_all"`
	Page        int    `json:"page"`
	PerPage     int    `json:"per_page"`
}

type GetListMasterCityRequest struct {
	StateUUID string `json:"state_uuid"`
	Name      string `json:"name"`
	FetchAll  bool   `json:"fetch_all"`
	Page      int    `json:"page"`
	PerPage   int    `json:"per_page"`
}

type GetListMasterCurrencyMappingRequest struct {
	FetchAll       bool   `json:"fetch_all"`
	Page           int    `json:"page"`
	PerPage        int    `json:"per_page"`
	CountryCode    string `json:"country_code"`
	TransferMethod string `json:"transfer_method"`
}

type GetListMasterIdentificationTypeRequest struct {
	FetchAll    bool   `json:"fetch_all"`
	AccountType string `json:"account_type"`
	Page        int    `json:"page"`
	PerPage     int    `json:"per_page"`
}

type GetListMasterAccountTypeRequest struct {
	FetchAll bool   `json:"fetch_all"`
	Code     string `json:"code"`
	Page     int    `json:"page"`
	PerPage  int    `json:"per_page"`
}

type GetListMasterPurposeRequest struct {
	FetchAll    bool   `json:"fetch_all"`
	Code        string `json:"code"`
	Page        int    `json:"page"`
	PerPage     int    `json:"per_page"`
	CountryCode string `json:"country_code"`
	RoutingCode string `json:"routing_code"`
}

type GetListMasterTransferMethodRequest struct {
	FetchAll bool   `json:"fetch_all"`
	Code     string `json:"code"`
	Page     int    `json:"page"`
	PerPage  int    `json:"per_page"`
}

type GetListMasterSourceOfIncomeRequest struct {
	FetchAll bool   `json:"fetch_all"`
	Name     string `json:"name"`
	Page     int    `json:"page"`
	PerPage  int    `json:"per_page"`
}

type GetListConfigSpreadRequest struct {
	Page       int    `json:"page"`
	PerPage    int    `json:"per_page"`
	MerchantID string `json:"merchant_id"`
}

type CreateConfigSpreadRequest struct {
	MerchantID          uuid.UUID       `json:"merchant_id" validate:"required"`
	SourceCurrency      string          `json:"source_currency" validate:"required"`
	DestinationCurrency string          `json:"destination_currency" validate:"required"`
	SpreadType          string          `json:"spread_type" validate:"required"`
	SpreadValue         decimal.Decimal `json:"spread_value"`
}

type UpdateConfigSpreadRequest struct {
	UUID                uuid.UUID        `json:"-"`
	SourceCurrency      *string          `json:"source_currency" validate:"required"`
	DestinationCurrency *string          `json:"destination_currency" validate:"required"`
	SpreadType          *string          `json:"spread_type" validate:"required"`
	SpreadValue         *decimal.Decimal `json:"spread_value" validate:"required"`
}
