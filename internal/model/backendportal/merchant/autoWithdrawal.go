package merchant

type MerchantWithActiveAutoWithdrawalStatus struct {
	MerchantId           string `json:"merchantId" db:"merchant_id"`
	MerchantName         string `json:"merchantName" db:"merchant_name"`
	AccountName          string `json:"accountName" db:"account_name"`
	BeneficiaryBankCode  string `json:"beneficiaryBankCode" db:"beneficiary_bank_code"`
	BeneficiaryAccountNo string `json:"beneficiaryAccountNo" db:"beneficiary_account_no"`
}

type ForceAutoWithdrawalProcessResponse struct {
	Total   int   `json:"total"`
	Skip    int64 `json:"skip"`
	Failed  int64 `json:"failed"`
	Notify  int64 `json:"notify"`
	Dormant int64 `json:"dormant"`
}

type MerchantWithdrawalDetails struct {
	MerchantId             string `json:"merchantId" db:"merchant_id"`
	MerchantName           string `json:"merchantName" db:"merchant_name"`
	MerchantEmail          string `json:"merchantEmail" db:"merchant_email"`
	AccountId              string `json:"accountId" db:"account_id"`
	AccountName            string `json:"accountName" db:"account_name"`
	BeneficiaryBankCode    string `json:"beneficiaryBankCode" db:"beneficiary_bank_code"`
	BeneficiaryBankName    string `json:"beneficiaryBankName" db:"beneficiary_bank_name"`
	BeneficiaryAccountNo   string `json:"beneficiaryAccountNo" db:"beneficiary_account_no"`
	BeneficiaryAccountName string `json:"beneficiaryAccountName" db:"beneficiary_account_name"`
}
