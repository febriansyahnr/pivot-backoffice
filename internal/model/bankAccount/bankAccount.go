package bankAccount

import "time"

type BankAccount struct {
	UUID                   string    `db:"id" example:"b3b3b3b3-3b3b-3b3b-3b3b-3b3b3b3b3b3b"`
	MerchantID             string    `db:"merchant_id" example:"a1a1a1a1-1a1a-1a1a-1a1a-1a1a1a1a1a1a"`
	BeneficiaryAccountNo   string    `db:"beneficiary_account_no" example:"8000800808"`
	BeneficiaryAccountName string    `db:"beneficiary_account_name" example:"Yories Yolanda"`
	BeneficiaryBankCode    string    `db:"beneficiary_bank_code" example:"008"`
	BeneficiaryBankName    string    `db:"beneficiary_bank_name" example:"Bank 008"`
	CreatedBy              string    `db:"created_by" example:"Yories Yolanda"`
	CreatedAt              time.Time `db:"created_at" example:"2021-01-01T00:00:00Z"`
	UpdatedBy              string    `db:"updated_by" example:"Yories Yolanda"`
	UpdatedAt              time.Time `db:"updated_at" example:"2021-01-01T00:00:00Z"`
	Deleted                bool      `db:"deleted" example:"false"`
	DeletedAt              time.Time `db:"deleted_at" example:"2021-01-01T00:00:00Z"`
}

func (b *BankAccount) ToResponse() *BankAccountResponse {
	return &BankAccountResponse{
		BeneficiaryBankCode:    b.BeneficiaryBankCode,
		BeneficiaryBankName:    b.BeneficiaryBankName,
		BeneficiaryAccountNo:   b.BeneficiaryAccountNo,
		BeneficiaryAccountName: b.BeneficiaryAccountName,
	}
}
