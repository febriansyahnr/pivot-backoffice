package xbModel

import (
	"mime/multipart"
	"time"

	"github.com/google/uuid"

	"github.com/shopspring/decimal"
)

type GetFxRateRequest struct {
	MerchantId          string `json:"-"`
	SourceCurrency      string `json:"sourceCurrency"`
	DestinationCurrency string `json:"destinationCurrency"`
}

type CreatePayoutSessionRequest struct {
	SenderID            string                    `json:"senderId" validate:"required_without=SenderData"`
	SenderData          *CreateSenderRequest      `json:"senderData"`
	BeneficiaryID       string                    `json:"beneficiaryId" validate:"required_without=BeneficiaryData"`
	BeneficiaryData     *CreateBeneficiaryRequest `json:"beneficiaryData"`
	ReferenceId         string                    `json:"referenceId" validate:"required"`
	SourceCurrency      string                    `json:"sourceCurrency" validate:"required"`
	DestinationCurrency string                    `json:"destinationCurrency" validate:"required"`
	DestinationAmount   decimal.Decimal           `json:"destinationAmount" validate:"min=0"`
	PurposeCode         string                    `json:"purposeCode" validate:"required"`
	Remark              string                    `json:"remark" validate:"required,max=20"`
	RoutingValue        string                    `json:"routingValue"`
	CNAPS               string                    `json:"cnaps"`

	MerchantId   string `json:"-"`
	MerchantName string `json:"-"`
	CreatedFrom  string `json:"-"`
	CreatedBy    string `json:"-"`
}

type UploadUnderlyingDocumentRequest struct {
	Document *multipart.FileHeader `validate:"-"`

	PayoutId   string `json:"-"`
	MerchantId string `json:"-"`
}

type SubmitRfiDetailsRequest struct {
	DocumentId string `json:"-" validate:"required"`
	Comment    string `json:"-" validate:"required"`
	Value      string `json:"-"`
	Document   *multipart.FileHeader

	PayoutId   string `json:"-" validate:"required"`
	MerchantId string `json:"-"`
}

type ConfirmPayoutRequest struct {
	PayoutId   string `json:"-"`
	MerchantId string `json:"-"`
	ApprovedBy string `json:"-"`
}

type CreateBeneficiaryRequest struct {
	MerchantId           string `json:"-"`
	Name                 string `json:"name" validate:"required"`
	AccountType          string `json:"accountType" validate:"required"`
	Address              string `json:"address" validate:"required"`
	City                 string `json:"city" validate:"required"`
	Postcode             string `json:"postcode" validate:"required"`
	State                string `json:"state" validate:"required"`
	CountryCode          string `json:"countryCode" validate:"required"`
	IdentificationType   string `json:"identificationType"`
	IdentificationNumber string `json:"identificationNumber"`
	AccountNumber        string `json:"accountNumber" validate:"required"`
	BankName             string `json:"bankName" validate:"required"`
	BankCode             string `json:"bankCode"`
	ContactCountryCode   string `json:"contactCountryCode"`
	ContactNumber        string `json:"contactNumber"`
	Email                string `json:"email"`
	PayoutMethod         string `json:"payoutMethod" validate:"xb_payout_method"`
}

type UpdateBeneficiaryRequest struct {
	BeneficiaryId        string `json:"-"`
	MerchantId           string `json:"-"`
	Name                 string `json:"name"`
	AccountType          string `json:"accountType"`
	Address              string `json:"address" `
	City                 string `json:"city" `
	Postcode             string `json:"postcode" `
	State                string `json:"state" `
	CountryCode          string `json:"countryCode" `
	IdentificationType   string `json:"identificationType"`
	IdentificationNumber string `json:"identificationNumber"`
	AccountNumber        string `json:"accountNumber" `
	BankName             string `json:"bankName" `
	BankCode             string `json:"bankCode"`
	ContactCountryCode   string `json:"contactCountryCode"`
	ContactNumber        string `json:"contactNumber"`
	Email                string `json:"email"`
}

