package qris

import (
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx/types"
)

type Registration struct {
	Id                       string             `db:"id"`
	ExternalId               string             `db:"external_id"`
	Acquirer                 string             `db:"acquirer"`
	MerchantType             string             `db:"merchant_type"`
	AcquirerParentMerchantId string             `db:"acquirer_parent_merchant_id"`
	MerchantName             string             `db:"merchant_name"`
	MerchantShortName        string             `db:"merchant_short_name"`
	AddressRaw               types.JSONText     `db:"address"`
	BusinessInfoRaw          types.JSONText     `db:"business_info"`
	BusinessDocumentRaw      types.JSONText     `db:"business_document"`
	Status                   string             `db:"status"`
	AcquirerMerchantId       *string            `db:"acquirer_merchant_id"`
	AcquirerTerminalId       *string            `db:"acquirer_terminal_id"`
	CallbackDetailRaw        types.NullJSONText `db:"callback_detail"`
	CallbackDatetime         sql.NullTime       `db:"callback_datetime"`
	CreatedAt                time.Time          `db:"created_at"`
	CreatedBy                string             `db:"created_by"`
	UpdatedAt                time.Time          `db:"updated_at"`
	// Internal Data
	Address          Address          `db:"-"`
	BusinessInfo     BusinessInfo     `db:"-"`
	BusinessDocument BusinessDocument `db:"-"`
	CallbackDetail   CallbackDetail   `db:"-"`
}

type Address struct {
	Province uint16 `json:"province" validate:"required"`
	City     uint16 `json:"city" validate:"required"`
	District uint16 `json:"district" validate:"required"`
	PostCode string `json:"postcode" validate:"required"`
	Detail   string `json:"detail" validate:"required"`
}

type BusinessInfo struct {
	NationalIdentityCardNumber string `json:"nationalIdentityCardNumber"`
	NationalIdentityCardFile   Media  `json:"nationalIdentityCardFile"`
	BusinessLicenseNumber      string `json:"businessLicenseNumber"`
	BusinessLicenseFile        Media  `json:"businessLicenseFile"`
	TaxIdentificationNumber    string `json:"taxIdentificationNumber"`
	TaxIdentificationFile      Media  `json:"taxIdentificationFile"`
	BusinessRegistrationNumber string `json:"businessRegistrationNumber"`
	BusinessRegistrationFile   Media  `json:"businessRegistrationFile"`
}

type BusinessDocument struct {
	CertificateIncorporation   Media `json:"certificateIncorporation"`
	CertificateNo40            Media `json:"certificateNo40"`
	CertificateLastAmendment   Media `json:"certificateLastAmendment"`
	CertificateDeedAmendment   Media `json:"certificateDeedAmendment"`
	CertificateAmendmentAct    Media `json:"certificateAmendmentAct"`
	CertificateEstablishment   Media `json:"certificateEstablishment"`
	CertificateTaxRegistration Media `json:"certificateTaxRegistration"`
	BusinessEnvironmentPhoto   Media `json:"businessEnvironmentPhoto"`
}

type Media struct {
	Internal Bucket `json:"internal"`
	External string `json:"external"`
}

type Bucket struct {
	Bucket string `json:"bucket"`
	Object string `json:"object"`
}

type UpdateDocument struct {
	Type   string `json:"type"`
	Number string `json:"number"`
	Media  Media  `json:"media"`
}

type RegistrationCallback struct {
	Id            string    `json:"applicationCode"`
	ApplymentCode string    `json:"applymentCode"`
	MerchantId    string    `json:"mId"`
	Status        string    `json:"auditStatus"`
	ResultCode    string    `json:"resultCode"`
	Datetime      time.Time `json:"dateTime"`

	AuditDetail []RegistrationCallbackAuditDetail `json:"auditDetail"`
}

type RegistrationCallbackAuditDetail struct {
	RejectCode   []string `json:"rejectCode"`
	RejectReason string   `json:"rejectReason"`
}

type CallbackDetail struct {
	ApplymentCode string `json:"applymentCode"`
	ResultCode    string `json:"resultCode"`

	AuditDetail []RegistrationCallbackAuditDetail `json:"auditDetail"`
}

type RegistrationMerchant struct {
	Id           string `db:"id"`
	ExternalId   string `db:"external_id"`
	MerchantId   string `db:"merchant_id"`
	Acquirer     string `db:"acquirer"`
	MerchantName string `db:"merchant_name"`
	Status       string `db:"status"`
}
