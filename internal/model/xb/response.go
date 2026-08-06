package xbModel

import (
	"time"

	pb "github.com/paper-indonesia/pivot-backoffice/internal/model/proto/messages/callback"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type GetFxRateResponse struct {
	FxRate            decimal.Decimal `json:"fxRate"`
	DestinationFxRate decimal.Decimal `json:"destinationFxRate"`
	ExpiryAt          time.Time       `json:"expiryAt"`
}

type CreatePayoutSessionResponse struct {
	Uuid                string                  `json:"uuid"`
	MerchantId          string                  `json:"merchantId"`
	ReferenceId         string                  `json:"referenceId"`
	SourceCurrency      string                  `json:"sourceCurrency"`
	DestinationCurrency string                  `json:"destinationCurrency"`
	DestinationAmount   decimal.Decimal         `json:"destinationAmount"`
	FxRate              decimal.Decimal         `json:"fxRate"`
	DestinationFxRate   decimal.Decimal         `json:"destinationFxRate"`
	Fee                 decimal.Decimal         `json:"fee"`
	TotalAmount         decimal.Decimal         `json:"totalAmount"`
	Remark              string                  `json:"remark"`
	CreatedAt           time.Time               `json:"createdAt"`
	ExpiredAt           time.Time               `json:"expiredAt"`
	Status              string                  `json:"status"`
	SenderId            string                  `json:"senderId"`
	BeneficiaryId       string                  `json:"beneficiaryId"`
	BeneficiaryData     BeneficiaryDataResponse `json:"beneficiaryData"`
	SenderData          SenderDataResponse      `json:"senderData"`
	RoutingCode         string                  `json:"routingCode"`
	RoutingValue        string                  `json:"routingValue"`
}

type UploadUnderlyingDocumentResponse struct {
	DocumentReference string `json:"documentReference"`
}

type SubmitRfiDetailsResponse struct {
	Uuid             string     `json:"uuid"`
	MerchantId       string     `json:"merchantId"`
	ReferenceId      string     `json:"referenceId"`
	DocumentID       string     `json:"documentId,omitempty"`       // -> rfi_id
	DocumentEntityID string     `json:"documentEntityId,omitempty"` // -> rfi_entity_id
	Actor            string     `json:"actor,omitempty"`            // BENEFICIARY, REMITTER
	Entity           string     `json:"entity,omitempty"`           // rfi_entity
	Type             string     `json:"type,omitempty"`             // rfi_type
	URL              string     `json:"url,omitempty"`              // public url
	Filename         string     `json:"filename,omitempty"`         // filename
	Value            string     `json:"value,omitempty"`            // rfi_value
	Comment          string     `json:"comment,omitempty"`          // comment
	Status           string     `json:"status,omitempty"`           // pending, received
	RequestedAt      *time.Time `json:"requestedAt,omitempty"`      // rfi_requested_datetime
}

type ConfirmPayoutResponse struct {
	Uuid                string                  `json:"uuid"`
	MerchantId          string                  `json:"merchantId"`
	ReferenceId         string                  `json:"referenceId"`
	SourceCurrency      string                  `json:"sourceCurrency"`
	DestinationCurrency string                  `json:"destinationCurrency"`
	DestinationAmount   decimal.Decimal         `json:"destinationAmount"`
	FxRate              decimal.Decimal         `json:"fxRate"`
	DestinationFxRate   decimal.Decimal         `json:"destinationFxRate"`
	Fee                 decimal.Decimal         `json:"fee"`
	TotalAmount         decimal.Decimal         `json:"totalAmount"`
	CreatedAt           time.Time               `json:"createdAt"`
	Remark              string                  `json:"remark"`
	Status              string                  `json:"status"`
	SenderId            string                  `json:"senderId"`
	BeneficiaryId       string                  `json:"beneficiaryId"`
	BeneficiaryData     BeneficiaryDataResponse `json:"beneficiaryData"`
	SenderData          SenderDataResponse      `json:"senderData"`
}

type CreateBeneficiaryResponse struct {
	UUID                 uuid.UUID  `json:"uuid,omitempty"`
	Name                 string     `json:"name"`
	AccountType          string     `json:"accountType"`
	Address              string     `json:"address"`
	City                 string     `json:"city"`
	Postcode             string     `json:"postcode"`
	State                string     `json:"state"`
	CountryCode          string     `json:"countryCode"`
	IdentificationType   string     `json:"identificationType"`
	IdentificationNumber string     `json:"identificationNumber"`
	AccountNumber        string     `json:"accountNumber"`
	BankName             string     `json:"bankName"`
	BankCode             string     `json:"bankCode"`
	ContactCountryCode   string     `json:"contactCountryCode"`
	ContactNumber        string     `json:"contactNumber"`
	Email                string     `json:"email"`
	CreatedAt            time.Time  `json:"createdAt"`
	UpdatedAt            *time.Time `json:"updatedAt,omitempty"`
	DeactivatedAt        *time.Time `json:"deactivatedAt,omitempty"`
	Created              bool       `json:"-"`
}

type CreateSenderResponse struct {
	UUID                 uuid.UUID  `json:"uuid,omitempty"`
	Name                 string     `json:"name"`
	CountryCode          string     `json:"countryCode"`
	State                string     `json:"state"`
	City                 string     `json:"city"`
	Address              string     `json:"address"`
	Postcode             string     `json:"postcode"`
	AccountType          string     `json:"accountType"`
	IdentificationType   string     `json:"identificationType"`
	IdentificationNumber string     `json:"identificationNumber"`
	BankAccountNumber    string     `json:"bankAccountNumber"`
	Dob                  string     `json:"dob"`
	ContactCountryCode   string     `json:"contactCountryCode"`
	ContactNumber        string     `json:"contactNumber"`
	SourceOfIncome       string     `json:"sourceOfIncome"`
	CreatedAt            time.Time  `json:"createdAt"`
	UpdatedAt            *time.Time `json:"updatedAt,omitempty"`
	DeactivatedAt        *time.Time `json:"deactivatedAt,omitempty"`
	Created              bool       `json:"-"`
}

type GetXbPayoutListResponse struct {
	UUID                   string          `json:"uuid"`
	ReferenceID            string          `json:"referenceId"`
	SourceCurrency         string          `json:"sourceCurrency"`
	DestinationCurrency    string          `json:"destinationCurrency"`
	DestinationAmount      decimal.Decimal `json:"destinationAmount"`
	SourceAmount           decimal.Decimal `json:"sourceAmount"`
	Fee                    decimal.Decimal `json:"fee"`
	TotalAmount            decimal.Decimal `json:"totalAmount"`
	CreatedAt              time.Time       `json:"createdAt"`
	Status                 string          `json:"status"`
	BeneficiaryAccountName string          `json:"beneficiaryAccountName"`
}

type GetXbPayoutDetailResponse struct {
	UUID                string                  `json:"uuid"`
	MerchantId          string                  `json:"merchantId"`
	ReferenceId         string                  `json:"referenceId"`
	SourceCurrency      string                  `json:"sourceCurrency"`
	DestinationCurrency string                  `json:"destinationCurrency"`
	DestinationAmount   decimal.Decimal         `json:"destinationAmount"`
	FxRate              decimal.Decimal         `json:"fxRate"`
	DestinationFxRate   decimal.Decimal         `json:"destinationFxRate"`
	SourceAmount        decimal.Decimal         `json:"sourceAmount"`
	Fee                 decimal.Decimal         `json:"fee"`
	TotalAmount         decimal.Decimal         `json:"totalAmount"`
	CreatedAt           time.Time               `json:"createdAt"`
	UpdatedAt           time.Time               `json:"updatedAt"`
	ExpiredAt           time.Time               `json:"expiredAt"`
	Status              string                  `json:"status"`
	StatusDescription   string                  `json:"statusDescription"`
	PurposeCode         string                  `json:"purposeCode"`
	Remark              string                  `json:"remark"`
	SenderData          SenderDataResponse      `json:"senderData"`
	BeneficiaryId       string                  `json:"beneficiaryId"`
	BeneficiaryData     BeneficiaryDataResponse `json:"beneficiaryData"`
	RoutingCode         string                  `json:"routingCode"`
	RoutingValue        string                  `json:"routingValue"`
	// Tracking transaction status
	StatusHistories []XbPayoutStatusHistoryResponse `json:"statusHistory,omitempty"`
}

type XbPayoutStatusHistoryResponse struct {
	Label          string    `json:"label"`
	Description    string    `json:"description"`
	Recommendation string    `json:"recommendation,omitempty"`
	Timestamp      time.Time `json:"timestamp"`
}

type BeneficiaryDataResponse struct {
	Name               string `json:"name"`
	CountryCode        string `json:"countryCode"`
	CountryName        string `json:"countryName,omitempty"`
	State              string `json:"state"`
	City               string `json:"city"`
	Address            string `json:"address"`
	Postcode           string `json:"postcode"`
	AccountType        string `json:"accountType"`
	AccountNumber      string `json:"accountNumber"`
	BankName           string `json:"bankName"`
	BankCode           string `json:"bankCode"`
	ContactCountryCode string `json:"contactCountryCode"`
	ContactNumber      string `json:"contactNumber"`
	Email              string `json:"email"`
	PayoutMethod       string `json:"payoutMethod"`
}

func (d *BeneficiaryDataResponse) ToProtoBeneficiaryXbDataCallback() *pb.BeneficiaryXbData {
	return &pb.BeneficiaryXbData{
		Name:               d.Name,
		Address:            d.Address,
		City:               d.City,
		Postcode:           d.Postcode,
		State:              d.State,
		Country:            d.CountryName,
		AccountNumber:      d.AccountNumber,
		BankName:           d.BankName,
		BankCode:           d.BankCode,
		ContactCountryCode: d.ContactCountryCode,
		ContactNumber:      d.ContactNumber,
		Email:              d.Email,
		AccountType:        d.AccountType,
		CountryCode:        d.CountryCode,
	}
}

type SenderDataResponse struct {
	Name                 string `json:"name"`
	CountryCode          string `json:"countryCode"`
	CountryName          string `json:"countryName,omitempty"`
	State                string `json:"state"`
	City                 string `json:"city"`
	Address              string `json:"address"`
	Postcode             string `json:"postcode"`
	AccountType          string `json:"accountType"`
	IdentificationType   string `json:"identificationType"`
	IdentificationNumber string `json:"identificationNumber"`
	BankAccountNumber    string `json:"bankAccountNumber"`
	Dob                  string `json:"dob"`
	ContactCountryCode   string `json:"contactCountryCode"`
	ContactNumber        string `json:"contactNumber"`
	SourceOfIncome       string `json:"sourceOfIncome"`
}

func (d *SenderDataResponse) ToProtoSenderDataCallback() *pb.SenderXbData {
	return &pb.SenderXbData{
		Name:                 d.Name,
		Country:              d.CountryName,
		State:                d.State,
		City:                 d.City,
		Address:              d.Address,
		Postcode:             d.Postcode,
		AccountType:          d.AccountType,
		IdentificationType:   d.IdentificationType,
		IdentificationNumber: d.IdentificationNumber,
		BankAccountNumber:    d.BankAccountNumber,
		ContactCountryCode:   d.ContactCountryCode,
		ContactNumber:        d.ContactNumber,
		Dob:                  d.Dob,
		SourceOfIncome:       d.SourceOfIncome,
		CountryCode:          d.CountryCode,
	}
}

type GetPayoutResponse struct {
	Uuid                string                  `json:"uuid"`
	MerchantId          string                  `json:"merchantId"`
	ReferenceId         string                  `json:"referenceId"`
	SourceCurrency      string                  `json:"sourceCurrency"`
	DestinationCurrency string                  `json:"destinationCurrency"`
	DestinationAmount   decimal.Decimal         `json:"destinationAmount"`
	FxRate              decimal.Decimal         `json:"fxRate"`
	DestinationFxRate   decimal.Decimal         `json:"destinationFxRate"`
	Fee                 decimal.Decimal         `json:"fee"`
	TotalAmount         decimal.Decimal         `json:"totalAmount"`
	Remark              string                  `json:"remark"`
	CreatedAt           time.Time               `json:"createdAt"`
	BeneficiaryData     BeneficiaryDataResponse `json:"beneficiaryData"`
	SenderData          SenderDataResponse      `json:"senderData"`
	Status              string                  `json:"status"`
	StatusDescription   string                  `json:"statusDescription,omitempty"`
	RfiDetails          []*RfiDetails           `json:"rfiDetails,omitempty"`
	RoutingCode         string                  `json:"routingCode"`
	RoutingValue        string                  `json:"routingValue"`
}

type GetRfiDetailsResponse struct {
	Uuid        string        `json:"uuid"`
	MerchantId  string        `json:"merchantId"`
	ReferenceId string        `json:"referenceId"`
	RfiDetails  []*RfiDetails `json:"rfiDetails,omitempty"`
}

type RfiDetails struct {
	PartnerDocumentID       string     `json:"documentId,omitempty"`       // -> rfi_id
	PartnerDocumentEntityID string     `json:"documentEntityId,omitempty"` // -> rfi_entity_id
	Actor                   string     `json:"actor,omitempty"`            // BENEFICIARY, REMITTER
	Entity                  string     `json:"entity,omitempty"`           // rfi_entity
	DocumentType            string     `json:"type,omitempty"`             // rfi_type
	DocumentURL             string     `json:"url,omitempty"`              // public url
	Filename                string     `json:"filename,omitempty"`         // filename
	Value                   string     `json:"value,omitempty"`            // rfi_value
	Comment                 string     `json:"comment,omitempty"`          // comment
	Status                  string     `json:"status,omitempty"`           // pending, received
	RequestedAt             *time.Time `json:"requestedAt,omitempty"`      // rfi_requested_datetime
}

type PaginationResponse struct {
	Results    interface{} `json:"results"`
	Pagination Pagination  `json:"pagination"`
}

type Pagination struct {
	Page       int  `json:"page" example:"1"`
	PerPage    int  `json:"perPage" example:"10"`
	TotalItems int  `json:"totalItems" example:"100"`
	TotalPages int  `json:"totalPages" example:"10"`
	FetchAll   bool `json:"fetchAll" example:"false"`
}

type GetConfigSpreadDetailResponse struct {
	UUID                uuid.UUID       `json:"uuid"`
	MerchantID          uuid.UUID       `json:"merchantId"`
	SourceCurrency      string          `json:"sourceCurrency"`
	DestinationCurrency string          `json:"destinationCurrency"`
	SpreadType          string          `json:"spreadType"`
	SpreadValue         decimal.Decimal `json:"spreadValue"`
	CreatedAt           time.Time       `json:"createdAt"`
	UpdatedAt           time.Time       `json:"updatedAt"`
}

type CreateConfigSpreadResponse struct {
	UUID    uuid.UUID `json:"uuid"`
	Created bool      `json:"created"`
}

type UpdateConfigSpreadResponse struct {
	UUID    uuid.UUID `json:"uuid"`
	Updated bool      `json:"updated"`
}

type ReConfirmEvent struct {
	NeedAutoConfirm bool   `json:"needAutoConfirm"`
	MerchantID      string `json:"merchantId"`
	PayoutId        string `json:"payoutId"`
}

type ExportXbPayoutResponse struct {
	Url string `json:"url"`
}
