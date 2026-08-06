package schemas

func BeneficiaryAccount() string {
	return `CREATE TABLE IF NOT EXISTS beneficiary_accounts (
		uuid varchar(100) NOT NULL,
		merchant_id varchar(100) NOT NULL,
		beneficiary_bank_code varchar(20) NOT NULL,
		beneficiary_bank_name varchar(150) DEFAULT NULL,
		beneficiary_account_no varchar(60) NOT NULL,
		beneficiary_account_name varchar(150) DEFAULT NULL,
		metadata json DEFAULT NULL,
		created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at timestamp(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
		deleted_at timestamp NULL DEFAULT NULL,
		PRIMARY KEY (uuid),
		UNIQUE KEY beneficiary_accounts_merchant_bank_code_account_no_comp_uniq_idx (merchant_id,beneficiary_bank_code,beneficiary_account_no,deleted_at)
	) ENGINE=InnoDB;`
}
