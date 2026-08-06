package withdrawal

import (
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx/types"
)

type Withdrawal struct {
	Id                     string             `db:"id"`
	MerchantId             string             `db:"merchant_id"`
	ReferenceId            string             `db:"reference_id"`
	BeneficiaryBankCode    string             `db:"beneficiary_bank_code"`
	BeneficiaryBankName    string             `db:"beneficiary_bank_name"`
	BeneficiaryAccountNo   string             `db:"beneficiary_account_no"`
	BeneficiaryAccountName string             `db:"beneficiary_account_name"`
	Type                   string             `db:"type"`
	Description            string             `db:"description"`
	Currency               string             `db:"currency"`
	Amount                 float64            `db:"amount"`
	RawMetadata            types.NullJSONText `db:"metadata"`
	CreatedBy              string             `db:"created_by"`
	CreatedAt              time.Time          `db:"created_at"`
	UpdatedAt              time.Time          `db:"updated_at"`
	DeletedAt              sql.NullTime       `db:"deleted_at"`
	// Internal data, not intended to store data in a database
	Metadata Metadata `db:"-"`
}

type Metadata struct {
	BankTransfer *BankTransfer `json:"bankTransfer,omitempty"`
	Source       string        `json:"source"`
	WithdrawType string        `json:"withdrawType"`
	BalanceType  string        `json:"balanceType"`
	IsFullAmount bool          `json:"isFullAmount"`
}

type BankTransfer struct {
	UUID               string `json:"uuid"`
	ExternalId         string `json:"externalId"`
	BankReferenceNo    string `json:"bankReferenceNo"`
	PartnerReferenceNo string `json:"partnerReferenceNo"`
}