type GetListBeneficiaryRequest struct {
	MerchantId      string `json:"-"`
	Page            int    `json:"page"`
	PerPage         int    `json:"perPage"`
	FetchAll        bool   `json:"fetchAll"`
	ShowDeactivated bool   `json:"showDeactivated"`
	Name            string `json:"name"`
	CountryCode     string `json:"countryCode"`
	AccountNumber   string `json:"accountNumber"`
	AccountType     string `json:"accountType"`
}

type GetBeneficiaryByIdRequest struct {
	MerchantId    string `json:"-"`
	BeneficiaryId string `json:"-"`
}

type GetPayoutRequest struct {
	MerchantId string `json:"-"`
	PayoutId   string `json:"id" validate:"required"`
}

type GetRfiDetailsRequest struct {
	MerchantId string `json:"merchantId" validate:"required"`
	PayoutId   string `json:"id"`
}

type CreateSenderRequest struct {
	MerchantId           string `json:"-"`
	Name                 string `json:"name" validate:"required"`
	CountryCode          string `json:"countryCode" validate:"required"`
	State                string `json:"state" validate:"required"`
	City                 string `json:"city" validate:"required"`
	Address              string `json:"address"  validate:"required"`
	Postcode             string `json:"postcode" validate:"required"`
	AccountType          string `json:"accountType" validate:"required"`
	IdentificationType   string `json:"identificationType" validate:"required"`
	IdentificationNumber string `json:"identificationNumber" validate:"required"`
	BankAccountNumber    string `json:"bankAccountNumber"`
	Dob                  string `json:"dob"`
	ContactCountryCode   string `json:"contactCountryCode"`
	ContactNumber        string `json:"contactNumber"`
	SourceOfIncome       string `json:"sourceOfIncome"`
}

type UpdateSenderRequest struct {
	SenderId             string `json:"-"`
	MerchantId           string `json:"-"`
	Name                 string `json:"name"`
	CountryCode          string `json:"countryCode"`
	State                string `json:"state"`
	City                 string `json:"city"`
	Address              string `json:"address"`
	Postcode             string `json:"postcode"`
	AccountType          string `json:"accountType"`
	IdentificationType   string `json:"identificationType" `
	IdentificationNumber string `json:"identificationNumber" `
	BankAccountNumber    string `json:"bankAccountNumber"`
	Dob                  string `json:"dob"`
	ContactCountryCode   string `json:"contactCountryCode"`
	ContactNumber        string `json:"contactNumber"`
	SourceOfIncome       string `json:"sourceOfIncome"`
}

type GetListSenderRequest struct {
	MerchantId      string `json:"-"`
	Page            int    `json:"page"`
	PerPage         int    `json:"perPage"`
	FetchAll        bool   `json:"fetchAll"`
	ShowDeactivated bool   `json:"showDeactivated"`
	Name            string `json:"name"`
	AccountType     string `json:"accountType"`
}

type GetSenderByIdRequest struct {
	MerchantId string `json:"-"`
	SenderId   string `json:"-"`
}

type ConsumePayoutStatusChangeRequest struct {
	AcquirerTransactionId string    `json:"acquirer_transaction_id"`
	PartnerTransactionId  string    `json:"partner_transaction_id"`
	Status                string    `json:"status"`
	Timestamp             time.Time `json:"timestamp"`
}

type GetListMasterCountryRequest struct {
	MerchantId   string `json:"-"`
	ActiveOnly   bool   `json:"activeOnly"`
	CountryCode  string `json:"countryCode"`
	CurrencyCode string `json:"currencyCode"`
	FetchAll     bool   `json:"fetchAll"`
	Page         int    `json:"page"`
	PerPage      int    `json:"perPage"`
}

type GetListMasterStateRequest struct {
	MerchantId  string `json:"-"`
	CountryCode string `json:"countryCode"`
	Name        string `json:"name"`
	FetchAll    bool   `json:"fetchAll"`
	Page        int    `json:"page"`
	PerPage     int    `json:"perPage"`
}

