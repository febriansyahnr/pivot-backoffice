package accountInquiries

import (
	"database/sql"
	"time"
)

type AccountInquiries struct {
	UUID                   string       `json:"uuid" db:"uuid" example:"b3b3b3b3-3b3b-3b3b-3b3b-3b3b3b3b3b3b"`
	BeneficiaryAccountNo   string       `json:"beneficiaryAccountNo" db:"beneficiary_account_no" example:"8000800808"`
	BeneficiaryAccountName string       `json:"beneficiaryAccountName" db:"beneficiary_account_name" example:"Yories Yolanda"`
	BeneficiaryBankCode    string       `json:"beneficiaryBankCode" db:"beneficiary_bank_code" example:"008"`
	BeneficiaryBankName    string       `json:"beneficiaryBankName" db:"beneficiary_bank_name" example:"Bank 008"`
	Response               string       `json:"response" db:"response" example:"{}"`
	CreatedAt              time.Time    `json:"createdAt" db:"created_at" example:"2021-01-01T00:00:00Z"`
	UpdatedAt              time.Time    `json:"updatedAt" db:"updated_at" example:"2021-01-01T00:00:00Z"`
	DeletedAt              sql.NullTime `json:"-" db:"deleted_at"`
}
