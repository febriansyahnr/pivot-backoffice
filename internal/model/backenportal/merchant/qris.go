package merchant

import (
	"github.com/paper-indonesia/pivot-backoffice/internal/model/qris"

	"github.com/jmoiron/sqlx/types"
)

type QrisMerchant struct {
	RegId                 string         `db:"registration_id" validate:"-"`
	Id                    string         `db:"uuid" validate:"-"`
	MID                   string         `db:"mid" validate:"required,numeric"`
	ExternalId            string         `db:"external_id" validate:"required" label:"external_id"`
	Type                  string         `db:"-" validate:"-" label:"merchant_type"`
	ParentId              string         `db:"parent_id" validate:"required_if=Type Franchisee,required_if=Type Sub-Merchant" label:"parent_id"`
	Name                  string         `db:"name" validate:"required" label:"name"`
	ShortName             string         `db:"short_name" validate:"required" label:"short_name"`
	ParentName            string         `db:"parent_name" validate:"-"`
	MCC                   string         `db:"mcc" json:"mcc"`
	AddressRaw            types.JSONText `db:"address" validate:"-"`
	Address               qris.Address   `db:"-" validate:"required" label:"address"`
	Documents             []QrisDocument `db:"-" validate:"-"`
	BODCount              int            `db:"-" validate:"-"`
	BOCCount              int            `db:"-" validate:"-"`
	BoardOfDirectors      []QrisBOD      `db:"-" validate:"-"`
	RegAcquirerMerchantId string         `db:"qr_acquirer_merchant_id" validate:"required_if=Type Franchisee,required_if=Type Sub-Merchant" label:"qr_acquirer_merchant_id"`
	RegStatus             string         `db:"qr_status" validate:"omitempty,oneof=FILLING_FORM FAILED" label:"qr_status"`
}

type QrisDocument struct {
	Type        string         `db:"type"`
	Number      string         `db:"number"`
	LocationRaw types.JSONText `db:"location"`
	Location    DocLocation    `db:"-"`
}

type QrisBOD struct {
	Position        string         `db:"position"`
	IdentityNumber  string         `db:"identity_number"`
	IdentityFileRaw types.JSONText `db:"identity_file"`
	IdentityFile    DocLocation    `db:"-"`
}