type GetListMasterCityRequest struct {
	MerchantId string `json:"-"`
	StateUUID  string `json:"stateUuid"`
	Name       string `json:"name"`
	FetchAll   bool   `json:"fetchAll"`
	Page       int    `json:"page"`
	PerPage    int    `json:"perPage"`
}

type GetListMasterCurrencyRequest struct {
	MerchantId string `json:"-"`
	Code       string `json:"code"`
	FetchAll   bool   `json:"fetchAll"`
	Page       int    `json:"page"`
	PerPage    int    `json:"perPage"`
}

type GetListMasterCurrencyMappingRequest struct {
	MerchantId     string `json:"-"`
	FetchAll       bool   `json:"fetchAll"`
	Page           int    `json:"page"`
	PerPage        int    `json:"perPage"`
	CountryCode    string `json:"countryCode"`
	TransferMethod string `json:"transferMethod"`
}

type GetListMasterIdentificationTypeRequest struct {
	MerchantId  string `json:"-"`
	FetchAll    bool   `json:"fetchAll"`
	AccountType string `json:"accountType"`
	Page        int    `json:"page"`
	PerPage     int    `json:"perPage"`
}

type GetListMasterAccountTypeRequest struct {
	MerchantId string `json:"-"`
	FetchAll   bool   `json:"fetchAll"`
	Code       string `json:"code"`
	Page       int    `json:"page"`
	PerPage    int    `json:"perPage"`
}

type GetListMasterPurposeRequest struct {
	MerchantId  string `json:"-"`
	FetchAll    bool   `json:"fetchAll"`
	Code        string `json:"code"`
	Page        int    `json:"page"`
	PerPage     int    `json:"perPage"`
	CountryCode string `json:"countryCode"`
	RoutingCode string `json:"routingCode"`
}

type GetListMasterTransferMethodRequest struct {
	MerchantId string `json:"-"`
	FetchAll   bool   `json:"fetchAll"`
	Code       string `json:"code"`
	Page       int    `json:"page"`
	PerPage    int    `json:"perPage"`
}

type GetListMasterSourceOfIncomeRequest struct {
	MerchantId string `json:"-"`
	FetchAll   bool   `json:"fetchAll"`
	Name       string `json:"name"`
	Page       int    `json:"page"`
	PerPage    int    `json:"perPage"`
}

type GetPayoutFilterRequest struct {
	MerchantID string    `json:"-"`
	UUID       string    `json:"uuid"`
	StartAt    time.Time `json:"startAt"`
	EndAt      time.Time `json:"endAt"`
	Status     string    `json:"status"`
	SortBy     string    `json:"sortBy"`
	Sort       string    `json:"sort"`
}

type GetListConfigSpreadRequest struct {
	Page       int       `json:"page"`
	PerPage    int       `json:"perPage"`
	MerchantID uuid.UUID `json:"merchantId"`
}

type CreateConfigSpreadRequest struct {
	MerchantID          uuid.UUID       `json:"merchantId" validate:"required,uuid"`
	SourceCurrency      string          `json:"sourceCurrency" validate:"required"`
	DestinationCurrency string          `json:"destinationCurrency" validate:"required"`
	SpreadType          string          `json:"spreadType" validate:"required"`
	SpreadValue         decimal.Decimal `json:"spreadValue"`
}

type UpdateConfigSpreadRequest struct {
	UUID                uuid.UUID        `json:"-"`
	SourceCurrency      *string          `json:"sourceCurrency" validate:"required"`
	DestinationCurrency *string          `json:"destinationCurrency" validate:"required"`
	SpreadType          *string          `json:"spreadType" validate:"required"`
	SpreadValue         *decimal.Decimal `json:"spreadValue" validate:"required"`
}

type ExportXbPayoutRequest struct {
	MerchantID string     `json:"-"`
	UUID       string     `json:"uuid"`
	StartAt    *time.Time `json:"startAt"`
	EndAt      *time.Time `json:"endAt"`
	Status     string     `json:"status"`
}
