package bankAccount

type BankAccountResponse struct {
	BeneficiaryBankCode    string `json:"beneficiaryBankCode" db:"beneficiary_bank_code"`
	BeneficiaryBankName    string `json:"beneficiaryBankName" db:"beneficiary_bank_name"`
	BeneficiaryAccountNo   string `json:"beneficiaryAccountNo" db:"beneficiary_account_no"`
	BeneficiaryAccountName string `json:"beneficiaryAccountName" db:"beneficiary_account_name"`
}
