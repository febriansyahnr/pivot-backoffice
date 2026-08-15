package merchant

import (
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx/types"
)

/*
Merchant Document Type:
- NationalIdentityCard
- BusinessLicense
- TaxIdentification
- BusinessRegistration
- CertificateIncorporation
- CertificateNo40
- CertificateLastAmendment
- CertificateDeedAmendment
- CertificateAmendmentAct
- CertificateEstablishment
- CertificateTaxRegistration
- BusinessEnvironmentPhoto
*/
type Document struct {
	Id         string         `db:"id"`
	MerchantId string         `db:"merchant_id"`
	Type       string         `db:"type"`
	Identifier string         `db:"identifier"`
	Location   types.JSONText `db:"location"`
	Hash       string         `db:"hash"`
	Status     string         `db:"status"`
	Notes      string         `db:"notes"`
	CreatedBy  string         `db:"created_by"`
	CreatedAt  time.Time      `db:"created_at"`
	ApprovedBy string         `db:"approved_by"`
	ApprovedAt sql.NullTime   `db:"approved_at"`
	UpdatedAt  time.Time      `db:"updated_at"`
	DeletedAt  sql.NullTime   `db:"deleted_at"`
	// Internal Data
	ObjLocation DocLocation `db:"-"`
}

type DocumentFilterResponse struct {
	DocumentID string    `json:"id" db:"id"`
	MerchantId string    `json:"merchantID" db:"merchant_id"`
	Type       string    `json:"type" db:"type"`
	Identifier string    `json:"identifier" db:"identifier"`
	BucketName *string   `json:"-" db:"bucket"`
	URL        *string   `json:"url" db:"object"`
	Status     string    `json:"status" db:"status"`
	Notes      string    `json:"notes" db:"notes"`
	CreatedBy  string    `json:"createdBy" db:"created_by"`
	CreatedAt  time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt  time.Time `json:"updatedAt" db:"updated_at"`
}

type DocLocation struct {
	Bucket string `json:"bucket"`
	Object string `json:"object"`
}
