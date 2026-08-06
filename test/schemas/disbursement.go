package schemas

func Disbursement() string {
	return `
	CREATE TABLE disbursements (
		uuid varchar(100) NOT NULL,
		reference_id varchar(100) NOT NULL,
		merchant_id varchar(100) NOT NULL,
		bulk_id varchar(100) DEFAULT NULL,
		purpose_id varchar(100) DEFAULT NULL,
		type VARCHAR(50) DEFAULT 'LOCAL_PAYOUT',
		sender_name varchar(60) NOT NULL,
		account_inquiry_id varchar(36) NULL,
		beneficiary_bank_code varchar(20) NOT NULL,
		beneficiary_bank_name varchar(60) DEFAULT NULL,
		beneficiary_account_no varchar(60) NOT NULL,
		beneficiary_account_name varchar(60) NOT NULL,
		processor_reference_id varchar(100) DEFAULT NULL,
		bank_reference_no varchar(60) DEFAULT NULL,
		currency varchar(3) NOT NULL COMMENT 'IDR, USD',
		amount decimal(18,2) NOT NULL,
		fee decimal(18,2) DEFAULT NULL,
		total_amount decimal(18,2) NOT NULL,
		status varchar(20) NOT NULL COMMENT 'PENDING, APPROVED, REJECTED',
		reason_type varchar(100) DEFAULT NULL,
		reason_description varchar(255) DEFAULT NULL,
		remark varchar(60) DEFAULT NULL,
		metadata JSON NULL,
		created_from varchar(60) DEFAULT NULL,
		created_by varchar(100) DEFAULT NULL,
		approved_by varchar(100) DEFAULT NULL,
		approved_at datetime DEFAULT NULL,
		created_at datetime NOT NULL,
		updated_at datetime NOT NULL,
		deleted_at datetime DEFAULT NULL,
		PRIMARY KEY (uuid),
		UNIQUE KEY disbursement_unique_reference_per_merchant (reference_id,merchant_id),
		KEY disbursements_merchant_id_IDX (merchant_id) USING BTREE,
		KEY disbursements_created_at_IDX (created_at) USING BTREE,
		KEY disbursements_bulk_id_IDX (bulk_id) USING BTREE
	) ENGINE=InnoDB`
}
